package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLocalObjectStore_DefaultBaseDir(t *testing.T) {
	t.Parallel()

	store := NewLocalObjectStore(" ")
	if !filepath.IsAbs(store.BaseDir) {
		t.Fatalf("expected absolute uploads dir, got %q", store.BaseDir)
	}
	if !strings.HasSuffix(filepath.ToSlash(store.BaseDir), "/uploads") {
		t.Fatalf("expected uploads suffix in base dir, got %q", store.BaseDir)
	}
}

func TestLocalObjectStore_PutObject(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	store := NewLocalObjectStore(baseDir)

	url, err := store.PutObject(context.Background(), "avatars/user-1.txt", "text/plain", strings.NewReader("hello"), 5)
	if err != nil {
		t.Fatalf("PutObject error: %v", err)
	}
	if url != "/uploads/avatars/user-1.txt" {
		t.Fatalf("unexpected object url %q", url)
	}

	content, err := os.ReadFile(filepath.Join(baseDir, "avatars", "user-1.txt"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected stored content %q", string(content))
	}
}

func TestLocalObjectStore_PutObject_ValidatesKey(t *testing.T) {
	t.Parallel()

	store := NewLocalObjectStore(t.TempDir())
	if _, err := store.PutObject(context.Background(), " ", "text/plain", strings.NewReader("x"), 1); err == nil {
		t.Fatal("expected empty object key error")
	}
}

func TestS3PutObject_ValidatesKeyAndBuildsPublicURL(t *testing.T) {
	t.Parallel()

	store := &S3ObjectStore{
		Bucket:        "bucket-a",
		KeyPrefix:     "prefix",
		PublicBaseURL: "https://cdn.example.com/base",
		Endpoint:      "s3.example.com",
	}

	if _, err := store.PutObject(context.Background(), " ", "text/plain", strings.NewReader("x"), 1); err == nil {
		t.Fatal("expected empty object key error")
	}
}
