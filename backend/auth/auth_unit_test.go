package auth_test

import (
	"testing"

	"flickey/go-backend/auth"
	"flickey/go-backend/config"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// OTP Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateOTP_Length(t *testing.T) {
	for _, length := range []int{4, 6, 8} {
		otp, err := auth.GenerateOTP(length)
		if err != nil {
			t.Fatalf("GenerateOTP(%d): %v", length, err)
		}
		if len(otp) != length {
			t.Errorf("expected OTP length %d, got %d (otp=%q)", length, len(otp), otp)
		}
	}
}

func TestGenerateOTP_OnlyDigits(t *testing.T) {
	for i := 0; i < 100; i++ {
		otp, err := auth.GenerateOTP(6)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range otp {
			if c < '0' || c > '9' {
				t.Errorf("OTP contains non-digit: %q", otp)
			}
		}
	}
}

func TestHashValue_VerifyHash_Success(t *testing.T) {
	for _, val := range []string{"123456", "000000", "999999", "abc", ""} {
		hashed, err := auth.HashValue(val)
		if err != nil {
			t.Fatalf("HashValue(%q): %v", val, err)
		}
		if !auth.VerifyHash(val, hashed) {
			t.Errorf("VerifyHash(%q) returned false, expected true", val)
		}
	}
}

func TestHashValue_DifferentSalts(t *testing.T) {
	// Same value must produce different hashes (salted).
	h1, _ := auth.HashValue("123456")
	h2, _ := auth.HashValue("123456")
	if h1 == h2 {
		t.Error("same value produced identical hashes — salt is not random")
	}
}

func TestVerifyHash_WrongValue(t *testing.T) {
	hashed, _ := auth.HashValue("123456")
	if auth.VerifyHash("654321", hashed) {
		t.Error("VerifyHash returned true for wrong value")
	}
}

func TestVerifyHash_Tampered(t *testing.T) {
	hashed, _ := auth.HashValue("123456")
	tampered := hashed[:len(hashed)-4] + "0000"
	if auth.VerifyHash("123456", tampered) {
		t.Error("VerifyHash returned true for tampered hash")
	}
}

func TestVerifyHash_ConstantTime(t *testing.T) {
	// Verify function doesn't panic on malformed hash.
	if auth.VerifyHash("value", "no-colon") {
		t.Error("expected false for malformed hash")
	}
	if auth.VerifyHash("value", "") {
		t.Error("expected false for empty hash")
	}
}

func TestHashPhoneForKey(t *testing.T) {
	// Same phone always produces same key.
	k1 := auth.HashPhoneForKey("+375291234567")
	k2 := auth.HashPhoneForKey("+375291234567")
	if k1 != k2 {
		t.Error("HashPhoneForKey is not deterministic")
	}
	// Different phones produce different keys.
	k3 := auth.HashPhoneForKey("+375291234568")
	if k1 == k3 {
		t.Error("different phones produced the same key")
	}
	// Always 16 chars.
	if len(k1) != 16 {
		t.Errorf("expected 16 chars, got %d", len(k1))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// JWT Tests
// ─────────────────────────────────────────────────────────────────────────────

func testCfg() *config.Settings {
	return &config.Settings{
		JWTSecret:           "a-super-secret-key-that-is-32-bytes-long!!",
		JWTAlgorithm:        "HS256",
		JWTAccessTTLSeconds: 900,
	}
}

func TestCreateDecodeAccessToken_Success(t *testing.T) {
	cfg := testCfg()
	userID := uuid.New()

	token, err := auth.CreateAccessToken(cfg, userID)
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := auth.DecodeAccessToken(cfg, token)
	if err != nil {
		t.Fatalf("DecodeAccessToken: %v", err)
	}
	if claims.Subject != userID.String() {
		t.Errorf("expected subject %s, got %s", userID, claims.Subject)
	}
	if claims.Type != auth.JWTTypeAccess {
		t.Errorf("expected type %q, got %q", auth.JWTTypeAccess, claims.Type)
	}
}

func TestDecodeAccessToken_WrongSecret(t *testing.T) {
	cfg := testCfg()
	token, _ := auth.CreateAccessToken(cfg, uuid.New())

	wrongCfg := &config.Settings{
		JWTSecret:           "a-completely-different-secret-key-32bytes!",
		JWTAlgorithm:        "HS256",
		JWTAccessTTLSeconds: 900,
	}
	_, err := auth.DecodeAccessToken(wrongCfg, token)
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestDecodeAccessToken_Malformed(t *testing.T) {
	cfg := testCfg()
	_, err := auth.DecodeAccessToken(cfg, "not.a.jwt.token")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestDecodeAccessToken_EmptyString(t *testing.T) {
	cfg := testCfg()
	_, err := auth.DecodeAccessToken(cfg, "")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phone Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestNormalizePhone_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+375291234567", "+375291234567"},
		{"+79991234567", "+79991234567"},
		{"+12125551234", "+12125551234"},
	}
	for _, tt := range tests {
		got, err := auth.NormalizePhone(tt.input)
		if err != nil {
			t.Errorf("NormalizePhone(%q): %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("NormalizePhone(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizePhone_Invalid(t *testing.T) {
	invalids := []string{"12345", "not-a-phone", "", "00000000000"}
	for _, phone := range invalids {
		_, err := auth.NormalizePhone(phone)
		if err == nil {
			t.Errorf("NormalizePhone(%q): expected error, got nil", phone)
		}
	}
}

func TestMaskPhone(t *testing.T) {
	masked := auth.MaskPhone("+375291234567")
	if masked == "" {
		t.Error("expected non-empty masked phone")
	}
	if masked == "+375291234567" {
		t.Error("masked phone should not equal original")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// URL Validation Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestValidateReturnURL(t *testing.T) {
	allowed := []string{"example.com", "app.flickey.by"}

	tests := []struct {
		url   string
		valid bool
	}{
		{"", true},
		{"http://example.com/path", true},
		{"https://app.flickey.by/callback", true},
		{"http://evil.com/steal", false},
		{"javascript:alert(1)", false},
	}
	for _, tt := range tests {
		got := auth.ValidateReturnURL(tt.url, allowed)
		if got != tt.valid {
			t.Errorf("ValidateReturnURL(%q) = %v, want %v", tt.url, got, tt.valid)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// IP Extraction Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		remoteAddr    string
		xForwardedFor string
		trustProxy    bool
		expected      string
	}{
		{"1.2.3.4:5678", "", false, "1.2.3.4"},
		{"1.2.3.4:5678", "9.9.9.9", false, "1.2.3.4"},          // no trust
		{"1.2.3.4:5678", "9.9.9.9", true, "9.9.9.9"},           // trust
		{"1.2.3.4:5678", "9.9.9.9, 10.0.0.1", true, "9.9.9.9"}, // first from chain
	}
	for _, tt := range tests {
		got := auth.ExtractClientIP(tt.remoteAddr, tt.xForwardedFor, tt.trustProxy)
		if got != tt.expected {
			t.Errorf("ExtractClientIP(%q, %q, %v) = %q, want %q",
				tt.remoteAddr, tt.xForwardedFor, tt.trustProxy, got, tt.expected)
		}
	}
}
