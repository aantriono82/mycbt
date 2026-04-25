package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	AppURL      string
	CORSOrigins string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       string
	RedisPrefix   string

	UploadProvider     string
	UploadLocalDir     string
	MinIOEndpoint      string
	MinIOAccessKey     string
	MinIOSecretKey     string
	MinIOBucket        string
	MinIOUseSSL        string
	MinIOPublicBaseURL string
	MinIOKeyPrefix     string

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
		CORSOrigins: getenv("CORS_ORIGINS", "http://localhost:5173"),

		DatabaseURL: getenv("DATABASE_URL", ""),

		RedisAddr:     getenv("REDIS_ADDR", ""),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		RedisDB:       getenv("REDIS_DB", "0"),
		RedisPrefix:   getenv("REDIS_PREFIX", "mycbt"),

		UploadProvider:     getenv("UPLOAD_PROVIDER", "local"),
		UploadLocalDir:     getenv("UPLOAD_LOCAL_DIR", "uploads"),
		MinIOEndpoint:      getenv("MINIO_ENDPOINT", ""),
		MinIOAccessKey:     getenv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey:     getenv("MINIO_SECRET_KEY", ""),
		MinIOBucket:        getenv("MINIO_BUCKET", ""),
		MinIOUseSSL:        getenv("MINIO_USE_SSL", "false"),
		MinIOPublicBaseURL: getenv("MINIO_PUBLIC_BASE_URL", ""),
		MinIOKeyPrefix:     getenv("MINIO_KEY_PREFIX", ""),

		JWTSecret:     getenv("JWT_SECRET", ""),
		JWTIssuer:     getenv("JWT_ISSUER", "mycbt"),
		JWTTTLMinutes: getenv("JWT_TTL_MINUTES", "120"),

		AdminUsername:      getenv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getenv("ADMIN_PASSWORD", "admin12345"),
		AdminName:          getenv("ADMIN_NAME", "Administrator"),
		AdminEmail:         getenv("ADMIN_EMAIL", "admin@example.com"),
		GoogleClientID:     getenv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URL", ""),
		SMTPHost:           getenv("SMTP_HOST", ""),
		SMTPPort:           getenv("SMTP_PORT", "587"),
		SMTPUser:           getenv("SMTP_USER", ""),
		SMTPPass:           getenv("SMTP_PASS", ""),
		SMTPFrom:           getenv("SMTP_FROM", "no-reply@mycbt.com"),
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
