// Package testutil provides shared test infrastructure for integration tests.
// Tests in this package require a running PostgreSQL and Redis instance.
// Set environment variable TEST_DATABASE_URL and TEST_REDIS_URL, or use the
// defaults matching the docker-compose.yml configuration.
package testutil

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"flickey/go-backend/auth"
	"flickey/go-backend/config"
	"flickey/go-backend/db"
	"flickey/go-backend/sms"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck
}

// TestConfig returns a config suitable for testing.
// Reads from environment, falls back to local docker-compose defaults.
func TestConfig() *config.Settings {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://rental:rental_dev@localhost:5432/flickey"
	}
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	return &config.Settings{
		Environment:                   "dev",
		Port:                          "8081",
		DatabaseURL:                   dbURL,
		RedisURL:                      redisURL,
		JWTSecret:                     "test-super-secret-jwt-key-must-be-32b!!",
		JWTAlgorithm:                  "HS256",
		JWTAccessTTLSeconds:           900,
		JWTRefreshTTLSeconds:          2592000,
		OTPTTLSeconds:                 300,
		OTPLength:                     6,
		OTPMaxVerifyAttempts:          50,
		OTPMaxRequestsPerPhonePerHour: 1000,
		OTPMaxRequestsPerIPPerHour:    1000,
		OTPResendCooldownSeconds:      0,
		CookieSecure:                  false,
		MediaMaxFileSizeBytes:         15728640,
		MediaMinCount:                 5,
		MediaMaxCount:                 25,
	}
}

// Suite holds shared test infrastructure for integration tests.
type Suite struct {
	T     *testing.T
	Cfg   *config.Settings
	DB    *gorm.DB
	Redis *redis.Client
	Auth  *auth.AuthService
}

// NewSuite creates and connects all test infrastructure.
// Skips the test if the database/Redis is not reachable.
func NewSuite(t *testing.T) *Suite {
	t.Helper()
	cfg := TestConfig()

	database, err := db.Open(cfg)
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to PostgreSQL: %v", err)
	}

	sqlDB, _ := database.DB()
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("skipping integration test: PostgreSQL ping failed: %v", err)
	}

	rdb, err := db.NewRedis(cfg)
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to Redis: %v", err)
	}

	authSvc := &auth.AuthService{
		DB:        database,
		Redis:     rdb,
		Cfg:       cfg,
		SMSSender: &sms.ConsoleSMSSender{},
		Logger:    newNopLogger(),
	}

	return &Suite{
		T:     t,
		Cfg:   cfg,
		DB:    database,
		Redis: rdb,
		Auth:  authSvc,
	}
}

// Cleanup cleans up test listings, drafts, amenities, and test state after each test.
func (s *Suite) Cleanup() {
	s.T.Helper()
	ctx := context.Background()
	tables := []string{
		"listing_amenities",
		"listing_drafts",
		"listings",
	}
	for _, table := range tables {
		s.DB.WithContext(ctx).Exec("DELETE FROM " + table)
	}
	s.DB.WithContext(ctx).Exec("UPDATE media SET listing_id = NULL")
	s.DB.WithContext(ctx).Exec("DELETE FROM notifications WHERE user_id != '8752b84c-7bd8-4926-a527-f97419a9865d'")
	s.DB.WithContext(ctx).Exec("DELETE FROM audit_logs")
	s.DB.WithContext(ctx).Exec("DELETE FROM users WHERE id != '8752b84c-7bd8-4926-a527-f97419a9865d'")
	s.Redis.FlushAll(ctx)
}

// CreateUser inserts a test user and returns it with access and refresh tokens.
func (s *Suite) CreateUser(phone string, role db.UserRole) (*db.User, string, string) {
	s.T.Helper()
	ctx := context.Background()
	firstName, lastName := "Test", "User"
	u := &db.User{
		ID:        uuid.New(),
		Phone:     phone,
		FirstName: &firstName,
		LastName:  &lastName,
		Role:      role,
		Status:    db.UserStatusActive,
	}
	if err := db.CreateUser(ctx, s.DB, u); err != nil {
		s.T.Fatalf("CreateUser: %v", err)
	}
	access, err := auth.CreateAccessToken(s.Cfg, u.ID)
	if err != nil {
		s.T.Fatalf("CreateAccessToken: %v", err)
	}
	refresh, err := auth.CreateRefreshToken(ctx, s.Redis, s.Cfg, u.ID, "", "")
	if err != nil {
		s.T.Fatalf("CreateRefreshToken: %v", err)
	}
	return u, access, refresh
}

// RandomPhone generates a random E.164 phone number for testing.
func RandomPhone() string {
	return fmt.Sprintf("+37529%07d", rand.Intn(9999999))
}

// DoRequest executes a request against the provided Gin engine and returns the response recorder.
func DoRequest(engine *gin.Engine, method, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, _ := http.NewRequest(method, path, bodyReader)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func newNopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
