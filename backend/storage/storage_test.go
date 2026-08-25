package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"flickey/go-backend/storage"
)

func TestLocalStorage_CRUD(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewLocalStorage(tempDir, "")
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	ctx := context.Background()
	fileKey := "raw/test-host/test-media.jpg"
	content := []byte("hello-image-content-here")

	// 1. PutObject
	if err := store.PutObject(ctx, fileKey, content, "image/jpeg"); err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	// 2. ObjectExists
	exists, err := store.ObjectExists(ctx, fileKey)
	if err != nil {
		t.Fatalf("ObjectExists: %v", err)
	}
	if !exists {
		t.Error("expected object to exist")
	}

	// 3. GetObjectMetadata
	meta, err := store.GetObjectMetadata(ctx, fileKey)
	if err != nil {
		t.Fatalf("GetObjectMetadata: %v", err)
	}
	if meta == nil {
		t.Fatal("expected non-nil metadata")
	}
	if meta["content_length"].(int64) != int64(len(content)) {
		t.Errorf("expected size %d, got %v", len(content), meta["content_length"])
	}

	// 4. GetObjectBytes
	readBytes, err := store.GetObjectBytes(ctx, fileKey, 5)
	if err != nil {
		t.Fatalf("GetObjectBytes: %v", err)
	}
	if string(readBytes) != "hello" {
		t.Errorf("expected 'hello', got %q", string(readBytes))
	}

	// 5. GeneratePresignedUploadURL
	url, err := store.GeneratePresignedUploadURL(ctx, fileKey, "image/jpeg", 300)
	if err != nil {
		t.Fatalf("GeneratePresignedUploadURL: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty presigned URL")
	}

	// 6. Delete
	if err := store.Delete(ctx, fileKey); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	exists, _ = store.ObjectExists(ctx, fileKey)
	if exists {
		t.Error("expected object to not exist after delete")
	}
}

func TestLocalStorage_NonExistent(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "nonexistent_storage_test")
	store, _ := storage.NewLocalStorage(tempDir, "")
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	exists, err := store.ObjectExists(ctx, "nonexistent.jpg")
	if err != nil {
		t.Errorf("unexpected error on ObjectExists: %v", err)
	}
	if exists {
		t.Error("expected false for nonexistent object")
	}

	meta, err := store.GetObjectMetadata(ctx, "nonexistent.jpg")
	if err != nil {
		t.Errorf("unexpected error on GetObjectMetadata: %v", err)
	}
	if meta != nil {
		t.Error("expected nil metadata for nonexistent object")
	}
}
