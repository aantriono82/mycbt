package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenBlocklist struct {
	client *redis.Client
	prefix string
}

func NewRedisTokenBlocklist(client *redis.Client, prefix string) *RedisTokenBlocklist {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "mycbt"
	}
	return &RedisTokenBlocklist{
		client: client,
		prefix: prefix,
	}
}

func (b *RedisTokenBlocklist) Revoke(ctx context.Context, tokenHash string, ttl time.Duration) error {
	if b == nil || b.client == nil {
		return nil
	}
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" {
		return fmt.Errorf("token hash is required")
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return b.client.Set(ctx, b.key(tokenHash), "1", ttl).Err()
}

func (b *RedisTokenBlocklist) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	if b == nil || b.client == nil {
		return false, nil
	}
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" {
		return false, nil
	}
	n, err := b.client.Exists(ctx, b.key(tokenHash)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (b *RedisTokenBlocklist) key(tokenHash string) string {
	return fmt.Sprintf("%s:jwt:blacklist:%s", b.prefix, tokenHash)
}
