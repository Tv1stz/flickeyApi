// Package listings provides HTTP handlers for listing and draft endpoints.
package listings

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"flickey/go-backend/auth"
	"flickey/go-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	active.POST("/drafts/from-listing/:listing_id", h.CreateDraftFromListing)
	active.GET("/drafts/:draft_id", h.GetDraft)
	active.PATCH("/drafts/:draft_id/step-2", h.UpdateStep2)
	active.PATCH("/drafts/:draft_id/step-3", h.UpdateStep3)
	active.PATCH("/drafts/:draft_id/step-4", h.UpdateStep4)
	active.PATCH("/drafts/:draft_id/step-5", h.UpdateStep5)
	active.PATCH("/drafts/:draft_id/step-6", h.UpdateStep6)
	active.POST("/drafts/:draft_id/submit", h.SubmitDraft)
	active.GET("/my", h.GetMyListings)
	active.PATCH("/:listing_id", h.UpdateListing)
	active.POST("/:listing_id/verification-video", h.AttachVerificationVideo)
	active.POST("/:listing_id/archive", h.ArchiveListing)
	active.POST("/:listing_id/unarchive", h.UnarchiveListing)
	active.DELETE("/:listing_id", h.DeleteListing)

	// Public routes — optional auth on :listing_id lets hosts preview their own unpublished listings.
	optionalAuthMw := auth.OptionalAuth(h.Service.Cfg, h.Service.DB)
	rg.GET("", h.BrowseListings)
	rg.GET("/:listing_id", optionalAuthMw, h.GetPublicListing)
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
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": formatBindingError(err)})
		return
	}

	draft, err := h.Service.CreateDraft(c.Request.Context(), user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	completedSteps := h.Service.registry().GetCompletedSteps(draft)
	c.JSON(http.StatusCreated, DraftCreateResponse{
		DraftID:         draft.ID,
		SourceListingID: draft.SourceListingID,
		Mode:            draft.Mode,
		CurrentStep:     draft.CurrentStep,
		CompletedSteps:  completedSteps,
		TotalSteps:      h.Service.registry().TotalSteps(),
		Status:          draft.Status,
	})
}

// CreateDraftFromListing handles POST /listings/drafts/from-listing/:listing_id.
//
//	@Summary		Create draft from existing/rejected listing
//	@Description	Creates a new editable draft populated with data from an existing or previously rejected listing.
//	@Tags			Listings
//	@Produce		json
//	@Param			listing_id	path		string	true	"Listing UUID"	format(uuid)
//	@Param			mode		query		string	false	"Draft mode: 'edit' or 'create'"
//	@Success		201			{object}	DraftCreateResponse
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_ID / VALIDATION_ERROR"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"LISTING_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/drafts/from-listing/{listing_id} [post]
func (h *Handler) CreateDraftFromListing(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	listingID, err := parseUUID(c, "listing_id")
	if err != nil {
		return
	}

	mode := c.Query("mode")

	draft, err := h.Service.CreateDraftFromListing(c.Request.Context(), listingID, user.ID, mode)
	if err != nil {
		handleListingError(c, err)
		return
	}

	completedSteps := h.Service.registry().GetCompletedSteps(draft)
	c.JSON(http.StatusCreated, DraftCreateResponse{
		DraftID:         draft.ID,
		SourceListingID: draft.SourceListingID,
		Mode:            draft.Mode,
		CurrentStep:     draft.CurrentStep,
		CompletedSteps:  completedSteps,
		TotalSteps:      h.Service.registry().TotalSteps(),
		Status:          draft.Status,
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

	resp := draftToResponse(draft, h.Service.registry())
	resp.Media = make([]DraftMediaItem, 0)
	if len(resp.MediaIDs) > 0 {
		uuids := make([]uuid.UUID, 0, len(resp.MediaIDs))
		for _, idStr := range resp.MediaIDs {
			if u, err := uuid.Parse(idStr); err == nil {
				uuids = append(uuids, u)
			}
		}
		if len(uuids) > 0 && h.Service.DB != nil {
			var mediaRecords []db.Media
			_ = h.Service.DB.WithContext(c.Request.Context()).Where("id IN ?", uuids).Find(&mediaRecords).Error
			mediaMap := make(map[string]string, len(mediaRecords))
			for _, m := range mediaRecords {
				url := m.FileKey
				if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://localhost:8000/api/v1/media/dev-upload/" + url
				}
				mediaMap[m.ID.String()] = url
			}
			for _, idStr := range resp.MediaIDs {
				if url, ok := mediaMap[idStr]; ok {
					resp.Media = append(resp.Media, DraftMediaItem{ID: idStr, URL: url})
				}
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateStep2 handles PATCH /listings/drafts/:draft_id/step-2.
//
//	@Summary		Update draft – step 2 (property details)
//	@Description	Sets name, area, floor, guest count, rooms, beds, and bathrooms.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep2Request	true	"Property details"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / STEP_NOT_ALLOWED"
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
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": formatBindingError(err)})
		return
	}

	draft, err := h.Service.UpdateStep2(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	completedSteps := h.Service.registry().GetCompletedSteps(draft)
	c.JSON(http.StatusOK, DraftStepResponse{
		DraftID:        draft.ID,
		CurrentStep:    draft.CurrentStep,
		CompletedSteps: completedSteps,
		TotalSteps:     h.Service.registry().TotalSteps(),
		Status:         draft.Status,
	})
}

// UpdateStep3 handles PATCH /listings/drafts/:draft_id/step-3.
//
//	@Summary		Update draft – step 3 (media)
//	@Description	Attaches uploaded media UUIDs to the draft.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep3Request	true	"Array of media UUIDs"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / STEP_NOT_ALLOWED"
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
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": formatBindingError(err)})
		return
	}

	draft, err := h.Service.UpdateStep3(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	completedSteps := h.Service.registry().GetCompletedSteps(draft)
	c.JSON(http.StatusOK, DraftStepResponse{
		DraftID:        draft.ID,
		CurrentStep:    draft.CurrentStep,
		CompletedSteps: completedSteps,
		TotalSteps:     h.Service.registry().TotalSteps(),
		Status:         draft.Status,
	})
}

// UpdateStep4 handles PATCH /listings/drafts/:draft_id/step-4.
//
//	@Summary		Update draft – step 4 (amenities)
//	@Description	Attaches one or more amenity IDs from the amenity registry to the draft.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep4Request	true	"Array of amenity ID strings"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / STEP_NOT_ALLOWED"
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
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": formatBindingError(err)})
		return
	}

	draft, err := h.Service.UpdateStep4(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	completedSteps := h.Service.registry().GetCompletedSteps(draft)
	c.JSON(http.StatusOK, DraftStepResponse{
		DraftID:        draft.ID,
		CurrentStep:    draft.CurrentStep,
		CompletedSteps: completedSteps,
		TotalSteps:     h.Service.registry().TotalSteps(),
		Status:         draft.Status,
	})
}

// UpdateStep5 handles PATCH /listings/drafts/:draft_id/step-5.
//
//	@Summary		Update draft – step 5 (pricing & rules)
//	@Description	Sets price per night, currency, minimum nights, check-in/check-out times, and house rules.
//	@Tags			Listings
//	@Accept			json
//	@Produce		json
//	@Param			draft_id	path		string				true	"Draft UUID"	format(uuid)
//	@Param			body		body		DraftStep5Request	true	"Pricing and house rules"
//	@Success		200			{object}	DraftStepResponse
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / STEP_NOT_ALLOWED"
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
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": formatBindingError(err)})
		return
	}

	draft, err := h.Service.UpdateStep5(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	completedSteps := h.Service.registry().GetCompletedSteps(draft)
	c.JSON(http.StatusOK, DraftStepResponse{
		DraftID:        draft.ID,
		CurrentStep:    draft.CurrentStep,
		CompletedSteps: completedSteps,
		TotalSteps:     h.Service.registry().TotalSteps(),
		Status:         draft.Status,
	})
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
//	@Failure		400			{object}	auth.ErrorResponse	"VALIDATION_ERROR / STEP_NOT_ALLOWED"
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
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": formatBindingError(err)})
		return
	}

	draft, err := h.Service.UpdateStep6(c.Request.Context(), draftID, user.ID, req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	completedSteps := h.Service.registry().GetCompletedSteps(draft)
	c.JSON(http.StatusOK, DraftStepResponse{
		DraftID:        draft.ID,
		CurrentStep:    draft.CurrentStep,
		CompletedSteps: completedSteps,
		TotalSteps:     h.Service.registry().TotalSteps(),
		Status:         draft.Status,
	})
}

// SubmitDraft handles POST /listings/drafts/:draft_id/submit.
//
//	@Summary		Submit draft for review
//	@Description	Finalises a completed 6-step draft and creates a published listing pending moderation.
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

// AttachVerificationVideo handles POST /listings/:listing_id/verification-video.
// Attaches a private apartment walkthrough video to the listing and updates its moderation status.
func (h *Handler) AttachVerificationVideo(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	listingID, err := parseUUID(c, "listing_id")
	if err != nil {
		return
	}

	var req AttachVerificationVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": formatBindingError(err)})
		return
	}

	listing, err := h.Service.AttachVerificationVideo(c.Request.Context(), listingID, user.ID, req.MediaID)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, listingToHostSchema(listing))
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
	var checkin, checkout time.Time
	if ci := c.Query("checkin"); ci != "" {
		if t, err := time.Parse("2006-01-02", ci); err == nil {
			checkin = t
		}
	}
	if co := c.Query("checkout"); co != "" {
		if t, err := time.Parse("2006-01-02", co); err == nil {
			checkout = t
		}
	}

	var hostID *uuid.UUID
	if hid := c.Query("host_id"); hid != "" {
		if u, err := uuid.Parse(hid); err == nil {
			hostID = &u
		}
	}

	listings, err := h.Service.GetPublishedListings(c.Request.Context(), checkin, checkout, hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	result := make([]ListingPublicSchema, len(listings))
	for i, l := range listings {
		s := listingToPublicSchema(&l)
		s.Status = ""
		result[i] = s
	}
	c.JSON(http.StatusOK, result)
}

// GetPublicHostProfile handles GET /users/:id.
//
//	@Summary		Get public host profile
//	@Description	Returns public profile information of a host along with published listings count.
//	@Tags			Users
//	@Produce		json
//	@Param			id	path		string	true	"User UUID"	format(uuid)
//	@Success		200	{object}	PublicUserProfileSchema
//	@Failure		400	{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		404	{object}	auth.ErrorResponse	"USER_NOT_FOUND"
//	@Failure		500	{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Router			/users/{id} [get]
func (h *Handler) GetPublicHostProfile(c *gin.Context) {
	userID, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	user, err := db.FindUserByIDCtx(c.Request.Context(), h.Service.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "USER_NOT_FOUND", "message": "User not found."})
		return
	}

	var count int64
	if err := h.Service.DB.WithContext(c.Request.Context()).
		Model(&db.Listing{}).
		Where("host_id = ? AND status = ?", userID, "published").
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	name := user.Phone
	if user.FirstName != nil && *user.FirstName != "" {
		name = *user.FirstName
		if user.LastName != nil && *user.LastName != "" {
			name += " " + *user.LastName
		}
	}

	c.JSON(http.StatusOK, PublicUserProfileSchema{
		ID:            user.ID,
		Name:          name,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Role:          string(user.Role),
		IsVerified:    user.Status == db.UserStatusActive,
		ListingsCount: int(count),
		CreatedAt:     user.CreatedAt,
	})
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

	user := auth.GetCurrentUser(c)
	var currentUserID *uuid.UUID
	isAdmin := false
	if user != nil {
		currentUserID = &user.ID
		if user.Role == db.UserRoleAdmin {
			isAdmin = true
		}
	}

	listing, err := h.Service.GetListingForView(c.Request.Context(), listingID, currentUserID, isAdmin)
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

// DeleteListing handles DELETE /listings/:listing_id.
//
//	@Summary		Delete host listing
//	@Description	Deletes a listing owned by the authenticated host.
//	@Tags			Listings
//	@Param			listing_id	path		string	true	"Listing UUID"	format(uuid)
//	@Success		204			"Listing deleted"
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"LISTING_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/{listing_id} [delete]
func (h *Handler) DeleteListing(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	listingID, err := parseUUID(c, "listing_id")
	if err != nil {
		return
	}

	if err := h.Service.DeleteListing(c.Request.Context(), listingID, user.ID); err != nil {
		handleListingError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateListing handles PATCH /listings/:listing_id.
//
//	@Summary		Update host listing
//	@Description	Updates fields of an existing listing owned by the authenticated host.
//	@Tags			Listings
//	@Param			listing_id	path		string					true	"Listing UUID"	format(uuid)
//	@Param			body		body		UpdateListingRequest	true	"Fields to update"
//	@Success		200			{object}	ListingHostReadSchema
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_ID / VALIDATION_ERROR"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"LISTING_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/{listing_id} [patch]
func (h *Handler) UpdateListing(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	listingID, err := parseUUID(c, "listing_id")
	if err != nil {
		return
	}

	var req UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_BODY", "message": err.Error()})
		return
	}

	listing, err := h.Service.UpdateListing(c.Request.Context(), listingID, user.ID, &req)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, listingToHostSchema(listing))
}

// ArchiveListing handles POST /listings/:listing_id/archive.
//
//	@Summary		Archive host listing
//	@Description	Archives an active listing owned by the host (removes it from public search).
//	@Tags			Listings
//	@Param			listing_id	path		string	true	"Listing UUID"	format(uuid)
//	@Success		200			{object}	ListingHostReadSchema
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"LISTING_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/{listing_id}/archive [post]
func (h *Handler) ArchiveListing(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	listingID, err := parseUUID(c, "listing_id")
	if err != nil {
		return
	}

	listing, err := h.Service.ArchiveListing(c.Request.Context(), listingID, user.ID)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, listingToHostSchema(listing))
}

// UnarchiveListing handles POST /listings/:listing_id/unarchive.
//
//	@Summary		Unarchive host listing
//	@Description	Restores an archived listing owned by the host to published status (or pending_review if unapproved changes exist).
//	@Tags			Listings
//	@Param			listing_id	path		string	true	"Listing UUID"	format(uuid)
//	@Success		200			{object}	ListingHostReadSchema
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		404			{object}	auth.ErrorResponse	"LISTING_NOT_FOUND"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/listings/{listing_id}/unarchive [post]
func (h *Handler) UnarchiveListing(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	listingID, err := parseUUID(c, "listing_id")
	if err != nil {
		return
	}

	listing, err := h.Service.UnarchiveListing(c.Request.Context(), listingID, user.ID)
	if err != nil {
		handleListingError(c, err)
		return
	}

	c.JSON(http.StatusOK, listingToHostSchema(listing))
}

// ─────────────────────────────────────────────────────────────────────────────
// Conversion helpers
// ─────────────────────────────────────────────────────────────────────────────

func draftToResponse(d *db.ListingDraft, reg *StepRegistry) DraftDetailResponse {
	if reg == nil {
		reg = DefaultRegistry
	}
	r := DraftDetailResponse{
		ID:              d.ID,
		HostID:          d.HostID,
		SourceListingID: d.SourceListingID,
		Mode:            d.Mode,
		CurrentStep:     d.CurrentStep,
		CompletedSteps:  reg.GetCompletedSteps(d),
		TotalSteps:      reg.TotalSteps(),
		Status:          d.Status,
		Type:            d.Type,
		Name:            d.Name,
		Address:         d.Address,
		City:            d.City,
		Street:          d.Street,
		HouseNumber:     d.HouseNumber,
		Latitude:        d.Latitude,
		Longitude:       d.Longitude,
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

	var videoURL *string
	if l.VerificationVideo != nil {
		u := l.VerificationVideo.PublicURL("")
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			u = "http://localhost:8000" + u
		}
		videoURL = &u
	}

	return ListingHostReadSchema{
		ID:                   l.ID,
		Status:               l.Status,
		Type:                 l.Type,
		Name:                 l.Name,
		Address:              l.Address,
		City:                 l.City,
		Street:               l.Street,
		HouseNumber:          l.HouseNumber,
		Latitude:             l.Latitude,
		Longitude:            l.Longitude,
		Square:               l.Square,
		Floor:                l.Floor,
		TotalFloors:          l.TotalFloors,
		MaxGuests:            l.MaxGuests,
		RoomsCount:           l.RoomsCount,
		BedsCount:            l.BedsCount,
		BathroomsCount:       l.BathroomsCount,
		PricePerNight:        l.PricePerNight,
		Currency:             l.Currency,
		MinNights:            l.MinNights,
		CheckinFrom:          l.CheckinFrom,
		CheckoutUntil:        l.CheckoutUntil,
		AllowChildren:        l.AllowChildren,
		AllowPets:            l.AllowPets,
		AllowSmoking:         l.AllowSmoking,
		AllowParties:         l.AllowParties,
		DepositRequired:      l.DepositRequired,
		WithInvoicing:        l.WithInvoicing,
		Description:          l.Description,
		Amenities:            amenitiesList,
		Media:                mediaList,
		VerificationVideoURL: videoURL,
		VerificationVideoID:  l.VerificationVideoID,
		RejectionReason:      l.RejectionReason,
		ModerationComment:    l.ModerationComment,
		CreatedAt:            l.CreatedAt,
		UpdatedAt:            l.UpdatedAt,
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
		HostID:          l.HostID,
		Host:            hostSchema,
		Status:          l.Status,
		Type:            l.Type,
		Name:            l.Name,
		Address:         l.Address,
		City:            l.City,
		Street:          l.Street,
		HouseNumber:     l.HouseNumber,
		Latitude:        l.Latitude,
		Longitude:       l.Longitude,
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

func formatBindingError(err error) string {
	var valErr validator.ValidationErrors
	if errors.As(err, &valErr) {
		for _, fe := range valErr {
			field := fe.Field()
			tag := fe.Tag()
			param := fe.Param()

			switch field {
			case "TotalFloors":
				if tag == "max" {
					return "Количество этажей не может превышать 150"
				}
				if tag == "min" || tag == "required" {
					return "Укажите количество этажей (минимум 1)"
				}
			case "Floor":
				if tag == "max" {
					return "Этаж не может превышать 150"
				}
				if tag == "min" || tag == "required" {
					return "Укажите этаж (минимум 1)"
				}
			case "Square":
				if tag == "gt" {
					return "Площадь должна быть больше 10 м²"
				}
				if tag == "lt" {
					return "Площадь должна быть меньше 1000 м²"
				}
				return "Укажите площадь жилья"
			case "Name":
				if tag == "min" {
					return "Название должно содержать минимум 10 символов"
				}
				if tag == "max" {
					return "Название не должно превышать 100 символов"
				}
				return "Укажите название объявления"
			case "Address":
				return "Укажите адрес объекта"
			case "Latitude", "Longitude":
				return "Укажите расположение на карте"
			case "MaxGuests":
				if tag == "max" {
					return "Максимальное число гостей: не более 50"
				}
				return "Укажите количество гостей (минимум 1)"
			case "RoomsCount":
				if tag == "max" {
					return "Количество комнат не может превышать 30"
				}
				return "Укажите количество комнат (минимум 1)"
			case "BedsCount":
				if tag == "max" {
					return "Количество спальных мест не может превышать 30"
				}
				return "Укажите количество спальных мест (минимум 1)"
			case "BathroomsCount":
				if tag == "max" {
					return "Количество ванных комнат не может превышать 20"
				}
				return "Укажите количество ванных комнат (минимум 1)"
			case "MediaIDs":
				if tag == "min" {
					return "Загрузите не менее 5 фотографий"
				}
				if tag == "max" {
					return "Можно загрузить не более 25 фотографий"
				}
				return "Добавьте фотографии жилья"
			case "Amenities":
				return "Выберите хотя бы одно удобство"
			case "PricePerNight":
				return "Укажите корректную стоимость за ночь (больше 0)"
			case "MinNights":
				if tag == "max" {
					return "Минимальный срок проживания: не более 30 ночей"
				}
				return "Укажите минимальный срок проживания (от 1 ночи)"
			case "CheckinFrom", "CheckoutUntil":
				return "Укажите время заезда и выезда"
			case "Description":
				if tag == "min" {
					return "Описание должно содержать не менее 30 символов"
				}
				if tag == "max" {
					return "Описание не должно превышать 5000 символов"
				}
				return "Заполните описание жилья"
			default:
				if param != "" {
					return fmt.Sprintf("Поле %s не соответствует требованию %s (%s)", field, tag, param)
				}
				return fmt.Sprintf("Некорректное значение поля %s (%s)", field, tag)
			}
		}
	}
	return "Проверьте корректность введённых данных"
}

func handleListingError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *ValidationError:
		resp := gin.H{"code": e.Code, "message": e.Message}
		if e.Details != nil {
			resp["details"] = e.Details
		}
		c.JSON(http.StatusBadRequest, resp)
	case *StepNotAllowedError:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "STEP_NOT_ALLOWED",
			"message": e.Error(),
			"step":    e.Step,
			"current": e.Current,
		})
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
