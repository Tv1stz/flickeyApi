package calendar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrBlockedIP      = errors.New("safe_http: destination IP is blocked (private or local network)")
	ErrInvalidScheme  = errors.New("safe_http: scheme must be http or https")
	ErrTooManyHops    = errors.New("safe_http: maximum redirects exceeded")
	ErrResponseTooBig = errors.New("safe_http: response body exceeded maximum allowed size (2MB)")
)

// MaxICalResponseBytes is the maximum bytes read from an external iCal feed (2 MB).
const MaxICalResponseBytes = 2 * 1024 * 1024

// IsBlockedIP returns true if an IP address belongs to loopback, private, link-local,
// unspecified, or well-known cloud metadata address ranges.
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// 1. Check standard Go net.IP classification
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	// 2. Specific IPv4 checks (e.g. 169.254.169.254 metadata service, carrier-grade NAT 100.64.0.0/10)
	if ipv4 := ip.To4(); ipv4 != nil {
		// Cloud metadata IP (AWS, GCP, Azure, OpenStack)
		if ipv4[0] == 169 && ipv4[1] == 254 {
			return true
		}
		// Carrier-grade NAT (100.64.0.0/10)
		if ipv4[0] == 100 && (ipv4[1]&0xc0) == 64 {
			return true
		}
		// Broadcast (255.255.255.255)
		if ipv4[0] == 255 && ipv4[1] == 255 && ipv4[2] == 255 && ipv4[3] == 255 {
			return true
		}
	} else {
		// IPv6 unique local (fc00::/7) or IPv4-mapped IPv6
		if strings.HasPrefix(ip.String(), "fc") || strings.HasPrefix(ip.String(), "fd") {
			return true
		}
	}

	return false
}

// ValidateURL performs initial scheme and hostname sanity checks.
func ValidateURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, ErrInvalidScheme
	}

	host := u.Hostname()
	if host == "" {
		return nil, errors.New("safe_http: empty host")
	}

	// Check literal IP if provided as host
	if ip := net.ParseIP(host); ip != nil {
		if IsBlockedIP(ip) {
			return nil, ErrBlockedIP
		}
	}

	return u, nil
}

// NewSafeHTTPClient creates an http.Client equipped with SSRF mitigation,
// DNS rebinding defense, strict timeouts, and redirect validation.
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 15 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 nil, // Do not inherit proxy for external calendar sync
		MaxIdleConns:          50,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			// Resolve host to IPs
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, fmt.Errorf("resolving %s: %w", host, err)
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no IP found for %s", host)
			}

			// Find first allowed public IP
			var allowedIP net.IP
			for _, ip := range ips {
				if !IsBlockedIP(ip) {
					allowedIP = ip
					break
				}
			}

			if allowedIP == nil {
				return nil, fmt.Errorf("%w: %s resolved to blocked addresses", ErrBlockedIP, host)
			}

			// Dial directly to the validated IP to prevent DNS rebinding attacks
			targetAddr := net.JoinHostPort(allowedIP.String(), port)
			return dialer.DialContext(ctx, network, targetAddr)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return ErrTooManyHops
			}
			// Validate target redirect URL
			if _, err := ValidateURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
	}

	return client
}

// FetchSafeURL performs an HTTP GET with SSRF protection, custom User-Agent,
// and enforces a 2MB maximum response body size limit.
func FetchSafeURL(ctx context.Context, client *http.Client, rawURL string) ([]byte, error) {
	parsedURL, err := ValidateURL(rawURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Flickey-CalendarSync/1.0 (+https://flickey.by)")
	req.Header.Set("Accept", "text/calendar, text/plain, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching calendar feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("calendar feed returned HTTP %d", resp.StatusCode)
	}

	// Read limited body (2MB max)
	limitedReader := io.LimitReader(resp.Body, MaxICalResponseBytes+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if len(data) > MaxICalResponseBytes {
		return nil, ErrResponseTooBig
	}

	return data, nil
}
