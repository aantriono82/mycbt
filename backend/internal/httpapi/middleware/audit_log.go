package middleware

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/repo/auditrepo"
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
		skipBody := shouldSkipAuditBody(c.Request.URL.Path)
		if !skipBody && c.Request != nil && c.Request.Body != nil {
			if shouldCaptureRequestBody(c.Request.Header.Get("Content-Type")) {
				raw, err := io.ReadAll(c.Request.Body)
				if err == nil {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
					if len(raw) > 0 {
						text := redactBody(c.Request.Header.Get("Content-Type"), raw)
						if len(text) > 8192 {
							text = text[:8192] + "...(truncated)"
						}
						payload["body"] = text
					}
				}
			}
		}
		if q := strings.TrimSpace(c.Request.URL.RawQuery); q != "" {
			payload["query"] = redactQueryString(q)
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

func shouldSkipAuditBody(path string) bool {
	p := strings.TrimSpace(path)
	if p == "" {
		return false
	}
	// Settings endpoints contain secrets (SMTP password, WhatsApp API key).
	if strings.HasPrefix(p, "/api/v1/settings/smtp") {
		return true
	}
	if strings.HasPrefix(p, "/api/v1/settings/whatsapp") {
		return true
	}
	// Student token entry endpoints.
	if strings.HasSuffix(p, "/student/exams/") {
		// no-op; guard for weird paths
	}
	if strings.HasSuffix(p, "/join") && strings.Contains(p, "/api/v1/student/exams/") {
		return true
	}
	if strings.HasSuffix(p, "/verify-token") && strings.Contains(p, "/api/v1/student/sessions/") {
		return true
	}
	return false
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
