package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	AppURL      string
	FrontendURL string
	CORSOrigins string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       string
	RedisPrefix   string

	UploadProvider      string
	UploadLocalDir      string
	RustFSEndpoint      string
	RustFSAccessKey     string
	RustFSSecretKey     string
	RustFSBucket        string
	RustFSUseSSL        string
	RustFSPublicBaseURL string
	RustFSKeyPrefix     string

	JWTSecret     string
	JWTIssuer     string
	JWTTTLMinutes string

	AdminUsername      string
	AdminPassword      string
	AdminName          string
	AdminEmail         string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	SMTPFrom           string
}

func Load() Config {
	// Support running the API from either repo root or from `backend/`.
	// - If CWD is `backend/`, `.env` will be found.
	// - If CWD is repo root, `backend/.env` will be found.
	_ = godotenv.Load()
	_ = godotenv.Load("backend/.env")
	return Config{
		Env:         getenv("GIN_MODE", "debug"),
		Port:        getenv("PORT", "8080"),
		AppURL:      getenv("APP_URL", "http://localhost:8080"),
		FrontendURL: getenv("FRONTEND_URL", ""),
		CORSOrigins: getenv("CORS_ORIGINS", "http://localhost:5173"),

		DatabaseURL: getenv("DATABASE_URL", ""),

		RedisAddr:     getenv("REDIS_ADDR", ""),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		RedisDB:       getenv("REDIS_DB", "0"),
		RedisPrefix:   getenv("REDIS_PREFIX", "atigacbt"),

		UploadProvider:      getenv("UPLOAD_PROVIDER", "local"),
		UploadLocalDir:      getenv("UPLOAD_LOCAL_DIR", "uploads"),
		RustFSEndpoint:      getenvAny([]string{"RUSTFS_ENDPOINT", "MINIO_ENDPOINT"}, ""),
		RustFSAccessKey:     getenvAny([]string{"RUSTFS_ACCESS_KEY", "MINIO_ACCESS_KEY"}, ""),
		RustFSSecretKey:     getenvAny([]string{"RUSTFS_SECRET_KEY", "MINIO_SECRET_KEY"}, ""),
		RustFSBucket:        getenvAny([]string{"RUSTFS_BUCKET", "MINIO_BUCKET"}, ""),
		RustFSUseSSL:        getenvAny([]string{"RUSTFS_USE_SSL", "MINIO_USE_SSL"}, "false"),
		RustFSPublicBaseURL: getenvAny([]string{"RUSTFS_PUBLIC_BASE_URL", "MINIO_PUBLIC_BASE_URL"}, ""),
		RustFSKeyPrefix:     getenvAny([]string{"RUSTFS_KEY_PREFIX", "MINIO_KEY_PREFIX"}, ""),

		JWTSecret:     getenv("JWT_SECRET", ""),
		JWTIssuer:     getenv("JWT_ISSUER", "atigacbt"),
		JWTTTLMinutes: getenv("JWT_TTL_MINUTES", "120"),

		AdminUsername:      getenv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getenv("ADMIN_PASSWORD", ""),
		AdminName:          getenv("ADMIN_NAME", "Administrator"),
		AdminEmail:         getenv("ADMIN_EMAIL", "admin@example.com"),
		GoogleClientID:     getenv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URL", ""),
		SMTPHost:           getenv("SMTP_HOST", ""),
		SMTPPort:           getenv("SMTP_PORT", "587"),
		SMTPUser:           getenv("SMTP_USER", ""),
		SMTPPass:           getenv("SMTP_PASS", ""),
		SMTPFrom:           getenv("SMTP_FROM", "no-reply@atigacbt.com"),
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func getenvAny(keys []string, fallback string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return fallback
}

func ResolveAppPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || filepath.IsAbs(path) {
		return path
	}

	if wd, err := os.Getwd(); err == nil {
		if filepath.Base(wd) == "backend" {
			return filepath.Join(wd, path)
		}
		if info, err := os.Stat(filepath.Join(wd, "backend")); err == nil && info.IsDir() {
			return filepath.Join(wd, "backend", path)
		}
		return filepath.Join(wd, path)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		if info, err := os.Stat(filepath.Join(exeDir, "backend")); err == nil && info.IsDir() {
			return filepath.Join(exeDir, "backend", path)
		}
		return filepath.Join(exeDir, path)
	}

	return path
}
