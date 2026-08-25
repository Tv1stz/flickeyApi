package notifications

import (
	"encoding/json"
	"testing"
	"time"

	"flickey/go-backend/db"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func TestNotification_JSON_SnakeCase(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()
	readAt := now.Add(time.Minute)

	notif := &db.Notification{
		ID:        id,
		UserID:    userID,
		Type:      TypeListingApproved,
		Title:     "Объявление одобрено",
		Message:   "Ваше объявление успешно прошло модерацию.",
		Payload:   datatypes.JSON(`{"listing_id":"123"}`),
		IsRead:    true,
		ReadAt:    &readAt,
		CreatedAt: now,
	}

	readDTO := FromModel(notif)
	data, err := json.Marshal(readDTO)
	if err != nil {
		t.Fatalf("failed to marshal NotificationRead: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	expectedKeys := []string{
		"id",
		"user_id",
		"type",
		"title",
		"message",
		"payload",
		"is_read",
		"read_at",
		"created_at",
	}

	for _, key := range expectedKeys {
		if _, exists := raw[key]; !exists {
			t.Errorf("expected JSON key %q to exist in serialized output, got keys: %v", key, raw)
		}
	}

	// Verify no PascalCase keys leaked
	prohibitedKeys := []string{"ID", "UserID", "Type", "Title", "Message", "Payload", "IsRead", "ReadAt", "CreatedAt"}
	for _, key := range prohibitedKeys {
		if _, exists := raw[key]; exists {
			t.Errorf("found prohibited PascalCase key %q in serialized JSON", key)
		}
	}
}

func TestNotificationListResponse_JSON_SnakeCase(t *testing.T) {
	resp := NotificationListResponse{
		Items:       []NotificationRead{},
		Total:       10,
		UnreadCount: 3,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal NotificationListResponse: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if _, exists := raw["items"]; !exists {
		t.Error("expected key 'items'")
	}
	if _, exists := raw["total"]; !exists {
		t.Error("expected key 'total'")
	}
	if _, exists := raw["unread_count"]; !exists {
		t.Error("expected key 'unread_count'")
	}
}
