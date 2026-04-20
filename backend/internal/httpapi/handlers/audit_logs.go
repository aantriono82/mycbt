package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/repo/auditrepo"
)

type AuditLogsHandler struct {
	repo *auditrepo.Repo
}

func NewAuditLogsHandler(repo *auditrepo.Repo) *AuditLogsHandler {
	return &AuditLogsHandler{repo: repo}
}

func (h *AuditLogsHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	q := strings.TrimSpace(c.Query("q"))
	role := strings.TrimSpace(c.Query("role"))
	method := strings.TrimSpace(c.Query("method"))
	path := strings.TrimSpace(c.Query("path"))
	ipStatus, _ := strconv.Atoi(strings.TrimSpace(c.Query("status")))
	var statusPtr *int
	if ipStatus > 0 {
		statusPtr = &ipStatus
	}

	from, fromOk := parseTimeQueryAudit(c.Query("from"))
	to, toOk := parseTimeQueryAudit(c.Query("to"))
	var fromPtr *time.Time
	var toPtr *time.Time
	if fromOk {
		fromPtr = &from
	}
	if toOk {
		toPtr = &to
	}

	items, total, err := h.repo.List(c.Request.Context(), auditrepo.ListFilter{
		Q:      q,
		Role:   role,
		Method: method,
		Path:   path,
		Status: statusPtr,
		From:   fromPtr,
		To:     toPtr,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal memuat audit log"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"items":  items,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *AuditLogsHandler) ExportCSV(c *gin.Context) {
	// Export without pagination (bounded).
	q := strings.TrimSpace(c.Query("q"))
	role := strings.TrimSpace(c.Query("role"))
	method := strings.TrimSpace(c.Query("method"))
	path := strings.TrimSpace(c.Query("path"))
	statusN, _ := strconv.Atoi(strings.TrimSpace(c.Query("status")))
	var statusPtr *int
	if statusN > 0 {
		statusPtr = &statusN
	}

	from, fromOk := parseTimeQueryAudit(c.Query("from"))
	to, toOk := parseTimeQueryAudit(c.Query("to"))
	var fromPtr *time.Time
	var toPtr *time.Time
	if fromOk {
		fromPtr = &from
	}
	if toOk {
		toPtr = &to
	}

	items, _, err := h.repo.List(c.Request.Context(), auditrepo.ListFilter{
		Q:      q,
		Role:   role,
		Method: method,
		Path:   path,
		Status: statusPtr,
		From:   fromPtr,
		To:     toPtr,
		Limit:  5000,
		Offset: 0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal export audit log"}})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="audit_logs.csv"`)
	c.Status(http.StatusOK)

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{
		"created_at",
		"username",
		"name",
		"role",
		"method",
		"path",
		"status_code",
		"ip",
		"user_agent",
		"request_id",
		"query",
		"payload_json",
	})

	for _, it := range items {
		roleStr := ""
		if it.Role != nil {
			roleStr = *it.Role
		}
		ipStr := ""
		if it.IP != nil {
			ipStr = *it.IP
		}
		uaStr := ""
		if it.UserAgent != nil {
			uaStr = *it.UserAgent
		}
		queryStr := ""
		if it.Query != nil {
			queryStr = *it.Query
		}
		payloadStr := "{}"
		if it.Payload != nil {
			if raw, mErr := json.Marshal(it.Payload); mErr == nil {
				// Avoid extremely large cells.
				if len(raw) > 8192 {
					raw = append(raw[:8192], []byte("...(truncated)")...)
				}
				payloadStr = string(raw)
			}
		}
		_ = w.Write([]string{
			it.CreatedAt.UTC().Format(time.RFC3339),
			it.Username,
			it.Name,
			roleStr,
			it.Method,
			it.Path,
			strconv.Itoa(it.StatusCode),
			ipStr,
			uaStr,
			it.RequestID,
			queryStr,
			payloadStr,
		})
	}
	w.Flush()
}

func (h *AuditLogsHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "id wajib"}})
		return
	}
	if err := h.repo.DeleteByID(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal menghapus audit log"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

func (h *AuditLogsHandler) ClearAll(c *gin.Context) {
	n, err := h.repo.ClearAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal menghapus semua audit log"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true, "deleted": n}})
}

func (h *AuditLogsHandler) Prune(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 30
	}
	n, err := h.repo.PruneOlderThan(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal prune audit log"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true, "deleted": n, "days": days}})
}

func parseTimeQueryAudit(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
	}
	return time.Time{}, false
}
