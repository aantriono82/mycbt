package middleware

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/repo/auditrepo"
)

func AuditLogger(audit *auditrepo.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		if audit == nil {
			c.Next()
			return
		}

		method := strings.ToUpper(strings.TrimSpace(c.Request.Method))
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
			c.Next()
			return
		}
		if !strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			c.Next()
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth/") {
			c.Next()
			return
		}

		payload := map[string]any{}
		if c.Request != nil && c.Request.Body != nil {
			if shouldCaptureRequestBody(c.Request.Header.Get("Content-Type")) {
				raw, err := io.ReadAll(c.Request.Body)
				if err == nil {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
					if len(raw) > 0 {
						text := string(raw)
						if len(text) > 8192 {
							text = text[:8192] + "...(truncated)"
						}
						payload["body"] = text
					}
				}
			}
		}
		if q := c.Request.URL.RawQuery; strings.TrimSpace(q) != "" {
			payload["query"] = q
		}

		c.Next()

		role := GetUserRole(c)
		if role != "admin" && role != "teacher" {
			return
		}

		_ = audit.Create(c.Request.Context(), auditrepo.CreateLogInput{
			RequestID:  GetRequestID(c),
			UserID:     GetUserID(c),
			Role:       role,
			Method:     method,
			Path:       c.Request.URL.Path,
			Query:      c.Request.URL.RawQuery,
			StatusCode: c.Writer.Status(),
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Payload:    payload,
		})
	}
}

func shouldCaptureRequestBody(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return true
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == "application/json" || mediaType == "application/x-www-form-urlencoded" || strings.HasPrefix(mediaType, "text/")
}
