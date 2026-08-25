// Package auth provides security utilities for phone handling, token generation,
// IP extraction, and URL validation.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// NormalizePhone parses and validates a phone number, returning E.164 format.
// Returns error if the phone number is invalid.
func NormalizePhone(phone string) (string, error) {
	parsed, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return "", fmt.Errorf("invalid phone number format: %w", err)
	}
	if !phonenumbers.IsValidNumber(parsed) {
		return "", fmt.Errorf("invalid phone number")
	}
	return phonenumbers.Format(parsed, phonenumbers.E164), nil
}

// MaskPhone returns a partially masked phone number for logging.
// Example: +375291234567 → +375XXXXX4567
func MaskPhone(phone string) string {
	if len(phone) <= 6 {
		return "****"
	}
	return phone[:3] + strings.Repeat("X", len(phone)-6) + phone[len(phone)-4:]
}

// GenerateURLSafeToken generates a URL-safe random token of the given byte length.
func GenerateURLSafeToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating URL-safe token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ValidateReturnURL checks that a return_url is safe to redirect to.
// Only allows URLs on the configured list of trusted hosts.
func ValidateReturnURL(rawURL string, allowedHosts []string) bool {
	if rawURL == "" {
		return true
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	for _, allowed := range allowedHosts {
		if strings.EqualFold(host, allowed) {
			return true
		}
	}
	return false
}

// ExtractClientIP extracts the real client IP from a request.
// When trustProxyHeaders is true, X-Forwarded-For is preferred.
func ExtractClientIP(remoteAddr string, xForwardedFor string, trustProxyHeaders bool) string {
	if trustProxyHeaders && xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
