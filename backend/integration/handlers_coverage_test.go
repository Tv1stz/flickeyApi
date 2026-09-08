package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"flickey/go-backend/db"

	"github.com/google/uuid"
)

// TestListing_DeleteListing verifies that a host can delete their own listing,
// and that anti-IDOR prevents another host from deleting it.
func TestListing_DeleteListing(t *testing.T) {
	app := setupApp(t)

	_, hostToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)
	_, otherToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)

	// Create and submit listing
	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, hostToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create draft failed: %d", w.Code)
	}
	var draftRes map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &draftRes)
	draftID := draftRes["draft_id"].(string)

	// Step 2
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-2", draftID), map[string]any{
		"name": "Listing to Delete", "address": "ул. Ленина 10",
		"latitude": 53.9006, "longitude": 27.5590, "square": 50,
		"floor": 3, "total_floors": 9, "max_guests": 2, "rooms_count": 1,
		"beds_count": 1, "bathrooms_count": 1,
	}, hostToken)

	// Step 3
	mediaIDs := createFakeMedia(t, app, hostToken, 5)
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-3", draftID), map[string]any{"media_ids": mediaIDs}, hostToken)

	// Step 4
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-4", draftID), map[string]any{"amenities": []string{"wifi"}}, hostToken)

	// Step 5
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-5", draftID), map[string]any{
		"price_per_night": 100, "currency": "BYN", "min_nights": 1,
		"checkin_from": "14:00", "checkout_until": "12:00",
		"rules": map[string]any{"allow_children": true},
	}, hostToken)

	// Step 6
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-6", draftID), map[string]any{
		"description": "This is a great apartment with lots of space and beautiful light.",
	}, hostToken)

	// Submit
	wSub := app.Do("POST", fmt.Sprintf("/api/v1/listings/drafts/%s/submit", draftID), nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Idempotency-Key": uuid.New().String(),
	})
	if wSub.Code != http.StatusOK && wSub.Code != http.StatusCreated {
		t.Fatalf("submit failed: %d: %s", wSub.Code, wSub.Body.String())
	}
	var subRes map[string]any
	_ = json.Unmarshal(wSub.Body.Bytes(), &subRes)
	listingID := subRes["listing_id"].(string)

	// Other host attempts to delete -> 404 (anti-IDOR)
	wDelOther := app.DoAuth("DELETE", fmt.Sprintf("/api/v1/listings/%s", listingID), nil, otherToken)
	if wDelOther.Code != http.StatusNotFound {
		t.Errorf("expected 404 on IDOR delete, got %d", wDelOther.Code)
	}

	// Owner deletes -> 204 No Content
	wDelOwner := app.DoAuth("DELETE", fmt.Sprintf("/api/v1/listings/%s", listingID), nil, hostToken)
	if wDelOwner.Code != http.StatusNoContent {
		t.Errorf("expected 204 on delete, got %d: %s", wDelOwner.Code, wDelOwner.Body.String())
	}

	// Verify listing is gone
	wGet := app.Do("GET", fmt.Sprintf("/api/v1/listings/%s", listingID), nil, nil)
	if wGet.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", wGet.Code)
	}
}

// TestListing_Step6_Description_Validation tests step 6 description requirements (min 30 chars).
func TestListing_Step6_Description_Validation(t *testing.T) {
	app := setupApp(t)
	_, hostToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)

	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, hostToken)
	var draftRes map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &draftRes)
	draftID := draftRes["draft_id"].(string)

	// Step 2
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-2", draftID), map[string]any{
		"name": "Validation Apartment", "address": "ул. Немига 5",
		"latitude": 53.904, "longitude": 27.551, "square": 60,
		"floor": 2, "total_floors": 5, "max_guests": 3, "rooms_count": 2,
		"beds_count": 2, "bathrooms_count": 1,
	}, hostToken)

	// Step 3
	mediaIDs := createFakeMedia(t, app, hostToken, 5)
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-3", draftID), map[string]any{"media_ids": mediaIDs}, hostToken)

	// Step 4
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-4", draftID), map[string]any{"amenities": []string{"wifi"}}, hostToken)

	// Step 5
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-5", draftID), map[string]any{
		"price_per_night": 120, "currency": "BYN", "min_nights": 1,
		"checkin_from": "14:00", "checkout_until": "12:00",
		"rules": map[string]any{"allow_children": true},
	}, hostToken)

	// Too short (< 30 characters)
	wShort := app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-6", draftID), map[string]any{
		"description": "Too short",
	}, hostToken)
	if wShort.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short description, got %d", wShort.Code)
	}

	// Valid (>= 30 characters)
	wValid := app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-6", draftID), map[string]any{
		"description": "This is a properly formatted description that exceeds thirty characters easily.",
	}, hostToken)
	if wValid.Code != http.StatusOK {
		t.Errorf("expected 200 for valid description, got %d: %s", wValid.Code, wValid.Body.String())
	}
}

// TestMedia_UploadFlow tests presign, dev upload, and complete handlers.
func TestMedia_UploadFlow(t *testing.T) {
	app := setupApp(t)
	_, userToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	// 1. Presign
	wPresign := app.DoAuth("POST", "/api/v1/media/presign", map[string]any{
		"content_type":    "image/jpeg",
		"file_size_bytes": 1024,
	}, userToken)
	if wPresign.Code != http.StatusOK {
		t.Fatalf("expected 200 for presign, got %d: %s", wPresign.Code, wPresign.Body.String())
	}
	var presignRes map[string]any
	_ = json.Unmarshal(wPresign.Body.Bytes(), &presignRes)
	mediaID := presignRes["media_id"].(string)
	uploadURL := presignRes["upload_url"].(string)

	// 2. Dev upload via PUT
	fakeImage := []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01\x01\x01\x00\x60\x00\x60\x00\x00\xFF\xDB\x00C\x00")
	req, _ := http.NewRequest("PUT", uploadURL, bytes.NewReader(fakeImage))
	req.Header.Set("Content-Type", "image/jpeg")
	wUpload := app.DoRaw(req)
	if wUpload.Code != http.StatusOK {
		t.Fatalf("expected 200 for dev upload, got %d: %s", wUpload.Code, wUpload.Body.String())
	}

	// 3. Complete
	wComp := app.DoAuth("POST", fmt.Sprintf("/api/v1/media/%s/complete", mediaID), nil, userToken)
	if wComp.Code != http.StatusOK {
		t.Fatalf("expected 200 for complete, got %d: %s", wComp.Code, wComp.Body.String())
	}
	var compRes map[string]any
	_ = json.Unmarshal(wComp.Body.Bytes(), &compRes)
	if compRes["status"] != "uploaded" {
		t.Errorf("expected status 'uploaded', got %v", compRes["status"])
	}
}

// TestNotifications_Flow tests notification listing, count, and mark as read endpoints.
func TestNotifications_Flow(t *testing.T) {
	app := setupApp(t)
	user, userToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	// Create test notification directly in DB
	notif := &db.Notification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      "test_alert",
		Title:     "Добро пожаловать",
		Message:   "Тестовое уведомление",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	if err := app.Suite.DB.Create(notif).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	// Unread count -> 1
	wCount := app.DoAuth("GET", "/api/v1/notifications/unread-count", nil, userToken)
	if wCount.Code != http.StatusOK {
		t.Fatalf("expected 200 for unread count, got %d", wCount.Code)
	}
	var countRes map[string]any
	_ = json.Unmarshal(wCount.Body.Bytes(), &countRes)
	if count, ok := countRes["unread_count"].(float64); !ok || int(count) != 1 {
		t.Errorf("expected count 1, got %v", countRes["unread_count"])
	}

	// List notifications
	wList := app.DoAuth("GET", "/api/v1/notifications", nil, userToken)
	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 for notifications list, got %d", wList.Code)
	}

	// Mark as read
	wRead := app.DoAuth("PATCH", fmt.Sprintf("/api/v1/notifications/%s/read", notif.ID), nil, userToken)
	if wRead.Code != http.StatusOK {
		t.Fatalf("expected 200 for mark as read, got %d: %s", wRead.Code, wRead.Body.String())
	}

	// Unread count -> 0
	wCount2 := app.DoAuth("GET", "/api/v1/notifications/unread-count", nil, userToken)
	var countRes2 map[string]any
	_ = json.Unmarshal(wCount2.Body.Bytes(), &countRes2)
	if count, ok := countRes2["unread_count"].(float64); !ok || int(count) != 0 {
		t.Errorf("expected count 0, got %v", countRes2["unread_count"])
	}

	// Mark all as read
	wReadAll := app.DoAuth("POST", "/api/v1/notifications/read-all", nil, userToken)
	if wReadAll.Code != http.StatusOK {
		t.Fatalf("expected 200 for mark all as read, got %d", wReadAll.Code)
	}
}

// TestAdmin_Overview_And_Moderation tests admin role enforcement, overview metrics,
// and approving / rejecting listings.
func TestAdmin_Overview_And_Moderation(t *testing.T) {
	app := setupApp(t)
	_, hostToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)
	_, adminToken, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleAdmin)

	// Non-admin cannot access admin overview -> 403
	wForbidden := app.DoAuth("GET", "/api/v1/admin/overview", nil, hostToken)
	if wForbidden.Code != http.StatusForbidden {
		t.Errorf("expected 403 for host accessing admin overview, got %d", wForbidden.Code)
	}

	// Admin accesses overview -> 200
	wOverview := app.DoAuth("GET", "/api/v1/admin/overview", nil, adminToken)
	if wOverview.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin overview, got %d: %s", wOverview.Code, wOverview.Body.String())
	}

	// Create a submitted listing (pending_review)
	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, hostToken)
	var draftRes map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &draftRes)
	draftID := draftRes["draft_id"].(string)

	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-2", draftID), map[string]any{
		"name": "Mod Apartment", "address": "ул. Немига 5",
		"latitude": 53.904, "longitude": 27.551, "square": 60,
		"floor": 2, "total_floors": 5, "max_guests": 3, "rooms_count": 2,
		"beds_count": 2, "bathrooms_count": 1,
	}, hostToken)

	mediaIDs := createFakeMedia(t, app, hostToken, 5)
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-3", draftID), map[string]any{"media_ids": mediaIDs}, hostToken)
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-4", draftID), map[string]any{"amenities": []string{"wifi"}}, hostToken)
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-5", draftID), map[string]any{
		"price_per_night": 120, "currency": "BYN", "min_nights": 1,
		"checkin_from": "14:00", "checkout_until": "12:00",
		"rules": map[string]any{"allow_children": true},
	}, hostToken)
	app.DoAuth("PATCH", fmt.Sprintf("/api/v1/listings/drafts/%s/step-6", draftID), map[string]any{
		"description": "Lovely renovated modern apartment located in the very center of Minsk.",
	}, hostToken)

	wSub := app.Do("POST", fmt.Sprintf("/api/v1/listings/drafts/%s/submit", draftID), nil, map[string]string{
		"Authorization":   "Bearer " + hostToken,
		"Idempotency-Key": uuid.New().String(),
	})
	if wSub.Code != http.StatusOK && wSub.Code != http.StatusCreated {
		t.Fatalf("submit draft failed: %d: %s", wSub.Code, wSub.Body.String())
	}
	var subRes map[string]any
	_ = json.Unmarshal(wSub.Body.Bytes(), &subRes)
	listingID := subRes["listing_id"].(string)

	// Admin views listings list
	wList := app.DoAuth("GET", "/api/v1/admin/listings?status=pending_review", nil, adminToken)
	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin listings, got %d", wList.Code)
	}

	// Admin approves listing
	wApprove := app.DoAuth("POST", fmt.Sprintf("/api/v1/admin/listings/%s/moderate", listingID), map[string]any{
		"action": "approve",
	}, adminToken)
	if wApprove.Code != http.StatusOK {
		t.Fatalf("expected 200 on approve, got %d: %s", wApprove.Code, wApprove.Body.String())
	}

	// Public can view published listing
	wPub := app.Do("GET", fmt.Sprintf("/api/v1/listings/%s", listingID), nil, nil)
	if wPub.Code != http.StatusOK {
		t.Errorf("expected 200 for published listing, got %d", wPub.Code)
	}

	// Admin rejects listing
	wReject := app.DoAuth("POST", fmt.Sprintf("/api/v1/admin/listings/%s/moderate", listingID), map[string]any{
		"action": "reject",
		"reason": "Temporary test rejection",
	}, adminToken)
	if wReject.Code != http.StatusOK {
		t.Fatalf("expected 200 on reject, got %d: %s", wReject.Code, wReject.Body.String())
	}

	// Public can no longer view rejected listing
	wPub2 := app.Do("GET", fmt.Sprintf("/api/v1/listings/%s", listingID), nil, nil)
	if wPub2.Code != http.StatusNotFound {
		t.Errorf("expected 404 for rejected listing from public endpoint, got %d", wPub2.Code)
	}
}

// TestGeo_Suggest_Validation tests query string validation on geocoding endpoints.
func TestGeo_Suggest_Validation(t *testing.T) {
	app := setupApp(t)

	// Query too short (< 2 chars) -> returns empty list without error
	wShort := app.Do("GET", "/api/v1/geo/suggest?q=a", nil, nil)
	if wShort.Code != http.StatusOK {
		t.Errorf("expected 200 for short query, got %d", wShort.Code)
	}
	var res map[string]any
	_ = json.Unmarshal(wShort.Body.Bytes(), &res)
	results, ok := res["results"].([]any)
	if !ok || len(results) != 0 {
		t.Errorf("expected empty results array, got %v", res)
	}

	// Reverse with invalid lat/lng -> returns 400
	wRev := app.Do("GET", "/api/v1/geo/reverse?lat=invalid&lng=invalid", nil, nil)
	if wRev.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid coordinates, got %d", wRev.Code)
	}
}
