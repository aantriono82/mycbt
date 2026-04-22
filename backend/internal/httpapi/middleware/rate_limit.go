package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

		globalRateLimiter.ensureJanitor()
		allowed, retryAfterSec := globalRateLimiter.allow(key, maxRequests, window, now)
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
