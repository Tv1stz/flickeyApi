// Package auth provides the AuthService and UserService for all auth operations.
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"flickey/go-backend/config"
	"flickey/go-backend/db"
	"flickey/go-backend/sms"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AuthService orchestrates the authentication flow.
type AuthService struct {
	DB        *gorm.DB
	Redis     *redis.Client
	Cfg       *config.Settings
	SMSSender sms.SMSSender
	Logger    *slog.Logger
}

func (s *AuthService) log() *slog.Logger {
	if s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

// RequestOTP validates the phone, enforces rate limits, generates and sends an OTP.
// Returns the challenge_id for the client to use in the verify step.
func (s *AuthService) RequestOTP(ctx context.Context, phone, ip, returnURL string) (string, error) {
	// Normalize phone to E.164.
	normalizedPhone, err := NormalizePhone(phone)
	if err != nil {
		return "", &ValidationError{Code: "INVALID_PHONE", Message: err.Error()}
	}

	// Validate return_url if provided.
	if returnURL != "" && !ValidateReturnURL(returnURL, s.Cfg.AllowedRedirectHosts) {
		return "", &ValidationError{Code: "INVALID_RETURN_URL", Message: "Return URL is not allowed."}
	}

	// Enforce per-phone rate limit (5/hr).
	if err := CheckPhoneRateLimit(ctx, s.Redis, s.Cfg, normalizedPhone); err != nil {
		return "", err
	}

	// Enforce per-IP rate limit (20/hr).
	if err := CheckIPRateLimit(ctx, s.Redis, s.Cfg, ip); err != nil {
		return "", err
	}

	// Enforce resend cooldown (60s).
	if err := CheckResendCooldown(ctx, s.Redis, s.Cfg, normalizedPhone); err != nil {
		return "", err
	}

	// Invalidate previous challenge for this phone.
	if err := InvalidatePreviousChallenges(ctx, s.Redis, normalizedPhone); err != nil {
		s.log().WarnContext(ctx, "failed to invalidate previous challenges", "error", err)
	}

	// Generate OTP and challenge ID.
	otp, err := GenerateOTP(s.Cfg.OTPLength)
	if err != nil {
		return "", fmt.Errorf("RequestOTP: %w", err)
	}

	challengeID, err := GenerateChallengeID()
	if err != nil {
		return "", fmt.Errorf("RequestOTP: %w", err)
	}

	otpHash, err := HashValue(otp)
	if err != nil {
		return "", fmt.Errorf("RequestOTP: %w", err)
	}

	// Store challenge in Redis.
	if err := StoreOTPChallenge(ctx, s.Redis, s.Cfg, challengeID, normalizedPhone, otpHash, returnURL); err != nil {
		return "", fmt.Errorf("RequestOTP store: %w", err)
	}

	// Store phone → challenge reverse mapping.
	if err := StorePhonesChallengeMappings(ctx, s.Redis, s.Cfg, normalizedPhone, challengeID); err != nil {
		s.log().WarnContext(ctx, "failed to store phone challenge mapping", "error", err)
	}

	// Send OTP via SMS.
	if err := s.SMSSender.SendOTP(ctx, normalizedPhone, otp); err != nil {
		s.log().ErrorContext(ctx, "failed to send OTP SMS",
			"phone", MaskPhone(normalizedPhone),
			"error", err,
		)
		return "", fmt.Errorf("RequestOTP send SMS: %w", err)
	}

	s.log().InfoContext(ctx, "otp_sent",
		"phone", MaskPhone(normalizedPhone),
		"challenge_id", challengeID[:8]+"...",
	)

	return challengeID, nil
}

// VerifyOTPResult is returned from VerifyOTP.
type VerifyOTPResult struct {
	AccessToken  string
	RefreshToken string
	IsNewUser    bool
	ReturnURL    string
	User         *db.User
}

// VerifyOTP validates an OTP code, creates/fetches the user, and issues tokens.
func (s *AuthService) VerifyOTP(ctx context.Context, challengeID, code string) (*VerifyOTPResult, error) {
	// Atomic attempt check + retrieval.
	res, err := VerifyOTPChallengeAtomic(ctx, s.Redis, s.Cfg, challengeID)
	if err != nil {
		return nil, err
	}

	if !res.Found {
		if res.Exhausted {
			return nil, &AuthError{Code: "OTP_EXPIRED_OR_MAX_ATTEMPTS", Message: "OTP has expired or maximum attempts exceeded."}
		}
		return nil, &AuthError{Code: "INVALID_CHALLENGE", Message: "Challenge not found. Request a new OTP."}
	}

	// Validate OTP with constant-time comparison.
	if !VerifyHash(code, res.HashVal) {
		return nil, &AuthError{Code: "INVALID_OTP", Message: "Incorrect OTP code."}
	}

	// Delete challenge immediately after successful verification.
	_ = DeleteOTPChallenge(ctx, s.Redis, challengeID)

	// Reset rate limits and resend cooldowns for this phone number upon successful login.
	ResetPhoneRateLimits(ctx, s.Redis, res.Phone)

	// Get or create user.
	user, isNew, err := s.getOrCreateByPhone(ctx, res.Phone)
	if err != nil {
		return nil, err
	}

	// Issue tokens.
	accessToken, err := CreateAccessToken(s.Cfg, user.ID)
	if err != nil {
		return nil, fmt.Errorf("VerifyOTP access token: %w", err)
	}

	refreshToken, err := CreateRefreshToken(ctx, s.Redis, s.Cfg, user.ID, "", "")
	if err != nil {
		return nil, fmt.Errorf("VerifyOTP refresh token: %w", err)
	}

	s.log().InfoContext(ctx, "otp_verified",
		"user_id", user.ID,
		"is_new", isNew,
	)

	return &VerifyOTPResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    isNew,
		ReturnURL:    res.ReturnURL,
		User:         user,
	}, nil
}

// RefreshTokens rotates a refresh token and issues new access + refresh tokens.
func (s *AuthService) RefreshTokens(ctx context.Context, rawRefreshToken string) (string, string, error) {
	if rawRefreshToken == "" {
		return "", "", &AuthError{Code: "MISSING_REFRESH_TOKEN", Message: "Refresh token not provided."}
	}

	newAccess, newRefresh, _, err := RotateRefreshToken(ctx, s.Redis, s.Cfg, rawRefreshToken)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenReuse) {
			return "", "", &AuthError{Code: "TOKEN_REUSE_DETECTED", Message: "Refresh token reuse detected. All sessions have been revoked."}
		}
		return "", "", &AuthError{Code: "INVALID_REFRESH_TOKEN", Message: "Invalid or expired refresh token."}
	}
	return newAccess, newRefresh, nil
}

// Logout revokes a single refresh token.
func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil // idempotent
	}
	return RevokeRefreshToken(ctx, s.Redis, rawRefreshToken)
}

// LogoutAll revokes all refresh tokens for a user.
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return RevokeAllUserTokens(ctx, s.Redis, userID)
}

// getOrCreateByPhone fetches or creates a user by phone number.
// Returns (user, isNew, error).
func (s *AuthService) getOrCreateByPhone(ctx context.Context, phone string) (*db.User, bool, error) {
	user, err := db.FindUserByPhone(ctx, s.DB, phone)
	if err != nil {
		return nil, false, fmt.Errorf("getOrCreateByPhone find: %w", err)
	}
	if user != nil {
		return user, false, nil
	}

	// Create new user.
	newUser := &db.User{
		ID:     uuid.New(),
		Phone:  phone,
		Role:   db.UserRoleGuest,
		Status: db.UserStatusPendingProfile,
	}

	if err := db.CreateUser(ctx, s.DB, newUser); err != nil {
		// Handle concurrent creation — retry fetch.
		existing, findErr := db.FindUserByPhone(ctx, s.DB, phone)
		if findErr != nil || existing == nil {
			return nil, false, fmt.Errorf("getOrCreateByPhone create: %w", err)
		}
		return existing, false, nil
	}

	return newUser, true, nil
}

// CompleteProfile transitions a user from pending_profile → active.
func (s *AuthService) CompleteProfile(ctx context.Context, userID uuid.UUID, req ProfileCompleteRequest) (*db.User, error) {
	user, err := db.FindUserByIDCtx(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &NotFoundError{Code: "USER_NOT_FOUND", Message: "User not found."}
	}
	if user.Status != db.UserStatusPendingProfile {
		return nil, &ValidationError{Code: "PROFILE_ALREADY_COMPLETE", Message: "Profile has already been completed."}
	}

	// Check email uniqueness if provided.
	if req.Email != "" {
		existing, err := db.FindUserByEmail(ctx, s.DB, req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != userID {
			return nil, &ConflictError{Code: "EMAIL_TAKEN", Message: "Email address is already in use."}
		}
	}

	var email *string
	if req.Email != "" {
		email = &req.Email
	}

	err = s.DB.WithContext(ctx).Model(&db.User{}).Where("id = ?", userID).Updates(map[string]any{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"email":      email,
		"status":     db.UserStatusActive,
		"updated_at": time.Now().UTC(),
	}).Error
	if err != nil {
		return nil, fmt.Errorf("CompleteProfile update: %w", err)
	}

	updated, err := db.FindUserByIDCtx(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DevLogin creates or updates a user for development purposes and returns instant tokens.
// Dev env only — returns 404 in production.
func (s *AuthService) DevLogin(ctx context.Context, req DevLoginIn) (*db.User, string, string, error) {
	// Normalize phone (lenient for dev).
	phone := req.Phone
	if normalized, err := NormalizePhone(phone); err == nil {
		phone = normalized
	}

	role := db.UserRoleGuest
	if req.Role == "host" {
		role = db.UserRoleHost
	} else if req.Role == "admin" {
		role = db.UserRoleAdmin
	}

	firstName := req.FirstName
	if firstName == "" {
		firstName = "Dev"
	}
	lastName := req.LastName
	if lastName == "" {
		lastName = "User"
	}

	user, err := db.FindUserByPhone(ctx, s.DB, phone)
	if err != nil {
		return nil, "", "", err
	}

	if user == nil {
		user = &db.User{
			ID:        uuid.New(),
			Phone:     phone,
			FirstName: &firstName,
			LastName:  &lastName,
			Role:      role,
			Status:    db.UserStatusActive,
		}
		if err := db.CreateUser(ctx, s.DB, user); err != nil {
			return nil, "", "", fmt.Errorf("DevLogin create: %w", err)
		}
	} else {
		updates := map[string]any{
			"first_name": firstName,
			"last_name":  lastName,
			"role":       role,
			"status":     db.UserStatusActive,
			"updated_at": time.Now().UTC(),
		}
		if err := s.DB.WithContext(ctx).Model(&db.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			return nil, "", "", fmt.Errorf("DevLogin update: %w", err)
		}
		user.FirstName = &firstName
		user.LastName = &lastName
		user.Role = role
		user.Status = db.UserStatusActive
	}

	accessToken, err := CreateAccessToken(s.Cfg, user.ID)
	if err != nil {
		return nil, "", "", err
	}
	refreshToken, err := CreateRefreshToken(ctx, s.Redis, s.Cfg, user.ID, "", "")
	if err != nil {
		return nil, "", "", err
	}

	ResetPhoneRateLimits(ctx, s.Redis, phone)
	return user, accessToken, refreshToken, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Domain Errors
// ─────────────────────────────────────────────────────────────────────────────

// AuthError is returned for authentication failures (401).
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string { return e.Message }

// ValidationError is returned for invalid input (400).
type ValidationError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *ValidationError) Error() string { return e.Message }

// ConflictError is returned when a unique constraint is violated (409).
type ConflictError struct {
	Code    string
	Message string
}

func (e *ConflictError) Error() string { return e.Message }

// NotFoundError is returned when a resource is not found (404).
type NotFoundError struct {
	Code    string
	Message string
}

func (e *NotFoundError) Error() string { return e.Message }

// AuthorizationError is returned for insufficient permissions (403).
type AuthorizationError struct {
	Code    string
	Message string
}

func (e *AuthorizationError) Error() string { return e.Message }
