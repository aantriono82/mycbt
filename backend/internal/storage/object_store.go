package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStore interface {
	PutObject(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) (string, error)
}

type LocalObjectStore struct {
	BaseDir string
}

func NewLocalObjectStore(baseDir string) *LocalObjectStore {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		baseDir = "uploads"
	}
	return &LocalObjectStore{BaseDir: baseDir}
}

func (s *LocalObjectStore) PutObject(_ context.Context, objectKey, _ string, body io.Reader, _ int64) (string, error) {
	objectKey = strings.TrimLeft(filepath.ToSlash(strings.TrimSpace(objectKey)), "/")
	if objectKey == "" {
		return "", fmt.Errorf("object key is required")
	}

	targetPath := filepath.Join(s.BaseDir, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, body); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return "/" + filepath.ToSlash(filepath.Join(s.BaseDir, objectKey)), nil
}

type MinIOObjectStore struct {
	Client        *minio.Client
	Bucket        string
	KeyPrefix     string
	PublicBaseURL string
	UseSSL        bool
	Endpoint      string
}

type MinIOConfig struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
	PublicBaseURL string
	KeyPrefix     string
}

func NewMinIOObjectStore(ctx context.Context, cfg MinIOConfig) (*MinIOObjectStore, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	bucket := strings.TrimSpace(cfg.Bucket)
	accessKey := strings.TrimSpace(cfg.AccessKey)
	secretKey := strings.TrimSpace(cfg.SecretKey)
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("incomplete minio configuration")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &MinIOObjectStore{
		Client:        client,
		Bucket:        bucket,
		KeyPrefix:     strings.Trim(strings.TrimSpace(cfg.KeyPrefix), "/"),
		PublicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
		UseSSL:        cfg.UseSSL,
		Endpoint:      endpoint,
	}, nil
}

func (s *MinIOObjectStore) PutObject(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) (string, error) {
	objectKey = strings.TrimLeft(filepath.ToSlash(strings.TrimSpace(objectKey)), "/")
	if objectKey == "" {
		return "", fmt.Errorf("object key is required")
	}

	fullKey := objectKey
	if s.KeyPrefix != "" {
		fullKey = s.KeyPrefix + "/" + objectKey
	}

	opts := minio.PutObjectOptions{}
	if strings.TrimSpace(contentType) != "" {
		opts.ContentType = strings.TrimSpace(contentType)
	}
	if _, err := s.Client.PutObject(ctx, s.Bucket, fullKey, body, size, opts); err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	if s.PublicBaseURL != "" {
		return s.PublicBaseURL + "/" + fullKey, nil
	}

	scheme := "http"
	if s.UseSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.Endpoint, s.Bucket, fullKey), nil
}
