// Package auth provides JWT token creation/validation and refresh token lifecycle.
//
// Access tokens: Short-lived JWTs (HS256) with minimal claims.
// Refresh tokens: Opaque secrets stored as SHA-256 hashes in Redis,
// with rotation and reuse detection via token families.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"flickey/go-backend/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// AccessTokenClaims are the minimal claims in a Flickey access token.
type AccessTokenClaims struct {
	Type string `json:"type"`
	jwt.RegisteredClaims
}

// CreateAccessToken issues a signed HS256 JWT with the minimum required claims.
func CreateAccessToken(cfg *config.Settings, userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := AccessTokenClaims{
		Type: JWTTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.JWTAccessTTLSeconds) * time.Second)),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}
	return signed, nil
}

// DecodeAccessToken validates a JWT and returns the claims.
// Explicitly restricts to HS256 to prevent algorithm confusion attacks.
func DecodeAccessToken(cfg *config.Settings, tokenStr string) (*AccessTokenClaims, error) {
	var claims AccessTokenClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != JWTTypeAccess {
		return nil, ErrInvalidTokenType
	}
	return &claims, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Refresh Token Lifecycle
// ─────────────────────────────────────────────────────────────────────────────

// GenerateSecureToken creates a 256-bit cryptographically secure URL-safe token.
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating secure token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// hashRefreshToken computes an unsalted SHA-256 hash for O(1) Redis key lookup.
// Security comes from the 256-bit entropy of the raw token itself.
func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// CreateRefreshToken generates a new refresh token and stores its hash in Redis.
// Returns the raw token (to be set as HttpOnly cookie).
func CreateRefreshToken(ctx context.Context, rdb *redis.Client, cfg *config.Settings, userID uuid.UUID, familyID, sessionID string) (string, error) {
	if familyID == "" {
		familyID = uuid.New().String()
	}
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	raw, err := GenerateSecureToken()
	if err != nil {
		return "", err
	}
	lookupHash := hashRefreshToken(raw)

	key := RefreshTokenPrefix + lookupHash
	userKey := RefreshUserPrefix + userID.String()
	ttl := time.Duration(cfg.JWTRefreshTTLSeconds) * time.Second

	pipe := rdb.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		RTFieldUserID:    userID.String(),
		RTFieldSessionID: sessionID,
		RTFieldFamilyID:  familyID,
		RTFieldRotated:   "0",
		RTFieldCreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	pipe.Expire(ctx, key, ttl)
	pipe.SAdd(ctx, userKey, lookupHash)
	pipe.Expire(ctx, userKey, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("CreateRefreshToken: %w", err)
	}

	return raw, nil
}

// rotateRefreshLua atomically marks a token as rotated and returns its metadata.
// Returns {status, user_id, family_id, session_id}
// status: 0=not found, -1=already rotated (reuse), 1=success
const rotateRefreshLua = `
local key = KEYS[1]
local rotated_ttl = tonumber(ARGV[1])

local exists = redis.call('EXISTS', key)
if exists == 0 then
    return {0, '', '', ''}
end

local is_rotated = redis.call('HGET', key, 'rotated') or '0'
local user_id = redis.call('HGET', key, 'user_id') or ''
local family_id = redis.call('HGET', key, 'family_id') or ''
local session_id = redis.call('HGET', key, 'session_id') or ''

if is_rotated == '1' then
    return {-1, user_id, family_id, session_id}
end

redis.call('HSET', key, 'rotated', '1')
redis.call('EXPIRE', key, rotated_ttl)

return {1, user_id, family_id, session_id}
`

// RotateRefreshToken atomically rotates a refresh token.
// On reuse detection, the entire token family is revoked.
// Returns (newAccessToken, newRawRefreshToken, userID).
func RotateRefreshToken(ctx context.Context, rdb *redis.Client, cfg *config.Settings, rawToken string) (string, string, uuid.UUID, error) {
	lookupHash := hashRefreshToken(rawToken)
	key := RefreshTokenPrefix + lookupHash
	rotatedTTL := min(cfg.JWTRefreshTTLSeconds, 86400)

	res, err := rdb.Eval(ctx, rotateRefreshLua, []string{key}, fmt.Sprintf("%d", rotatedTTL)).Result()
	if err != nil {
		return "", "", uuid.Nil, fmt.Errorf("RotateRefreshToken eval: %w", err)
	}

	vals, ok := res.([]any)
	if !ok || len(vals) < 4 {
		return "", "", uuid.Nil, fmt.Errorf("RotateRefreshToken: unexpected result")
	}

	status := toInt64(vals[0])
	userIDStr := toString(vals[1])
	familyID := toString(vals[2])
	sessionID := toString(vals[3])

	if status == 0 {
		return "", "", uuid.Nil, ErrInvalidRefreshToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		_ = rdb.Del(ctx, key)
		return "", "", uuid.Nil, ErrInvalidRefreshToken
	}

	if status == -1 {
		// Reuse detected — revoke entire family.
		slog.WarnContext(ctx, "refresh_token_reuse_detected",
			"user_id", userID,
			"family_id", familyID,
		)
		_ = revokeTokenFamily(ctx, rdb, userID, familyID)
		return "", "", uuid.Nil, ErrRefreshTokenReuse
	}

	// Issue new tokens in the same family.
	newRaw, err := CreateRefreshToken(ctx, rdb, cfg, userID, familyID, sessionID)
	if err != nil {
		return "", "", uuid.Nil, err
	}

	newAccess, err := CreateAccessToken(cfg, userID)
	if err != nil {
		return "", "", uuid.Nil, err
	}

	slog.InfoContext(ctx, "refresh_token_rotated", "user_id", userID, "session_id", sessionID)
	return newAccess, newRaw, userID, nil
}

// RevokeRefreshToken revokes a single refresh token (logout).
func RevokeRefreshToken(ctx context.Context, rdb *redis.Client, rawToken string) error {
	lookupHash := hashRefreshToken(rawToken)
	key := RefreshTokenPrefix + lookupHash

	data, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("RevokeRefreshToken hgetall: %w", err)
	}
	if len(data) == 0 {
		return nil // Already gone.
	}

	userID := data[RTFieldUserID]
	userKey := RefreshUserPrefix + userID
	pipe := rdb.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, userKey, lookupHash)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("RevokeRefreshToken pipe: %w", err)
	}
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user (logout-all).
func RevokeAllUserTokens(ctx context.Context, rdb *redis.Client, userID uuid.UUID) error {
	userKey := RefreshUserPrefix + userID.String()
	hashes, err := rdb.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("RevokeAllUserTokens smembers: %w", err)
	}
	if len(hashes) == 0 {
		return nil
	}

	keysToDelete := make([]string, len(hashes)+1)
	for i, h := range hashes {
		keysToDelete[i] = RefreshTokenPrefix + h
	}
	keysToDelete[len(hashes)] = userKey

	if err := rdb.Del(ctx, keysToDelete...).Err(); err != nil {
		return fmt.Errorf("RevokeAllUserTokens del: %w", err)
	}
	return nil
}

// revokeTokenFamily revokes all tokens belonging to a specific token family.
func revokeTokenFamily(ctx context.Context, rdb *redis.Client, userID uuid.UUID, familyID string) error {
	userKey := RefreshUserPrefix + userID.String()
	hashes, err := rdb.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("revokeTokenFamily smembers: %w", err)
	}
	if len(hashes) == 0 {
		return nil
	}

	var toDelete []string
	var toRemove []any
	for _, h := range hashes {
		tokenKey := RefreshTokenPrefix + h
		fid, err := rdb.HGet(ctx, tokenKey, RTFieldFamilyID).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			continue
		}
		if fid == familyID {
			toDelete = append(toDelete, tokenKey)
			toRemove = append(toRemove, h)
		}
	}

	if len(toDelete) == 0 {
		return nil
	}

	pipe := rdb.Pipeline()
	pipe.Del(ctx, toDelete...)
	pipe.SRem(ctx, userKey, toRemove...)
	_, err = pipe.Exec(ctx)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// Sentinel errors
// ─────────────────────────────────────────────────────────────────────────────

var (
	ErrTokenExpired        = errors.New("TOKEN_EXPIRED")
	ErrInvalidToken        = errors.New("INVALID_TOKEN")
	ErrInvalidTokenType    = errors.New("INVALID_TOKEN_TYPE")
	ErrMissingToken        = errors.New("MISSING_TOKEN")
	ErrInvalidRefreshToken = errors.New("INVALID_REFRESH_TOKEN")
	ErrRefreshTokenReuse   = errors.New("REFRESH_TOKEN_REUSE")
)
