package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/repo/loginlogrepo"
)

type LoginLogsHandler struct {
	repo *loginlogrepo.Repo
}

func NewLoginLogsHandler(repo *loginlogrepo.Repo) *LoginLogsHandler {
	return &LoginLogsHandler{repo: repo}
}

func (h *LoginLogsHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	q := strings.TrimSpace(c.Query("q"))
	role := strings.TrimSpace(c.Query("role"))
	ip := strings.TrimSpace(c.Query("ip"))
	from, fromOk := parseTimeQuery(c.Query("from"))
	to, toOk := parseTimeQuery(c.Query("to"))

	var fromPtr *time.Time
	var toPtr *time.Time
	if fromOk {
		fromPtr = &from
	}
	if toOk {
		toPtr = &to
	}

	items, total, err := h.repo.List(c.Request.Context(), loginlogrepo.ListFilter{
		Q:      q,
		Role:   role,
		IP:     ip,
		From:   fromPtr,
		To:     toPtr,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal memuat log aktivitas"}})
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

func (h *LoginLogsHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "bad_request", "message": "id wajib"}})
		return
	}

	if err := h.repo.DeleteByID(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal menghapus log"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

func (h *LoginLogsHandler) ClearAll(c *gin.Context) {
	n, err := h.repo.ClearAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal menghapus semua log"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true, "deleted": n}})
}

func (h *LoginLogsHandler) Prune(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 30
	}
	n, err := h.repo.PruneOlderThan(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal", "message": "gagal prune log"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true, "deleted": n, "days": days}})
}

func parseTimeQuery(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	// Accept RFC3339 or YYYY-MM-DD
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		// Interpret as UTC midnight to keep comparison stable.
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
	}
	return time.Time{}, false
}
