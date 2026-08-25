// Package storage provides the StorageProvider interface for object storage abstraction.
package storage

import (
	"context"
	"io"
)

// StorageProvider is the interface for asynchronous object storage operations.
// Implementations: S3Storage (Selectel-compatible AWS S3), LocalStorage (dev).
type StorageProvider interface {
	// GeneratePresignedUploadURL creates a presigned PUT URL for client-side direct upload.
	GeneratePresignedUploadURL(ctx context.Context, fileKey, contentType string, expiresIn int) (string, error)

	// ObjectExists checks if an object exists in storage.
	ObjectExists(ctx context.Context, fileKey string) (bool, error)

	// GetObjectMetadata fetches object metadata (ContentLength, ContentType, ETag).
	GetObjectMetadata(ctx context.Context, fileKey string) (map[string]any, error)

	// GetObjectBytes reads the first maxBytes bytes of an object for content inspection.
	GetObjectBytes(ctx context.Context, fileKey string, maxBytes int) ([]byte, error)

	// Delete removes an object from storage.
	Delete(ctx context.Context, fileKey string) error

	// PutObject stores raw bytes (used in dev uploads).
	PutObject(ctx context.Context, fileKey string, data []byte, contentType string) error
}

// ReadAll is a helper for reading from an io.ReadCloser with a size limit.
func ReadAll(r io.ReadCloser, maxBytes int) ([]byte, error) {
	defer r.Close()
	buf := make([]byte, maxBytes)
	n, err := io.ReadFull(r, buf)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		return buf[:n], nil
	}
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
