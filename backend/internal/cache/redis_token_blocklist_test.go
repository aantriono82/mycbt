package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewRedisTokenBlocklist_DefaultPrefix(t *testing.T) {
	t.Parallel()

	b := NewRedisTokenBlocklist(nil, " ")
	if b.prefix != "atigacbt" {
		t.Fatalf("expected default prefix, got %q", b.prefix)
	}
}

func TestRedisTokenBlocklist_NilClientAndValidation(t *testing.T) {
	t.Parallel()

	b := NewRedisTokenBlocklist(nil, "custom")
	if err := b.Revoke(context.Background(), "token-hash", time.Minute); err != nil {
		t.Fatalf("expected nil client revoke to be noop, got %v", err)
	}
	revoked, err := b.IsRevoked(context.Background(), "token-hash")
	if err != nil || revoked {
		t.Fatalf("expected nil client not revoked, revoked=%v err=%v", revoked, err)
	}
}

func TestRedisTokenBlocklist_Key(t *testing.T) {
	t.Parallel()

	b := NewRedisTokenBlocklist(nil, "prefix-a")
	if got := b.key("abc123"); got != "prefix-a:jwt:blacklist:abc123" {
		t.Fatalf("unexpected key %q", got)
	}
}
