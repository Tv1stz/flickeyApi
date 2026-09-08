// Package media provides the media upload service with S3 and magic byte validation.
package media

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"flickey/go-backend/config"
	"flickey/go-backend/db"
	"flickey/go-backend/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaService handles the presigned upload workflow and S3 object validation.
type MediaService struct {
	DB      *gorm.DB
	Storage storage.StorageProvider
	Cfg     *config.Settings
}

// NewMediaService creates a MediaService.
func NewMediaService(database *gorm.DB, store storage.StorageProvider, cfg *config.Settings) *MediaService {
	return &MediaService{DB: database, Storage: store, Cfg: cfg}
}

// AllowedContentTypes maps MIME type to file extension.
var AllowedContentTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
	"video/webm":      ".webm",
	"application/pdf": ".pdf",
}

func (s *MediaService) getMaxFileSize(contentType string) int64 {
	if strings.HasPrefix(contentType, "video/") || contentType == "application/pdf" {
		if s.Cfg != nil && s.Cfg.MediaVideoMaxFileSizeBytes > 0 {
			return int64(s.Cfg.MediaVideoMaxFileSizeBytes)
		}
		return 104857600 // 100 MB
	}
	if s.Cfg != nil && s.Cfg.MediaMaxFileSizeBytes > 0 {
		return int64(s.Cfg.MediaMaxFileSizeBytes)
	}
	return 15728640 // 15 MB
}

// CreatePresignedUpload validates the request, creates a media record, and returns a presigned URL.
func (s *MediaService) CreatePresignedUpload(ctx context.Context, hostID uuid.UUID, contentType string, fileSizeBytes int64) (*PresignResponse, error) {
	if _, ok := AllowedContentTypes[contentType]; !ok {
		return nil, &ValidationError{
			Code:    "INVALID_CONTENT_TYPE",
			Message: fmt.Sprintf("Content type '%s' is not allowed.", contentType),
		}
	}

	maxBytes := s.getMaxFileSize(contentType)
	if fileSizeBytes <= 0 || fileSizeBytes > maxBytes {
		return nil, &ValidationError{
			Code:    "INVALID_FILE_SIZE",
			Message: fmt.Sprintf("File size must be between 1 and %d bytes.", maxBytes),
		}
	}

	mediaID := uuid.New()
	ext := AllowedContentTypes[contentType]
	fileKey := fmt.Sprintf("raw/%s/%s%s", hostID, mediaID, ext)

	m := &db.Media{
		ID:            mediaID,
		HostID:        hostID,
		FileKey:       fileKey,
		Status:        db.MediaStatusPending,
		ContentType:   contentType,
		FileSizeBytes: fileSizeBytes,
	}

	if err := db.CreateMedia(ctx, s.DB, m); err != nil {
		return nil, fmt.Errorf("CreatePresignedUpload create record: %w", err)
	}

	uploadURL, err := s.Storage.GeneratePresignedUploadURL(ctx, fileKey, contentType, s.Cfg.S3PresignedURLTTL)
	if err != nil {
		return nil, fmt.Errorf("CreatePresignedUpload presign: %w", err)
	}

	return &PresignResponse{
		MediaID:   mediaID,
		UploadURL: uploadURL,
		FileKey:   fileKey,
		ExpiresIn: s.Cfg.S3PresignedURLTTL,
	}, nil
}

// CompleteUpload verifies that the object was successfully uploaded and marks it as uploaded.
func (s *MediaService) CompleteUpload(ctx context.Context, hostID, mediaID uuid.UUID) (*CompleteResponse, error) {
	m, err := db.FindMediaByIDAndHost(ctx, s.DB, mediaID, hostID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, &NotOwnedError{Code: "MEDIA_NOT_OWNED", Message: "Media not found or not owned by you."}
	}

	// Idempotency: already uploaded is a success.
	if m.Status == db.MediaStatusUploaded {
		return &CompleteResponse{MediaID: m.ID, Status: string(m.Status), FileKey: m.FileKey}, nil
	}

	if m.Status != db.MediaStatusPending {
		return nil, &ValidationError{
			Code:    "INVALID_MEDIA_STATUS",
			Message: fmt.Sprintf("Media is in '%s' status and cannot be completed.", m.Status),
		}
	}

	// 1. Verify object exists and size via HEAD.
	meta, err := s.Storage.GetObjectMetadata(ctx, m.FileKey)
	if err != nil {
		return nil, fmt.Errorf("CompleteUpload metadata: %w", err)
	}
	if meta == nil {
		return nil, &StorageError{Code: "OBJECT_NOT_FOUND", Message: "Uploaded file not found in storage. Upload may have failed."}
	}

	var contentLength int64
	if cl, ok := meta["content_length"]; ok {
		switch v := cl.(type) {
		case int64:
			contentLength = v
		case int32:
			contentLength = int64(v)
		}
	}
	maxBytes := s.getMaxFileSize(m.ContentType)
	if contentLength <= 0 || contentLength > maxBytes {
		return nil, &InvalidMediaError{
			Code:    "INVALID_FILE_SIZE",
			Message: fmt.Sprintf("Uploaded object size (%d bytes) is invalid or exceeds maximum allowed.", contentLength),
		}
	}

	// 2. Inspect magic bytes.
	header, err := s.Storage.GetObjectBytes(ctx, m.FileKey, 32)
	if err != nil {
		return nil, fmt.Errorf("CompleteUpload get bytes: %w", err)
	}
	if !IsValidMagicBytes(header, m.ContentType) {
		return nil, &InvalidMediaError{
			Code:    "INVALID_FILE_SIGNATURE",
			Message: "Uploaded file content does not match the declared file signature.",
		}
	}

	// Update status and actual size.
	if err := db.UpdateMediaStatus(ctx, s.DB, m.ID, db.MediaStatusUploaded, contentLength); err != nil {
		return nil, fmt.Errorf("CompleteUpload update status: %w", err)
	}

	return &CompleteResponse{
		MediaID: m.ID,
		Status:  string(db.MediaStatusUploaded),
		FileKey: m.FileKey,
	}, nil
}

// IsValidMagicBytes verifies that file data matches expected binary signature.
func IsValidMagicBytes(data []byte, contentType string) bool {
	if len(data) == 0 {
		return false
	}

	switch strings.ToLower(contentType) {
	case "image/jpeg":
		return bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF})
	case "image/png":
		return bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	case "image/webp":
		return len(data) >= 12 &&
			bytes.HasPrefix(data, []byte("RIFF")) &&
			bytes.Equal(data[8:12], []byte("WEBP"))
	case "video/mp4":
		return len(data) >= 8 && bytes.Equal(data[4:8], []byte("ftyp"))
	case "video/quicktime":
		if len(data) >= 8 {
			tag := string(data[4:8])
			return tag == "ftyp" || tag == "moov" || tag == "mdat" || tag == "wide"
		}
		return false
	case "video/webm":
		return bytes.HasPrefix(data, []byte{0x1A, 0x45, 0xDF, 0xA3})
	case "application/pdf":
		return bytes.HasPrefix(data, []byte("%PDF-"))
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// PresignRequest is the request for POST /media/presign.
type PresignRequest struct {
	ContentType   string `json:"content_type" binding:"required"`
	FileSizeBytes int64  `json:"file_size_bytes" binding:"required,min=1"`
}

// PresignResponse is the response from POST /media/presign.
type PresignResponse struct {
	MediaID   uuid.UUID `json:"media_id"`
	UploadURL string    `json:"upload_url"`
	FileKey   string    `json:"file_key"`
	ExpiresIn int       `json:"expires_in"`
}

// CompleteResponse is the response from POST /media/{id}/complete.
type CompleteResponse struct {
	MediaID uuid.UUID `json:"media_id"`
	Status  string    `json:"status"`
	FileKey string    `json:"file_key"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Domain Errors
// ─────────────────────────────────────────────────────────────────────────────

// ValidationError is returned for invalid input (400).
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// NotOwnedError is returned when media is not owned by the requesting user (403/404).
type NotOwnedError struct {
	Code    string
	Message string
}

func (e *NotOwnedError) Error() string { return e.Message }

// StorageError is returned when the S3 object is missing (422).
type StorageError struct {
	Code    string
	Message string
}

func (e *StorageError) Error() string { return e.Message }

// InvalidMediaError is returned for magic byte failures (422).
type InvalidMediaError struct {
	Code    string
	Message string
}

func (e *InvalidMediaError) Error() string { return e.Message }
