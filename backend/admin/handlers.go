// Package admin provides HTTP handlers for admin endpoints.
package admin

import (
	"net/http"

	"flickey/go-backend/auth"
	"flickey/go-backend/config"
	"flickey/go-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *AdminService
	Cfg     *config.Settings
}

func NewHandler(svc *AdminService, cfg *config.Settings) *Handler {
	return &Handler{Service: svc, Cfg: cfg}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc, activeMw gin.HandlerFunc) {
	adminMw := auth.RequireRole(db.UserRoleAdmin)

	admin := rg.Group("")
	admin.Use(authMw, activeMw, adminMw)

	// Dashboard Overview
	admin.GET("/overview", h.GetOverview)

	// Listing Moderation
	admin.GET("/listings", h.GetListings)
	admin.GET("/listings/:id", h.GetListingDetail)
	admin.POST("/listings/:id/moderate", h.ModerateListing)

	// User Directory & Enforcement
	admin.GET("/users", h.GetUsers)
	admin.GET("/users/:id", h.GetUserDetail)
	admin.POST("/users/:id/enforce", h.EnforceUser)
	admin.POST("/users/:id/role", h.UpdateUserRole)

	// Verifications
	admin.GET("/verification", h.GetVerifications)
	admin.GET("/verification/:id", h.GetVerificationDetail)
	admin.POST("/verification/:id/review", h.ReviewVerification)

	// Reports
	admin.GET("/reports", h.GetReports)
	admin.POST("/reports/:id/resolve", h.ResolveReport)

	// Audit Logs
	admin.GET("/audit", h.GetAuditLogs)
}

// GetOverview handles GET /admin/overview.
//
//	@Summary		Get admin overview metrics
//	@Description	Returns counts of total users, listings, pending verifications, active listings, etc.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	OverviewMetrics
//	@Failure		401	{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403	{object}	auth.ErrorResponse	"INSUFFICIENT_ROLE"
//	@Failure		500	{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/admin/overview [get]
func (h *Handler) GetOverview(c *gin.Context) {
	metrics, err := h.Service.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func formatAdminListing(l *db.Listing, uploadBaseURL string) gin.H {
	amenitiesList := make([]string, len(l.ListingAmenities))
	for i, a := range l.ListingAmenities {
		amenitiesList[i] = a.AmenityID
	}

	mediaList := make([]string, len(l.Media))
	for i, m := range l.Media {
		mediaList[i] = m.PublicURL(uploadBaseURL)
	}

	var verificationVideoURL *string
	if l.VerificationVideo != nil {
		u := l.VerificationVideo.PublicURL(uploadBaseURL)
		verificationVideoURL = &u
	}

	var hostMap gin.H
	if l.Host != nil {
		hostName := l.Host.Phone
		if l.Host.FirstName != nil && *l.Host.FirstName != "" {
			hostName = *l.Host.FirstName
			if l.Host.LastName != nil && *l.Host.LastName != "" {
				hostName += " " + *l.Host.LastName
			}
		}
		hostMap = gin.H{
			"id":         l.Host.ID,
			"name":       hostName,
			"first_name": l.Host.FirstName,
			"last_name":  l.Host.LastName,
			"phone":      l.Host.Phone,
			"email":      l.Host.Email,
			"role":       l.Host.Role,
		}
	}

	return gin.H{
		"id":                     l.ID,
		"host_id":                l.HostID,
		"host":                   hostMap,
		"verification_video_url": verificationVideoURL,
		"verification_video_id":  l.VerificationVideoID,
		"status":                 l.Status,
		"type":             l.Type,
		"name":             l.Name,
		"address":          l.Address,
		"city":             l.City,
		"street":           l.Street,
		"house_number":     l.HouseNumber,
		"latitude":         l.Latitude,
		"longitude":        l.Longitude,
		"square":           l.Square,
		"floor":            l.Floor,
		"total_floors":     l.TotalFloors,
		"max_guests":       l.MaxGuests,
		"rooms_count":      l.RoomsCount,
		"beds_count":       l.BedsCount,
		"bathrooms_count":  l.BathroomsCount,
		"price_per_night":  l.PricePerNight,
		"currency":         l.Currency,
		"min_nights":       l.MinNights,
		"checkin_from":     l.CheckinFrom,
		"checkout_until":   l.CheckoutUntil,
		"allow_children":   l.AllowChildren,
		"allow_pets":       l.AllowPets,
		"allow_smoking":    l.AllowSmoking,
		"allow_parties":    l.AllowParties,
		"deposit_required": l.DepositRequired,
		"with_invoicing":   l.WithInvoicing,
		"description":      l.Description,
		"amenities":        amenitiesList,
		"media":            mediaList,
		"created_at":       l.CreatedAt,
		"updated_at":       l.UpdatedAt,
	}
}

// GetListings handles GET /admin/listings.
//
//	@Summary		List all listings
//	@Description	Returns a list of listings, optionally filtered by status.
//	@Tags			Admin
//	@Produce		json
//	@Param			status	query		string	false	"Listing status filter (pending, active, rejected, suspended, archived)"	Enums(pending, active, rejected, suspended, archived)
//	@Success		200		{array}		object	"List of listings with host details"
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/listings [get]
func (h *Handler) GetListings(c *gin.Context) {
	status := c.Query("status")
	listings, err := h.Service.GetListings(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	baseURL := ""
	if h.Cfg != nil {
		baseURL = h.Cfg.S3PublicBaseURL
	}
	out := make([]gin.H, len(listings))
	for i, l := range listings {
		out[i] = formatAdminListing(&l, baseURL)
	}

	c.JSON(http.StatusOK, out)
}

// GetListingDetail handles GET /admin/listings/:id.
//
//	@Summary		Get listing details
//	@Description	Returns full listing details for moderation.
//	@Tags			Admin
//	@Produce		json
//	@Param			id	path		string	true	"Listing UUID"	format(uuid)
//	@Success		200	{object}	object
//	@Failure		400	{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401	{object}	auth.ErrorResponse
//	@Failure		403	{object}	auth.ErrorResponse
//	@Failure		404	{object}	auth.ErrorResponse	"LISTING_NOT_FOUND"
//	@Security		BearerAuth
//	@Router			/admin/listings/{id} [get]
func (h *Handler) GetListingDetail(c *gin.Context) {
	idStr := c.Param("id")
	listingID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing ID"})
		return
	}

	listing, err := h.Service.GetListingByID(c.Request.Context(), listingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "LISTING_NOT_FOUND", "message": "Listing not found"})
		return
	}

	baseURL := ""
	if h.Cfg != nil {
		baseURL = h.Cfg.S3PublicBaseURL
	}
	c.JSON(http.StatusOK, formatAdminListing(listing, baseURL))
}

type ModerateRequest struct {
	Action string `json:"action" binding:"required"` // approve, reject, request_changes, suspend
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

// ModerateListing handles POST /admin/listings/:id/moderate.
//
//	@Summary		Moderate a listing
//	@Description	Approve, reject, or suspend a listing. Transitions state and records an audit log.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Listing UUID"	format(uuid)
//	@Param			body	body		ModerateRequest	true	"Moderation action (approve/reject/request_changes/suspend) and reason"
//	@Success		200		{object}	object
//	@Failure		400		{object}	auth.ErrorResponse	"VALIDATION_ERROR"
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/listings/{id}/moderate [post]
func (h *Handler) ModerateListing(c *gin.Context) {
	admin := auth.GetCurrentUser(c)
	idStr := c.Param("id")
	listingID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing ID"})
		return
	}

	var req ModerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	listing, err := h.Service.ModerateListing(c.Request.Context(), admin, listingID, req.Action, req.Reason, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "MODERATION_ERROR", "message": err.Error()})
		return
	}

	baseURL := ""
	if h.Cfg != nil {
		baseURL = h.Cfg.S3PublicBaseURL
	}
	c.JSON(http.StatusOK, formatAdminListing(listing, baseURL))
}

// GetUsers handles GET /admin/users.
//
//	@Summary		List users
//	@Description	Returns a list of users, optionally filtered by role and status.
//	@Tags			Admin
//	@Produce		json
//	@Param			role	query		string	false	"Filter by role"
//	@Param			status	query		string	false	"Filter by status"
//	@Success		200		{array}		db.User
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	role := c.Query("role")
	status := c.Query("status")
	users, err := h.Service.GetUsers(c.Request.Context(), role, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUserDetail handles GET /admin/users/:id.
//
//	@Summary		Get user details
//	@Description	Returns a user's details and their listings.
//	@Tags			Admin
//	@Produce		json
//	@Param			id	path		string	true	"User UUID"	format(uuid)
//	@Success		200	{object}	object
//	@Failure		400	{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401	{object}	auth.ErrorResponse
//	@Failure		403	{object}	auth.ErrorResponse
//	@Failure		404	{object}	auth.ErrorResponse	"USER_NOT_FOUND"
//	@Security		BearerAuth
//	@Router			/admin/users/{id} [get]
func (h *Handler) GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid user ID"})
		return
	}

	var user db.User
	if err := h.Service.DB.WithContext(c.Request.Context()).Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "USER_NOT_FOUND", "message": "User not found"})
		return
	}

	var listings []db.Listing
	h.Service.DB.WithContext(c.Request.Context()).
		Preload("ListingAmenities").
		Preload("Media").
		Where("host_id = ?", userID).
		Order("created_at DESC").
		Find(&listings)

	baseURL := ""
	if h.Cfg != nil {
		baseURL = h.Cfg.S3PublicBaseURL
	}
	formattedListings := make([]gin.H, len(listings))
	for i, l := range listings {
		formattedListings[i] = formatAdminListing(&l, baseURL)
	}

	var restrictions []db.UserRestriction
	h.Service.DB.WithContext(c.Request.Context()).Where("user_id = ?", userID).Order("created_at DESC").Find(&restrictions)

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"listings":     formattedListings,
		"restrictions": restrictions,
	})
}

type EnforceRequest struct {
	EnforcementType string `json:"enforcement_type" binding:"required"` // warning, temporary_restriction, temporary_block, permanent_block, unblock
	Reason          string `json:"reason" binding:"required"`
	Note            string `json:"note"`
	DurationDays    int    `json:"duration_days"`
}

// EnforceUser handles POST /admin/users/:id/enforce.
//
//	@Summary		Enforce user restrictions
//	@Description	Applies a warning, temporary block, or ban to a user. Records an audit log.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"User UUID"	format(uuid)
//	@Param			body	body		EnforceRequest	true	"Enforcement action and reason"
//	@Success		200		{object}	db.User
//	@Failure		400		{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_ID"
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/enforce [post]
func (h *Handler) EnforceUser(c *gin.Context) {
	admin := auth.GetCurrentUser(c)
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid user ID"})
		return
	}

	var req EnforceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	user, err := h.Service.EnforceUser(c.Request.Context(), admin, userID, req.EnforcementType, req.Reason, req.Note, req.DurationDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ENFORCEMENT_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

type RoleUpdateRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateUserRole handles POST /admin/users/:id/role.
//
//	@Summary		Update user role
//	@Description	Changes a user's role (guest, host, admin, moderator, support).
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"User UUID"	format(uuid)
//	@Param			body	body		RoleUpdateRequest	true	"Target user role"
//	@Success		200		{object}	db.User
//	@Failure		400		{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_ID"
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/role [post]
func (h *Handler) UpdateUserRole(c *gin.Context) {
	admin := auth.GetCurrentUser(c)
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid user ID"})
		return
	}

	var req RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	user, err := h.Service.UpdateUserRole(c.Request.Context(), admin, userID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ROLE_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetVerifications handles GET /admin/verification.
//
//	@Summary		List identity verification requests
//	@Description	Returns verification requests optionally filtered by status.
//	@Tags			Admin
//	@Produce		json
//	@Param			status	query		string	false	"Verification status filter (pending, approved, rejected, changes_requested)"	Enums(pending, approved, rejected, changes_requested)
//	@Success		200		{array}		db.VerificationRequest
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/verification [get]
func (h *Handler) GetVerifications(c *gin.Context) {
	status := c.Query("status")
	items, err := h.Service.GetVerifications(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetVerificationDetail handles GET /admin/verification/:id.
//
//	@Summary		Get verification request details
//	@Description	Returns the details of a single verification request.
//	@Tags			Admin
//	@Produce		json
//	@Param			id	path		string	true	"Verification request UUID"	format(uuid)
//	@Success		200	{object}	db.VerificationRequest
//	@Failure		400	{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401	{object}	auth.ErrorResponse
//	@Failure		403	{object}	auth.ErrorResponse
//	@Failure		404	{object}	auth.ErrorResponse	"VERIFICATION_NOT_FOUND"
//	@Security		BearerAuth
//	@Router			/admin/verification/{id} [get]
func (h *Handler) GetVerificationDetail(c *gin.Context) {
	idStr := c.Param("id")
	reqID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid request ID"})
		return
	}

	var item db.VerificationRequest
	if err := h.Service.DB.WithContext(c.Request.Context()).Where("id = ?", reqID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "VERIFICATION_NOT_FOUND", "message": "Verification request not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

type ReviewVerificationRequest struct {
	Action string `json:"action" binding:"required"` // approve, reject, request_changes
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

// ReviewVerification handles POST /admin/verification/:id/review.
//
//	@Summary		Review identity verification request
//	@Description	Approves, rejects, or requests changes for a verification request. Updates user status and logs audit entry.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Verification request UUID"	format(uuid)
//	@Param			body	body		ReviewVerificationRequest	true	"Review action (approve/reject/request_changes) and reason"
//	@Success		200		{object}	db.VerificationRequest
//	@Failure		400		{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_ID"
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/verification/{id}/review [post]
func (h *Handler) ReviewVerification(c *gin.Context) {
	admin := auth.GetCurrentUser(c)
	idStr := c.Param("id")
	reqID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid request ID"})
		return
	}

	var req ReviewVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	item, err := h.Service.ReviewVerification(c.Request.Context(), admin, reqID, req.Action, req.Reason, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "REVIEW_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetReports handles GET /admin/reports.
//
//	@Summary		List reports
//	@Description	Returns user/listing reports optionally filtered by status.
//	@Tags			Admin
//	@Produce		json
//	@Param			status	query		string	false	"Report status filter (pending, resolved, dismissed)"	Enums(pending, resolved, dismissed)
//	@Success		200		{array}		db.Report
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/reports [get]
func (h *Handler) GetReports(c *gin.Context) {
	status := c.Query("status")
	reports, err := h.Service.GetReports(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
}

type ResolveReportRequest struct {
	Action string `json:"action" binding:"required"` // resolve, dismiss
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

// ResolveReport handles POST /admin/reports/:id/resolve.
//
//	@Summary		Resolve or dismiss a report
//	@Description	Marks a moderation report as resolved or dismissed, recording note and reason.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Report UUID"	format(uuid)
//	@Param			body	body		ResolveReportRequest	true	"Resolution action (resolve/dismiss) and details"
//	@Success		200		{object}	db.Report
//	@Failure		400		{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_ID"
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		403		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/reports/{id}/resolve [post]
func (h *Handler) ResolveReport(c *gin.Context) {
	admin := auth.GetCurrentUser(c)
	idStr := c.Param("id")
	reportID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid report ID"})
		return
	}

	var req ResolveReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	report, err := h.Service.ResolveReport(c.Request.Context(), admin, reportID, req.Action, req.Reason, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "RESOLUTION_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetAuditLogs handles GET /admin/audit.
//
//	@Summary		List audit logs
//	@Description	Returns recent administrative audit log entries.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{array}		db.AuditLog
//	@Failure		401	{object}	auth.ErrorResponse
//	@Failure		403	{object}	auth.ErrorResponse
//	@Failure		500	{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/audit [get]
func (h *Handler) GetAuditLogs(c *gin.Context) {
	logs, err := h.Service.GetAuditLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
