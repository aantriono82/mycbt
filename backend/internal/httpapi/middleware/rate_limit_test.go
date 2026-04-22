package middleware

import (
	"testing"
	"time"
)

func TestInMemoryRateLimiterAllowAndBlock(t *testing.T) {
	limiter := &inMemoryRateLimiter{
		store: map[string]rateLimitEntry{},
	}
	now := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)

	allowed, retry := limiter.allow("auth:1.2.3.4", 2, time.Minute, now)
	if !allowed || retry != 0 {
		t.Fatalf("first request should pass, got allowed=%v retry=%d", allowed, retry)
	}

	allowed, retry = limiter.allow("auth:1.2.3.4", 2, time.Minute, now.Add(5*time.Second))
	if !allowed || retry != 0 {
		t.Fatalf("second request should pass, got allowed=%v retry=%d", allowed, retry)
	}

	allowed, retry = limiter.allow("auth:1.2.3.4", 2, time.Minute, now.Add(10*time.Second))
	if allowed {
		t.Fatalf("third request should be blocked")
	}
	if retry <= 0 {
		t.Fatalf("retry-after should be positive, got %d", retry)
	}
}

func TestInMemoryRateLimiterResetAfterWindow(t *testing.T) {
	limiter := &inMemoryRateLimiter{
		store: map[string]rateLimitEntry{},
	}
	now := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)

	allowed, _ := limiter.allow("join:5.6.7.8", 1, 30*time.Second, now)
	if !allowed {
		t.Fatalf("first request should pass")
	}

	allowed, _ = limiter.allow("join:5.6.7.8", 1, 30*time.Second, now.Add(5*time.Second))
	if allowed {
		t.Fatalf("second request within window should be blocked")
	}

	allowed, retry := limiter.allow("join:5.6.7.8", 1, 30*time.Second, now.Add(31*time.Second))
	if !allowed {
		t.Fatalf("request after window should pass, got retry=%d", retry)
	}
}

func TestRateLimitIdentity(t *testing.T) {
	if got := rateLimitIdentity("u1", "1.2.3.4"); got != "u:u1" {
		t.Fatalf("expected user identity, got %q", got)
	}
	if got := rateLimitIdentity("", "1.2.3.4"); got != "ip:1.2.3.4" {
		t.Fatalf("expected ip identity, got %q", got)
	}
	if got := rateLimitIdentity("  ", ""); got != "ip:unknown" {
		t.Fatalf("expected unknown ip identity, got %q", got)
	}
}
