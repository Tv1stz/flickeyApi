// Package notifications provides HTTP and SSE stream handlers for notification endpoints.
package notifications

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flickey/go-backend/auth"
	"flickey/go-backend/config"
	"flickey/go-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler provides HTTP endpoints for user notifications.
type Handler struct {
	Service *NotificationService
	Cfg     *config.Settings
	DB      *gorm.DB
}

// NewHandler constructs a new Handler.
func NewHandler(svc *NotificationService, cfg *config.Settings, database *gorm.DB) *Handler {
	return &Handler{
		Service: svc,
		Cfg:     cfg,
		DB:      database,
	}
}

// RegisterRoutes registers all notification endpoints on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc, activeMw gin.HandlerFunc) {
	// SSE Stream endpoint handles its own auth (supporting ?token= query parameter for EventSource)
	rg.GET("/stream", h.StreamNotifications)

	// REST endpoints require standard Bearer auth and active user status
	protected := rg.Group("")
	protected.Use(authMw, activeMw)
	{
		protected.GET("", h.GetNotifications)
		protected.GET("/unread-count", h.GetUnreadCount)
		protected.PATCH("/:id/read", h.MarkAsRead)
		protected.POST("/read-all", h.MarkAllAsRead)
	}
}

// authenticateSSE extracts and validates token from either Authorization header or query parameter.
func (h *Handler) authenticateSSE(c *gin.Context) (*db.User, error) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		header := c.GetHeader("Authorization")
		if header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = strings.TrimSpace(parts[1])
			}
		}
	}

	if tokenStr == "" {
		return nil, fmt.Errorf("missing authentication token")
	}

	claims, err := auth.DecodeAccessToken(h.Cfg, tokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("invalid token subject")
	}

	user, err := db.FindUserByIDCtx(c.Request.Context(), h.DB, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Status == db.UserStatusSuspended || user.Status == db.UserStatusBanned {
		return nil, fmt.Errorf("account suspended")
	}

	return user, nil
}

// StreamNotifications handles GET /notifications/stream via Server-Sent Events (SSE).
//
//	@Summary		Stream real-time notifications (SSE)
//	@Description	Establishes a persistent SSE stream to deliver real-time notifications to the client. Token can be passed in Authorization header or as ?token= query param.
//	@Tags			Notifications
//	@Produce		text/event-stream
//	@Param			token	query		string	false	"JWT access token (for EventSource)"
//	@Success		200		{string}	string	"event: notification\ndata: {...}\n\n"
//	@Failure		401		{object}	auth.ErrorResponse
//	@Failure		500		{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/stream [get]
func (h *Handler) StreamNotifications(c *gin.Context) {
	user, err := h.authenticateSSE(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
		})
		return
	}

	if h.Service.Redis == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "REDIS_UNAVAILABLE",
			"message": "Real-time notification broker is unavailable",
		})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Subscribe to user channel in Redis
	ctx := c.Request.Context()
	channel := h.Service.ChannelForUser(user.ID)
	pubsub := h.Service.Redis.Subscribe(ctx, channel)
	defer pubsub.Close()

	redisCh := pubsub.Channel()

	// 20-second keep-alive ticker to prevent intermediate proxies from dropping the connection
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	// Send initial connection event
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"status\":\"connected\",\"user_id\":\"%s\"}\n\n", user.ID.String())
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false

		case <-ticker.C:
			// Ping comment
			fmt.Fprintf(w, ": ping\n\n")
			return true

		case msg, ok := <-redisCh:
			if !ok {
				return false
			}
			// Forward notification event
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", msg.Payload)
			return true
		}
	})
}

// GetNotifications handles GET /notifications.
//
//	@Summary		Get user notifications
//	@Description	Returns a paginated list of notifications for the authenticated user.
//	@Tags			Notifications
//	@Produce		json
//	@Param			limit		query		int		false	"Limit (default 20, max 100)"
//	@Param			offset		query		int		false	"Offset (default 0)"
//	@Param			unread_only	query		bool	false	"Filter by unread only"
//	@Success		200			{object}	NotificationListResponse
//	@Failure		401			{object}	auth.ErrorResponse
//	@Failure		500			{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications [get]
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	unreadOnly := c.Query("unread_only") == "true"

	items, total, unreadCount, err := h.Service.GetUserNotifications(c.Request.Context(), userID, limit, offset, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	readItems := make([]NotificationRead, len(items))
	for i, item := range items {
		readItems[i] = FromModel(&item)
	}

	c.JSON(http.StatusOK, NotificationListResponse{
		Items:       readItems,
		Total:       total,
		UnreadCount: unreadCount,
	})
}

// GetUnreadCount handles GET /notifications/unread-count.
//
//	@Summary		Get unread notifications count
//	@Description	Returns total count of unread notifications for the authenticated user.
//	@Tags			Notifications
//	@Produce		json
//	@Success		200	{object}	UnreadCountResponse
//	@Failure		401	{object}	auth.ErrorResponse
//	@Failure		500	{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	count, err := h.Service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, UnreadCountResponse{
		UnreadCount: count,
	})
}

// MarkAsRead handles PATCH /notifications/:id/read.
//
//	@Summary		Mark notification as read
//	@Description	Marks a specific notification as read.
//	@Tags			Notifications
//	@Produce		json
//	@Param			id	path		string	true	"Notification UUID"	format(uuid)
//	@Success		200	{object}	SuccessResponse
//	@Failure		400	{object}	auth.ErrorResponse	"INVALID_ID"
//	@Failure		401	{object}	auth.ErrorResponse
//	@Failure		404	{object}	auth.ErrorResponse	"NOT_FOUND"
//	@Failure		500	{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/{id}/read [patch]
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	idStr := c.Param("id")
	notifID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_ID", "message": "Invalid notification ID format"})
		return
	}

	if err := h.Service.MarkAsRead(c.Request.Context(), notifID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

// MarkAllAsRead handles POST /notifications/read-all.
//
//	@Summary		Mark all notifications as read
//	@Description	Marks all unread notifications for the user as read.
//	@Tags			Notifications
//	@Produce		json
//	@Success		200	{object}	SuccessResponse
//	@Failure		401	{object}	auth.ErrorResponse
//	@Failure		500	{object}	auth.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notifications/read-all [post]
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := auth.GetCurrentUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"})
		return
	}

	if err := h.Service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true})
}
