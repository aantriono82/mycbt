package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"atigacbt/backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStore interface {
	PutObject(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) (string, error)
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
}

type LocalObjectStore struct {
	BaseDir string
}

func NewLocalObjectStore(baseDir string) *LocalObjectStore {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		baseDir = "uploads"
	}
	baseDir = config.ResolveAppPath(baseDir)
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

	return "/uploads/" + strings.TrimLeft(filepath.ToSlash(objectKey), "/"), nil
}

func (s *LocalObjectStore) GetObject(_ context.Context, objectKey string) (io.ReadCloser, string, error) {
	objectKey = strings.TrimLeft(filepath.ToSlash(strings.TrimSpace(objectKey)), "/")
	if objectKey == "" {
		return nil, "", fmt.Errorf("object key is required")
	}

	targetPath := filepath.Join(s.BaseDir, filepath.FromSlash(objectKey))
	f, err := os.Open(targetPath)
	if err != nil {
		return nil, "", err
	}
	contentType := mimeFromExt(filepath.Ext(targetPath))
	return f, contentType, nil
}

type S3ObjectStore struct {
	Client        *minio.Client
	Bucket        string
	KeyPrefix     string
	PublicBaseURL string
	UseSSL        bool
	Endpoint      string
}

type S3Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
	PublicBaseURL string
	KeyPrefix     string
}

func NewS3ObjectStore(ctx context.Context, cfg S3Config) (*S3ObjectStore, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	bucket := strings.TrimSpace(cfg.Bucket)
	accessKey := strings.TrimSpace(cfg.AccessKey)
	secretKey := strings.TrimSpace(cfg.SecretKey)
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("incomplete s3-compatible object storage configuration")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 client: %w", err)
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

	return &S3ObjectStore{
		Client:        client,
		Bucket:        bucket,
		KeyPrefix:     strings.Trim(strings.TrimSpace(cfg.KeyPrefix), "/"),
		PublicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
		UseSSL:        cfg.UseSSL,
		Endpoint:      endpoint,
	}, nil
}

func (s *S3ObjectStore) PutObject(ctx context.Context, objectKey, contentType string, body io.Reader, size int64) (string, error) {
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

	return "/uploads/" + fullKey, nil
}

func (s *S3ObjectStore) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	objectKey = strings.TrimLeft(filepath.ToSlash(strings.TrimSpace(objectKey)), "/")
	if objectKey == "" {
		return nil, "", fmt.Errorf("object key is required")
	}

	obj, err := s.Client.GetObject(ctx, s.Bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("get object: %w", err)
	}

	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, "", fmt.Errorf("stat object: %w", err)
	}

	return obj, strings.TrimSpace(info.ContentType), nil
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
