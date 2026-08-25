// Package auth provides request/response types for authentication endpoints.
package auth

import (
	"time"

	"flickey/go-backend/db"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// OTP Request/Response
// ─────────────────────────────────────────────────────────────────────────────

// OTPRequestIn is the request body for POST /auth/request-otp.
type OTPRequestIn struct {
	Phone     string `json:"phone" binding:"required"`
	ReturnURL string `json:"return_url"`
}

// OTPRequestOut is the response for POST /auth/request-otp.
type OTPRequestOut struct {
	ChallengeID string `json:"challenge_id"`
	ExpiresIn   int    `json:"expires_in"`
}

// OTPVerifyIn is the request body for POST /auth/verify-otp.
type OTPVerifyIn struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	Code        string `json:"code" binding:"required,len=6"`
}

// OTPVerifyOut is the response for POST /auth/verify-otp.
type OTPVerifyOut struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IsNewUser   bool   `json:"is_new_user"`
	ReturnURL   string `json:"return_url,omitempty"`
}

// RefreshOut is the response for POST /auth/refresh.
type RefreshOut struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// MessageOut is a generic message response.
type MessageOut struct {
	Message string `json:"message"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Profile
// ─────────────────────────────────────────────────────────────────────────────

// ProfileCompleteRequest is the body for POST /auth/complete-profile.
type ProfileCompleteRequest struct {
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name" binding:"required,min=1,max=100"`
	Email     string `json:"email" binding:"omitempty,email,max=255"`
}

// ProfileUpdateRequest is the body for PATCH /auth/profile (future use).
type ProfileUpdateRequest struct {
	FirstName string `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email     string `json:"email" binding:"omitempty,email,max=255"`
}

// ─────────────────────────────────────────────────────────────────────────────
// User Read
// ─────────────────────────────────────────────────────────────────────────────

// UserRead is the read-only user representation returned from /me.
type UserRead struct {
	ID        uuid.UUID     `json:"id"`
	Phone     string        `json:"phone"`
	FirstName *string       `json:"first_name"`
	LastName  *string       `json:"last_name"`
	Email     *string       `json:"email"`
	Role      db.UserRole   `json:"role"`
	Status    db.UserStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// UserReadFromModel converts a db.User model to a UserRead response.
func UserReadFromModel(u *db.User) UserRead {
	return UserRead{
		ID:        u.ID,
		Phone:     u.Phone,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Dev Login
// ─────────────────────────────────────────────────────────────────────────────

// DevLoginIn is the request body for POST /auth/dev-login (dev env only).
type DevLoginIn struct {
	Phone     string `json:"phone" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// DevLoginOut is the response for POST /auth/dev-login.
type DevLoginOut struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	User        UserRead `json:"user"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Shared error envelope
// ─────────────────────────────────────────────────────────────────────────────

// ErrorResponse is the standard error body returned on 4xx / 5xx responses.
type ErrorResponse struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Human-readable error description."`
	// Details is present only for VALIDATION_ERROR responses.
	Details any `json:"details,omitempty"`
}
