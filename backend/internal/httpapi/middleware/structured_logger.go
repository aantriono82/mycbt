package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type structuredLogEntry struct {
	Timestamp string `json:"ts"`
	Level     string `json:"level"`
	Message   string `json:"msg"`
	RequestID string `json:"request_id,omitempty"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Query     string `json:"query,omitempty"`
	Status    int    `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	ResponseB int    `json:"response_bytes"`
	ClientIP  string `json:"client_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"role,omitempty"`
	Error     string `json:"error,omitempty"`
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		entry := structuredLogEntry{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Level:     "info",
			Message:   "http_request",
			RequestID: GetRequestID(c),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Query:     redactQueryString(c.Request.URL.RawQuery),
			Status:    c.Writer.Status(),
			LatencyMS: time.Since(start).Milliseconds(),
			ResponseB: c.Writer.Size(),
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			UserID:    GetUserID(c),
			Role:      GetUserRole(c),
			Error:     c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}

		if entry.Status >= 500 {
			entry.Level = "error"
		} else if entry.Status >= 400 {
			entry.Level = "warn"
		}

		b, err := json.Marshal(entry)
		if err != nil {
			log.Printf(`{"ts":"%s","level":"error","msg":"structured_log_marshal_failed","error":"%s"}`, time.Now().UTC().Format(time.RFC3339), err.Error())
			return
		}
		log.Println(string(b))
	}
}
