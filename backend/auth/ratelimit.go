// Package auth provides atomic Lua-based rate limiting for OTP requests.
// All rate limits are identical to the FastAPI implementation.
package auth

import (
	"context"
	"fmt"
	"time"

	"flickey/go-backend/config"

	"github.com/redis/go-redis/v9"
)

// RateLimitError is returned when a rate limit is exceeded.
type RateLimitError struct {
	Code              string
	Message           string
	RetryAfterSeconds int64
}

func (e *RateLimitError) Error() string { return e.Message }

// ─────────────────────────────────────────────────────────────────────────────
// Atomic Lua Rate Limiting
// ─────────────────────────────────────────────────────────────────────────────

// rateLimitLua atomically checks and increments a rate limit counter.
// KEYS[1] = counter key, ARGV[1] = limit, ARGV[2] = window seconds
// Returns: current count (positive) or -1 if limit exceeded (with TTL in seconds in element 2)
const rateLimitLua = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call('GET', key)
if current and tonumber(current) >= limit then
    local ttl = redis.call('TTL', key)
    return {-1, ttl}
end

local count = redis.call('INCR', key)
if count == 1 then
    redis.call('EXPIRE', key, window)
end

if count > limit then
    local ttl = redis.call('TTL', key)
    return {-1, ttl}
end

return {count, 0}
`

func checkRateLimit(ctx context.Context, rdb *redis.Client, key string, limit, windowSeconds int) (int64, error) {
	res, err := rdb.Eval(ctx, rateLimitLua, []string{key}, fmt.Sprintf("%d", limit), fmt.Sprintf("%d", windowSeconds)).Result()
	if err != nil {
		return 0, fmt.Errorf("rate limit eval: %w", err)
	}
	vals, ok := res.([]any)
	if !ok || len(vals) < 2 {
		return 0, fmt.Errorf("rate limit: unexpected result")
	}
	count := toInt64(vals[0])
	ttl := toInt64(vals[1])
	if count == -1 {
		return ttl, &RateLimitError{Code: "RATE_LIMIT_EXCEEDED", RetryAfterSeconds: ttl}
	}
	return 0, nil
}

// ResetPhoneRateLimits clears phone rate limit and resend cooldown keys upon successful verification/login.
func ResetPhoneRateLimits(ctx context.Context, rdb *redis.Client, phone string) {
	if rdb == nil {
		return
	}
	phoneHash := HashPhoneForKey(phone)
	phoneRateKey := OTPPhoneRatePrefix + phoneHash
	cooldownKey := OTPResendCooldownPrefix + phoneHash
	_ = rdb.Del(ctx, phoneRateKey, cooldownKey).Err()
}

// CheckPhoneRateLimit enforces per-phone OTP request rate limiting (5 requests/hour).
func CheckPhoneRateLimit(ctx context.Context, rdb *redis.Client, cfg *config.Settings, phone string) error {
	key := OTPPhoneRatePrefix + HashPhoneForKey(phone)
	_, err := checkRateLimit(ctx, rdb, key, cfg.OTPMaxRequestsPerPhonePerHour, 3600)
	if err != nil {
		if rl, ok := err.(*RateLimitError); ok {
			rl.Message = "Too many OTP requests for this phone number. Try again later."
			return rl
		}
		return err
	}
	return nil
}

// CheckIPRateLimit enforces per-IP OTP request rate limiting (20 requests/hour).
func CheckIPRateLimit(ctx context.Context, rdb *redis.Client, cfg *config.Settings, ip string) error {
	key := OTPIPRatePrefix + ip
	_, err := checkRateLimit(ctx, rdb, key, cfg.OTPMaxRequestsPerIPPerHour, 3600)
	if err != nil {
		if rl, ok := err.(*RateLimitError); ok {
			rl.Message = "Too many OTP requests from this IP address. Try again later."
			return rl
		}
		return err
	}
	return nil
}

// CheckResendCooldown enforces a per-phone resend cooldown (60 seconds, NX lock).
// Returns error if the phone is still within the cooldown window.
func CheckResendCooldown(ctx context.Context, rdb *redis.Client, cfg *config.Settings, phone string) error {
	key := OTPResendCooldownPrefix + HashPhoneForKey(phone)
	ttl := time.Duration(cfg.OTPResendCooldownSeconds) * time.Second

	// SET NX returns true if set (no cooldown), false if already set (in cooldown).
	set, err := rdb.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return fmt.Errorf("CheckResendCooldown: %w", err)
	}
	if !set {
		// Already set — get remaining TTL.
		remaining, err := rdb.TTL(ctx, key).Result()
		if err != nil {
			remaining = ttl
		}
		return &RateLimitError{
			Code:              "RESEND_COOLDOWN",
			Message:           fmt.Sprintf("Please wait %d seconds before requesting another OTP.", int(remaining.Seconds())),
			RetryAfterSeconds: int64(remaining.Seconds()),
		}
	}
	return nil
}
