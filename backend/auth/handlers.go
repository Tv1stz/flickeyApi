// Package auth provides Gin HTTP handlers for authentication endpoints.
package auth

import (
	"net/http"
	"time"

	"flickey/go-backend/config"
	"flickey/go-backend/db"

	"github.com/gin-gonic/gin"
)

const refreshCookieName = "refresh_token"

// Handler holds dependencies for auth HTTP handlers.
type Handler struct {
	Service *AuthService
	Cfg     *config.Settings
}

// NewHandler creates an auth Handler.
func NewHandler(svc *AuthService, cfg *config.Settings) *Handler {
	return &Handler{Service: svc, Cfg: cfg}
}

// RegisterRoutes registers all /auth routes under the provided RouterGroup.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc, activeMw gin.HandlerFunc) {
	rg.POST("/request-otp", h.RequestOTP)
	rg.POST("/verify-otp", h.VerifyOTP)
	rg.POST("/refresh", h.Refresh)
	rg.POST("/logout", h.Logout)

	// Authenticated routes (accessible to active and pending_profile users).
	auth := rg.Group("")
	auth.Use(authMw)
	auth.GET("/me", h.GetMe)
	auth.POST("/complete-profile", h.CompleteProfile)
	auth.POST("/logout-all", h.LogoutAll)

	// Dev-only route — registered conditionally in main.go.
	if h.Cfg.IsDev() {
		rg.POST("/dev-login", h.DevLogin)
	}
}

// RequestOTP handles POST /auth/request-otp.
//
//	@Summary		Request an OTP code
//	@Description	Sends a one-time password to the provided phone number and returns a challenge ID used in the verify step.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		OTPRequestIn	true	"Phone number (E.164) and optional return URL"
//	@Success		200		{object}	OTPRequestOut
//	@Failure		400		{object}	ErrorResponse	"VALIDATION_ERROR – missing or invalid phone"
//	@Failure		429		{object}	ErrorResponse	"RATE_LIMIT_EXCEEDED – too many OTP requests"
//	@Failure		500		{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Router			/auth/request-otp [post]
func (h *Handler) RequestOTP(c *gin.Context) {
	var req OTPRequestIn
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	ip := ExtractClientIP(
		c.Request.RemoteAddr,
		c.GetHeader("X-Forwarded-For"),
		h.Cfg.TrustProxyHeaders,
	)

	challengeID, err := h.Service.RequestOTP(c.Request.Context(), req.Phone, ip, req.ReturnURL)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, OTPRequestOut{
		ChallengeID: challengeID,
		ExpiresIn:   h.Cfg.OTPTTLSeconds,
	})
}

// VerifyOTP handles POST /auth/verify-otp.
//
//	@Summary		Verify OTP and authenticate
//	@Description	Validates the 6-digit OTP against the challenge ID. On success issues an access token in the body and sets an HttpOnly refresh_token cookie.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		OTPVerifyIn		true	"Challenge ID and 6-digit OTP code"
//	@Success		200		{object}	OTPVerifyOut	"access_token (JWT bearer) + is_new_user flag"
//	@Failure		400		{object}	ErrorResponse	"VALIDATION_ERROR – missing fields or wrong OTP format"
//	@Failure		401		{object}	ErrorResponse	"INVALID_OTP / OTP_EXPIRED / CHALLENGE_NOT_FOUND"
//	@Failure		429		{object}	ErrorResponse	"RATE_LIMIT_EXCEEDED"
//	@Failure		500		{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Router			/auth/verify-otp [post]
func (h *Handler) VerifyOTP(c *gin.Context) {
	var req OTPVerifyIn
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	result, err := h.Service.VerifyOTP(c.Request.Context(), req.ChallengeID, req.Code)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	h.setRefreshCookie(c, result.RefreshToken)

	c.JSON(http.StatusOK, OTPVerifyOut{
		AccessToken: result.AccessToken,
		TokenType:   "bearer",
		IsNewUser:   result.IsNewUser,
		ReturnURL:   result.ReturnURL,
	})
}

// CompleteProfile handles POST /auth/complete-profile.
//
//	@Summary		Complete user profile
//	@Description	Sets first name, last name, and optional email for a newly registered user. Transitions status from pending_profile to active.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		ProfileCompleteRequest	true	"Profile fields (first_name and last_name required, email optional)"
//	@Success		200		{object}	UserRead
//	@Failure		400		{object}	ErrorResponse	"VALIDATION_ERROR"
//	@Failure		401		{object}	ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN / TOKEN_EXPIRED"
//	@Failure		409		{object}	ErrorResponse	"EMAIL_TAKEN – email already used by another account"
//	@Failure		500		{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/auth/complete-profile [post]
func (h *Handler) CompleteProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		respondUnauthorized(c, "MISSING_USER", "User not in context.")
		return
	}

	var req ProfileCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	updated, err := h.Service.CompleteProfile(c.Request.Context(), user.ID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, UserReadFromModel(updated))
}

// Refresh handles POST /auth/refresh.
//
//	@Summary		Refresh access token
//	@Description	Exchanges a valid refresh_token cookie (or Authorization header in dev) for a new access token and rotates the refresh token.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	RefreshOut
//	@Failure		401	{object}	ErrorResponse	"MISSING_REFRESH_TOKEN / TOKEN_EXPIRED / INVALID_TOKEN"
//	@Failure		500	{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Security		CookieAuth
//	@Router			/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	rawToken := c.GetHeader("Authorization") // For dev-UI testing.
	if rawToken == "" {
		rawToken, _ = c.Cookie(refreshCookieName)
	}

	if rawToken == "" {
		respondUnauthorized(c, "MISSING_REFRESH_TOKEN", "Refresh token not provided.")
		return
	}

	newAccess, newRefresh, err := h.Service.RefreshTokens(c.Request.Context(), rawToken)
	if err != nil {
		h.clearRefreshCookie(c)
		handleServiceError(c, err)
		return
	}

	h.setRefreshCookie(c, newRefresh)

	c.JSON(http.StatusOK, RefreshOut{
		AccessToken: newAccess,
		TokenType:   "bearer",
	})
}

// Logout handles POST /auth/logout.
//
//	@Summary		Log out (current session)
//	@Description	Revokes the current refresh token and clears the HttpOnly cookie.
//	@Tags			Auth
//	@Produce		json
//	@Success		204	"No Content"
//	@Failure		500	{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Security		CookieAuth
//	@Router			/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	rawToken, _ := c.Cookie(refreshCookieName)
	_ = h.Service.Logout(c.Request.Context(), rawToken)
	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

// LogoutAll handles POST /auth/logout-all.
//
//	@Summary		Log out all sessions
//	@Description	Revokes all refresh tokens for the current user across all devices.
//	@Tags			Auth
//	@Produce		json
//	@Success		204	"No Content"
//	@Failure		401	{object}	ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		500	{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/auth/logout-all [post]
func (h *Handler) LogoutAll(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		respondUnauthorized(c, "MISSING_USER", "User not in context.")
		return
	}
	_ = h.Service.LogoutAll(c.Request.Context(), user.ID)
	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

// GetMe handles GET /auth/me.
// Auto-upgrades guest → host if user has listings in DB.
//
//	@Summary		Get current user
//	@Description	Returns the authenticated user's profile. Automatically promotes guest → host if the user already has published listings.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	UserRead
//	@Failure		401	{object}	ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN / TOKEN_EXPIRED"
//	@Failure		500	{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		respondUnauthorized(c, "MISSING_USER", "User not in context.")
		return
	}

	// Auto-sync: if user is guest but has listings, promote to host.
	if user.Role == db.UserRoleGuest {
		hasListings, err := db.FindListingByHostExists(c.Request.Context(), h.Service.DB, user.ID)
		if err == nil && hasListings {
			_ = db.UpdateUserRole(c.Request.Context(), h.Service.DB, user.ID, db.UserRoleHost)
			user.Role = db.UserRoleHost
		}
	}

	c.JSON(http.StatusOK, UserReadFromModel(user))
}

// DevLogin handles POST /auth/dev-login (dev environment only).
//
//	@Summary		Dev-only instant login
//	@Description	Creates or finds a user by phone and returns tokens without OTP verification. **Only available when ENVIRONMENT=dev.**
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		DevLoginIn	true	"Phone and optional profile fields"
//	@Success		200		{object}	DevLoginOut
//	@Failure		400		{object}	ErrorResponse	"VALIDATION_ERROR"
//	@Failure		404		{object}	ErrorResponse	"NOT_FOUND – endpoint disabled in non-dev environments"
//	@Failure		500		{object}	ErrorResponse	"INTERNAL_ERROR"
//	@Router			/auth/dev-login [post]
func (h *Handler) DevLogin(c *gin.Context) {
	if !h.Cfg.IsDev() {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Not found."})
		return
	}

	var req DevLoginIn
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	user, accessToken, refreshToken, err := h.Service.DevLogin(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)

	c.JSON(http.StatusOK, DevLoginOut{
		AccessToken: accessToken,
		TokenType:   "bearer",
		User:        UserReadFromModel(user),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Cookie helpers
// ─────────────────────────────────────────────────────────────────────────────

func (h *Handler) setRefreshCookie(c *gin.Context, rawToken string) {
	sameSite := http.SameSiteLaxMode
	c.SetSameSite(sameSite)
	c.SetCookie(
		refreshCookieName,
		rawToken,
		h.Cfg.JWTRefreshTTLSeconds,
		"/api/v1/auth",
		h.Cfg.CookieDomain,
		h.Cfg.CookieSecure,
		true, // HttpOnly
	)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshCookieName,
		"",
		-1,
		"/api/v1/auth",
		h.Cfg.CookieDomain,
		h.Cfg.CookieSecure,
		true,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Error handling helpers
// ─────────────────────────────────────────────────────────────────────────────

func handleServiceError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *AuthError:
		c.Header("WWW-Authenticate", "Bearer")
		c.JSON(http.StatusUnauthorized, gin.H{"code": e.Code, "message": e.Message})
	case *ValidationError:
		resp := gin.H{"code": e.Code, "message": e.Message}
		if e.Details != nil {
			resp["details"] = e.Details
		}
		c.JSON(http.StatusBadRequest, resp)
	case *ConflictError:
		c.JSON(http.StatusConflict, gin.H{"code": e.Code, "message": e.Message})
	case *NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"code": e.Code, "message": e.Message})
	case *AuthorizationError:
		c.JSON(http.StatusForbidden, gin.H{"code": e.Code, "message": e.Message})
	case *RateLimitError:
		if e.RetryAfterSeconds > 0 {
			c.Header("Retry-After", time.Duration(e.RetryAfterSeconds*int64(time.Second)).String())
		}
		c.JSON(http.StatusTooManyRequests, gin.H{"code": e.Code, "message": e.Message})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "An internal error occurred."})
	}
}

func respondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
}
