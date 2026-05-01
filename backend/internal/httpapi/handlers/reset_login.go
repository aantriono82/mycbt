package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/repo/examrepo"
	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/repo/studentexamrepo"
)

type ResetLoginHandler struct {
	ex       *examrepo.Repo
	st       *studentexamrepo.Repo
	settings *masterrepo.SettingsRepo
}

func NewResetLoginHandler(ex *examrepo.Repo, st *studentexamrepo.Repo, settings *masterrepo.SettingsRepo) *ResetLoginHandler {
	return &ResetLoginHandler{ex: ex, st: st, settings: settings}
}

func (h *ResetLoginHandler) ResetSession(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	if h.settings != nil {
		sys, err := h.settings.GetSystem(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !sys.AllowResetLogin {
			c.JSON(403, gin.H{"error": gin.H{"code": "policy_disabled", "message": "reset login is disabled by system settings"}})
			return
		}
	}

	examID := c.Param("id")
	sessionID := c.Param("sessionId")

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

	// Only reset sessions that are not submitted.
	const q = `
SELECT s.status
FROM exam_sessions s
WHERE s.id = $1 AND s.exam_id = $2
LIMIT 1`
	var status string
	if err := h.ex.Pool().QueryRow(c.Request.Context(), q, sessionID, examID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if status == "submitted" {
		c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "cannot reset submitted session"}})
		return
	}

	ct, err := h.ex.Pool().Exec(c.Request.Context(), `DELETE FROM exam_sessions WHERE id = $1 AND exam_id = $2`, sessionID, examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if ct.RowsAffected() == 0 {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *ResetLoginHandler) ForceSubmitSession(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	examID := c.Param("id")
	sessionID := c.Param("sessionId")

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

	if h.st == nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	const q = `
SELECT EXISTS (
	SELECT 1
	FROM exam_sessions s
	WHERE s.id = $1 AND s.exam_id = $2
)`
	var exists bool
	if err := h.ex.Pool().QueryRow(c.Request.Context(), q, sessionID, examID).Scan(&exists); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !exists {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if err := h.st.ForceSubmitSession(c.Request.Context(), sessionID); err != nil {
		switch err {
		case studentexamrepo.ErrSessionNotFound:
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		case studentexamrepo.ErrSessionNotActive:
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session not active"}})
		default:
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		}
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}
