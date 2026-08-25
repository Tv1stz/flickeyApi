// Package auth provides OTP generation, hashing, storage, and verification.
//
// Security model:
//   - OTPs are generated using crypto/rand (OS CSPRNG).
//   - Stored as salted SHA-256 hashes: "salt:sha256(salt+':'+otp)".
//   - Verification uses constant-time comparison (crypto/subtle).
//   - Attempt counting is atomic via Lua to prevent TOCTOU races.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"flickey/go-backend/config"

	"github.com/redis/go-redis/v9"
)

// GenerateOTP produces a cryptographically secure N-digit OTP (zero-padded).
func GenerateOTP(length int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generating OTP: %w", err)
	}
	return fmt.Sprintf("%0*s", length, n.String()), nil
}

// GenerateChallengeID produces a URL-safe random string for the OTP challenge.
func GenerateChallengeID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating challenge ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashValue hashes a value with a random salt: "salt:sha256(salt+':'+value)".
// Used for OTPs — salted to prevent precomputation attacks.
func HashValue(value string) (string, error) {
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	salt := hex.EncodeToString(saltBytes)
	sum := sha256.Sum256([]byte(salt + ":" + value))
	hash := hex.EncodeToString(sum[:])
	return salt + ":" + hash, nil
}

// VerifyHash checks a value against a stored "salt:hash" string.
// Uses constant-time comparison to prevent timing attacks.
func VerifyHash(value, stored string) bool {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, expectedHash := parts[0], parts[1]
	sum := sha256.Sum256([]byte(salt + ":" + value))
	candidate := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(expectedHash)) == 1
}

// HashPhoneForKey produces a deterministic 16-char hex of the phone for Redis keys.
// Prevents rate-limit bypass via phone format variations.
func HashPhoneForKey(phone string) string {
	sum := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(sum[:])[:16]
}

// StoreOTPChallenge writes the OTP challenge to Redis as a HASH with TTL.
func StoreOTPChallenge(ctx context.Context, rdb *redis.Client, cfg *config.Settings, challengeID, phone, otpHash string, returnURL string) error {
	key := OTPChallengePrefix + challengeID
	mapping := map[string]any{
		OTPFieldHash:     otpHash,
		OTPFieldPhone:    phone,
		OTPFieldAttempts: "0",
	}
	if returnURL != "" {
		mapping[OTPFieldReturnURL] = returnURL
	}

	pipe := rdb.Pipeline()
	pipe.HSet(ctx, key, mapping)
	pipe.Expire(ctx, key, time.Duration(cfg.OTPTTLSeconds)*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("StoreOTPChallenge: %w", err)
	}
	return nil
}

// StorePhonesChallengeMappings stores a phone→challengeID mapping for later invalidation.
func StorePhonesChallengeMappings(ctx context.Context, rdb *redis.Client, cfg *config.Settings, phone, challengeID string) error {
	key := OTPPhoneChallengePrefix + phone
	if err := rdb.Set(ctx, key, challengeID, time.Duration(cfg.OTPTTLSeconds)*time.Second).Err(); err != nil {
		return fmt.Errorf("StorePhonesChallengeMappings: %w", err)
	}
	return nil
}

// InvalidatePreviousChallenges deletes any previous OTP challenge for this phone.
func InvalidatePreviousChallenges(ctx context.Context, rdb *redis.Client, phone string) error {
	mappingKey := OTPPhoneChallengePrefix + phone
	prevChallengeID, err := rdb.Get(ctx, mappingKey).Result()
	if err == redis.Nil {
		return nil // No previous challenge.
	}
	if err != nil {
		return fmt.Errorf("InvalidatePreviousChallenges get: %w", err)
	}
	if prevChallengeID != "" {
		if err := rdb.Del(ctx, OTPChallengePrefix+prevChallengeID).Err(); err != nil {
			return fmt.Errorf("InvalidatePreviousChallenges del: %w", err)
		}
	}
	return nil
}

// verifyOTPLua is the atomic Lua script for OTP attempt counting and retrieval.
// Returns: {status, hash_val, phone, return_url}
// status: 0=not found, -1=max attempts exceeded, 1=proceed
const verifyOTPLua = `
local key = KEYS[1]
local max_attempts = tonumber(ARGV[1])

local exists = redis.call('EXISTS', key)
if exists == 0 then
    return {0, '', '', ''}
end

local attempts = tonumber(redis.call('HGET', key, 'attempts') or '0')
if attempts >= max_attempts then
    redis.call('DEL', key)
    return {-1, '', '', ''}
end

redis.call('HINCRBY', key, 'attempts', 1)

local hash_val = redis.call('HGET', key, 'hash') or ''
local phone = redis.call('HGET', key, 'phone') or ''
local return_url = redis.call('HGET', key, 'return_url') or ''

return {1, hash_val, phone, return_url}
`

// VerifyOTPChallengeResult is the result of an OTP challenge verification.
type VerifyOTPChallengeResult struct {
	Found     bool
	Exhausted bool
	HashVal   string
	Phone     string
	ReturnURL string
}

// VerifyOTPChallengeAtomic atomically checks, increments, and returns OTP data.
func VerifyOTPChallengeAtomic(ctx context.Context, rdb *redis.Client, cfg *config.Settings, challengeID string) (*VerifyOTPChallengeResult, error) {
	key := OTPChallengePrefix + challengeID
	res, err := rdb.Eval(ctx, verifyOTPLua, []string{key}, fmt.Sprintf("%d", cfg.OTPMaxVerifyAttempts)).Result()
	if err != nil {
		return nil, fmt.Errorf("VerifyOTPChallengeAtomic: %w", err)
	}

	vals, ok := res.([]any)
	if !ok || len(vals) < 4 {
		return nil, fmt.Errorf("VerifyOTPChallengeAtomic: unexpected Lua result type")
	}

	status := toInt64(vals[0])
	result := &VerifyOTPChallengeResult{
		HashVal:   toString(vals[1]),
		Phone:     toString(vals[2]),
		ReturnURL: toString(vals[3]),
	}

	switch status {
	case 0:
		return result, nil // not found
	case -1:
		result.Exhausted = true
		return result, nil
	default:
		result.Found = true
		return result, nil
	}
}

// DeleteOTPChallenge removes an OTP challenge after successful verification.
func DeleteOTPChallenge(ctx context.Context, rdb *redis.Client, challengeID string) error {
	return rdb.Del(ctx, OTPChallengePrefix+challengeID).Err()
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	}
	return 0
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	}
	return fmt.Sprintf("%v", v)
}
