package calendar

import (
	"errors"
	"net/http"
	"time"

	"flickey/go-backend/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for calendar export, import, and availability.
type Handler struct {
	Service *CalendarService
}

// NewHandler creates a new calendar Handler.
func NewHandler(svc *CalendarService) *Handler {
	return &Handler{Service: svc}
}

// RegisterRoutes registers all calendar endpoints under the listings router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMw, activeMw gin.HandlerFunc) {
	// ── Public Routes ──────────────────────────────────────────────────────────
	rg.GET("/:listing_id/calendar.ics", h.GetICalFeed)
	rg.GET("/:listing_id/availability", h.GetPublicAvailability)

	// ── Protected Host Routes ──────────────────────────────────────────────────
	hostGroup := rg.Group("/:listing_id/calendar")
	hostGroup.Use(authMw, activeMw)
	{
		hostGroup.GET("", h.GetHostCalendar)
		hostGroup.POST("/block", h.BlockDates)
		hostGroup.DELETE("/reservations/:res_id", h.UnblockDates)
		hostGroup.POST("/syncs", h.AddSyncFeed)
		hostGroup.PATCH("/syncs/:sync_id", h.UpdateSyncFeed)
		hostGroup.DELETE("/syncs/:sync_id", h.DeleteSyncFeed)
		hostGroup.POST("/sync-now", h.SyncNow)
	}
}

// GetICalFeed handles GET /api/v1/listings/:listing_id/calendar.ics
func (h *Handler) GetICalFeed(c *gin.Context) {
	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	data, err := h.Service.GetICalFeed(c.Request.Context(), listingID)
	if err != nil {
		if errors.Is(err, ErrListingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Listing not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", "inline; filename=\"calendar.ics\"")
	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "text/calendar; charset=utf-8", data)
}

// GetPublicAvailability handles GET /api/v1/listings/:listing_id/availability
func (h *Handler) GetPublicAvailability(c *gin.Context) {
	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	ranges, err := h.Service.GetPublicAvailability(c.Request.Context(), listingID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ranges)
}

// GetHostCalendar handles GET /api/v1/listings/:listing_id/calendar
func (h *Handler) GetHostCalendar(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	resp, err := h.Service.GetHostCalendar(c.Request.Context(), user.ID, listingID, from, to)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "You do not own this listing"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type BlockDatesRequest struct {
	StartDate string `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate   string `json:"end_date" binding:"required"`   // YYYY-MM-DD
	Note      string `json:"note"`
}

// BlockDates handles POST /api/v1/listings/:listing_id/calendar/block
func (h *Handler) BlockDates(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	var req BlockDatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "Invalid request body"})
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "start_date must be in YYYY-MM-DD format"})
		return
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "end_date must be in YYYY-MM-DD format"})
		return
	}

	res, err := h.Service.BlockDates(c.Request.Context(), user.ID, listingID, start, end, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "You do not own this listing"})
		case errors.Is(err, ErrInvalidDates):
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATES", "message": "Check-out date must be after check-in date"})
		case errors.Is(err, ErrDateConflict):
			c.JSON(http.StatusConflict, gin.H{"code": "DATE_CONFLICT", "message": "Selected dates conflict with an existing booking or block"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, res)
}

// UnblockDates handles DELETE /api/v1/listings/:listing_id/calendar/reservations/:res_id
func (h *Handler) UnblockDates(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	resID, err := uuid.Parse(c.Param("res_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid reservation UUID"})
		return
	}

	if err := h.Service.UnblockDates(c.Request.Context(), user.ID, listingID, resID); err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "You do not own this listing"})
		case errors.Is(err, ErrReservationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Block not found or cannot be removed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dates unblocked successfully"})
}

type AddSyncFeedRequest struct {
	Name    string `json:"name" binding:"required"`
	FeedURL string `json:"feed_url" binding:"required"`
	Color   string `json:"color"`
}

// AddSyncFeed handles POST /api/v1/listings/:listing_id/calendar/syncs
func (h *Handler) AddSyncFeed(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	var req AddSyncFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "name and feed_url are required"})
		return
	}

	feed, err := h.Service.AddSyncFeed(c.Request.Context(), user.ID, listingID, req.Name, req.FeedURL, req.Color)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "You do not own this listing"})
		case errors.Is(err, ErrBlockedIP), errors.Is(err, ErrInvalidScheme):
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_URL", "message": "Provided calendar URL is invalid or blocked for security reasons"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, feed)
}

type UpdateSyncFeedRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// UpdateSyncFeed handles PATCH /api/v1/listings/:listing_id/calendar/syncs/:sync_id
func (h *Handler) UpdateSyncFeed(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	syncID, err := uuid.Parse(c.Param("sync_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid sync UUID"})
		return
	}

	var req UpdateSyncFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "Invalid request body"})
		return
	}

	feed, err := h.Service.UpdateSyncFeed(c.Request.Context(), user.ID, listingID, syncID, req.Name, req.Color)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "You do not own this listing"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feed)
}

// DeleteSyncFeed handles DELETE /api/v1/listings/:listing_id/calendar/syncs/:sync_id
func (h *Handler) DeleteSyncFeed(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	syncID, err := uuid.Parse(c.Param("sync_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid sync UUID"})
		return
	}

	if err := h.Service.DeleteSyncFeed(c.Request.Context(), user.ID, listingID, syncID); err != nil {
		if errors.Is(err, ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "You do not own this listing"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sync feed deleted successfully"})
}

// SyncNow handles POST /api/v1/listings/:listing_id/calendar/sync-now
func (h *Handler) SyncNow(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	listingID, err := uuid.Parse(c.Param("listing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid listing UUID"})
		return
	}

	if err := h.Service.verifyHostListing(c.Request.Context(), user.ID, listingID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "You do not own this listing"})
		return
	}

	if err := h.Service.SyncListingFeeds(c.Request.Context(), listingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "SYNC_FAILED", "message": err.Error()})
		return
	}

	resp, err := h.Service.GetHostCalendar(c.Request.Context(), user.ID, listingID, time.Time{}, time.Time{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
