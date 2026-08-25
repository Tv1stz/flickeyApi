// Package listings provides the ListingService: multi-step draft state machine
// and atomic listing submission with role promotion.
package listings

import (
	"context"
	"encoding/json"
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
)

// ListingService orchestrates draft lifecycle and final submission.
type ListingService struct {
	DB    *gorm.DB
	Redis *redis.Client
	Cfg   *config.Settings
}

// NewListingService creates a ListingService.
func NewListingService(database *gorm.DB, rdb *redis.Client, cfg *config.Settings) *ListingService {
	return &ListingService{DB: database, Redis: rdb, Cfg: cfg}
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
		ID:          uuid.New(),
		HostID:      hostID,
		Type:        &t,
		CurrentStep: 2,
		Status:      "draft",
	}

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

// UpdateStep2 saves property name and parameters.
func (s *ListingService) UpdateStep2(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep2Request) (*db.ListingDraft, error) {
	draft, err := s.GetDraft(ctx, draftID, hostID)
	if err != nil {
		return nil, err
	}
	if err := ensureStepAccessible(draft, 2); err != nil {
		return nil, err
	}

	// Sanitize and validate name.
	name := strings.TrimSpace(sanitizer.Sanitize(req.Name))
	if len(name) < 10 || len(name) > 100 {
		return nil, &ValidationError{Code: "INVALID_NAME", Message: "Property name must be between 10 and 100 characters."}
	}

	if req.Floor > req.TotalFloors {
		return nil, &ValidationError{Code: "INVALID_FLOOR", Message: "Floor must be less than or equal to total floors."}
	}

	// Null out downstream steps if going back.
	if draft.CurrentStep > 2 {
		nullDownstreamFrom3(draft)
	}

	draft.Name = &name
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
	draft.CurrentStep = 3

	if err := db.SaveDraft(ctx, s.DB, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 3: Media
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep3 attaches uploaded media IDs to the draft.
func (s *ListingService) UpdateStep3(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep3Request) (*db.ListingDraft, error) {
	draft, err := s.GetDraft(ctx, draftID, hostID)
	if err != nil {
		return nil, err
	}
	if err := ensureStepAccessible(draft, 3); err != nil {
		return nil, err
	}

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

	// Validate ownership + uploaded status.
	validMedia, err := db.FindUploadedUnattachedMediaByIDsAndHost(ctx, s.DB, req.MediaIDs, hostID)
	if err != nil {
		return nil, err
	}
	if len(validMedia) != len(req.MediaIDs) {
		return nil, &MediaNotOwnedError{Code: "MEDIA_NOT_OWNED", Message: "One or more media items are not found, not uploaded, or not owned by you."}
	}

	// Null out downstream steps.
	if draft.CurrentStep > 3 {
		nullDownstreamFrom4(draft)
	}

	// Store as JSON array of UUID strings.
	uuidStrs := make([]string, len(req.MediaIDs))
	for i, id := range req.MediaIDs {
		uuidStrs[i] = id.String()
	}
	mediaJSON, _ := json.Marshal(uuidStrs)
	draft.MediaIDs = mediaJSON
	draft.CurrentStep = 4

	if err := db.SaveDraft(ctx, s.DB, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 4: Amenities
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep4 saves validated amenities for the draft.
func (s *ListingService) UpdateStep4(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep4Request) (*db.ListingDraft, error) {
	draft, err := s.GetDraft(ctx, draftID, hostID)
	if err != nil {
		return nil, err
	}
	if err := ensureStepAccessible(draft, 4); err != nil {
		return nil, err
	}

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

	// Validate each amenity against registry + housing type.
	housingType := amenities.HousingTypeApartment
	if draft.Type != nil {
		housingType = amenities.HousingType(*draft.Type)
	}

	for _, aid := range req.Amenities {
		def, ok := amenities.GetAmenity(aid)
		if !ok {
			return nil, &ValidationError{
				Code:    "INVALID_AMENITY",
				Message: fmt.Sprintf("Amenity '%s' does not exist.", aid),
			}
		}
		if !def.IsAllowedFor(housingType) {
			return nil, &ValidationError{
				Code:    "AMENITY_NOT_ALLOWED",
				Message: fmt.Sprintf("Amenity '%s' is not allowed for housing type '%s'.", aid, housingType),
			}
		}
	}

	// Null out downstream steps.
	if draft.CurrentStep > 4 {
		nullDownstreamFrom5(draft)
	}

	amenitiesJSON, _ := json.Marshal(req.Amenities)
	draft.Amenities = amenitiesJSON
	draft.CurrentStep = 5

	if err := db.SaveDraft(ctx, s.DB, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 5: Price & Rules
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep5 saves pricing, currency, and house rules.
func (s *ListingService) UpdateStep5(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep5Request) (*db.ListingDraft, error) {
	draft, err := s.GetDraft(ctx, draftID, hostID)
	if err != nil {
		return nil, err
	}
	if err := ensureStepAccessible(draft, 5); err != nil {
		return nil, err
	}

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

	// Null out downstream step.
	if draft.CurrentStep > 5 {
		draft.Description = nil
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
	draft.CurrentStep = 6

	if err := db.SaveDraft(ctx, s.DB, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Step 6: Description
// ─────────────────────────────────────────────────────────────────────────────

// UpdateStep6 saves and sanitizes the property description.
func (s *ListingService) UpdateStep6(ctx context.Context, draftID, hostID uuid.UUID, req DraftStep6Request) (*db.ListingDraft, error) {
	draft, err := s.GetDraft(ctx, draftID, hostID)
	if err != nil {
		return nil, err
	}
	if err := ensureStepAccessible(draft, 6); err != nil {
		return nil, err
	}

	desc := strings.TrimSpace(sanitizer.Sanitize(req.Description))
	if len(desc) < 30 || len(desc) > 5000 {
		return nil, &ValidationError{Code: "INVALID_DESCRIPTION", Message: "Description must be between 30 and 5000 characters."}
	}

	draft.Description = &desc
	// Note: current_step stays at 6 (as per original implementation).
	draft.CurrentStep = 6

	if err := db.SaveDraft(ctx, s.DB, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Submit Draft (Atomic)
// ─────────────────────────────────────────────────────────────────────────────

// SubmitDraft atomically converts a completed draft into a finalized listing.
// Uses Redis idempotency + PostgreSQL transaction with row-level locking.
func (s *ListingService) SubmitDraft(ctx context.Context, draftID, hostID uuid.UUID, idempotencyKey uuid.UUID) (*ListingSubmitResponse, error) {
	redisKey := fmt.Sprintf("idempotency:listing-submit:%s:%s", hostID, idempotencyKey)

	// ── Idempotency lock ──────────────────────────────────────────────────────
	acquired, err := s.Redis.SetNX(ctx, redisKey, "PROCESSING", 60*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("SubmitDraft redis lock: %w", err)
	}
	if !acquired {
		// Check existing value.
		val, err := s.Redis.Get(ctx, redisKey).Result()
		if err == nil && strings.HasPrefix(val, "COMPLETED:") {
			var resp ListingSubmitResponse
			if parseErr := json.Unmarshal([]byte(val[len("COMPLETED:"):]), &resp); parseErr == nil {
				return &resp, nil
			}
		}
		return nil, &IdempotencyConflictError{Code: "IDEMPOTENCY_CONFLICT", Message: "A submission with this idempotency key is already in progress."}
	}

	// On any failure, remove the lock to allow retries.
	rollbackLock := func() {
		cur, _ := s.Redis.Get(ctx, redisKey).Result()
		if cur == "PROCESSING" {
			s.Redis.Del(ctx, redisKey)
		}
	}

	// ── Transaction ───────────────────────────────────────────────────────────
	var listing db.Listing

	txErr := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Lock draft row — prevents concurrent submissions.
		draft, err := db.LockDraftByIDAndHost(ctx, tx, draftID, hostID)
		if err != nil {
			return &DraftNotFoundError{}
		}
		if draft == nil {
			return &DraftNotFoundError{}
		}
		if draft.Status == "submitted" {
			return &DraftAlreadySubmittedError{}
		}
		if draft.CurrentStep != 6 {
			return &DraftIncompleteError{Message: "Draft must complete all 6 steps before submission."}
		}

		// 2. Full draft re-validation.
		if err := validateFullDraft(draft, s.Cfg); err != nil {
			return err
		}

		// 3. Check business verification (stub: always false).
		isVerified, _ := db.HasApprovedBusinessVerification(ctx, tx, hostID)
		listingStatus := "awaiting_company_verification"
		if isVerified {
			listingStatus = "pending_review"
		}

		// 4. Parse media IDs from JSON.
		var mediaIDStrs []string
		if draft.MediaIDs != nil {
			_ = json.Unmarshal(draft.MediaIDs, &mediaIDStrs)
		}
		mediaIDs := make([]uuid.UUID, len(mediaIDStrs))
		for i, s := range mediaIDStrs {
			mediaIDs[i], _ = uuid.Parse(s)
		}

		// 5. Parse amenities from JSON.
		var amenityIDs []string
		if draft.Amenities != nil {
			_ = json.Unmarshal(draft.Amenities, &amenityIDs)
		}

		// 6. INSERT listing.
		listing = db.Listing{
			ID:              uuid.New(),
			HostID:          hostID,
			Status:          listingStatus,
			Type:            derefStr(draft.Type),
			Name:            derefStr(draft.Name),
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
		if err := db.CreateListing(ctx, tx, &listing); err != nil {
			return fmt.Errorf("create listing: %w", err)
		}

		// 7. INSERT amenities.
		if err := db.AddListingAmenities(ctx, tx, listing.ID, amenityIDs); err != nil {
			return fmt.Errorf("add amenities: %w", err)
		}

		// 8. Attach media — raw SQL with rowcount check (security-critical).
		affected, err := db.AttachMediaToListing(ctx, tx, mediaIDs, hostID, listing.ID)
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

		return nil // COMMIT
	})

	if txErr != nil {
		rollbackLock()
		return nil, txErr
	}

	// ── Cache result in Redis ─────────────────────────────────────────────────
	resp := ListingSubmitResponse{
		ListingID: listing.ID,
		Status:    listing.Status,
	}
	if payload, err := json.Marshal(resp); err == nil {
		s.Redis.Set(ctx, redisKey, "COMPLETED:"+string(payload), 86400*time.Second)
	}

	return &resp, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Read Operations
// ─────────────────────────────────────────────────────────────────────────────

// GetMyListings returns all listings belonging to the authenticated host.
func (s *ListingService) GetMyListings(ctx context.Context, hostID uuid.UUID) ([]db.Listing, error) {
	return db.FindListingsByHost(ctx, s.DB, hostID)
}

// GetPublishedListings returns all published listings for public browsing.
func (s *ListingService) GetPublishedListings(ctx context.Context) ([]db.Listing, error) {
	return db.FindPublishedListings(ctx, s.DB)
}

// GetPublishedByID returns a single published listing for public view.
func (s *ListingService) GetPublishedByID(ctx context.Context, listingID uuid.UUID) (*db.Listing, error) {
	return db.FindPublishedListingByID(ctx, s.DB, listingID)
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
	if len(mediaIDStrs) < cfg.MediaMinCount || len(mediaIDStrs) > cfg.MediaMaxCount {
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
// Step accessibility guard
// ─────────────────────────────────────────────────────────────────────────────

func ensureStepAccessible(draft *db.ListingDraft, expected int) error {
	if draft.Status == "submitted" {
		return &DraftAlreadySubmittedError{}
	}
	if draft.CurrentStep < expected {
		return &InvalidStepTransitionError{Message: fmt.Sprintf("Step %d is not yet available. Current step: %d.", expected, draft.CurrentStep)}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Cascade null helpers (preserve Python state machine behavior)
// ─────────────────────────────────────────────────────────────────────────────

func nullDownstreamFrom3(d *db.ListingDraft) {
	d.MediaIDs = nil
	nullDownstreamFrom4(d)
}

func nullDownstreamFrom4(d *db.ListingDraft) {
	d.Amenities = nil
	nullDownstreamFrom5(d)
}

func nullDownstreamFrom5(d *db.ListingDraft) {
	d.PricePerNight = nil
	d.Currency = nil
	d.MinNights = nil
	d.CheckinFrom = nil
	d.CheckoutUntil = nil
	d.AllowChildren = nil
	d.AllowPets = nil
	d.AllowSmoking = nil
	d.AllowParties = nil
	d.DepositRequired = nil
	d.WithInvoicing = nil
	d.Description = nil
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

// ─────────────────────────────────────────────────────────────────────────────
// Domain Errors
// ─────────────────────────────────────────────────────────────────────────────

type ValidationError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *ValidationError) Error() string { return e.Message }

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
