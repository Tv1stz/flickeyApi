// Package auth provides Gin middleware for authentication and authorization.
package auth

import (
	"strings"

	"flickey/go-backend/config"
	"flickey/go-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ctxUserKey   = "current_user"
	ctxUserIDKey = "current_user_id"
)

// RequireAuth validates the Bearer token and sets the user in Gin context.
// Returns 401 if the token is missing, invalid, or the user does not exist/is banned.
func RequireAuth(cfg *config.Settings, database *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			respondUnauthorized(c, "MISSING_TOKEN", "Authentication token required.")
			return
		}

		claims, err := DecodeAccessToken(cfg, tokenStr)
		if err != nil {
			if err == ErrTokenExpired {
				respondUnauthorized(c, "TOKEN_EXPIRED", "Access token has expired.")
			} else {
				respondUnauthorized(c, "INVALID_TOKEN", "Invalid authentication token.")
			}
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			respondUnauthorized(c, "INVALID_TOKEN", "Invalid token subject.")
			return
		}

		user, err := db.FindUserByIDCtx(c.Request.Context(), database, userID)
		if err != nil || user == nil {
			respondUnauthorized(c, "USER_NOT_FOUND", "Authenticated user not found.")
			return
		}

		if user.Status == db.UserStatusSuspended || user.Status == db.UserStatusBanned {
			respondUnauthorized(c, "ACCOUNT_SUSPENDED", "Account is suspended or banned.")
			return
		}

		c.Set(ctxUserKey, user)
		c.Set(ctxUserIDKey, userID)
		c.Next()
	}
}

// OptionalAuth parses the Bearer token if present and populates the user in context.
// Requests without token or with invalid tokens proceed without error.
func OptionalAuth(cfg *config.Settings, database *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}

		claims, err := DecodeAccessToken(cfg, tokenStr)
		if err != nil {
			c.Next()
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.Next()
			return
		}

		user, err := db.FindUserByIDCtx(c.Request.Context(), database, userID)
		if err == nil && user != nil && user.Status != db.UserStatusSuspended && user.Status != db.UserStatusBanned {
			c.Set(ctxUserKey, user)
			c.Set(ctxUserIDKey, userID)
		}
		c.Next()
	}
}

// RequireActiveUser rejects users with pending_profile status.
// Must be chained after RequireAuth.
func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			respondUnauthorized(c, "MISSING_USER", "User not in context.")
			return
		}
		if user.Status == db.UserStatusPendingProfile {
			respondForbidden(c, "PROFILE_INCOMPLETE",
				"Please complete your profile before performing this action.")
			return
		}
		c.Next()
	}
}

// RequireRole enforces that the current user has one of the specified roles.
func RequireRole(roles ...db.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			respondUnauthorized(c, "MISSING_USER", "User not in context.")
			return
		}
		for _, r := range roles {
			if user.Role == r {
				c.Next()
				return
			}
		}
		respondForbidden(c, "INSUFFICIENT_ROLE", "You do not have permission to perform this action.")
	}
}

// GetCurrentUser retrieves the authenticated user from Gin context.
func GetCurrentUser(c *gin.Context) *db.User {
	if u, ok := c.Get(ctxUserKey); ok {
		if user, ok := u.(*db.User); ok {
			return user
		}
	}
	return nil
}

// GetCurrentUserID retrieves the authenticated user's UUID from Gin context.
func GetCurrentUserID(c *gin.Context) uuid.UUID {
	if id, ok := c.Get(ctxUserIDKey); ok {
		if uid, ok := id.(uuid.UUID); ok {
			return uid
		}
	}
	return uuid.Nil
}

// extractBearerToken extracts the token from Authorization: Bearer <token> header.
func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func respondUnauthorized(c *gin.Context, code, message string) {
	c.Header("WWW-Authenticate", "Bearer")
	c.AbortWithStatusJSON(401, gin.H{"code": code, "message": message})
}

func respondForbidden(c *gin.Context, code, message string) {
	c.AbortWithStatusJSON(403, gin.H{"code": code, "message": message})
}
