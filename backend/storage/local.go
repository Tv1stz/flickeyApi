// Package storage provides a local filesystem storage implementation for development.
// This replaces S3 when no S3 credentials are configured.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// LocalStorage stores files in a local directory tree.
// The dev upload handler (PUT /media/dev-upload/{key}) writes to this storage.
type LocalStorage struct {
	BaseDir string
	BaseURL string
}

// NewLocalStorage creates a LocalStorage rooted at baseDir with a base upload URL.
func NewLocalStorage(baseDir string, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating local storage dir: %w", err)
	}
	if baseURL == "" {
		baseURL = "/api/v1/media/dev-upload/"
	}
	return &LocalStorage{BaseDir: baseDir, BaseURL: baseURL}, nil
}

func (s *LocalStorage) path(fileKey string) string {
	return filepath.Join(s.BaseDir, filepath.Clean("/"+fileKey))
}

// GeneratePresignedUploadURL returns a local dev-upload URL.
func (s *LocalStorage) GeneratePresignedUploadURL(_ context.Context, fileKey, _ string, _ int) (string, error) {
	return s.BaseURL + fileKey, nil
}

// ObjectExists checks if the file exists on disk.
func (s *LocalStorage) ObjectExists(_ context.Context, fileKey string) (bool, error) {
	_, err := os.Stat(s.path(fileKey))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// GetObjectMetadata returns basic file metadata.
func (s *LocalStorage) GetObjectMetadata(_ context.Context, fileKey string) (map[string]any, error) {
	info, err := os.Stat(s.path(fileKey))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"content_length": info.Size(),
		"content_type":   "application/octet-stream",
	}, nil
}

// GetObjectBytes reads the first maxBytes of a stored file.
func (s *LocalStorage) GetObjectBytes(_ context.Context, fileKey string, maxBytes int) ([]byte, error) {
	f, err := os.Open(s.path(fileKey))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, maxBytes)
	n, err := f.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return buf[:n], nil
}

// Delete removes a file from local storage.
func (s *LocalStorage) Delete(_ context.Context, fileKey string) error {
	err := os.Remove(s.path(fileKey))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// PutObject writes raw bytes to a local file.
func (s *LocalStorage) PutObject(_ context.Context, fileKey string, data []byte, _ string) error {
	p := s.path(fileKey)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, bytes.Clone(data), 0o644)
}
