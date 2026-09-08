// Package config loads application settings from environment variables or .env file.
// All field names match the existing FastAPI settings exactly to preserve behavior.
package config

import (
	"fmt"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

// Settings holds all application configuration.
type Settings struct {
	// Server
	Environment string `env:"ENVIRONMENT" env-default:"dev"`
	Port        string `env:"PORT" env-default:"8080"`

	// Database
	DatabaseURL string `env:"DATABASE_URL" env-default:"postgresql://rental:rental_dev@localhost:5432/flickey"`

	// Redis
	RedisURL string `env:"REDIS_URL" env-default:"redis://localhost:6379/0"`

	// Geocoder (Photon OSM)
	GeocoderURL string `env:"GEOCODER_URL" env-default:"http://localhost:2322"`

	// TileServer (TileServer GL)
	TileServerURL string `env:"TILESERVER_URL" env-default:"http://localhost:8081"`

	// JWT
	JWTSecret            string `env:"JWT_SECRET"`
	JWTAlgorithm         string `env:"JWT_ALGORITHM" env-default:"HS256"`
	JWTAccessTTLSeconds  int    `env:"JWT_ACCESS_TTL_SECONDS" env-default:"900"`
	JWTRefreshTTLSeconds int    `env:"JWT_REFRESH_TTL_SECONDS" env-default:"2592000"`

	// OTP
	OTPTTLSeconds                 int `env:"OTP_TTL_SECONDS" env-default:"300"`
	OTPLength                     int `env:"OTP_LENGTH" env-default:"6"`
	OTPMaxVerifyAttempts          int `env:"OTP_MAX_VERIFY_ATTEMPTS" env-default:"5"`
	OTPMaxRequestsPerPhonePerHour int `env:"OTP_MAX_REQUESTS_PER_PHONE_PER_HOUR" env-default:"5"`
	OTPMaxRequestsPerIPPerHour    int `env:"OTP_MAX_REQUESTS_PER_IP_PER_HOUR" env-default:"20"`
	OTPResendCooldownSeconds      int `env:"OTP_RESEND_COOLDOWN_SECONDS" env-default:"60"`

	// Security
	AllowedRedirectHostsRaw string `env:"ALLOWED_REDIRECT_HOSTS" env-default:""`
	CookieDomain            string `env:"COOKIE_DOMAIN" env-default:""`
	CookieSecure            bool   `env:"COOKIE_SECURE" env-default:"true"`
	TrustProxyHeaders       bool   `env:"TRUST_PROXY_HEADERS" env-default:"false"`

	// Storage (Selectel S3-compatible)
	S3EndpointURL     string `env:"S3_ENDPOINT_URL" env-default:"s3.gis-1.storage.selcloud.ru"`
	S3AccessKey       string `env:"S3_ACCESS_KEY" env-default:""`
	S3SecretKey       string `env:"S3_SECRET_KEY" env-default:""`
	S3Bucket          string `env:"S3_BUCKET" env-default:"flickey"`
	S3Region          string `env:"S3_REGION" env-default:"gis-1"`
	S3PublicBaseURL   string `env:"S3_PUBLIC_BASE_URL" env-default:""`
	S3PresignedURLTTL int    `env:"S3_PRESIGNED_URL_TTL_SECONDS" env-default:"300"`

	// Media constraints
	MediaMaxFileSizeBytes      int `env:"MEDIA_MAX_FILE_SIZE_BYTES" env-default:"15728640"`       // 15 MB (images)
	MediaVideoMaxFileSizeBytes int `env:"MEDIA_VIDEO_MAX_FILE_SIZE_BYTES" env-default:"104857600"` // 100 MB (videos / documents)
	MediaMinCount              int `env:"MEDIA_MIN_COUNT" env-default:"5"`
	MediaMaxCount              int `env:"MEDIA_MAX_COUNT" env-default:"25"`

	// Derived (parsed from raw)
	AllowedRedirectHosts []string `env:"-"`
}

// Load reads settings from .env file (if present) then environment variables.
func Load() (*Settings, error) {
	var cfg Settings

	// Try .env file in the current directory; ignore if missing.
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		// Fall back to environment-only if .env not found.
		if err2 := cleanenv.ReadEnv(&cfg); err2 != nil {
			return nil, fmt.Errorf("loading config: %w", err2)
		}
	}

	// Validate required fields.
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len([]byte(cfg.JWTSecret)) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes for secure HS256 signing")
	}

	allowed := []string{"HS256", "HS384", "HS512"}
	valid := false
	for _, a := range allowed {
		if cfg.JWTAlgorithm == a {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("JWT_ALGORITHM must be one of %v", allowed)
	}

	// Parse comma-separated redirect hosts.
	if cfg.AllowedRedirectHostsRaw != "" {
		for _, h := range strings.Split(cfg.AllowedRedirectHostsRaw, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				cfg.AllowedRedirectHosts = append(cfg.AllowedRedirectHosts, h)
			}
		}
	}

	return &cfg, nil
}

// IsDev returns true when running in development mode.
func (s *Settings) IsDev() bool {
	return s.Environment == "dev"
}
