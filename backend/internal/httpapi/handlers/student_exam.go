package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/httpapi/params"
	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/repo/studentexamrepo"
)

type studentExamRepo interface {
	StudentByUserID(ctx context.Context, userID string) (studentexamrepo.StudentInfo, bool, error)
	ListAvailableForStudent(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentExamsFilter) ([]studentexamrepo.StudentExam, int, error)
	DismissExamCard(ctx context.Context, examID, studentID string) error
	VerifyExamToken(ctx context.Context, examID, token string, nowUTC time.Time) error
	GetExamForStudentJoin(ctx context.Context, examID, studentID, levelID, groupID string, nowUTC time.Time, loc *time.Location) (studentexamrepo.ExamForJoin, error)
	GetSessionByExamStudent(ctx context.Context, examID, studentID string) (studentexamrepo.Session, bool, error)
	CountAttemptsByExamStudent(ctx context.Context, examID, studentID string) (int, error)
	CountInProgressSessionsByStudent(ctx context.Context, studentID string) (int, error)
	CreateSessionAttempt(ctx context.Context, examID, studentID string, clientIP net.IP, userAgent string) (studentexamrepo.Session, error)
	GetOrCreateSession(ctx context.Context, examID, studentID string, clientIP net.IP, userAgent string) (studentexamrepo.Session, error)
	EnsureSessionQuestions(ctx context.Context, sessionID, examID string, shuffleQuestions bool) (int, error)
	GetSessionState(ctx context.Context, sessionID, studentID string, nowUTC time.Time) (studentexamrepo.SessionState, bool, error)
	SessionExists(ctx context.Context, sessionID string) (bool, error)
	ListSessionQuestions(ctx context.Context, sessionID, studentID string, shuffleOptions bool) ([]studentexamrepo.StudentQuestion, error)
	ListSessionAnswers(ctx context.Context, sessionID, studentID string) ([]studentexamrepo.SessionAnswer, error)
	UpsertAnswer(ctx context.Context, sessionID, studentID, questionID string, answerJSON json.RawMessage, nowUTC time.Time) error
	SubmitSession(ctx context.Context, sessionID, studentID string) error
	ComputeAutoScore(ctx context.Context, sessionID, studentID string, nowUTC time.Time) (studentexamrepo.AutoScoreSummary, error)
	Heartbeat(ctx context.Context, sessionID, studentID string, payload json.RawMessage) error
	ListStudentResults(ctx context.Context, studentID string, f studentexamrepo.ListStudentResultsFilter) ([]studentexamrepo.StudentResultSummary, int, error)
	ListStudentAnnouncements(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentAnnouncementsFilter) ([]studentexamrepo.StudentAnnouncement, int, error)
	EnsureStudentCanAttendExam(ctx context.Context, examID, studentID, levelID, groupID string) (bool, error)
	UpsertAttendance(ctx context.Context, examID, studentID, note string, clientIP net.IP, nowUTC time.Time, opts ...studentexamrepo.AttendanceOption) (studentexamrepo.AttendanceSubmission, error)
	ListAttendanceHistory(ctx context.Context, studentID string, f studentexamrepo.ListAttendanceHistoryFilter) ([]studentexamrepo.AttendanceHistoryItem, int, error)
}

type systemSettingsRepo interface {
	GetSystem(ctx context.Context) (masterrepo.SystemSettings, error)
}

type StudentExamHandler struct {
	repo     studentExamRepo
	settings systemSettingsRepo
}

func NewStudentExamHandler(repo studentExamRepo, settings systemSettingsRepo) *StudentExamHandler {
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

func (h *StudentExamHandler) DismissExamCard(c *gin.Context) {
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

	examID := strings.TrimSpace(c.Param("id"))
	if examID == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "exam id required"}})
		return
	}

	if err := h.repo.DismissExamCard(c.Request.Context(), examID, st.StudentID); err != nil {
		switch err {
		case studentexamrepo.ErrExamNotDismissible:
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "exam card can only be dismissed after exam has finished"}})
		default:
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		}
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

type joinReq struct {
	Token string `json:"token"`
}

func (h *StudentExamHandler) respondSessionNotOwnedOrMissing(c *gin.Context, sessionID string) {
	exists, err := h.repo.SessionExists(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if exists {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
		return
	}
	c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
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
			// finished/expired/forced: continue and create a new attempt if still allowed.
		} else {
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
	attempts, err := h.repo.CountAttemptsByExamStudent(c.Request.Context(), examID, st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if ex.MaxAttempts <= 0 {
		ex.MaxAttempts = 1
	}
	if attempts >= ex.MaxAttempts {
		c.JSON(409, gin.H{
			"error": gin.H{
				"code":    "max_attempts_reached",
				"message": "maximum exam attempts reached",
			},
			"meta": gin.H{
				"attempts_used": attempts,
				"max_attempts":  ex.MaxAttempts,
			},
		})
		return
	}

	ip := net.ParseIP(c.ClientIP())
	sess, err := h.repo.CreateSessionAttempt(c.Request.Context(), examID, st.StudentID, ip, c.GetHeader("User-Agent"))
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

// StartCompat provides compatibility with POST /api/v1/exams/:id/start.
// Body may contain user_id but this handler relies on authenticated user context.
func (h *StudentExamHandler) StartCompat(c *gin.Context) {
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

	// Best effort: accept legacy body payload { "user_id": ... } without enforcing it.
	var _req map[string]any
	_ = c.ShouldBindJSON(&_req)

	examID := c.Param("id")
	existing, exists, err := h.repo.GetSessionByExamStudent(c.Request.Context(), examID, st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if exists && existing.Status == "in_progress" {
		c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session already active"}})
		return
	}

	sys := masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}
	if h.settings != nil {
		if stg, stgErr := h.settings.GetSystem(c.Request.Context()); stgErr == nil {
			sys = stg
		}
	}
	loc, _ := time.LoadLocation(sys.Timezone)

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

	attempts, err := h.repo.CountAttemptsByExamStudent(c.Request.Context(), examID, st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if ex.MaxAttempts <= 0 {
		ex.MaxAttempts = 1
	}
	if attempts >= ex.MaxAttempts {
		c.JSON(409, gin.H{
			"error": gin.H{"code": "max_attempts_reached", "message": "maximum exam attempts reached"},
			"meta":  gin.H{"attempts_used": attempts, "max_attempts": ex.MaxAttempts},
		})
		return
	}

	ip := net.ParseIP(c.ClientIP())
	sess, err := h.repo.CreateSessionAttempt(c.Request.Context(), examID, st.StudentID, ip, c.GetHeader("User-Agent"))
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

	state, ok, err := h.repo.GetSessionState(c.Request.Context(), sess.ID, st.StudentID, nowUTC)
	if err != nil || !ok {
		c.JSON(200, gin.H{"data": gin.H{"session_id": sess.ID, "session_token": sess.ID}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"session_id": sess.ID, "session_token": sess.ID, "expired_at": state.DeadlineAt}})
}

// GetQuestionsByExamCompat provides compatibility with GET /api/v1/exams/:id/questions.
func (h *StudentExamHandler) GetQuestionsByExamCompat(c *gin.Context) {
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

	sess, ok, err := h.repo.GetSessionByExamStudent(c.Request.Context(), c.Param("id"), st.StudentID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if sess.Status != "in_progress" {
		c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session not active"}})
		return
	}

	state, ok, err := h.repo.GetSessionState(c.Request.Context(), sess.ID, st.StudentID, nowUTC)
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

	limit := params.IntQuery(c, "limit", 50, 1, 500)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)
	total := len(items)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	c.JSON(200, gin.H{
		"data": items[offset:end],
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// SaveAnswerCompat provides compatibility with POST /api/v1/sessions/:session_id/answers.
func (h *StudentExamHandler) SaveAnswerCompat(c *gin.Context) {
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

	if err := h.repo.UpsertAnswer(c.Request.Context(), c.Param("session_id"), st.StudentID, req.QuestionID, req.AnswerJSON, nowUTC); err != nil {
		switch err {
		case studentexamrepo.ErrQuestionNotInSession:
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
		case studentexamrepo.ErrSessionNotActive:
			c.JSON(422, gin.H{"error": gin.H{"code": "exam_time_expired", "message": "exam time expired"}})
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

// FinishCompat provides compatibility with POST /api/v1/sessions/:session_id/finish.
func (h *StudentExamHandler) FinishCompat(c *gin.Context) {
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

	sessionID := c.Param("session_id")
	if err := h.repo.SubmitSession(c.Request.Context(), sessionID, st.StudentID); err != nil {
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

	sum, err := h.repo.ComputeAutoScore(c.Request.Context(), sessionID, st.StudentID, nowUTC)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	score := sum.Score

	state, ok, err := h.repo.GetSessionState(c.Request.Context(), sessionID, st.StudentID, nowUTC)
	if err != nil || !ok {
		c.JSON(200, gin.H{"data": gin.H{"ok": true, "score": score}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true, "score": score, "finished_at": state.Session.FinishedAt}})
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
		h.respondSessionNotOwnedOrMissing(c, c.Param("id"))
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
		h.respondSessionNotOwnedOrMissing(c, c.Param("id"))
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

	// Ensure session belongs to this student; do not leak other users' sessions.
	if _, ok, err := h.repo.GetSessionState(c.Request.Context(), c.Param("id"), st.StudentID, nowUTC); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	} else if !ok {
		h.respondSessionNotOwnedOrMissing(c, c.Param("id"))
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

type verifyTokenReq struct {
	Token string `json:"token"`
}

func (h *StudentExamHandler) VerifyToken(c *gin.Context) {
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

	// Always require token entry for this endpoint (UI gate).
	var req verifyTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "token required"}})
		return
	}

	state, ok, err := h.repo.GetSessionState(c.Request.Context(), c.Param("id"), st.StudentID, nowUTC)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		h.respondSessionNotOwnedOrMissing(c, c.Param("id"))
		return
	}
	if state.Session.Status != "in_progress" {
		c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "session not active"}})
		return
	}

	if err := h.repo.VerifyExamToken(c.Request.Context(), state.Session.ExamID, req.Token, nowUTC); err != nil {
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

	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
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
