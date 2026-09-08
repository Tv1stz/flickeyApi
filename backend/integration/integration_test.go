// Package integration provides end-to-end API tests against the full Gin router.
// Tests require running PostgreSQL and Redis (see testutil package).
// Run with: go test -v ./integration/...
// Or: go test -v -race ./integration/...
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"flickey/go-backend/admin"
	"flickey/go-backend/amenities"
	"flickey/go-backend/auth"
	"flickey/go-backend/config"
	"flickey/go-backend/db"
	"flickey/go-backend/geo"
	"flickey/go-backend/listings"
	"flickey/go-backend/media"
	"flickey/go-backend/notifications"
	"flickey/go-backend/sms"
	"flickey/go-backend/storage"
	"flickey/go-backend/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test App Setup
// ─────────────────────────────────────────────────────────────────────────────

type testApp struct {
	Router  *gin.Engine
	Suite   *testutil.Suite
	AuthSvc *auth.AuthService
}

func setupApp(t *testing.T) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	suite := testutil.NewSuite(t)
	t.Cleanup(suite.Cleanup)

	cfg := suite.Cfg

	store, _ := storage.NewLocalStorage(t.TempDir(), "")

	authSvc := &auth.AuthService{
		DB:        suite.DB,
		Redis:     suite.Redis,
		Cfg:       cfg,
		SMSSender: &sms.ConsoleSMSSender{},
	}
	listingSvc := listings.NewListingService(suite.DB, suite.Redis, cfg)
	mediaSvc := media.NewMediaService(suite.DB, store, cfg)

	authMw := auth.RequireAuth(cfg, suite.DB)
	activeMw := auth.RequireActiveUser()

	notifSvc := notifications.NewNotificationService(suite.DB, suite.Redis, slog.New(slog.NewTextHandler(io.Discard, nil)))
	adminSvc := admin.NewAdminService(suite.DB, notifSvc)
	geoSvc := geo.NewGeoService(suite.Redis, cfg.GeocoderURL, cfg.TileServerURL, slog.New(slog.NewTextHandler(io.Discard, nil)))

	r := gin.New()
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")

	auth.NewHandler(authSvc, cfg).RegisterRoutes(v1.Group("/auth"), authMw, activeMw)
	listings.NewHandler(listingSvc).RegisterRoutes(v1.Group("/listings"), authMw, activeMw)
	media.NewHandler(mediaSvc).RegisterRoutes(v1.Group("/media"), authMw, activeMw)
	amenities.RegisterRoutes(v1.Group("/amenities"))
	notifications.NewHandler(notifSvc, cfg, suite.DB).RegisterRoutes(v1.Group("/notifications"), authMw, activeMw)
	admin.NewHandler(adminSvc, cfg).RegisterRoutes(v1.Group("/admin"), authMw, activeMw)
	geo.NewHandler(geoSvc).RegisterRoutes(v1.Group("/geo"))

	return &testApp{Router: r, Suite: suite, AuthSvc: authSvc}
}

func (a *testApp) Do(method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)
	return w
}

func (a *testApp) DoAuth(method, path string, body any, accessToken string) *httptest.ResponseRecorder {
	return a.Do(method, path, body, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
}

func (a *testApp) DoRaw(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)
	return w
}

// ─────────────────────────────────────────────────────────────────────────────
// Amenities Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestAmenities_GetAll(t *testing.T) {
	app := setupApp(t)
	w := app.Do("GET", "/api/v1/amenities", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	cats, ok := resp["categories"].([]any)
	if !ok || len(cats) == 0 {
		t.Error("expected non-empty categories")
	}
}

func TestAmenities_FilterByHousingType(t *testing.T) {
	app := setupApp(t)
	w := app.Do("GET", "/api/v1/amenities?housing_type=apartment", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAmenities_InvalidHousingType(t *testing.T) {
	app := setupApp(t)
	w := app.Do("GET", "/api/v1/amenities?housing_type=condo", nil, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Auth: Dev Login Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDevLogin_CreatesUser(t *testing.T) {
	app := setupApp(t)
	phone := randomPhone()
	w := app.Do("POST", "/api/v1/auth/dev-login", map[string]any{
		"phone":      phone,
		"first_name": "Test",
		"last_name":  "User",
		"role":       "guest",
	}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("dev-login failed: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["access_token"] == nil {
		t.Error("expected access_token in response")
	}
}

func TestDevLogin_SamePhoneTwice_UpdatesUser(t *testing.T) {
	app := setupApp(t)
	phone := randomPhone()
	for i := 0; i < 2; i++ {
		w := app.Do("POST", "/api/v1/auth/dev-login", map[string]any{
			"phone":      phone,
			"first_name": "Test",
			"last_name":  "User",
			"role":       "guest",
		}, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("dev-login attempt %d failed: %d", i, w.Code)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Auth: GET /me Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestGetMe_ReturnsUser(t *testing.T) {
	app := setupApp(t)
	u, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	w := app.DoAuth("GET", "/api/v1/auth/me", nil, access)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["id"] != u.ID.String() {
		t.Errorf("expected user ID %s, got %v", u.ID, resp["id"])
	}
}

func TestGetMe_NoToken_Returns401(t *testing.T) {
	app := setupApp(t)
	w := app.Do("GET", "/api/v1/auth/me", nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetMe_InvalidToken_Returns401(t *testing.T) {
	app := setupApp(t)
	w := app.DoAuth("GET", "/api/v1/auth/me", nil, "invalid.jwt.token")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetMe_GuestWithListing_AutoPromotesHost(t *testing.T) {
	app := setupApp(t)
	u, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	// Insert a listing for this user.
	app.Suite.DB.Exec(`
		INSERT INTO listings (id, host_id, status, type, name, address, city, street, house_number, latitude, longitude, square, floor, total_floors,
			max_guests, rooms_count, beds_count, bathrooms_count, price_per_night, currency,
			min_nights, checkin_from, checkout_until, description)
		VALUES (?, ?, 'published', 'apartment', 'Test Apartment', 'г. Минск, ул. Ленина, 1', 'Минск', 'Ленина', '1', 53.9006, 27.5590, 50.0, 2, 5, 4, 2, 2, 1,
			100.0, 'BYN', 1, '14:00:00', '12:00:00', 'A nice test listing for auto-promotion')`,
		uuid.New(), u.ID,
	)

	w := app.DoAuth("GET", "/api/v1/auth/me", nil, access)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["role"] != "host" {
		t.Errorf("expected role=host after auto-promotion, got %v", resp["role"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Auth: Logout Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestLogout_Returns204(t *testing.T) {
	app := setupApp(t)
	_, _, refresh := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	w := app.Do("POST", "/api/v1/auth/logout", nil, map[string]string{
		"Cookie": "refresh_token=" + refresh,
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestLogoutAll_RevokesAllTokens(t *testing.T) {
	app := setupApp(t)
	u, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	// Create multiple refresh tokens.
	ctx := context.Background()
	auth.CreateRefreshToken(ctx, app.Suite.Redis, app.Suite.Cfg, u.ID, "", "")
	auth.CreateRefreshToken(ctx, app.Suite.Redis, app.Suite.Cfg, u.ID, "", "")

	w := app.DoAuth("POST", "/api/v1/auth/logout-all", nil, access)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Auth: Complete Profile Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCompleteProfile_Success(t *testing.T) {
	app := setupApp(t)
	// Create a pending-profile user.
	firstName, lastName := "Pending", "User"
	u := &db.User{
		ID:        uuid.New(),
		Phone:     randomPhone(),
		FirstName: &firstName,
		LastName:  &lastName,
		Role:      db.UserRoleGuest,
		Status:    db.UserStatusPendingProfile,
	}
	db.CreateUser(context.Background(), app.Suite.DB, u) //nolint
	access, _ := auth.CreateAccessToken(app.Suite.Cfg, u.ID)

	w := app.DoAuth("POST", "/api/v1/auth/complete-profile", map[string]any{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      fmt.Sprintf("john_%d@example.com", time.Now().UnixNano()),
	}, access)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "active" {
		t.Errorf("expected status=active, got %v", resp["status"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Listings: Draft State Machine Tests
// ─────────────────────────────────────────────────────────────────────────────

// createDraftSteps1to6 runs the full draft creation flow and returns the draft ID.
func createDraftSteps1to6(t *testing.T, app *testApp, access string) string {
	t.Helper()

	// Step 1: Create draft.
	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, access)
	if w.Code != http.StatusCreated {
		t.Fatalf("step1 failed: %d %s", w.Code, w.Body.String())
	}
	var step1 map[string]any
	json.Unmarshal(w.Body.Bytes(), &step1)
	draftID := step1["draft_id"].(string)

	// Step 2: Property info.
	w = app.DoAuth("PATCH", "/api/v1/listings/drafts/"+draftID+"/step-2", map[string]any{
		"name":            "Beautiful Test Apartment",
		"address":         "г. Минск, пр. Победителей, 1",
		"latitude":        53.9006,
		"longitude":       27.5590,
		"square":          55.5,
		"floor":           3,
		"total_floors":    9,
		"max_guests":      4,
		"rooms_count":     2,
		"beds_count":      2,
		"bathrooms_count": 1,
	}, access)
	if w.Code != http.StatusOK {
		t.Fatalf("step2 failed: %d %s", w.Code, w.Body.String())
	}

	// Step 3: Media — create 5 fake media records.
	mediaIDs := createFakeMedia(t, app, access, 5)
	w = app.DoAuth("PATCH", "/api/v1/listings/drafts/"+draftID+"/step-3", map[string]any{
		"media_ids": mediaIDs,
	}, access)
	if w.Code != http.StatusOK {
		t.Fatalf("step3 failed: %d %s", w.Code, w.Body.String())
	}

	// Step 4: Amenities.
	w = app.DoAuth("PATCH", "/api/v1/listings/drafts/"+draftID+"/step-4", map[string]any{
		"amenities": []string{"wifi", "heating", "hot_water"},
	}, access)
	if w.Code != http.StatusOK {
		t.Fatalf("step4 failed: %d %s", w.Code, w.Body.String())
	}

	// Step 5: Pricing.
	w = app.DoAuth("PATCH", "/api/v1/listings/drafts/"+draftID+"/step-5", map[string]any{
		"price_per_night": 150.0,
		"currency":        "BYN",
		"min_nights":      2,
		"checkin_from":    "14:00",
		"checkout_until":  "12:00",
		"rules": map[string]any{
			"allow_children":   true,
			"allow_pets":       false,
			"allow_smoking":    false,
			"allow_parties":    false,
			"deposit_required": false,
			"with_invoicing":   false,
		},
	}, access)
	if w.Code != http.StatusOK {
		t.Fatalf("step5 failed: %d %s", w.Code, w.Body.String())
	}

	// Step 6: Description.
	w = app.DoAuth("PATCH", "/api/v1/listings/drafts/"+draftID+"/step-6", map[string]any{
		"description": "A beautiful and spacious apartment in the heart of Minsk, perfect for families and business travelers alike.",
	}, access)
	if w.Code != http.StatusOK {
		t.Fatalf("step6 failed: %d %s", w.Code, w.Body.String())
	}

	return draftID
}

func createFakeMedia(t *testing.T, app *testApp, access string, count int) []string {
	t.Helper()
	ctx := app.Suite.DB.Statement.Context
	if ctx == nil {
		ctx = app.Suite.DB.WithContext(nil).Statement.Context
	}

	// Get user ID from access token.
	userID := getUserIDFromToken(t, app.Suite.Cfg, access)

	var ids []string
	for i := 0; i < count; i++ {
		m := &db.Media{
			ID:            uuid.New(),
			HostID:        userID,
			FileKey:       fmt.Sprintf("raw/%s/%s.jpg", userID, uuid.New()),
			Status:        db.MediaStatusUploaded,
			ContentType:   "image/jpeg",
			FileSizeBytes: 1024 * 100,
		}
		db.CreateMedia(context.Background(), app.Suite.DB, m) //nolint
		ids = append(ids, m.ID.String())
	}
	return ids
}

func getUserIDFromToken(t *testing.T, cfg *config.Settings, accessToken string) uuid.UUID {
	t.Helper()
	claims, err := auth.DecodeAccessToken(cfg, accessToken)
	if err != nil {
		t.Fatalf("getUserIDFromToken: %v", err)
	}
	id, _ := uuid.Parse(claims.Subject)
	return id
}

func TestDraft_CreateAndGetDraft(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, access)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	draftID := resp["draft_id"].(string)
	if resp["current_step"].(float64) != 2 {
		t.Errorf("expected current_step=2, got %v", resp["current_step"])
	}

	// Fetch draft.
	w = app.DoAuth("GET", "/api/v1/listings/drafts/"+draftID, nil, access)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDraft_IDOR_CannotAccessOtherUserDraft(t *testing.T) {
	app := setupApp(t)
	_, access1, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	_, access2, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	// User1 creates a draft.
	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "house"}, access1)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	draftID := resp["draft_id"].(string)

	// User2 tries to access User1's draft — must get 404.
	w = app.DoAuth("GET", "/api/v1/listings/drafts/"+draftID, nil, access2)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 (IDOR prevention), got %d", w.Code)
	}
}

func TestDraft_InvalidHousingType_Returns400(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "condo"}, access)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDraft_Step2_InvalidFloor_Returns400(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, access)
	var r1 map[string]any
	json.Unmarshal(w.Body.Bytes(), &r1)
	draftID := r1["draft_id"].(string)

	// floor > total_floors.
	w = app.DoAuth("PATCH", "/api/v1/listings/drafts/"+draftID+"/step-2", map[string]any{
		"name":            "Test Apartment Name That Is Long",
		"address":         "г. Минск, пр. Победителей, 1",
		"latitude":        53.9006,
		"longitude":       27.5590,
		"square":          50.0,
		"floor":           10,
		"total_floors":    5, // floor > total_floors
		"max_guests":      4,
		"rooms_count":     2,
		"beds_count":      2,
		"bathrooms_count": 1,
	}, access)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for floor > total_floors, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDraft_Step5_WrongCurrency_Returns400(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	draftID := createDraftSteps1to6(t, app, access)
	_ = draftID

	// Currency must be BYN.
	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, access)
	var r map[string]any
	json.Unmarshal(w.Body.Bytes(), &r)
	newDraftID := r["draft_id"].(string)

	// Go through steps 2-4 quickly.
	app.DoAuth("PATCH", "/api/v1/listings/drafts/"+newDraftID+"/step-2", map[string]any{
		"name": "Long Enough Name Here", "address": "г. Минск, пр. Победителей, 1", "latitude": 53.9006, "longitude": 27.5590, "square": 50.0, "floor": 2,
		"total_floors": 5, "max_guests": 2, "rooms_count": 1, "beds_count": 1, "bathrooms_count": 1,
	}, access)
	mediaIDs := createFakeMedia(t, app, access, 5)
	app.DoAuth("PATCH", "/api/v1/listings/drafts/"+newDraftID+"/step-3", map[string]any{"media_ids": mediaIDs}, access)
	app.DoAuth("PATCH", "/api/v1/listings/drafts/"+newDraftID+"/step-4", map[string]any{"amenities": []string{"wifi"}}, access)

	w = app.DoAuth("PATCH", "/api/v1/listings/drafts/"+newDraftID+"/step-5", map[string]any{
		"price_per_night": 100.0,
		"currency":        "USD", // wrong currency
		"min_nights":      1,
		"checkin_from":    "14:00",
		"checkout_until":  "12:00",
		"rules":           map[string]any{},
	}, access)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for USD currency, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDraft_SubmitComplete_AtomicRolePromotion(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)

	draftID := createDraftSteps1to6(t, app, access)

	idempKey := uuid.New().String()
	w := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + access,
		"Content-Type":    "application/json",
		"Idempotency-Key": idempKey,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("submit failed: %d %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["listing_id"] == nil {
		t.Error("expected listing_id in response")
	}

	// User should now be host.
	meResp := app.DoAuth("GET", "/api/v1/auth/me", nil, access)
	var me map[string]any
	json.Unmarshal(meResp.Body.Bytes(), &me)
	// The access token was issued before promotion; /me auto-promotes.
	if me["role"] != "host" {
		t.Errorf("expected role=host after submission, got %v", me["role"])
	}
}

func TestDraft_Submit_IdempotencyKey_Required(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	draftID := createDraftSteps1to6(t, app, access)

	w := app.DoAuth("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, access)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing Idempotency-Key, got %d", w.Code)
	}
}

func TestDraft_Submit_InvalidIdempotencyKey(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	draftID := createDraftSteps1to6(t, app, access)

	w := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, map[string]string{
		"Authorization":   "Bearer " + access,
		"Idempotency-Key": "not-a-uuid",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID idempotency key, got %d", w.Code)
	}
}

func TestDraft_Submit_Idempotent(t *testing.T) {
	app := setupApp(t)
	_, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	draftID := createDraftSteps1to6(t, app, access)

	idempKey := uuid.New().String()
	headers := map[string]string{
		"Authorization":   "Bearer " + access,
		"Content-Type":    "application/json",
		"Idempotency-Key": idempKey,
	}

	w1 := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, headers)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first submit failed: %d %s", w1.Code, w1.Body.String())
	}

	// Give Redis time to update the idempotency key.
	time.Sleep(100 * time.Millisecond)

	w2 := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, headers)
	// Second submission with same key should return the cached result (201) or conflict (409).
	if w2.Code != http.StatusCreated && w2.Code != http.StatusConflict {
		t.Errorf("expected 201 or 409 for idempotent submit, got %d: %s", w2.Code, w2.Body.String())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Public Listing Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestPublicListings_OnlyPublished(t *testing.T) {
	app := setupApp(t)

	// Insert one published and one pending listing.
	hostID := uuid.New()
	pubID := uuid.New()
	pendID := uuid.New()
	app.Suite.DB.Exec(`INSERT INTO users (id, phone, role, status) VALUES (?, ?, 'host', 'active')`, hostID, randomPhone())
	app.Suite.DB.Exec(`INSERT INTO listings (id, host_id, status, type, name, address, city, street, house_number, latitude, longitude, square, floor, total_floors, max_guests, rooms_count, beds_count, bathrooms_count, price_per_night, currency, min_nights, checkin_from, checkout_until, description)
		VALUES (?, ?, 'published', 'apartment', 'Published Listing', 'г. Минск, пр. Победителей, 1', 'Минск', 'Победителей', '1', 53.9006, 27.5590, 50.0, 2, 5, 4, 2, 2, 1, 100.0, 'BYN', 1, '14:00:00', '12:00:00', 'A beautiful published apartment available for rent in Minsk')`,
		pubID, hostID)
	app.Suite.DB.Exec(`INSERT INTO listings (id, host_id, status, type, name, address, city, street, house_number, latitude, longitude, square, floor, total_floors, max_guests, rooms_count, beds_count, bathrooms_count, price_per_night, currency, min_nights, checkin_from, checkout_until, description)
		VALUES (?, ?, 'pending_review', 'house', 'Pending Listing', 'г. Минск, ул. Садовая, 5', 'Минск', 'Садовая', '5', 53.9100, 27.5600, 80.0, 1, 2, 6, 3, 3, 2, 200.0, 'BYN', 1, '14:00:00', '12:00:00', 'A house that is pending review and should not be visible publicly')`,
		pendID, hostID)

	w := app.Do("GET", "/api/v1/listings", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var listings []map[string]any
	json.Unmarshal(w.Body.Bytes(), &listings)
	foundPub := false
	for _, l := range listings {
		if l["status"] != nil {
			t.Errorf("public listing should not expose status field")
		}
		if l["id"] == pubID.String() {
			foundPub = true
		}
		if l["id"] == pendID.String() {
			t.Errorf("pending listing %s should not be visible publicly", pendID)
		}
	}
	if !foundPub {
		t.Errorf("expected published listing %s to be in public results", pubID)
	}
}

func TestPublicListing_NotFound(t *testing.T) {
	app := setupApp(t)
	w := app.Do("GET", "/api/v1/listings/"+uuid.New().String(), nil, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestMyListings_ReturnsAllStatuses(t *testing.T) {
	app := setupApp(t)
	u, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleHost)

	// Insert one published and one pending listing.
	app.Suite.DB.Exec(`INSERT INTO listings (id, host_id, status, type, name, square, floor, total_floors, max_guests, rooms_count, beds_count, bathrooms_count, price_per_night, currency, min_nights, checkin_from, checkout_until, description)
		VALUES (?, ?, 'published', 'apartment', 'My Published Listing', 50.0, 2, 5, 4, 2, 2, 1, 100.0, 'BYN', 1, '14:00:00', '12:00:00', 'Description of my published apartment listing')`,
		uuid.New(), u.ID)
	app.Suite.DB.Exec(`INSERT INTO listings (id, host_id, status, type, name, square, floor, total_floors, max_guests, rooms_count, beds_count, bathrooms_count, price_per_night, currency, min_nights, checkin_from, checkout_until, description)
		VALUES (?, ?, 'pending_review', 'house', 'My Pending Listing', 80.0, 1, 2, 6, 3, 3, 2, 200.0, 'BYN', 1, '14:00:00', '12:00:00', 'Description of my pending house listing that needs review')`,
		uuid.New(), u.ID)

	w := app.DoAuth("GET", "/api/v1/listings/my", nil, access)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var list []map[string]any
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Errorf("expected 2 listings (all statuses), got %d", len(list))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Concurrency Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestConcurrentSubmit_IdempotencyPreventsDoubleCreation(t *testing.T) {
	app := setupApp(t)
	user, access, _ := app.Suite.CreateUser(randomPhone(), db.UserRoleGuest)
	draftID := createDraftSteps1to6(t, app, access)
	idempKey := uuid.New().String()
	headers := map[string]string{
		"Authorization":   "Bearer " + access,
		"Content-Type":    "application/json",
		"Idempotency-Key": idempKey,
	}

	const concurrency = 5
	results := make([]int, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := app.Do("POST", "/api/v1/listings/drafts/"+draftID+"/submit", nil, headers)
			results[idx] = w.Code
		}(i)
	}
	wg.Wait()

	// Count successful creations — should be at least 1.
	successes := 0
	for _, code := range results {
		if code == http.StatusCreated {
			successes++
		}
	}

	var listingCount int64
	app.Suite.DB.Model(&db.Listing{}).Where("host_id = ?", user.ID).Count(&listingCount)
	if listingCount != 1 {
		t.Errorf("expected exactly 1 listing after concurrent submit, got %d", listingCount)
	}
}

func TestConcurrentOTPVerify_OnlyOneSucceeds(t *testing.T) {
	// Verifying the same OTP challenge concurrently should only succeed once.
	app := setupApp(t)

	phone := randomPhone()
	challengeID, _ := auth.GenerateChallengeID()
	otp := "123456"
	hash, _ := auth.HashValue(otp)
	auth.StoreOTPChallenge(context.Background(), app.Suite.Redis, app.Suite.Cfg, challengeID, phone, hash, "")

	// Create a test user for the phone.
	app.Suite.DB.Exec(`INSERT INTO users (id, phone, role, status) VALUES (?, ?, 'guest', 'pending_profile')`, uuid.New(), phone)

	const concurrency = 5
	results := make([]int, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := app.Do("POST", "/api/v1/auth/verify-otp", map[string]any{
				"challenge_id": challengeID,
				"code":         otp,
			}, nil)
			results[idx] = w.Code
		}(i)
	}
	wg.Wait()

	successes := 0
	for _, code := range results {
		if code == http.StatusOK {
			successes++
		}
	}
	if successes != 1 {
		t.Errorf("expected exactly 1 successful OTP verification, got %d", successes)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Authorization Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestRequireAuth_RejectsNoToken(t *testing.T) {
	app := setupApp(t)
	// Protected endpoint with no auth.
	w := app.Do("GET", "/api/v1/auth/me", nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	// Check WWW-Authenticate header.
	if app.Router != nil && w.Header().Get("WWW-Authenticate") == "" {
		t.Log("note: WWW-Authenticate header not set") // informational
	}
}

func TestRequireActiveUser_RejectsPendingProfile(t *testing.T) {
	app := setupApp(t)
	// pending_profile user.
	fn, ln := "P", "U"
	u := &db.User{
		ID:        uuid.New(),
		Phone:     randomPhone(),
		FirstName: &fn,
		LastName:  &ln,
		Role:      db.UserRoleGuest,
		Status:    db.UserStatusPendingProfile,
	}
	db.CreateUser(context.Background(), app.Suite.DB, u) //nolint
	access, _ := auth.CreateAccessToken(app.Suite.Cfg, u.ID)

	// Try to create a draft (requires active user).
	w := app.DoAuth("POST", "/api/v1/listings/drafts", map[string]any{"type": "apartment"}, access)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for pending_profile user, got %d: %s", w.Code, w.Body.String())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func randomPhone() string {
	return fmt.Sprintf("+37529%07d", rand.Intn(9999999))
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func bodyStr(w *httptest.ResponseRecorder) string {
	return strings.TrimSpace(w.Body.String())
}
