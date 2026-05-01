package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type rateLimitEntry struct {
	Count   int
	ResetAt time.Time
}

type inMemoryRateLimiter struct {
	mu    sync.Mutex
	store map[string]rateLimitEntry
	once  sync.Once
}

var globalRateLimiter = &inMemoryRateLimiter{
	store: map[string]rateLimitEntry{},
}

var (
	globalRedisRateLimiterMu sync.RWMutex
	globalRedisRateLimiter   *redisRateLimiter
)

var redisRateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("TTL", KEYS[1])
return {current, ttl}
`)

func RateLimit(scope string, maxRequests int, window time.Duration) gin.HandlerFunc {
	if strings.TrimSpace(scope) == "" {
		scope = "global"
	}
	if maxRequests < 1 {
		maxRequests = 1
	}
	if window <= 0 {
		window = time.Minute
	}

	return func(c *gin.Context) {
		userID := strings.TrimSpace(GetUserID(c))
		ip := strings.TrimSpace(c.ClientIP())
		key := fmt.Sprintf("%s:%s", scope, rateLimitIdentity(userID, ip))
		now := time.Now()

		allowed, retryAfterSec := allowRateLimit(c, key, maxRequests, window, now)
		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfterSec))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "rate_limited",
					"message": "too many requests",
				},
			})
			return
		}

		c.Next()
	}
}

func UseRedisRateLimiter(client *redis.Client, prefix string) {
	globalRedisRateLimiterMu.Lock()
	defer globalRedisRateLimiterMu.Unlock()
	if client == nil {
		globalRedisRateLimiter = nil
		return
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "atigacbt"
	}
	globalRedisRateLimiter = &redisRateLimiter{
		client: client,
		prefix: prefix,
	}
}

func allowRateLimit(c *gin.Context, key string, maxRequests int, window time.Duration, now time.Time) (bool, int) {
	globalRedisRateLimiterMu.RLock()
	rr := globalRedisRateLimiter
	globalRedisRateLimiterMu.RUnlock()

	if rr != nil {
		allowed, retryAfterSec, err := rr.allow(c, key, maxRequests, window)
		if err == nil {
			return allowed, retryAfterSec
		}
	}

	globalRateLimiter.ensureJanitor()
	return globalRateLimiter.allow(key, maxRequests, window, now)
}

func rateLimitIdentity(userID, ip string) string {
	if strings.TrimSpace(userID) != "" {
		return "u:" + strings.TrimSpace(userID)
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}
	return "ip:" + ip
}

func (l *inMemoryRateLimiter) ensureJanitor() {
	l.once.Do(func() {
		go func() {
			t := time.NewTicker(60 * time.Second)
			defer t.Stop()
			for now := range t.C {
				l.cleanupExpired(now)
			}
		}()
	})
}

func (l *inMemoryRateLimiter) allow(key string, maxRequests int, window time.Duration, now time.Time) (allowed bool, retryAfterSec int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.store[key]
	if !ok || now.After(entry.ResetAt) {
		l.store[key] = rateLimitEntry{
			Count:   1,
			ResetAt: now.Add(window),
		}
		return true, 0
	}

	if entry.Count >= maxRequests {
		retry := int(entry.ResetAt.Sub(now).Seconds())
		if retry < 1 {
			retry = 1
		}
		return false, retry
	}

	entry.Count++
	l.store[key] = entry
	return true, 0
}

func (l *inMemoryRateLimiter) cleanupExpired(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range l.store {
		if now.After(v.ResetAt) {
			delete(l.store, k)
		}
	}
}

type redisRateLimiter struct {
	client *redis.Client
	prefix string
}

func (r *redisRateLimiter) allow(c *gin.Context, key string, maxRequests int, window time.Duration) (bool, int, error) {
	ctx := c.Request.Context()
	fullKey := fmt.Sprintf("%s:ratelimit:%s", r.prefix, key)
	windowSec := int(window.Seconds())
	if windowSec < 1 {
		windowSec = 1
	}

	res, err := redisRateLimitScript.Run(ctx, r.client, []string{fullKey}, windowSec).Result()
	if err != nil {
		return false, 0, err
	}

	arr, ok := res.([]any)
	if !ok || len(arr) < 2 {
		return false, 0, fmt.Errorf("unexpected redis rate-limit response")
	}

	current, err := toInt64(arr[0])
	if err != nil {
		return false, 0, err
	}
	ttlSec, err := toInt64(arr[1])
	if err != nil {
		return false, 0, err
	}
	if current <= int64(maxRequests) {
		return true, 0, nil
	}
	if ttlSec < 1 {
		ttlSec = 1
	}
	return false, int(ttlSec), nil
}

func toInt64(v any) (int64, error) {
	switch t := v.(type) {
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case string:
		return strconv.ParseInt(strings.TrimSpace(t), 10, 64)
	case []byte:
		return strconv.ParseInt(strings.TrimSpace(string(t)), 10, 64)
	default:
		return 0, fmt.Errorf("unexpected number type %T", v)
	}
}
