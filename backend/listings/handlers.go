// Package listings provides HTTP handlers for listing and draft endpoints.
package listings

import (
	"encoding/json"
	"net/http"
	"strings"

	"flickey/go-backend/auth"
	"flickey/go-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds dependencies for listing HTTP handlers.
type Handler struct {
	Service *ListingService
}

// NewHandler creates a listings Handler.
func NewHandler(svc *ListingService) *Handler {
	return &Handler{Service: svc}
}

// RegisterRoutes registers listing and draft routes under the provided RouterGroup.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc, activeMw gin.HandlerFunc) {
	// Active-user-authenticated routes.
	active := rg.Group("")
	active.Use(authMw, activeMw)
	active.POST("/drafts", h.CreateDraft)
	active.GET("/drafts/:draft_id", h.GetDraft)
	active.PATCH("/drafts/:draft_id/step-2", h.UpdateStep2)
	active.PATCH("/drafts/:draft_id/step-3", h.UpdateStep3)
	active.PATCH("/drafts/:draft_id/step-4", h.UpdateStep4)
	active.PATCH("/drafts/:draft_id/step-5", h.UpdateStep5)
	active.PATCH("/drafts/:draft_id/step-6", h.UpdateStep6)
	active.POST("/drafts/:draft_id/submit", h.SubmitDraft)
	active.GET("/my", h.GetMyListings)

	// Public routes — no auth.
	rg.GET("", h.BrowseListings)
	rg.GET("/:listing_id", h.GetPublicListing)
}

// CreateDraft handles POST /listings/drafts.
//
//	@Summary		Create a new listing draft (step 1)
//	@Description	Initialises a new multi-step wizard draft with the property type. Returns the draft ID and current step number.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		DraftCreateRequest	true	"Property type (apartment | house | manor)"
//	@Success		201		{object}	DraftCreateResponse
//	@Failure		400		{object}	auth.ErrorResponse	"VALIDATION_ERROR"
//	@Failure		401		{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403		{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		500		{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts [post]
func (h *Handler) CreateDraft(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	var req DraftCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	draft, err := h.Service.CreateDraft(c.Request.Context(), user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, DraftCreateResponse{
		DraftID:     draft.ID,
		CurrentStep: draft.CurrentStep,
	})
}

// GetDraft handles GET /listings/drafts/:draft_id.
//
//	@Summary		Get a draft
//	@Description	Fetches the full state of a listing draft. Only the owning host can access it.
//	@Tags			Listings
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Success		200			{object}	DraftDetailResponse
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"DRAFT_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/{draft_id} [get]
func (h *Handler) GetDraft(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	draftID, err := parseUUID(c, "draft_id")
	if err != nil {
		return
	}

	draft, err := h.Service.GetDraft(c.Request.Context(), draftID, user.ID)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, draftToResponse(draft))
}

// UpdateStep2 handles PATCH /listings/drafts/:draft_id/step-2.
//
//	@Summary		Update draft – step 2 (property details)
//	@Description	Sets name, area, floor, guest count, rooms, beds, and bathrooms. All fields required. name: 10–100 chars; square: 10–1000 m²; floor/total_floors: 1–150; max_guests: 1–50; rooms/beds: 1–30; bathrooms: 1–20.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep2Request	true	"Property details"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_STEP_TRANSITION"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"DRAFT_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/{draft_id}/step-2 [patch]
func (h *Handler) UpdateStep2(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	draftID, err := parseUUID(c, "draft_id")
	if err != nil {
		return
	}

	var req DraftStep2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	draft, err := h.Service.UpdateStep2(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, DraftStepResponse{DraftID: draft.ID, CurrentStep: draft.CurrentStep, Status: draft.Status})
}

// UpdateStep3 handles PATCH /listings/drafts/:draft_id/step-3.
//
//	@Summary		Update draft – step 3 (media)
//	@Description	Attaches 5–15 uploaded media UUIDs to the draft. All referenced media must belong to the current user and be in uploaded status.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep3Request	true	"Array of media UUIDs (5–15 items)"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_STEP_TRANSITION"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE / MEDIA_NOT_OWNED"
//	@Failure		404			{object}	auth.ErrorResponse	"DRAFT_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/{draft_id}/step-3 [patch]
func (h *Handler) UpdateStep3(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	draftID, err := parseUUID(c, "draft_id")
	if err != nil {
		return
	}

	var req DraftStep3Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	draft, err := h.Service.UpdateStep3(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, DraftStepResponse{DraftID: draft.ID, CurrentStep: draft.CurrentStep, Status: draft.Status})
}

// UpdateStep4 handles PATCH /listings/drafts/:draft_id/step-4.
//
//	@Summary		Update draft – step 4 (amenities)
//	@Description	Attaches one or more amenity IDs from the amenity registry to the draft.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep4Request	true	"Array of amenity ID strings (at least 1)"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_STEP_TRANSITION"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"DRAFT_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/{draft_id}/step-4 [patch]
func (h *Handler) UpdateStep4(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	draftID, err := parseUUID(c, "draft_id")
	if err != nil {
		return
	}

	var req DraftStep4Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	draft, err := h.Service.UpdateStep4(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, DraftStepResponse{DraftID: draft.ID, CurrentStep: draft.CurrentStep, Status: draft.Status})
}

// UpdateStep5 handles PATCH /listings/drafts/:draft_id/step-5.
//
//	@Summary		Update draft – step 5 (pricing & rules)
//	@Description	Sets price per night, currency, minimum nights, check-in/check-out times (HH:MM format), and house rules (allow_children, allow_pets, allow_smoking, allow_parties, deposit_required, with_invoicing).
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep5Request	true	"Pricing and house rules"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_STEP_TRANSITION"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"DRAFT_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/{draft_id}/step-5 [patch]
func (h *Handler) UpdateStep5(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	draftID, err := parseUUID(c, "draft_id")
	if err != nil {
		return
	}

	var req DraftStep5Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	draft, err := h.Service.UpdateStep5(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, DraftStepResponse{DraftID: draft.ID, CurrentStep: draft.CurrentStep, Status: draft.Status})
}

// UpdateStep6 handles PATCH /listings/drafts/:draft_id/step-6.
//
//	@Summary		Update draft – step 6 (description)
//	@Description	Sets the free-text description for the listing. 30–5000 characters.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep6Request	true	"Description text"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_STEP_TRANSITION"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"DRAFT_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/{draft_id}/step-6 [patch]
func (h *Handler) UpdateStep6(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	draftID, err := parseUUID(c, "draft_id")
	if err != nil {
		return
	}

	var req DraftStep6Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	draft, err := h.Service.UpdateStep6(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, DraftStepResponse{DraftID: draft.ID, CurrentStep: draft.CurrentStep, Status: draft.Status})
}

// SubmitDraft handles POST /listings/drafts/:draft_id/submit.
//
//	@Summary		Submit draft for review
//	@Description	Finalises a completed 6-step draft and creates a published listing pending moderation. Requires the Idempotency-Key header (UUIDv4) to prevent duplicate submissions.
//	@Tags			Listings
//	@Produce		json
//	@Param			draft_id		path	string	true	"Draft UUID"	format(uuid)
//	@Param			Idempotency-Key	header	string	true	"Client-generated UUIDv4 idempotency key"	format(uuid)
//	@Success		201	{object}	ListingSubmitResponse
//	@Failure		400	{object}	auth.ErrorResponse	"MISSING_IDEMPOTENCY_KEY / INVALID_IDEMPOTENCY_KEY / DRAFT_INCOMPLETE / VALIDATION_ERROR"
//	@Failure		401	{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403	{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404	{object}	auth.ErrorResponse	"DRAFT_NOT_FOUND"
//	@Failure		409	{object}	auth.ErrorResponse	"DRAFT_ALREADY_SUBMITTED / IDEMPOTENCY_CONFLICT"
//	@Failure		500	{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/{draft_id}/submit [post]
func (h *Handler) SubmitDraft(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	draftID, err := parseUUID(c, "draft_id")
	if err != nil {
		return
	}

	// Validate Idempotency-Key header.
	idempKeyStr := c.GetHeader("Idempotency-Key")
	if idempKeyStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "MISSING_IDEMPOTENCY_KEY", "message": "Idempotency-Key header is required."})
		return
	}
	idempKey, parseErr := uuid.Parse(idempKeyStr)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_IDEMPOTENCY_KEY", "message": "Idempotency-Key header must be a valid UUIDv4."})
		return
	}

	resp, err := h.Service.SubmitDraft(c.Request.Context(), draftID, user.ID, idempKey)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetMyListings handles GET /listings/my.
//
//	@Summary		Get host's own listings
//	@Description	Returns all published (and pending) listings owned by the authenticated host.
//	@Tags			Listings
//	@Produce		json
//	@Success		200	{array}		ListingHostReadSchema
//	@Failure		401	{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403	{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		500	{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/my [get]
func (h *Handler) GetMyListings(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	listings, err := h.Service.GetMyListings(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	result := make([]ListingHostReadSchema, len(listings))
	for i, l := range listings {
		result[i] = listingToHostSchema(&l)
	}
	c.JSON(http.StatusOK, result)
}

// BrowseListings handles GET /listings.
//
//	@Summary		Browse published listings
//	@Description	Returns all currently published listings. No authentication required.
//	@Tags			Listings
//	@Produce		json
//	@Success		200	{array}		ListingPublicSchema
//	@Failure		500	{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Router			/listings [get]
func (h *Handler) BrowseListings(c *gin.Context) {
	listings, err := h.Service.GetPublishedListings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	result := make([]ListingPublicSchema, len(listings))
	for i, l := range listings {
		result[i] = listingToPublicSchema(&l)
	}
	c.JSON(http.StatusOK, result)
}

// GetPublicListing handles GET /listings/:listing_id.
//
//	@Summary		Get a single published listing
//	@Description	Returns the public-facing details of a specific listing by UUID. No authentication required.
//	@Tags			Listings
//	@Produce		json
//	@Param			listing_id	path		string				true	"Listing UUID"	format(uuid)
//	@Success		200			{object}	ListingPublicSchema
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		404			{object}	auth.ErrorResponse	"LISTING_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Router			/listings/{listing_id} [get]
func (h *Handler) GetPublicListing(c *gin.Context) {
	listingID, err := parseUUID(c, "listing_id")
	if err != nil {
		return
	}

	listing, err := h.Service.GetPublishedByID(c.Request.Context(), listingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	if listing == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "LISTING_NOT_FOUND", "message": "Listing not found."})
		return
	}

	c.JSON(http.StatusOK, listingToPublicSchema(listing))
}

// ─────────────────────────────────────────────────────────────────────────────
// Conversion helpers
// ─────────────────────────────────────────────────────────────────────────────

func draftToResponse(d *db.ListingDraft) DraftDetailResponse {
	r := DraftDetailResponse{
		ID:              d.ID,
		HostID:          d.HostID,
		CurrentStep:     d.CurrentStep,
		Status:          d.Status,
		Type:            d.Type,
		Name:            d.Name,
		Square:          d.Square,
		Floor:           d.Floor,
		TotalFloors:     d.TotalFloors,
		MaxGuests:       d.MaxGuests,
		RoomsCount:      d.RoomsCount,
		BedsCount:       d.BedsCount,
		BathroomsCount:  d.BathroomsCount,
		PricePerNight:   d.PricePerNight,
		Currency:        d.Currency,
		MinNights:       d.MinNights,
		CheckinFrom:     d.CheckinFrom,
		CheckoutUntil:   d.CheckoutUntil,
		AllowChildren:   d.AllowChildren,
		AllowPets:       d.AllowPets,
		AllowSmoking:    d.AllowSmoking,
		AllowParties:    d.AllowParties,
		DepositRequired: d.DepositRequired,
		WithInvoicing:   d.WithInvoicing,
		Description:     d.Description,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
	if d.MediaIDs != nil {
		_ = json.Unmarshal(d.MediaIDs, &r.MediaIDs)
	}
	if d.Amenities != nil {
		_ = json.Unmarshal(d.Amenities, &r.Amenities)
	}
	return r
}

func extractAmenitiesAndMedia(l *db.Listing) ([]string, []string) {
	amenitiesList := make([]string, 0, len(l.ListingAmenities))
	for _, a := range l.ListingAmenities {
		amenitiesList = append(amenitiesList, a.AmenityID)
	}

	mediaList := make([]string, 0, len(l.Media))
	for _, m := range l.Media {
		url := m.FileKey
		if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://localhost:8000/api/v1/media/dev-upload/" + url
		}
		mediaList = append(mediaList, url)
	}
	return amenitiesList, mediaList
}

func listingToHostSchema(l *db.Listing) ListingHostReadSchema {
	amenitiesList, mediaList := extractAmenitiesAndMedia(l)
	return ListingHostReadSchema{
		ID:              l.ID,
		Status:          l.Status,
		Type:            l.Type,
		Name:            l.Name,
		Square:          l.Square,
		Floor:           l.Floor,
		TotalFloors:     l.TotalFloors,
		MaxGuests:       l.MaxGuests,
		RoomsCount:      l.RoomsCount,
		BedsCount:       l.BedsCount,
		BathroomsCount:  l.BathroomsCount,
		PricePerNight:   l.PricePerNight,
		Currency:        l.Currency,
		MinNights:       l.MinNights,
		CheckinFrom:     l.CheckinFrom,
		CheckoutUntil:   l.CheckoutUntil,
		AllowChildren:   l.AllowChildren,
		AllowPets:       l.AllowPets,
		AllowSmoking:    l.AllowSmoking,
		AllowParties:    l.AllowParties,
		DepositRequired: l.DepositRequired,
		WithInvoicing:   l.WithInvoicing,
		Description:     l.Description,
		Amenities:       amenitiesList,
		Media:           mediaList,
		CreatedAt:       l.CreatedAt,
		UpdatedAt:       l.UpdatedAt,
	}
}

func listingToPublicSchema(l *db.Listing) ListingPublicSchema {
	amenitiesList, mediaList := extractAmenitiesAndMedia(l)
	var hostSchema *HostSchema
	if l.Host != nil {
		hostName := l.Host.Phone
		if l.Host.FirstName != nil && *l.Host.FirstName != "" {
			hostName = *l.Host.FirstName
			if l.Host.LastName != nil && *l.Host.LastName != "" {
				hostName += " " + *l.Host.LastName
			}
		}
		hostSchema = &HostSchema{
			ID:        l.Host.ID,
			Name:      hostName,
			FirstName: l.Host.FirstName,
			LastName:  l.Host.LastName,
			Phone:     l.Host.Phone,
			Role:      string(l.Host.Role),
		}
	}
	return ListingPublicSchema{
		ID:              l.ID,
		Host:            hostSchema,
		Type:            l.Type,
		Name:            l.Name,
		Square:          l.Square,
		Floor:           l.Floor,
		TotalFloors:     l.TotalFloors,
		MaxGuests:       l.MaxGuests,
		RoomsCount:      l.RoomsCount,
		BedsCount:       l.BedsCount,
		BathroomsCount:  l.BathroomsCount,
		PricePerNight:   l.PricePerNight,
		Currency:        l.Currency,
		MinNights:       l.MinNights,
		CheckinFrom:     l.CheckinFrom,
		CheckoutUntil:   l.CheckoutUntil,
		AllowChildren:   l.AllowChildren,
		AllowPets:       l.AllowPets,
		AllowSmoking:    l.AllowSmoking,
		AllowParties:    l.AllowParties,
		DepositRequired: l.DepositRequired,
		WithInvoicing:   l.WithInvoicing,
		Description:     l.Description,
		Amenities:       amenitiesList,
		Media:           mediaList,
		CreatedAt:       l.CreatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Error handling
// ─────────────────────────────────────────────────────────────────────────────

func handleListingError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *ValidationError:
		resp := gin.H{"code": e.Code, "message": e.Message}
		if e.Details != nil {
			resp["details"] = e.Details
		}
		c.JSON(http.StatusBadRequest, resp)
	case *DraftNotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"code": "DRAFT_NOT_FOUND", "message": e.Error()})
	case *DraftAlreadySubmittedError:
		c.JSON(http.StatusConflict, gin.H{"code": "DRAFT_ALREADY_SUBMITTED", "message": e.Error()})
	case *DraftIncompleteError:
		c.JSON(http.StatusBadRequest, gin.H{"code": "DRAFT_INCOMPLETE", "message": e.Error()})
	case *InvalidStepTransitionError:
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_STEP_TRANSITION", "message": e.Error()})
	case *MediaNotOwnedError:
		c.JSON(http.StatusForbidden, gin.H{"code": e.Code, "message": e.Message})
	case *IdempotencyConflictError:
		c.JSON(http.StatusConflict, gin.H{"code": e.Code, "message": e.Message})
	case *ListingNotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"code": "LISTING_NOT_FOUND", "message": e.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "An internal error occurred."})
	}
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid ID format."})
		return uuid.Nil, err
	}
	return id, nil
}
