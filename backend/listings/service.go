// Package listings provides the ListingService: multi-step draft state machine
// and atomic listing submission with role promotion.
package listings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"flickey/go-backend/amenities"
	"flickey/go-backend/config"
	"flickey/go-backend/db"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListingService orchestrates draft lifecycle and final submission.
type ListingService struct {
	DB       *gorm.DB
	Redis    *redis.Client
	Cfg      *config.Settings
	Registry *StepRegistry
}

// NewListingService creates a ListingService.
func NewListingService(database *gorm.DB, rdb *redis.Client, cfg *config.Settings) *ListingService {
	return &ListingService{
		DB:       database,
		Redis:    rdb,
		Cfg:      cfg,
		Registry: DefaultRegistry,
	}
}

func (s *ListingService) registry() *StepRegistry {
	if s.Registry != nil {
		return s.Registry
	}
	return DefaultRegistry
}

var sanitizer = bluemonday.StrictPolicy()

// ─────────────────────────────────────────────────────────────────────────────
// Step 1: Create Draft
// ─────────────────────────────────────────────────────────────────────────────

// CreateDraft creates a new listing draft (Step 1: housing type).
func (s *ListingService) CreateDraft(ctx context.Context, hostID uuid.UUID, req DraftCreateRequest) (*db.ListingDraft, error) {
	if !amenities.IsValidHousingType(req.Type) {
		return nil, &ValidationError{Code: "INVALID_TYPE", Message: "type must be apartment, house, or manor."}
	}

	t := req.Type
	draft := &db.ListingDraft{
		ID:     uuid.New(),
		HostID: hostID,
		Type:   &t,
		Status: "draft",
	}
	draft.CurrentStep = s.registry().CalculateNextStep(draft)

	if err := db.CreateDraft(ctx, s.DB, draft); err != nil {
		return nil, fmt.Errorf("CreateDraft: %w", err)
	}

	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Get Draft
// ─────────────────────────────────────────────────────────────────────────────

// GetDraft fetches a draft enforcing host ownership (Anti-IDOR).
func (s *ListingService) GetDraft(ctx context.Context, draftID, hostID uuid.UUID) (*db.ListingDraft, error) {
	draft, err := db.FindDraftByIDAndHost(ctx, s.DB, draftID, hostID)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, &DraftNotFoundError{}
	}
	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 2: Property Parameters
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep2 saves property name, address, coordinates, and parameters.
func (s *ListingService) UpdateStep2(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep2Request) (*db.ListingDraft, error) {
	// Sanitize and validate name before tx.
	name := strings.TrimSpace(sanitizer.Sanitize(req.Name))
	if len(name) < 10 || len(name) > 100 {
		return nil, &ValidationError{Code: "INVALID_NAME", Message: "Property name must be between 10 and 100 characters."}
	}

	// Sanitize address fields.
	address := strings.TrimSpace(sanitizer.Sanitize(req.Address))
	if len(address) == 0 {
		return nil, &ValidationError{Code: "INVALID_ADDRESS", Message: "Address is required."}
	}
	city := strings.TrimSpace(sanitizer.Sanitize(req.City))
	street := strings.TrimSpace(sanitizer.Sanitize(req.Street))
	houseNumber := strings.TrimSpace(sanitizer.Sanitize(req.HouseNumber))

	if req.Floor > req.TotalFloors {
		return nil, &ValidationError{Code: "INVALID_FLOOR", Message: "Floor must be less than or equal to total floors."}
	}

	var updatedDraft *db.ListingDraft
	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		draft, err := db.LockDraftByIDAndHost(ctx, tx, draftID, hostID)
		if err != nil {
			return err
		}
		if draft == nil {
			return &DraftNotFoundError{}
		}
		if draft.Status == "submitted" {
			return &DraftAlreadySubmittedError{}
		}
		if !s.registry().CanAccessStep(draft, 2) {
			return &StepNotAllowedError{Step: 2, Current: draft.CurrentStep}
		}

		draft.Name = &name
		draft.Address = &address
		draft.City = &city
		draft.Street = &street
		draft.HouseNumber = &houseNumber
		draft.Latitude = &req.Latitude
		draft.Longitude = &req.Longitude
		draft.Square = &req.Square
		floor := req.Floor
		draft.Floor = &floor
		tf := req.TotalFloors
		draft.TotalFloors = &tf
		mg := req.MaxGuests
		draft.MaxGuests = &mg
		rc := req.RoomsCount
		draft.RoomsCount = &rc
		bc := req.BedsCount
		draft.BedsCount = &bc
		bac := req.BathroomsCount
		draft.BathroomsCount = &bac

		s.registry().InvalidateDependents(draft, 2)
		draft.CurrentStep = s.registry().CalculateNextStep(draft)

		if err := db.SaveDraft(ctx, tx, draft); err != nil {
			return err
		}
		updatedDraft = draft
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return updatedDraft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 3: Media
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep3 attaches uploaded media IDs to the draft.
func (s *ListingService) UpdateStep3(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep3Request) (*db.ListingDraft, error) {
	if len(req.MediaIDs) < s.Cfg.MediaMinCount || len(req.MediaIDs) > s.Cfg.MediaMaxCount {
		return nil, &ValidationError{
			Code:    "INVALID_MEDIA_COUNT",
			Message: fmt.Sprintf("Media count must be between %d and %d.", s.Cfg.MediaMinCount, s.Cfg.MediaMaxCount),
		}
	}

	// Validate uniqueness.
	seen := map[uuid.UUID]struct{}{}
	for _, id := range req.MediaIDs {
		if _, dup := seen[id]; dup {
			return nil, &ValidationError{Code: "DUPLICATE_MEDIA_IDS", Message: "Duplicate media IDs are not allowed."}
		}
		seen[id] = struct{}{}
	}

	// Store as JSON array of UUID strings.
	uuidStrs := make([]string, len(req.MediaIDs))
	for i, id := range req.MediaIDs {
		uuidStrs[i] = id.String()
	}
	mediaJSON, _ := json.Marshal(uuidStrs)

	var updatedDraft *db.ListingDraft
	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		draft, err := db.LockDraftByIDAndHost(ctx, tx, draftID, hostID)
		if err != nil {
			return err
		}
		if draft == nil {
			return &DraftNotFoundError{}
		}
		if draft.Status == "submitted" {
			return &DraftAlreadySubmittedError{}
		}
		if !s.registry().CanAccessStep(draft, 3) {
			return &StepNotAllowedError{Step: 3, Current: draft.CurrentStep}
		}

		// Validate ownership + uploaded status (allowing attached media for source listing in edit mode).
		validMedia, err := db.FindValidMediaForDraft(ctx, tx, req.MediaIDs, hostID, draft.SourceListingID)
		if err != nil {
			return err
		}
		if len(validMedia) != len(req.MediaIDs) {
			return &MediaNotOwnedError{Code: "MEDIA_NOT_OWNED", Message: "One or more media items are not found, not uploaded, or not owned by you."}
		}

		draft.MediaIDs = mediaJSON

		s.registry().InvalidateDependents(draft, 3)
		draft.CurrentStep = s.registry().CalculateNextStep(draft)

		if err := db.SaveDraft(ctx, tx, draft); err != nil {
			return err
		}
		updatedDraft = draft
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return updatedDraft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 4: Amenities
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep4 saves validated amenities for the draft.
func (s *ListingService) UpdateStep4(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep4Request) (*db.ListingDraft, error) {
	if len(req.Amenities) == 0 {
		return nil, &ValidationError{Code: "EMPTY_AMENITIES", Message: "At least one amenity must be selected."}
	}

	// Validate uniqueness.
	seen := map[string]struct{}{}
	for _, a := range req.Amenities {
		if _, dup := seen[a]; dup {
			return nil, &ValidationError{Code: "DUPLICATE_AMENITIES", Message: "Duplicate amenities are not allowed."}
		}
		seen[a] = struct{}{}
	}

	amenitiesJSON, _ := json.Marshal(req.Amenities)

	var updatedDraft *db.ListingDraft
	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		draft, err := db.LockDraftByIDAndHost(ctx, tx, draftID, hostID)
		if err != nil {
			return err
		}
		if draft == nil {
			return &DraftNotFoundError{}
		}
		if draft.Status == "submitted" {
			return &DraftAlreadySubmittedError{}
		}
		if !s.registry().CanAccessStep(draft, 4) {
			return &StepNotAllowedError{Step: 4, Current: draft.CurrentStep}
		}

		// Validate each amenity against registry + housing type.
		housingType := amenities.HousingTypeApartment
		if draft.Type != nil {
			housingType = amenities.HousingType(*draft.Type)
		}

		for _, aid := range req.Amenities {
			def, ok := amenities.GetAmenity(aid)
			if !ok {
				return &ValidationError{
					Code:    "INVALID_AMENITY",
					Message: fmt.Sprintf("Amenity '%s' does not exist.", aid),
				}
			}
			if !def.IsAllowedFor(housingType) {
				return &ValidationError{
					Code:    "AMENITY_NOT_ALLOWED",
					Message: fmt.Sprintf("Amenity '%s' is not allowed for housing type '%s'.", aid, housingType),
				}
			}
		}

		draft.Amenities = amenitiesJSON

		s.registry().InvalidateDependents(draft, 4)
		draft.CurrentStep = s.registry().CalculateNextStep(draft)

		if err := db.SaveDraft(ctx, tx, draft); err != nil {
			return err
		}
		updatedDraft = draft
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return updatedDraft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 5: Price & Rules
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep5 saves pricing, currency, and house rules.
func (s *ListingService) UpdateStep5(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep5Request) (*db.ListingDraft, error) {
	if req.PricePerNight <= 0 || req.PricePerNight > 10000 {
		return nil, &ValidationError{Code: "INVALID_PRICE", Message: "Price per night must be between 0.01 and 10000."}
	}
	if req.Currency != "BYN" {
		return nil, &ValidationError{Code: "INVALID_CURRENCY", Message: "Currency must be BYN."}
	}
	if req.MinNights < 1 || req.MinNights > 30 {
		return nil, &ValidationError{Code: "INVALID_MIN_NIGHTS", Message: "Minimum nights must be between 1 and 30."}
	}

	// Normalize time strings.
	checkin := normalizeTimeStr(req.CheckinFrom)
	checkout := normalizeTimeStr(req.CheckoutUntil)

	var updatedDraft *db.ListingDraft
	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		draft, err := db.LockDraftByIDAndHost(ctx, tx, draftID, hostID)
		if err != nil {
			return err
		}
		if draft == nil {
			return &DraftNotFoundError{}
		}
		if draft.Status == "submitted" {
			return &DraftAlreadySubmittedError{}
		}
		if !s.registry().CanAccessStep(draft, 5) {
			return &StepNotAllowedError{Step: 5, Current: draft.CurrentStep}
		}

		draft.PricePerNight = &req.PricePerNight
		cur := req.Currency
		draft.Currency = &cur
		mn := req.MinNights
		draft.MinNights = &mn
		draft.CheckinFrom = &checkin
		draft.CheckoutUntil = &checkout
		draft.AllowChildren = &req.Rules.AllowChildren
		draft.AllowPets = &req.Rules.AllowPets
		draft.AllowSmoking = &req.Rules.AllowSmoking
		draft.AllowParties = &req.Rules.AllowParties
		draft.DepositRequired = &req.Rules.DepositRequired
		draft.WithInvoicing = &req.Rules.WithInvoicing

		s.registry().InvalidateDependents(draft, 5)
		draft.CurrentStep = s.registry().CalculateNextStep(draft)

		if err := db.SaveDraft(ctx, tx, draft); err != nil {
			return err
		}
		updatedDraft = draft
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return updatedDraft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 6: Description
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep6 saves and sanitizes the property description.
func (s *ListingService) UpdateStep6(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep6Request) (*db.ListingDraft, error) {
	desc := strings.TrimSpace(sanitizer.Sanitize(req.Description))
	if len(desc) < 30 || len(desc) > 5000 {
		return nil, &ValidationError{Code: "INVALID_DESCRIPTION", Message: "Description must be between 30 and 5000 characters."}
	}

	var updatedDraft *db.ListingDraft
	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		draft, err := db.LockDraftByIDAndHost(ctx, tx, draftID, hostID)
		if err != nil {
			return err
		}
		if draft == nil {
			return &DraftNotFoundError{}
		}
		if draft.Status == "submitted" {
			return &DraftAlreadySubmittedError{}
		}
		if !s.registry().CanAccessStep(draft, 6) {
			return &StepNotAllowedError{Step: 6, Current: draft.CurrentStep}
		}

		draft.Description = &desc

		s.registry().InvalidateDependents(draft, 6)
		draft.CurrentStep = s.registry().CalculateNextStep(draft)

		if err := db.SaveDraft(ctx, tx, draft); err != nil {
			return err
		}
		updatedDraft = draft
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return updatedDraft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Create Draft From Existing Listing (for re-submitting rejected listings)
// ─────────────────────────────────────────────────────────────────────────────

// CreateDraftFromListing clones or creates an edit draft from an existing listing owned by hostID.
func (s *ListingService) CreateDraftFromListing(ctx context.Context, listingID, hostID uuid.UUID, mode string) (*db.ListingDraft, error) {
	listing, err := db.FindListingByIDAndHost(ctx, s.DB, listingID, hostID)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, &ListingNotFoundError{}
	}

	if mode == "" {
		if listing.Status == "rejected" {
			mode = "create"
		} else {
			mode = "edit"
		}
	}

	// Copy media IDs
	var mediaIDStrs []string
	for _, m := range listing.Media {
		mediaIDStrs = append(mediaIDStrs, m.ID.String())
	}
	var mediaJSON []byte
	if len(mediaIDStrs) > 0 {
		mediaJSON, _ = json.Marshal(mediaIDStrs)
	}

	// Copy amenity IDs
	var amenityIDs []string
	for _, a := range listing.ListingAmenities {
		amenityIDs = append(amenityIDs, a.AmenityID)
	}
	var amenitiesJSON []byte
	if len(amenityIDs) > 0 {
		amenitiesJSON, _ = json.Marshal(amenityIDs)
	}

	draft := &db.ListingDraft{
		ID:              uuid.New(),
		HostID:          hostID,
		SourceListingID: &listing.ID,
		Mode:            mode,
		Status:          "draft",
		Type:            &listing.Type,
		Name:            &listing.Name,
		Address:         &listing.Address,
		City:            &listing.City,
		Street:          &listing.Street,
		HouseNumber:     &listing.HouseNumber,
		Latitude:        &listing.Latitude,
		Longitude:       &listing.Longitude,
		Square:          &listing.Square,
		Floor:           &listing.Floor,
		TotalFloors:     &listing.TotalFloors,
		MaxGuests:       &listing.MaxGuests,
		RoomsCount:      &listing.RoomsCount,
		BedsCount:       &listing.BedsCount,
		BathroomsCount:  &listing.BathroomsCount,
		MediaIDs:        mediaJSON,
		Amenities:       amenitiesJSON,
		PricePerNight:   &listing.PricePerNight,
		Currency:        &listing.Currency,
		MinNights:       &listing.MinNights,
		CheckinFrom:     &listing.CheckinFrom,
		CheckoutUntil:   &listing.CheckoutUntil,
		AllowChildren:   &listing.AllowChildren,
		AllowPets:       &listing.AllowPets,
		AllowSmoking:    &listing.AllowSmoking,
		AllowParties:    &listing.AllowParties,
		DepositRequired: &listing.DepositRequired,
		WithInvoicing:   &listing.WithInvoicing,
		Description:     &listing.Description,
	}

	draft.CurrentStep = s.registry().CalculateNextStep(draft)

	if err := db.CreateDraft(ctx, s.DB, draft); err != nil {
		return nil, fmt.Errorf("CreateDraftFromListing: %w", err)
	}

	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Submit Draft (Atomic)
// ─────────────────────────────────────────────────────────────────────────────

// SubmitDraft atomically converts a completed draft into a finalized listing or submits an edit.
// Uses Redis idempotency + PostgreSQL transaction with row-level locking.
func (s *ListingService) SubmitDraft(ctx context.Context, draftID, hostID uuid.UUID, idempotencyKey uuid.UUID) (*ListingSubmitResponse, error) {
	redisKey := fmt.Sprintf("idempotency:listing-submit:%s:%s", hostID, idempotencyKey)

	// ── Idempotency lock ──────────────────────────────────────────────────────
	if s.Redis != nil {
		acquired, err := s.Redis.SetNX(ctx, redisKey, "PROCESSING", 60*time.Second).Result()
		if err != nil {
			return nil, fmt.Errorf("SubmitDraft redis lock: %w", err)
		}
		if !acquired {
			val, err := s.Redis.Get(ctx, redisKey).Result()
			if err == nil && strings.HasPrefix(val, "COMPLETED:") {
				var resp ListingSubmitResponse
				if parseErr := json.Unmarshal([]byte(val[len("COMPLETED:"):]), &resp); parseErr == nil {
					return &resp, nil
				}
			}
			return nil, &IdempotencyConflictError{Code: "IDEMPOTENCY_CONFLICT", Message: "A submission with this idempotency key is already in progress."}
		}
	}

	rollbackLock := func() {
		if s.Redis != nil {
			cur, _ := s.Redis.Get(ctx, redisKey).Result()
			if cur == "PROCESSING" {
				s.Redis.Del(ctx, redisKey)
			}
		}
	}

	// ── Transaction ───────────────────────────────────────────────────────────
	var resp ListingSubmitResponse

	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Lock draft row — prevents concurrent submissions.
		draft, err := db.LockDraftByIDAndHost(ctx, tx, draftID, hostID)
		if err != nil {
			return &DraftNotFoundError{}
		}
		if draft == nil {
			return &DraftNotFoundError{}
		}
		if draft.Status == "submitted" || draft.Status == "approved" {
			return &DraftAlreadySubmittedError{}
		}

		completed := s.registry().GetCompletedSteps(draft)
		if len(completed) < s.registry().TotalSteps() {
			return &DraftIncompleteError{Message: "All wizard steps must be completed before submission."}
		}

		// 2. Full draft re-validation.
		if err := validateFullDraft(draft, s.Cfg); err != nil {
			return err
		}

		// 3. Handle Edit Draft mode
		if draft.Mode == "edit" && draft.SourceListingID != nil {
			sourceListing, err := db.FindListingByIDAndHost(ctx, tx, *draft.SourceListingID, hostID)
			if err != nil {
				return err
			}
			if sourceListing == nil {
				return &ListingNotFoundError{}
			}

			// Extract original amenities
			originalAmenities, _ := extractAmenitiesAndMedia(sourceListing)

			// Determine if meaningful changes were made
			pubAction, _ := DeterminePublicationAction(sourceListing, draft, originalAmenities, sourceListing.Media)

			if pubAction == PublicationActionNoChange {
				// No meaningful changes: listing remains in its current status (published / archived)
				draft.Status = "submitted"
				if err := db.SaveDraft(ctx, tx, draft); err != nil {
					return fmt.Errorf("save draft: %w", err)
				}

				resp = ListingSubmitResponse{
					ListingID:         sourceListing.ID,
					DraftID:           &draft.ID,
					Status:            sourceListing.Status,
					PublicationAction: string(PublicationActionNoChange),
				}
				return nil
			}

			// Meaningful changes detected: draft enters moderation review
			draft.Status = "pending_review"
			if err := db.SaveDraft(ctx, tx, draft); err != nil {
				return fmt.Errorf("save draft pending review: %w", err)
			}

			// If source listing was rejected, archived, or had changes requested, align its status to pending_review
			if sourceListing.Status == "rejected" || sourceListing.Status == "archived" || sourceListing.Status == "changes_requested" {
				_ = tx.Model(sourceListing).Updates(map[string]any{
					"status":             "pending_review",
					"rejection_reason":   nil,
					"moderation_comment": nil,
				}).Error
				sourceListing.Status = "pending_review"
			}

			resp = ListingSubmitResponse{
				ListingID:         sourceListing.ID,
				DraftID:           &draft.ID,
				Status:            "pending_review",
				PublicationAction: string(PublicationActionRequireModeration),
			}
			return nil
		}

		// 4. Handle New Listing Creation
		// New listings without an apartment verification video are created in 'draft_video_required'.
		listingStatus := "draft_video_required"
		if draft.VerificationVideoID != nil {
			isVerified, _ := db.HasApprovedBusinessVerification(ctx, tx, hostID)
			if isVerified {
				listingStatus = "pending_review"
			} else {
				existingReq, _ := db.GetUserVerificationRequest(ctx, tx, hostID)
				if existingReq != nil && existingReq.Status == "pending" {
					listingStatus = "awaiting_company_verification"
				} else {
					listingStatus = "draft"
				}
			}
		}

		// Parse media IDs from JSON.
		var mediaIDStrs []string
		if draft.MediaIDs != nil {
			_ = json.Unmarshal(draft.MediaIDs, &mediaIDStrs)
		}
		mediaIDs := make([]uuid.UUID, len(mediaIDStrs))
		for i, s := range mediaIDStrs {
			mediaIDs[i], _ = uuid.Parse(s)
		}

		// Parse amenities from JSON.
		var amenityIDs []string
		if draft.Amenities != nil {
			_ = json.Unmarshal(draft.Amenities, &amenityIDs)
		}

		// INSERT listing.
		newListing := db.Listing{
			ID:                  uuid.New(),
			HostID:              hostID,
			Status:              listingStatus,
			VerificationVideoID: draft.VerificationVideoID,
			Type:            derefStr(draft.Type),
			Name:            derefStr(draft.Name),
			Address:         derefStr(draft.Address),
			City:            derefStr(draft.City),
			Street:          derefStr(draft.Street),
			HouseNumber:     derefStr(draft.HouseNumber),
			Latitude:        derefFloat(draft.Latitude),
			Longitude:       derefFloat(draft.Longitude),
			Square:          derefFloat(draft.Square),
			Floor:           derefInt(draft.Floor),
			TotalFloors:     derefInt(draft.TotalFloors),
			MaxGuests:       derefInt(draft.MaxGuests),
			RoomsCount:      derefInt(draft.RoomsCount),
			BedsCount:       derefInt(draft.BedsCount),
			BathroomsCount:  derefInt(draft.BathroomsCount),
			PricePerNight:   derefFloat(draft.PricePerNight),
			Currency:        derefStr(draft.Currency),
			MinNights:       derefInt(draft.MinNights),
			CheckinFrom:     derefStr(draft.CheckinFrom),
			CheckoutUntil:   derefStr(draft.CheckoutUntil),
			AllowChildren:   derefBool(draft.AllowChildren),
			AllowPets:       derefBool(draft.AllowPets),
			AllowSmoking:    derefBool(draft.AllowSmoking),
			AllowParties:    derefBool(draft.AllowParties),
			DepositRequired: derefBool(draft.DepositRequired),
			WithInvoicing:   derefBool(draft.WithInvoicing),
			Description:     derefStr(draft.Description),
		}
		if err := db.CreateListing(ctx, tx, &newListing); err != nil {
			return fmt.Errorf("create listing: %w", err)
		}

		if err := db.AddListingAmenities(ctx, tx, newListing.ID, amenityIDs); err != nil {
			return fmt.Errorf("add amenities: %w", err)
		}

		// 8. Attach media — raw SQL with rowcount check (security-critical).
		affected, err := db.AttachMediaToListing(ctx, tx, mediaIDs, hostID, newListing.ID)
		if err != nil {
			return fmt.Errorf("attach media: %w", err)
		}
		if affected != int64(len(mediaIDs)) {
			return &MediaNotOwnedError{
				Code:    "MEDIA_ATTACH_FAILED",
				Message: "One or more media items could not be attached (must be uploaded, unattached, and owned by host).",
			}
		}

		// 9. Mark draft submitted.
		draft.Status = "submitted"
		if err := db.SaveDraft(ctx, tx, draft); err != nil {
			return fmt.Errorf("mark draft submitted: %w", err)
		}

		// 10. Atomic role transition: guest → host (with row lock).
		user, err := db.LockUserByID(ctx, tx, hostID)
		if err != nil {
			return fmt.Errorf("lock user: %w", err)
		}
		if user != nil && user.Role == db.UserRoleGuest {
			if err := db.UpdateUserRole(ctx, tx, hostID, db.UserRoleHost); err != nil {
				return fmt.Errorf("promote to host: %w", err)
			}
		}

		resp = ListingSubmitResponse{
			ListingID:         newListing.ID,
			DraftID:           &draft.ID,
			Status:            newListing.Status,
			PublicationAction: string(PublicationActionRequireModeration),
		}
		return nil // COMMIT
	})

	if txErr != nil {
		rollbackLock()
		return nil, txErr
	}

	// ── Cache result in Redis ─────────────────────────────────────────────────
	if s.Redis != nil {
		if payload, err := json.Marshal(resp); err == nil {
			s.Redis.Set(ctx, redisKey, "COMPLETED:"+string(payload), 86400*time.Second)
		}
	}

	return &resp, nil
}

// ArchiveListing archives an active published listing owned by hostID.
func (s *ListingService) ArchiveListing(ctx context.Context, listingID, hostID uuid.UUID) (*db.Listing, error) {
	listing, err := db.FindListingByIDAndHost(ctx, s.DB, listingID, hostID)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, &ListingNotFoundError{}
	}
	if listing.Status == "archived" {
		return listing, nil
	}
	if listing.Status != "published" && listing.Status != "active" {
		return nil, &ValidationError{
			Code:    "CANNOT_ARCHIVE_UNPUBLISHED",
			Message: "Only published listings can be moved to archive.",
		}
	}

	return db.ArchiveListingByIDAndHost(ctx, s.DB, listingID, hostID)
}

// UnarchiveListing restores an archived listing or submits a draft for moderation.
func (s *ListingService) UnarchiveListing(ctx context.Context, listingID, hostID uuid.UUID) (*db.Listing, error) {
	listing, err := db.FindListingByIDAndHost(ctx, s.DB, listingID, hostID)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, &ListingNotFoundError{}
	}
	if listing.Status != "archived" && listing.Status != "draft" {
		return nil, &ValidationError{
			Code:    "INVALID_STATUS_FOR_RESTORE",
			Message: "Only archived listings or drafts can be restored.",
		}
	}

	targetStatus := "published"
	if listing.Status == "draft" {
		// A draft MUST go through moderation. It can NEVER bypass moderation directly to published.
		targetStatus = "pending_review"
	} else {
		// If listing was previously rejected or changes were requested, it must go to moderation
		if listing.RejectionReason != nil || listing.ModerationComment != nil {
			targetStatus = "pending_review"
		} else {
			// Check if there is an active submitted draft with pending changes under moderation
			pendingDraft, err := db.FindPendingDraftBySourceListingID(ctx, s.DB, listingID)
			if err != nil {
				return nil, err
			}
			if pendingDraft != nil {
				// If there is a submitted draft pending moderation, unarchiving routes through moderation
				targetStatus = "awaiting_company_verification"
			}
		}
	}

	return db.UnarchiveListingByIDAndHost(ctx, s.DB, listingID, hostID, targetStatus)
}

// ─────────────────────────────────────────────────────────────────────────────
// Read Operations
// ─────────────────────────────────────────────────────────────────────────────

// GetMyListings returns all listings belonging to the authenticated host.
func (s *ListingService) GetMyListings(ctx context.Context, hostID uuid.UUID) ([]db.Listing, error) {
	return db.FindListingsByHost(ctx, s.DB, hostID)
}

// GetPublishedListings returns all published listings for public browsing, optionally filtered by checkin, checkout dates and host ID.
func (s *ListingService) GetPublishedListings(ctx context.Context, checkin, checkout time.Time, hostID *uuid.UUID) ([]db.Listing, error) {
	return db.FindPublishedListingsWithDatesAndHost(ctx, s.DB, checkin, checkout, hostID)
}

// GetPublishedListingsByHost returns all published listings belonging to a specific host.
func (s *ListingService) GetPublishedListingsByHost(ctx context.Context, hostID uuid.UUID) ([]db.Listing, error) {
	return db.FindPublishedListingsWithDatesAndHost(ctx, s.DB, time.Time{}, time.Time{}, &hostID)
}

// GetPublishedByID returns a single published listing for public view.
func (s *ListingService) GetPublishedByID(ctx context.Context, listingID uuid.UUID) (*db.Listing, error) {
	return db.FindPublishedListingByID(ctx, s.DB, listingID)
}

// GetListingForView returns a listing for viewing:
// - If user is admin: returns the listing regardless of status
// - If user is the owner (host): returns the listing regardless of status
// - Otherwise: returns only if status is "published"
func (s *ListingService) GetListingForView(ctx context.Context, listingID uuid.UUID, userID *uuid.UUID, isAdmin bool) (*db.Listing, error) {
	if isAdmin {
		return db.FindListingByID(ctx, s.DB, listingID)
	}
	if userID != nil {
		listing, err := db.FindListingByIDAndHost(ctx, s.DB, listingID, *userID)
		if err != nil {
			return nil, err
		}
		if listing != nil {
			return listing, nil
		}
	}
	return db.FindPublishedListingByID(ctx, s.DB, listingID)
}

// DeleteListing deletes a host's own listing (Anti-IDOR).
func (s *ListingService) DeleteListing(ctx context.Context, listingID, hostID uuid.UUID) error {
	deleted, err := db.DeleteListingByIDAndHost(ctx, s.DB, listingID, hostID)
	if err != nil {
		return err
	}
	if !deleted {
		return &ListingNotFoundError{}
	}
	return nil
}



// ─────────────────────────────────────────────────────────────────────────────
// Full Draft Validation (pre-submission)
// ─────────────────────────────────────────────────────────────────────────────

func validateFullDraft(draft *db.ListingDraft, cfg *config.Settings) error {
	if draft.Type == nil || !amenities.IsValidHousingType(*draft.Type) {
		return &DraftIncompleteError{Message: "Invalid or missing housing type."}
	}
	if draft.Name == nil || len(*draft.Name) < 10 || len(*draft.Name) > 100 {
		return &DraftIncompleteError{Message: "Invalid or missing property name."}
	}
	if draft.Address == nil || len(strings.TrimSpace(*draft.Address)) == 0 {
		return &DraftIncompleteError{Message: "Invalid or missing address."}
	}
	if draft.Latitude == nil || draft.Longitude == nil || *draft.Latitude == 0 || *draft.Longitude == 0 {
		return &DraftIncompleteError{Message: "Invalid or missing map coordinates."}
	}
	if draft.Square == nil || *draft.Square <= 10 || *draft.Square >= 1000 {
		return &DraftIncompleteError{Message: "Square must be between 10 and 1000."}
	}
	if draft.Floor == nil || draft.TotalFloors == nil || *draft.Floor > *draft.TotalFloors {
		return &DraftIncompleteError{Message: "Floor must be <= total floors."}
	}
	if *draft.Floor < 1 || *draft.Floor > 150 || *draft.TotalFloors < 1 || *draft.TotalFloors > 150 {
		return &DraftIncompleteError{Message: "Floor values must be between 1 and 150."}
	}
	if draft.MaxGuests == nil || *draft.MaxGuests < 1 || *draft.MaxGuests > 50 {
		return &DraftIncompleteError{Message: "Max guests must be between 1 and 50."}
	}
	if draft.RoomsCount == nil || *draft.RoomsCount < 1 || *draft.RoomsCount > 30 {
		return &DraftIncompleteError{Message: "Rooms count must be between 1 and 30."}
	}
	if draft.BedsCount == nil || *draft.BedsCount < 1 || *draft.BedsCount > 30 {
		return &DraftIncompleteError{Message: "Beds count must be between 1 and 30."}
	}
	if draft.BathroomsCount == nil || *draft.BathroomsCount < 1 || *draft.BathroomsCount > 20 {
		return &DraftIncompleteError{Message: "Bathrooms count must be between 1 and 20."}
	}

	var mediaIDStrs []string
	if draft.MediaIDs != nil {
		_ = json.Unmarshal(draft.MediaIDs, &mediaIDStrs)
	}
	if cfg != nil && (len(mediaIDStrs) < cfg.MediaMinCount || len(mediaIDStrs) > cfg.MediaMaxCount) {
		return &DraftIncompleteError{Message: fmt.Sprintf("Media count must be between %d and %d.", cfg.MediaMinCount, cfg.MediaMaxCount)}
	}

	var amenityIDs []string
	if draft.Amenities != nil {
		_ = json.Unmarshal(draft.Amenities, &amenityIDs)
	}
	if len(amenityIDs) < 1 {
		return &DraftIncompleteError{Message: "Amenities list cannot be empty."}
	}
	ht := amenities.HousingType(*draft.Type)
	for _, aid := range amenityIDs {
		def, ok := amenities.GetAmenity(aid)
		if !ok || !def.IsAllowedFor(ht) {
			return &DraftIncompleteError{Message: fmt.Sprintf("Invalid amenity '%s' for housing type '%s'.", aid, *draft.Type)}
		}
	}

	if draft.PricePerNight == nil || *draft.PricePerNight <= 0 || *draft.PricePerNight > 10000 {
		return &DraftIncompleteError{Message: "Price per night must be between 0.01 and 10000."}
	}
	if draft.Currency == nil || *draft.Currency != "BYN" {
		return &DraftIncompleteError{Message: "Currency must be BYN."}
	}
	if draft.MinNights == nil || *draft.MinNights < 1 || *draft.MinNights > 30 {
		return &DraftIncompleteError{Message: "Min nights must be between 1 and 30."}
	}
	if draft.CheckinFrom == nil || draft.CheckoutUntil == nil {
		return &DraftIncompleteError{Message: "Check-in and check-out times must be specified."}
	}
	if draft.AllowChildren == nil || draft.AllowPets == nil || draft.AllowSmoking == nil ||
		draft.AllowParties == nil || draft.DepositRequired == nil || draft.WithInvoicing == nil {
		return &DraftIncompleteError{Message: "House rules must be fully specified."}
	}
	if draft.Description == nil || len(*draft.Description) < 30 || len(*draft.Description) > 5000 {
		return &DraftIncompleteError{Message: "Description must be between 30 and 5000 characters."}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Utility helpers
// ─────────────────────────────────────────────────────────────────────────────

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// normalizeTimeStr ensures time strings are in "HH:MM:SS" format.
func normalizeTimeStr(t string) string {
	parts := strings.Split(t, ":")
	switch len(parts) {
	case 2:
		return t + ":00"
	case 3:
		return t
	default:
		return t
	}
}

// AttachVerificationVideo atomically attaches a verification video to a listing and updates its status
// inside a PostgreSQL transaction using row-level locking (SELECT FOR UPDATE) to prevent race conditions.
func (s *ListingService) AttachVerificationVideo(ctx context.Context, listingID, hostID, mediaID uuid.UUID) (*db.Listing, error) {
	var updatedListing *db.Listing

	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Lock listing row (SELECT FOR UPDATE)
		listing, err := db.LockListingByIDAndHost(ctx, tx, listingID, hostID)
		if err != nil {
			return err
		}
		if listing == nil {
			return &ListingNotFoundError{}
		}

		// 2. Validate media exists, is owned by host, and is a video
		var media db.Media
		if err := tx.WithContext(ctx).Where("id = ? AND host_id = ?", mediaID, hostID).First(&media).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &ValidationError{Code: "MEDIA_NOT_FOUND", Message: "Verification media not found or not owned by you."}
			}
			return err
		}

		if media.Status != db.MediaStatusUploaded && media.Status != db.MediaStatusAttached {
			return &ValidationError{Code: "MEDIA_NOT_READY", Message: "Media upload must be completed first."}
		}

		if !strings.HasPrefix(media.ContentType, "video/") {
			return &ValidationError{Code: "INVALID_MEDIA_TYPE", Message: "Verification media must be a video file (MP4, MOV, WebM)."}
		}

		// 3. Mark media as attached
		if err := tx.WithContext(ctx).Model(&media).Update("status", db.MediaStatusAttached).Error; err != nil {
			return err
		}

		// 4. Attach verification video
		listing.VerificationVideoID = &mediaID
		listing.VerificationVideo = &media

		// 5. Determine new status based on partner verification
		isVerified, err := db.HasApprovedBusinessVerification(ctx, tx, hostID)
		if err != nil {
			return err
		}

		if isVerified {
			listing.Status = "pending_review"
		} else {
			existingReq, err := db.GetUserVerificationRequest(ctx, tx, hostID)
			if err != nil {
				return err
			}
			if existingReq != nil && existingReq.Status == "pending" {
				listing.Status = "awaiting_company_verification"
			} else {
				listing.Status = "draft"
			}
		}

		if err := tx.WithContext(ctx).Save(listing).Error; err != nil {
			return err
		}

		updatedListing = listing
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return updatedListing, nil
}

// UpdateListing updates an existing listing owned by hostID.
func (s *ListingService) UpdateListing(ctx context.Context, listingID, hostID uuid.UUID, req *UpdateListingRequest) (*db.Listing, error) {
	listing, err := db.FindListingByIDAndHost(ctx, s.DB, listingID, hostID)
	if err != nil {
		return nil, fmt.Errorf("FindListingByIDAndHost: %w", err)
	}
	if listing == nil {
		return nil, &ListingNotFoundError{}
	}

	var updatedListing *db.Listing

	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update scalar fields if provided
		if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
			listing.Name = strings.TrimSpace(*req.Title)
		} else if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
			listing.Name = strings.TrimSpace(*req.Name)
		}

		if req.Description != nil {
			listing.Description = strings.TrimSpace(*req.Description)
		}

		if req.PropertyType != nil && strings.TrimSpace(*req.PropertyType) != "" {
			listing.Type = strings.TrimSpace(*req.PropertyType)
		} else if req.Type != nil && strings.TrimSpace(*req.Type) != "" {
			listing.Type = strings.TrimSpace(*req.Type)
		}

		if req.Address != nil {
			listing.Address = strings.TrimSpace(*req.Address)
		}
		if req.City != nil {
			listing.City = strings.TrimSpace(*req.City)
		}
		if req.Street != nil {
			listing.Street = strings.TrimSpace(*req.Street)
		}
		if req.HouseNumber != nil {
			listing.HouseNumber = strings.TrimSpace(*req.HouseNumber)
		}
		if req.Latitude != nil {
			listing.Latitude = *req.Latitude
		}
		if req.Longitude != nil {
			listing.Longitude = *req.Longitude
		}

		if req.Area != nil && *req.Area > 0 {
			listing.Square = *req.Area
		} else if req.Square != nil && *req.Square > 0 {
			listing.Square = *req.Square
		}

		if req.Floor != nil {
			listing.Floor = *req.Floor
		}
		if req.TotalFloors != nil && *req.TotalFloors > 0 {
			listing.TotalFloors = *req.TotalFloors
		}
		if req.MaxGuests != nil && *req.MaxGuests > 0 {
			listing.MaxGuests = *req.MaxGuests
		}
		if req.Bedrooms != nil {
			listing.RoomsCount = *req.Bedrooms
		} else if req.RoomsCount != nil {
			listing.RoomsCount = *req.RoomsCount
		}
		if req.Beds != nil {
			listing.BedsCount = *req.Beds
		} else if req.BedsCount != nil {
			listing.BedsCount = *req.BedsCount
		}
		if req.Bathrooms != nil {
			listing.BathroomsCount = *req.Bathrooms
		} else if req.BathroomsCount != nil {
			listing.BathroomsCount = *req.BathroomsCount
		}

		if req.PricePerNight != nil && *req.PricePerNight > 0 {
			listing.PricePerNight = *req.PricePerNight
		}
		if req.Currency != nil && strings.TrimSpace(*req.Currency) != "" {
			listing.Currency = strings.TrimSpace(*req.Currency)
		}
		if req.MinNights != nil && *req.MinNights > 0 {
			listing.MinNights = *req.MinNights
		}
		if req.CheckinFrom != nil && strings.TrimSpace(*req.CheckinFrom) != "" {
			listing.CheckinFrom = strings.TrimSpace(*req.CheckinFrom)
		}
		if req.CheckoutUntil != nil && strings.TrimSpace(*req.CheckoutUntil) != "" {
			listing.CheckoutUntil = strings.TrimSpace(*req.CheckoutUntil)
		}

		if req.AllowChildren != nil {
			listing.AllowChildren = *req.AllowChildren
		}
		if req.AllowPets != nil {
			listing.AllowPets = *req.AllowPets
		}
		if req.AllowSmoking != nil {
			listing.AllowSmoking = *req.AllowSmoking
		}
		if req.AllowParties != nil {
			listing.AllowParties = *req.AllowParties
		}
		if req.DepositRequired != nil {
			listing.DepositRequired = *req.DepositRequired
		}

		// Handle status transitions (Draft vs Moderation)
		if req.SubmitForModeration != nil && *req.SubmitForModeration {
			listing.Status = "pending_review"
			listing.RejectionReason = nil
			listing.ModerationComment = nil
		} else if req.SaveAsDraft != nil && *req.SaveAsDraft {
			listing.Status = "draft"
		} else if req.Status != nil && strings.TrimSpace(*req.Status) == "draft" {
			listing.Status = "draft"
		} else if req.Status != nil && strings.TrimSpace(*req.Status) == "pending_review" {
			listing.Status = "pending_review"
			listing.RejectionReason = nil
			listing.ModerationComment = nil
		} else if listing.Status == "changes_requested" || listing.Status == "rejected" || listing.Status == "published" || listing.Status == "active" || listing.Status == "archived" {
			listing.Status = "pending_review"
			listing.RejectionReason = nil
			listing.ModerationComment = nil
		}

		// Update Amenities if provided
		if req.Amenities != nil {
			if err := tx.Where("listing_id = ?", listing.ID).Delete(&db.ListingAmenity{}).Error; err != nil {
				return fmt.Errorf("delete old amenities: %w", err)
			}
			for _, amenityID := range *req.Amenities {
				if strings.TrimSpace(amenityID) == "" {
					continue
				}
				la := db.ListingAmenity{
					ListingID: listing.ID,
					AmenityID: strings.TrimSpace(amenityID),
				}
				if err := tx.Create(&la).Error; err != nil {
					return fmt.Errorf("create listing amenity: %w", err)
				}
			}
		}

		// Update Media if provided
		mediaList := req.MediaIDs
		if mediaList == nil {
			mediaList = req.Images
		}
		if mediaList != nil {
			for _, mStr := range *mediaList {
				mUUID, parseErr := uuid.Parse(mStr)
				if parseErr == nil {
					_ = tx.Model(&db.Media{}).Where("id = ? AND host_id = ?", mUUID, hostID).Update("listing_id", listing.ID).Error
				} else {
					cleanKey := strings.TrimPrefix(mStr, "http://localhost:8000/api/v1/media/dev-upload/")
					cleanKey = strings.TrimPrefix(cleanKey, "/api/v1/media/dev-upload/")
					_ = tx.Model(&db.Media{}).Where("(file_key = ? OR file_key LIKE ?) AND host_id = ?", cleanKey, "%"+cleanKey, hostID).Update("listing_id", listing.ID).Error
				}
			}
		}

		// Clear loaded associations from in-memory struct so GORM does not cascade-insert old items
		listing.ListingAmenities = nil
		listing.Media = nil
		listing.VerificationVideo = nil
		listing.Host = nil

		if err := tx.Omit(clause.Associations).Save(listing).Error; err != nil {
			return fmt.Errorf("save updated listing: %w", err)
		}

		updatedListing = listing
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	freshListing, err := db.FindListingByIDAndHost(ctx, s.DB, listingID, hostID)
	if err != nil || freshListing == nil {
		return updatedListing, nil
	}
	return freshListing, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Domain Errors
// ─────────────────────────────────────────────────────────────────────────────

type ValidationError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *ValidationError) Error() string { return e.Message }

type StepNotAllowedError struct {
	Step    int
	Current int
	Message string
}

func (e *StepNotAllowedError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("Step %d is not available. Current step: %d.", e.Step, e.Current)
}

type DraftNotFoundError struct{}

func (e *DraftNotFoundError) Error() string { return "Draft not found." }

type DraftAlreadySubmittedError struct{}

func (e *DraftAlreadySubmittedError) Error() string { return "Draft has already been submitted." }

type DraftIncompleteError struct {
	Message string
}

func (e *DraftIncompleteError) Error() string { return e.Message }

type InvalidStepTransitionError struct {
	Message string
}

func (e *InvalidStepTransitionError) Error() string { return e.Message }

type MediaNotOwnedError struct {
	Code    string
	Message string
}

func (e *MediaNotOwnedError) Error() string { return e.Message }

type IdempotencyConflictError struct {
	Code    string
	Message string
}

func (e *IdempotencyConflictError) Error() string { return e.Message }

type ListingNotFoundError struct{}

func (e *ListingNotFoundError) Error() string { return "Listing not found." }
