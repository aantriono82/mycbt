package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/repo/studentexamrepo"
)

type notificationStudentRepo interface {
	StudentByUserID(ctx context.Context, userID string) (studentexamrepo.StudentInfo, bool, error)
	ListAvailableForStudent(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentExamsFilter) ([]studentexamrepo.StudentExam, int, error)
	ListStudentAnnouncements(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentAnnouncementsFilter) ([]studentexamrepo.StudentAnnouncement, int, error)
}

type NotificationHandler struct {
	st notificationStudentRepo
}

func NewNotificationHandler(st *studentexamrepo.Repo) *NotificationHandler {
	return &NotificationHandler{st: st}
}

func (h *NotificationHandler) buildSnapshot(ctx context.Context, userID string, nowUTC time.Time) (string, error) {
	st, ok, err := h.st.StudentByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if !ok || !st.IsActive {
		return "", nil
	}

	anns, annTotal, err := h.st.ListStudentAnnouncements(ctx, st.StudentID, st.LevelID, st.GroupID, studentexamrepo.ListStudentAnnouncementsFilter{
		Limit:      5,
		Offset:     0,
		NowUTC:     nowUTC,
		UnreadOnly: true,
	})
	if err != nil {
		return "", err
	}

	exams, examTotal, err := h.st.ListAvailableForStudent(ctx, st.StudentID, st.LevelID, st.GroupID, studentexamrepo.ListStudentExamsFilter{
		Limit:  5,
		Offset: 0,
		NowUTC: nowUTC,
	})
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "ann_total=%d;", annTotal)
	for _, ann := range anns {
		fmt.Fprintf(&b, "ann:%s:%s:%s;", ann.ID, ann.PublishedAt, ann.ExpiresAt)
	}
	fmt.Fprintf(&b, "exam_total=%d;", examTotal)
	for _, exam := range exams {
		fmt.Fprintf(
			&b,
			"exam:%s:%s:%s:%s:%s:%t;",
			exam.ID,
			exam.StartsAt,
			exam.EndsAt,
			exam.SessionStatus,
			exam.ActiveToken,
			exam.CanJoin,
		)
	}
	return b.String(), nil
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

	userID := middleware.GetUserID(c)
	lastSnapshot, err := h.buildSnapshot(c.Request.Context(), userID, time.Now().UTC())
	if err != nil {
		lastSnapshot = ""
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			currentSnapshot, err := h.buildSnapshot(c.Request.Context(), userID, time.Now().UTC())
			if err == nil {
				if currentSnapshot != lastSnapshot {
					_ = writeEvent("update", "notifications")
					lastSnapshot = currentSnapshot
				}
			}

			_ = writeEvent("heartbeat", "heartbeat")
		}
	}
}
