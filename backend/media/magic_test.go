package media_test

import (
	"testing"

	"flickey/go-backend/media"
)

func TestIsValidMagicBytes_JPEG(t *testing.T) {
	// Valid JPEG: starts with FF D8 FF
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	if !media.IsValidMagicBytes(data, "image/jpeg") {
		t.Error("expected valid JPEG magic bytes")
	}
}

func TestIsValidMagicBytes_JPEG_Wrong(t *testing.T) {
	data := []byte{0xFF, 0xD8, 0xAA} // wrong 3rd byte
	if media.IsValidMagicBytes(data, "image/jpeg") {
		t.Error("expected invalid JPEG magic bytes")
	}
}

func TestIsValidMagicBytes_PNG(t *testing.T) {
	// Valid PNG: 89 50 4E 47 0D 0A 1A 0A
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	if !media.IsValidMagicBytes(data, "image/png") {
		t.Error("expected valid PNG magic bytes")
	}
}

func TestIsValidMagicBytes_PNG_Wrong(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x01} // last byte wrong
	if media.IsValidMagicBytes(data, "image/png") {
		t.Error("expected invalid PNG magic bytes")
	}
}

func TestIsValidMagicBytes_WebP_Valid(t *testing.T) {
	// RIFF????WEBP
	data := make([]byte, 12)
	copy(data[0:4], []byte("RIFF"))
	data[4] = 0x00
	data[5] = 0x00
	data[6] = 0x00
	data[7] = 0x00
	copy(data[8:12], []byte("WEBP"))
	if !media.IsValidMagicBytes(data, "image/webp") {
		t.Error("expected valid WebP magic bytes")
	}
}

func TestIsValidMagicBytes_WebP_Wrong(t *testing.T) {
	data := make([]byte, 12)
	copy(data[0:4], []byte("RIFF"))
	copy(data[8:12], []byte("WAVE")) // wrong 4CC
	if media.IsValidMagicBytes(data, "image/webp") {
		t.Error("expected invalid WebP (WAVE not WEBP)")
	}
}

func TestIsValidMagicBytes_WebP_TooShort(t *testing.T) {
	data := []byte("RIFF")
	if media.IsValidMagicBytes(data, "image/webp") {
		t.Error("expected false for too-short WebP data")
	}
}

func TestIsValidMagicBytes_Empty(t *testing.T) {
	for _, ct := range []string{"image/jpeg", "image/png", "image/webp"} {
		if media.IsValidMagicBytes([]byte{}, ct) {
			t.Errorf("expected false for empty data with content-type %s", ct)
		}
	}
}

func TestIsValidMagicBytes_UnknownType(t *testing.T) {
	data := []byte{0xFF, 0xD8, 0xFF}
	if media.IsValidMagicBytes(data, "application/octet-stream") {
		t.Error("expected false for unknown content type")
	}
}

func TestIsValidMagicBytes_ContentTypeMismatch(t *testing.T) {
	// JPEG magic bytes but declared as PNG.
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	if media.IsValidMagicBytes(jpegData, "image/png") {
		t.Error("expected false: JPEG bytes declared as PNG")
	}
}
