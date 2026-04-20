package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
)

const ctxRequestIDKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if reqID == "" {
			reqID = generateRequestID()
		}
		c.Set(ctxRequestIDKey, reqID)
		c.Header("X-Request-ID", reqID)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	v, _ := c.Get(ctxRequestIDKey)
	s, _ := v.(string)
	return s
}

func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-fallback"
	}
	return "req-" + hex.EncodeToString(b)
}
