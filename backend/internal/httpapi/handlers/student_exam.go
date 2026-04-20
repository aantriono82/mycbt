package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/httpapi/params"
	"mycbt/backend/internal/repo/masterrepo"
	"mycbt/backend/internal/repo/studentexamrepo"
)

type StudentExamHandler struct {
	repo     *studentexamrepo.Repo
	settings *masterrepo.SettingsRepo
}

func NewStudentExamHandler(repo *studentexamrepo.Repo, settings *masterrepo.SettingsRepo) *StudentExamHandler {
	return &StudentExamHandler{repo: repo, settings: settings}
}

func (h *StudentExamHandler) ListExams(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	sys := masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}
	if h.settings != nil {
		if stg, stgErr := h.settings.GetSystem(c.Request.Context()); stgErr == nil {
			sys = stg
		}
	}
	loc, _ := time.LoadLocation(sys.Timezone)

	items, total, err := h.repo.ListAvailableForStudent(c.Request.Context(), st.StudentID, st.LevelID, st.GroupID, studentexamrepo.ListStudentExamsFilter{
		Q:      q,
		Limit:  limit,
		Offset: offset,
		NowUTC: time.Now().UTC(),
		Loc:    loc,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": items, "meta": gin.H{"q": q, "limit": limit, "offset": offset, "total": total}})
}

type joinReq struct {
	Token string `json:"token"`
}

func (h *StudentExamHandler) Join(c *gin.Context) {
	userID := middleware.GetUserID(c)
	nowUTC := time.Now().UTC()

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	sys := masterrepo.SystemSettings{Timezone: "Asia/Jakarta", TokenRequired: true}
	if h.settings != nil {
		if stg, stgErr := h.settings.GetSystem(c.Request.Context()); stgErr == nil {
			sys = stg
		}
	}
	loc, _ := time.LoadLocation(sys.Timezone) // best-effort, if fails loc stays nil (UTC)

	examID := c.Param("id")
	ex, err := h.repo.GetExamForStudentJoin(c.Request.Context(), examID, st.StudentID, st.LevelID, st.GroupID, nowUTC, loc)
	if err != nil {
		if err == studentexamrepo.ErrExamNotFound {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
			return
		}
		if err == studentexamrepo.ErrExamNotJoinable {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "exam not joinable (outside schedule)"}})
			return
		}
		if errors.Is(err, studentexamrepo.ErrSessionTimeMismatch) {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": err.Error()}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	// If an in-progress session already exists, allow re-enter without token.
	existing, ok, err := h.repo.GetSessionByExamStudent(c.Request.Context(), examID, st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if ok {
		if existing.Status != "in_progress" {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session already finished"}})
			return
		}
		if _, err := h.repo.EnsureSessionQuestions(c.Request.Context(), existing.ID, examID, ex.ShuffleQuestions); err != nil {
			if errors.Is(err, studentexamrepo.ErrNoQuestionSets) {
				c.JSON(409, gin.H{"error": gin.H{"code": "no_questions", "message": "Ujian ini belum memiliki bank soal. Silakan hubungi pengampu."}})
				return
			}
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		c.JSON(200, gin.H{"data": gin.H{"session": existing, "session_id": existing.ID, "exam": gin.H{"id": ex.ID, "title": ex.Title}}})
		return
	}

	var req joinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if !(errors.Is(err, io.EOF) && !sys.TokenRequired) {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
			return
		}
	}
	req.Token = strings.TrimSpace(req.Token)
	if sys.TokenRequired && req.Token == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "token required"}})
		return
	}

	if req.Token != "" {
		if err := h.repo.VerifyExamToken(c.Request.Context(), examID, req.Token, nowUTC); err != nil {
			switch err {
			case studentexamrepo.ErrTokenNotFound:
				c.JSON(400, gin.H{"error": gin.H{"code": "invalid_token", "message": "invalid token"}})
			case studentexamrepo.ErrTokenInactive:
				c.JSON(403, gin.H{"error": gin.H{"code": "token_inactive", "message": "token inactive"}})
			case studentexamrepo.ErrTokenNotStarted:
				c.JSON(403, gin.H{"error": gin.H{"code": "token_not_started", "message": "token not started"}})
			case studentexamrepo.ErrTokenExpired:
				c.JSON(403, gin.H{"error": gin.H{"code": "token_expired", "message": "token expired"}})
			default:
				c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			}
			return
		}
	}

	maxActive := sys.MaxActiveSessions
	if maxActive < 1 {
		maxActive = 1
	}
	activeSessions, err := h.repo.CountInProgressSessionsByStudent(c.Request.Context(), st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if activeSessions >= maxActive {
		c.JSON(409, gin.H{
			"error": gin.H{
				"code":    "max_active_sessions_reached",
				"message": "maximum active sessions reached",
			},
			"meta": gin.H{
				"active_sessions": activeSessions,
				"max_allowed":     maxActive,
			},
		})
		return
	}

	ip := net.ParseIP(c.ClientIP())
	sess, err := h.repo.GetOrCreateSession(c.Request.Context(), examID, st.StudentID, ip, c.GetHeader("User-Agent"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if sess.Status != "in_progress" {
		c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session already finished"}})
		return
	}

	if _, err := h.repo.EnsureSessionQuestions(c.Request.Context(), sess.ID, examID, ex.ShuffleQuestions); err != nil {
		if errors.Is(err, studentexamrepo.ErrNoQuestionSets) {
			c.JSON(409, gin.H{"error": gin.H{"code": "no_questions", "message": "Ujian ini belum memiliki bank soal. Silakan hubungi pengampu."}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"session": sess, "session_id": sess.ID, "exam": gin.H{"id": ex.ID, "title": ex.Title}}})
}

func (h *StudentExamHandler) GetActiveSessionByExam(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	examID := c.Param("id")
	sess, ok, err := h.repo.GetSessionByExamStudent(c.Request.Context(), examID, st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if sess.Status != "in_progress" {
		c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session already finished"}})
		return
	}

	c.JSON(200, gin.H{"data": sess})
}

func (h *StudentExamHandler) GetSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	nowUTC := time.Now().UTC()

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	state, ok, err := h.repo.GetSessionState(c.Request.Context(), c.Param("id"), st.StudentID, nowUTC)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": state})
}

func (h *StudentExamHandler) GetQuestions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	nowUTC := time.Now().UTC()

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	state, ok, err := h.repo.GetSessionState(c.Request.Context(), c.Param("id"), st.StudentID, nowUTC)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	items, err := h.repo.ListSessionQuestions(c.Request.Context(), state.Session.ID, st.StudentID, state.Exam.ShuffleOptions)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"session": state.Session, "questions": items}})
}

func (h *StudentExamHandler) GetAnswers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	items, err := h.repo.ListSessionAnswers(c.Request.Context(), c.Param("id"), st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type answerReq struct {
	QuestionID string          `json:"question_id"`
	AnswerJSON json.RawMessage `json:"answer_json"`
}

func (h *StudentExamHandler) SaveAnswer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	nowUTC := time.Now().UTC()

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	var req answerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	req.QuestionID = strings.TrimSpace(req.QuestionID)
	if req.QuestionID == "" || len(req.AnswerJSON) == 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "question_id and answer_json required"}})
		return
	}

	if err := h.repo.UpsertAnswer(c.Request.Context(), c.Param("id"), st.StudentID, req.QuestionID, req.AnswerJSON, nowUTC); err != nil {
		switch err {
		case studentexamrepo.ErrQuestionNotInSession:
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "question not found"}})
		case studentexamrepo.ErrSessionNotActive:
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session not active"}})
		default:
			if strings.Contains(err.Error(), "invalid answer_json") {
				c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid answer_json"}})
				return
			}
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		}
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *StudentExamHandler) Submit(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	if err := h.repo.SubmitSession(c.Request.Context(), c.Param("id"), st.StudentID); err != nil {
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

func (h *StudentExamHandler) Heartbeat(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	body, _ := c.GetRawData()
	var payload json.RawMessage
	if len(body) > 0 {
		payload = json.RawMessage(body)
	}

	if err := h.repo.Heartbeat(c.Request.Context(), c.Param("id"), st.StudentID, payload); err != nil {
		switch err {
		case studentexamrepo.ErrSessionNotFound:
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		default:
			if strings.Contains(err.Error(), "invalid payload_json") {
				c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
				return
			}
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		}
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

func (h *StudentExamHandler) ListResults(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.repo.ListStudentResults(c.Request.Context(), st.StudentID, studentexamrepo.ListStudentResultsFilter{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": items, "meta": gin.H{"limit": limit, "offset": offset, "total": total}})
}

func (h *StudentExamHandler) ListAnnouncements(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 20, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.repo.ListStudentAnnouncements(
		c.Request.Context(),
		st.StudentID,
		st.LevelID,
		st.GroupID,
		studentexamrepo.ListStudentAnnouncementsFilter{
			Q:      q,
			Limit:  limit,
			Offset: offset,
			NowUTC: time.Now().UTC(),
		},
	)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": items, "meta": gin.H{"q": q, "limit": limit, "offset": offset, "total": total}})
}

type attendanceReq struct {
	ExamID string `json:"exam_id"`
	Note   string `json:"note"`
}

func (h *StudentExamHandler) SubmitAttendance(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	var req attendanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	req.ExamID = strings.TrimSpace(req.ExamID)
	req.Note = strings.TrimSpace(req.Note)
	if req.ExamID == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "exam_id required"}})
		return
	}
	if len(req.Note) > 1000 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "note too long (max 1000)"}})
		return
	}

	allowed, err := h.repo.EnsureStudentCanAttendExam(c.Request.Context(), req.ExamID, st.StudentID, st.LevelID, st.GroupID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !allowed {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "exam not found"}})
		return
	}

	sys := masterrepo.SystemSettings{}
	if h.settings != nil {
		if stg, stgErr := h.settings.GetSystem(c.Request.Context()); stgErr == nil {
			sys = stg
		}
	}
	clientIP := net.ParseIP(c.ClientIP())
	if sys.AttendanceRequireIP && clientIP == nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "client_ip_required", "message": "client ip required"}})
		return
	}

	item, err := h.repo.UpsertAttendance(c.Request.Context(), req.ExamID, st.StudentID, req.Note, clientIP, time.Now().UTC())
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": item})
}

func (h *StudentExamHandler) ListAttendanceHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)

	st, ok, err := h.repo.StudentByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "student not registered"}})
		return
	}
	if !st.IsActive {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "user inactive"}})
		return
	}

	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.repo.ListAttendanceHistory(c.Request.Context(), st.StudentID, studentexamrepo.ListAttendanceHistoryFilter{
		Q:      q,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	c.JSON(200, gin.H{"data": items, "meta": gin.H{"q": q, "limit": limit, "offset": offset, "total": total}})
}
