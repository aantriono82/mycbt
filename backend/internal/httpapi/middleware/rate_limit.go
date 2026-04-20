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
		ip := strings.TrimSpace(c.ClientIP())
		if ip == "" {
			ip = "unknown"
		}
		key := fmt.Sprintf("%s:%s", scope, ip)
		now := time.Now()

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

func (l *inMemoryRateLimiter) allow(key string, maxRequests int, window time.Duration, now time.Time) (allowed bool, retryAfterSec int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Lazy cleanup to keep map bounded.
	for k, v := range l.store {
		if now.After(v.ResetAt) {
			delete(l.store, k)
		}
	}

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
