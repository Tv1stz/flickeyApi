// Package notifications provides models, service, and HTTP/SSE handlers for in-app notifications.
package notifications

import (
	"time"

	"flickey/go-backend/db"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Notification types.
const (
	TypeListingApproved      = "listing_approved"
	TypeListingRejected      = "listing_rejected"
	TypeEnforcementIssued    = "enforcement_issued"
	TypeBookingCreated       = "booking_created"
	TypeBookingStatusChanged = "booking_status_changed"
	TypeNewMessage           = "new_message"
)

// NotificationRead represents the serialized notification payload for clients.
type NotificationRead struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Payload   datatypes.JSON `json:"payload"`
	IsRead    bool           `json:"is_read"`
	ReadAt    *time.Time     `json:"read_at"`
	CreatedAt time.Time      `json:"created_at"`
}

// FromModel converts a db.Notification to NotificationRead.
func FromModel(n *db.Notification) NotificationRead {
	return NotificationRead{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Payload:   n.Payload,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}

// NotificationListResponse is the paginated response for GET /notifications.
type NotificationListResponse struct {
	Items       []NotificationRead `json:"items"`
	Total       int64              `json:"total"`
	UnreadCount int64              `json:"unread_count"`
}

// UnreadCountResponse is the response for GET /notifications/unread-count.
type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}

// SuccessResponse is a generic success response.
type SuccessResponse struct {
	Success bool `json:"success"`
}
