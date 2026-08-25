package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"flickey/go-backend/db"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// NotificationService handles notification persistence, querying, and real-time distribution.
type NotificationService struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Logger *slog.Logger
}

// NewNotificationService constructs a new NotificationService instance.
func NewNotificationService(database *gorm.DB, rdb *redis.Client, logger *slog.Logger) *NotificationService {
	return &NotificationService{
		DB:     database,
		Redis:  rdb,
		Logger: logger,
	}
}

// ChannelForUser returns the Redis pub/sub channel for a specific user.
func (s *NotificationService) ChannelForUser(userID uuid.UUID) string {
	return "notifications:user:" + userID.String()
}

// CreateNotification persists a notification in PostgreSQL and publishes it to Redis Pub/Sub.
func (s *NotificationService) CreateNotification(
	ctx context.Context,
	userID uuid.UUID,
	notifType, title, message string,
	payload any,
) (*db.Notification, error) {
	var payloadBytes []byte
	var err error

	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			payloadBytes = []byte("{}")
		}
	} else {
		payloadBytes = []byte("{}")
	}

	notif := &db.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Payload:   datatypes.JSON(payloadBytes),
		IsRead:    false,
		ReadAt:    nil,
		CreatedAt: time.Now(),
	}

	if err := s.DB.WithContext(ctx).Create(notif).Error; err != nil {
		return nil, fmt.Errorf("creating notification: %w", err)
	}

	// Publish to Redis channel for real-time delivery
	if s.Redis != nil {
		readModel := FromModel(notif)
		data, marshalErr := json.Marshal(readModel)
		if marshalErr == nil {
			channel := s.ChannelForUser(userID)
			if pubErr := s.Redis.Publish(ctx, channel, data).Err(); pubErr != nil {
				if s.Logger != nil {
					s.Logger.Warn("failed to publish notification to Redis", "channel", channel, "error", pubErr)
				}
			}
		}
	}

	return notif, nil
}

// GetUserNotifications returns a paginated list of notifications for a user, along with total and unread counts.
func (s *NotificationService) GetUserNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	unreadOnly bool,
) ([]db.Notification, int64, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var items []db.Notification
	var total int64
	var unreadCount int64

	// Count unread notifications
	if err := s.DB.WithContext(ctx).
		Model(&db.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&unreadCount).Error; err != nil {
		return nil, 0, 0, fmt.Errorf("counting unread notifications: %w", err)
	}

	// Base query
	query := s.DB.WithContext(ctx).
		Model(&db.Notification{}).
		Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = false")
	}

	// Count total matching
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, fmt.Errorf("counting notifications: %w", err)
	}

	// Fetch items ordered by created_at DESC
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, 0, fmt.Errorf("fetching notifications: %w", err)
	}

	return items, total, unreadCount, nil
}

// GetUnreadCount returns the count of unread notifications for a user.
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := s.DB.WithContext(ctx).
		Model(&db.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("counting unread notifications: %w", err)
	}
	return count, nil
}

// MarkAsRead marks a single notification as read for a given user.
func (s *NotificationService) MarkAsRead(ctx context.Context, notifID, userID uuid.UUID) error {
	now := time.Now()
	res := s.DB.WithContext(ctx).
		Model(&db.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Updates(map[string]any{
			"is_read": true,
			"read_at": now,
		})
	if res.Error != nil {
		return fmt.Errorf("marking notification as read: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("notification not found or not owned by user")
	}
	return nil
}

// MarkAllAsRead marks all unread notifications as read for a given user.
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	err := s.DB.WithContext(ctx).
		Model(&db.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]any{
			"is_read": true,
			"read_at": now,
		}).Error
	if err != nil {
		return fmt.Errorf("marking all notifications as read: %w", err)
	}
	return nil
}
