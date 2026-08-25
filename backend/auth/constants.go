// Package auth provides authentication constants for Redis key prefixes.
// These must match the Python constants exactly for cross-compatibility.
package auth

const (
	// OTP challenge storage: HASH {hash, phone, attempts, return_url?}
	OTPChallengePrefix = "otp:challenge:"
	// Phone → challenge reverse mapping for invalidation.
	OTPPhoneChallengePrefix = "otp:phone_challenge:"
	// Per-phone OTP rate limit counter.
	OTPPhoneRatePrefix = "otp:phone_rate:"
	// Per-IP OTP rate limit counter.
	OTPIPRatePrefix = "otp:ip_rate:"
	// Resend cooldown NX lock.
	OTPResendCooldownPrefix = "otp:resend_cooldown:"

	// Refresh token HASH: {user_id, session_id, family_id, rotated, created_at}
	RefreshTokenPrefix = "refresh:"
	// Set of all refresh token hashes for a user (logout-all).
	RefreshUserPrefix = "refresh:user:"

	// OTP challenge hash fields.
	OTPFieldHash      = "hash"
	OTPFieldPhone     = "phone"
	OTPFieldAttempts  = "attempts"
	OTPFieldReturnURL = "return_url"

	// Refresh token hash fields.
	RTFieldUserID    = "user_id"
	RTFieldSessionID = "session_id"
	RTFieldFamilyID  = "family_id"
	RTFieldRotated   = "rotated"
	RTFieldCreatedAt = "created_at"

	// JWT claim keys.
	JWTClaimType  = "type"
	JWTTypeAccess = "access"
)
