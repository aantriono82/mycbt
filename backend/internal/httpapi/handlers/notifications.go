package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/repo/studentexamrepo"
)

type NotificationHandler struct {
	ann *masterrepo.AnnouncementsRepo
	st  *studentexamrepo.Repo
}

func NewNotificationHandler(ann *masterrepo.AnnouncementsRepo, st *studentexamrepo.Repo) *NotificationHandler {
	return &NotificationHandler{ann: ann, st: st}
}

func (h *NotificationHandler) Stream(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(500, gin.H{"error": "streaming unsupported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	writeEvent := func(event string, data string) error {
		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	_ = writeEvent("connect", "connected")

	// Get initial latest IDs to detect changes
	lastAnnID := ""
	lastExamID := ""

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			// 1. Check for new announcements
			anns, _, err := h.ann.List(c.Request.Context(), "", "true", 1, 0)
			if err == nil && len(anns) > 0 {
				currentAnnID := fmt.Sprintf("%v", anns[0].ID)
				if lastAnnID != "" && currentAnnID != lastAnnID {
					_ = writeEvent("update", "announcement")
				}
				lastAnnID = currentAnnID
			}

			// 2. Check for new exams
			userID := middleware.GetUserID(c)
			st, ok, err := h.st.StudentByUserID(c.Request.Context(), userID)
			if err == nil && ok && st.IsActive {
				exams, _, err := h.st.ListAvailableForStudent(c.Request.Context(), st.StudentID, st.LevelID, st.GroupID, studentexamrepo.ListStudentExamsFilter{
					Limit:  1,
					Offset: 0,
					NowUTC: time.Now().UTC(),
				})
				if err == nil && len(exams) > 0 {
					currentExamID := fmt.Sprintf("%v", exams[0].ID)
					if lastExamID != "" && currentExamID != lastExamID {
						_ = writeEvent("update", "exam")
					}
					lastExamID = currentExamID
				}
			}

			// Ping to keep connection alive
			_ = writeEvent("heartbeat", "heartbeat")
		}
	}
}
