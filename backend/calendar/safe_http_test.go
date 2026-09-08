package calendar

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestIsBlockedIP(t *testing.T) {
	blockedCases := []string{
		"127.0.0.1",
		"127.0.1.1",
		"10.0.0.1",
		"10.255.255.255",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"192.168.100.200",
		"169.254.169.254", // AWS/GCP metadata
		"169.254.1.1",     // Link-local
		"100.64.0.1",      // Carrier grade NAT
		"0.0.0.0",
		"::1",     // IPv6 loopback
		"fc00::1", // IPv6 ULA
		"fe80::1", // IPv6 Link-local
	}

	for _, ipStr := range blockedCases {
		ip := net.ParseIP(ipStr)
		if !IsBlockedIP(ip) {
			t.Errorf("expected IP %s to be blocked, but was allowed", ipStr)
		}
	}

	allowedCases := []string{
		"8.8.8.8",
		"1.1.1.1",
		"93.184.216.34", // example.com
		"140.82.121.3",  // github.com
	}

	for _, ipStr := range allowedCases {
		ip := net.ParseIP(ipStr)
		if IsBlockedIP(ip) {
			t.Errorf("expected IP %s to be allowed, but was blocked", ipStr)
		}
	}
}

func TestValidateURL(t *testing.T) {
	invalidURLs := []string{
		"ftp://example.com/cal.ics",
		"file:///etc/passwd",
		"gopher://example.com",
		"http://127.0.0.1/cal.ics",
		"http://192.168.1.50/cal.ics",
		"http://169.254.169.254/latest/meta-data/",
	}

	for _, u := range invalidURLs {
		if _, err := ValidateURL(u); err == nil {
			t.Errorf("expected URL %s to be rejected, but it was accepted", u)
		}
	}

	validURLs := []string{
		"https://calendar.google.com/calendar/ical/xyz/basic.ics",
		"https://airbnb.com/calendar/ical/12345.ics",
		"http://example.com/calendar.ics",
	}

	for _, u := range validURLs {
		if _, err := ValidateURL(u); err != nil {
			t.Errorf("expected URL %s to be valid, but got error: %v", u, err)
		}
	}
}

func TestSafeHTTPClient_BlocksLocalhost(t *testing.T) {
	client := NewSafeHTTPClient(2 * time.Second)
	ctx := context.Background()

	// Attempting to fetch localhost should be blocked by dialer
	_, err := FetchSafeURL(ctx, client, "http://127.0.0.1:8000/api/v1/listings")
	if err == nil {
		t.Fatal("expected request to 127.0.0.1 to be blocked, but got no error")
	}
}
