package handlers

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"mycbt/backend/internal/storage"
)

func uploadImageToStore(ctx context.Context, store storage.ObjectStore, file *multipart.FileHeader, dirPrefix, namePrefix string) (string, error) {
	if store == nil {
		return "", fmt.Errorf("object store is not configured")
	}
	if file == nil {
		return "", fmt.Errorf("file is required")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	objectKey := strings.Trim(filepath.ToSlash(filepath.Join(strings.TrimSpace(dirPrefix), fmt.Sprintf("%s_%d%s", strings.TrimSpace(namePrefix), time.Now().UnixNano(), ext))), "/")
	contentType := file.Header.Get("Content-Type")

	url, err := store.PutObject(ctx, objectKey, contentType, src, file.Size)
	if err != nil {
		return "", err
	}
	return url, nil
}
