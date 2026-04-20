package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LimitBodyBytes(maxBytes int64) gin.HandlerFunc {
	if maxBytes < 1 {
		maxBytes = 1
	}
	return func(c *gin.Context) {
		if c.Request == nil || c.Request.Body == nil {
			c.Next()
			return
		}

		// Fast reject when Content-Length is provided and exceeds limit.
		if c.Request.ContentLength > 0 && c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": gin.H{
					"code":    "payload_too_large",
					"message": fmt.Sprintf("request body exceeds limit (%d bytes)", maxBytes),
				},
			})
			return
		}

		// Hard limit for chunked/unknown length bodies.
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
