package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/service/authsvc"
)

const (
	ctxUserIDKey   = "user_id"
	ctxUserRoleKey = "user_role"
)

func RequireAuth(auth *authsvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "unauthorized", "message": "missing Authorization header"}})
			return
		}

		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "unauthorized", "message": "invalid Authorization header"}})
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "unauthorized", "message": "invalid token"}})
			return
		}

		c.Set(ctxUserIDKey, claims.Subject)
		c.Set(ctxUserRoleKey, claims.Role)
		c.Next()
	}
}

// RequireAuthHeaderOrQueryToken allows browser EventSource (SSE) clients which cannot set custom headers.
// It accepts either:
// - `Authorization: Bearer <token>` header (preferred)
// - `?access_token=<token>` query param
//
// WARNING: Query token can leak via logs/referrers. Prefer header/cookie auth in production.
func RequireAuthHeaderOrQueryToken(auth *authsvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		raw := ""

		if h != "" {
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "unauthorized", "message": "invalid Authorization header"}})
				return
			}
			raw = parts[1]
		} else {
			raw = strings.TrimSpace(c.Query("access_token"))
			if raw == "" {
				c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "unauthorized", "message": "missing token"}})
				return
			}
		}

		claims, err := auth.ParseToken(raw)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "unauthorized", "message": "invalid token"}})
			return
		}

		c.Set(ctxUserIDKey, claims.Subject)
		c.Set(ctxUserRoleKey, claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get(ctxUserRoleKey)
		roleStr, _ := role.(string)
		if roleStr == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "unauthorized", "message": "unauthorized"}})
			return
		}
		if _, ok := allowed[roleStr]; !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	v, _ := c.Get(ctxUserIDKey)
	s, _ := v.(string)
	return s
}

func GetUserRole(c *gin.Context) string {
	v, _ := c.Get(ctxUserRoleKey)
	s, _ := v.(string)
	return s
}
