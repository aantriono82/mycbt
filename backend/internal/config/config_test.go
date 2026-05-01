package config

import "testing"

func TestGetenv(t *testing.T) {
	t.Setenv("CFG_TEST_ONE", "value-1")
	if got := getenv("CFG_TEST_ONE", "fallback"); got != "value-1" {
		t.Fatalf("expected env value, got %q", got)
	}
	if got := getenv("CFG_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestGetenvAny(t *testing.T) {
	t.Setenv("CFG_TEST_ANY_B", "value-b")
	if got := getenvAny([]string{"CFG_TEST_ANY_A", "CFG_TEST_ANY_B"}, "fallback"); got != "value-b" {
		t.Fatalf("expected second env value, got %q", got)
	}
	if got := getenvAny([]string{"CFG_TEST_ANY_X", "CFG_TEST_ANY_Y"}, "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestLoad_PrefersExplicitEnvAndAliases(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	t.Setenv("PORT", "9090")
	t.Setenv("APP_URL", "https://app.example.com")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("MINIO_ENDPOINT", "minio.example.com")
	t.Setenv("MINIO_ACCESS_KEY", "miniouser")
	t.Setenv("MINIO_SECRET_KEY", "miniosecret")
	t.Setenv("MINIO_BUCKET", "bucket-1")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("MINIO_PUBLIC_BASE_URL", "https://cdn.example.com")
	t.Setenv("MINIO_KEY_PREFIX", "files")
	t.Setenv("SMTP_PORT", "2525")

	cfg := Load()
	if cfg.Env != "release" || cfg.Port != "9090" || cfg.AppURL != "https://app.example.com" {
		t.Fatalf("unexpected basic config: %+v", cfg)
	}
	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("unexpected DATABASE_URL: %q", cfg.DatabaseURL)
	}
	if cfg.RustFSEndpoint != "minio.example.com" || cfg.RustFSBucket != "bucket-1" || cfg.RustFSUseSSL != "true" {
		t.Fatalf("unexpected rustfs/minio alias mapping: %+v", cfg)
	}
	if cfg.SMTPPort != "2525" {
		t.Fatalf("unexpected smtp port: %q", cfg.SMTPPort)
	}
}
