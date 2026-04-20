package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/httpapi/params"
	"mycbt/backend/internal/repo/examrepo"
	"mycbt/backend/internal/repo/studentexamrepo"
)

type MonitorHandler struct {
	ex *examrepo.Repo
	st *studentexamrepo.Repo
}

func NewMonitorHandler(ex *examrepo.Repo, st *studentexamrepo.Repo) *MonitorHandler {
	return &MonitorHandler{ex: ex, st: st}
}

func (h *MonitorHandler) ListSessions(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	examID := c.Param("id")
	exam, ok, err := h.ex.Get(c.Request.Context(), examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || exam.TeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	limit := params.IntQuery(c, "limit", 100, 1, 500)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	nowUTC := time.Now().UTC()
	items, total, err := h.st.ListMonitorSessions(c.Request.Context(), examID, studentexamrepo.MonitorSessionsFilter{
		NowUTC:       nowUTC,
		OnlineWindow: 30 * time.Second,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{
		"data": items,
		"meta": gin.H{
			"exam":              exam,
			"limit":             limit,
			"offset":            offset,
			"total":             total,
			"online_window_sec": 30,
			"server_time":       nowUTC.Format(time.RFC3339),
		},
	})
}

func (h *MonitorHandler) Stream(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	examID := c.Param("id")
	exam, ok, err := h.ex.Get(c.Request.Context(), examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if role == "teacher" {
		tid, ok, err := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || exam.TeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	view := params.StringQueryTrim(c, "view")
	if view == "" {
		view = "sessions"
	}
	if view != "sessions" && view != "participants" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "view must be sessions or participants"}})
		return
	}
	q := params.StringQueryTrim(c, "q")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "streaming unsupported"}})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // nginx

	writeEvent := func(event string, payload any) error {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		// SSE format: event + data + blank line
		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", b); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	_ = writeEvent("hello", gin.H{
		"exam_id":           examID,
		"view":              view,
		"q":                 q,
		"server_time":       time.Now().UTC().Format(time.RFC3339),
		"online_window_sec": 30,
	})

	sendSnapshot := func(nowUTC time.Time) error {
		if view == "sessions" {
			items, total, err := h.st.ListMonitorSessions(c.Request.Context(), examID, studentexamrepo.MonitorSessionsFilter{
				NowUTC:       nowUTC,
				OnlineWindow: 30 * time.Second,
				Limit:        500,
				Offset:       0,
			})
			if err != nil {
				return writeEvent("error", gin.H{"message": "internal error"})
			}
			return writeEvent("snapshot", gin.H{
				"data": items,
				"meta": gin.H{
					"total":       total,
					"server_time": nowUTC.Format(time.RFC3339),
				},
			})
		}

		items, total, err := h.st.ListMonitorParticipants(c.Request.Context(), examID, studentexamrepo.MonitorParticipantsFilter{
			Q:            q,
			NowUTC:       nowUTC,
			OnlineWindow: 30 * time.Second,
			Limit:        500,
			Offset:       0,
		})
		if err != nil {
			return writeEvent("error", gin.H{"message": "internal error"})
		}
		return writeEvent("snapshot", gin.H{
			"data": items,
			"meta": gin.H{
				"q":           q,
				"total":       total,
				"server_time": nowUTC.Format(time.RFC3339),
			},
		})
	}

	if err := sendSnapshot(time.Now().UTC()); err != nil {
		return
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			_ = sendSnapshot(time.Now().UTC())
		}
	}
}

func (h *MonitorHandler) ListParticipants(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	examID := c.Param("id")
	exam, ok, err := h.ex.Get(c.Request.Context(), examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || exam.TeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 100, 1, 500)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	nowUTC := time.Now().UTC()
	items, total, err := h.st.ListMonitorParticipants(c.Request.Context(), examID, studentexamrepo.MonitorParticipantsFilter{
		Q:            q,
		NowUTC:       nowUTC,
		OnlineWindow: 30 * time.Second,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{
		"data": items,
		"meta": gin.H{
			"exam":              exam,
			"q":                 q,
			"limit":             limit,
			"offset":            offset,
			"total":             total,
			"online_window_sec": 30,
			"server_time":       nowUTC.Format(time.RFC3339),
		},
	})
}
