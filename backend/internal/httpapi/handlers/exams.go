package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"atigacbt/backend/internal/httpapi/middleware"
	"atigacbt/backend/internal/httpapi/params"
	"atigacbt/backend/internal/httpapi/pgerr"
	"atigacbt/backend/internal/repo/examrepo"
	"atigacbt/backend/internal/repo/masterrepo"
)

type ExamsHandler struct {
	ex            *examrepo.Repo
	teacherSubs   *masterrepo.TeacherSubjectsRepo
	teacherGroups *masterrepo.TeacherGroupsRepo
	teacherLevels *masterrepo.TeacherLevelsRepo
}

func NewExamsHandler(
	ex *examrepo.Repo,
	teacherSubs *masterrepo.TeacherSubjectsRepo,
	teacherGroups *masterrepo.TeacherGroupsRepo,
	teacherLevels *masterrepo.TeacherLevelsRepo,
) *ExamsHandler {
	return &ExamsHandler{
		ex:            ex,
		teacherSubs:   teacherSubs,
		teacherGroups: teacherGroups,
		teacherLevels: teacherLevels,
	}
}

func (h *ExamsHandler) List(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	status := params.StringQueryTrim(c, "status")
	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	teacherID := ""
	if role == "teacher" {
		tid, ok, err := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "teacher not registered"}})
			return
		}
		teacherID = tid
	}

	items, total, err := h.ex.List(c.Request.Context(), role, teacherID, examrepo.ListFilter{
		Status: status,
		Q:      q,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items, "meta": gin.H{"status": status, "q": q, "limit": limit, "offset": offset, "total": total}})
}

type createExamReq struct {
	SubjectID                string `json:"subject_id"`
	TeacherID                string `json:"teacher_id"` // required for admin
	SessionID                string `json:"session_id"`
	Title                    string `json:"title"`
	StartsAt                 string `json:"starts_at"` // RFC3339
	EndsAt                   string `json:"ends_at"`   // RFC3339
	DurationMinutes          *int   `json:"duration_minutes"`
	ShuffleQuestions         bool   `json:"shuffle_questions"`
	ShuffleOptions           bool   `json:"shuffle_options"`
	ScoringMode              string `json:"scoring_mode"` // partial|absolute
	MaxAttempts              *int   `json:"max_attempts"`
	PassingScore             *int   `json:"passing_score"`
	ShowDiscussionToStudents *bool  `json:"show_discussion_to_students"`
}

func (h *ExamsHandler) Create(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	var req createExamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	req.SubjectID = strings.TrimSpace(req.SubjectID)
	req.TeacherID = strings.TrimSpace(req.TeacherID)
	req.Title = strings.TrimSpace(req.Title)
	if req.SubjectID == "" || req.Title == "" || strings.TrimSpace(req.StartsAt) == "" || strings.TrimSpace(req.EndsAt) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "subject_id, title, starts_at, ends_at required"}})
		return
	}
	if req.DurationMinutes != nil && *req.DurationMinutes <= 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "duration_minutes must be > 0"}})
		return
	}
	maxAttempts := 1
	if req.MaxAttempts != nil {
		maxAttempts = *req.MaxAttempts
	}
	if maxAttempts <= 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "max_attempts must be > 0"}})
		return
	}
	passingScore := 75
	if req.PassingScore != nil {
		passingScore = *req.PassingScore
	}
	if passingScore < 0 || passingScore > 100 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "passing_score must be between 0 and 100"}})
		return
	}
	showDiscussion := false
	if req.ShowDiscussionToStudents != nil {
		showDiscussion = *req.ShowDiscussionToStudents
	}
	scoringMode := strings.TrimSpace(strings.ToLower(req.ScoringMode))
	if scoringMode == "" {
		scoringMode = "partial"
	}
	if scoringMode != "partial" && scoringMode != "absolute" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid scoring_mode"}})
		return
	}

	startsAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.StartsAt))
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid starts_at (RFC3339)"}})
		return
	}
	endsAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EndsAt))
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid ends_at (RFC3339)"}})
		return
	}
	if !endsAt.After(startsAt) {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "ends_at must be after starts_at"}})
		return
	}

	teacherID := ""
	if role == "teacher" {
		tid, ok, err := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "teacher not registered"}})
			return
		}
		teacherID = tid

		allowed, err := h.teacherSubs.Has(c.Request.Context(), teacherID, req.SubjectID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !allowed {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "subject not assigned to teacher"}})
			return
		}
	} else {
		if req.TeacherID == "" {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "teacher_id required for admin"}})
			return
		}
		teacherID = req.TeacherID
	}

	var sessID *string
	if strings.TrimSpace(req.SessionID) != "" {
		s := strings.TrimSpace(req.SessionID)
		sessID = &s
	}

	it, err := h.ex.Create(c.Request.Context(), examrepo.CreateInput{
		SubjectID:                req.SubjectID,
		TeacherID:                teacherID,
		SessionID:                sessID,
		Title:                    req.Title,
		StartsAt:                 startsAt,
		EndsAt:                   endsAt,
		DurationMinutes:          req.DurationMinutes,
		ShuffleQuestions:         req.ShuffleQuestions,
		ShuffleOptions:           req.ShuffleOptions,
		ScoringMode:              scoringMode,
		MaxAttempts:              maxAttempts,
		PassingScore:             passingScore,
		ShowDiscussionToStudents: showDiscussion,
	})
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid subject_id or teacher_id"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *ExamsHandler) Get(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	it, ok, err := h.ex.Get(c.Request.Context(), c.Param("id"))
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
		if !ok || it.TeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	c.JSON(200, gin.H{"data": it})
}

type patchExamReq struct {
	SessionID                *string `json:"session_id"`
	Title                    *string `json:"title"`
	StartsAt                 *string `json:"starts_at"`
	EndsAt                   *string `json:"ends_at"`
	DurationMinutes          **int   `json:"duration_minutes"` // can set null
	ShuffleQuestions         *bool   `json:"shuffle_questions"`
	ShuffleOptions           *bool   `json:"shuffle_options"`
	ScoringMode              *string `json:"scoring_mode"` // partial|absolute
	MaxAttempts              *int    `json:"max_attempts"`
	PassingScore             *int    `json:"passing_score"`
	ShowDiscussionToStudents *bool   `json:"show_discussion_to_students"`
	Status                   *string `json:"status"` // draft|published|archived
}

func (h *ExamsHandler) Patch(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	cur, ok, err := h.ex.Get(c.Request.Context(), c.Param("id"))
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
		if !ok || cur.TeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	var req patchExamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	title := cur.Title
	if req.Title != nil {
		title = strings.TrimSpace(*req.Title)
	}
	if title == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "title required"}})
		return
	}

	// Parse current times to time.Time for update.
	startsAt, err := time.Parse(time.RFC3339, cur.StartsAt)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	endsAt, err := time.Parse(time.RFC3339, cur.EndsAt)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if req.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.StartsAt))
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid starts_at (RFC3339)"}})
			return
		}
		startsAt = t
	}
	if req.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.EndsAt))
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid ends_at (RFC3339)"}})
			return
		}
		endsAt = t
	}
	if !endsAt.After(startsAt) {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "ends_at must be after starts_at"}})
		return
	}

	dur := cur.DurationMinutes
	if req.DurationMinutes != nil {
		dur = *req.DurationMinutes
	}
	if dur != nil && *dur <= 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "duration_minutes must be > 0 or null"}})
		return
	}

	shQ := cur.ShuffleQuestions
	if req.ShuffleQuestions != nil {
		shQ = *req.ShuffleQuestions
	}
	shO := cur.ShuffleOptions
	if req.ShuffleOptions != nil {
		shO = *req.ShuffleOptions
	}
	scoringMode := cur.ScoringMode
	if scoringMode == "" {
		scoringMode = "partial"
	}
	if req.ScoringMode != nil {
		scoringMode = strings.TrimSpace(strings.ToLower(*req.ScoringMode))
	}
	if scoringMode != "partial" && scoringMode != "absolute" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid scoring_mode"}})
		return
	}
	maxAttempts := cur.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	if req.MaxAttempts != nil {
		maxAttempts = *req.MaxAttempts
	}
	if maxAttempts <= 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "max_attempts must be > 0"}})
		return
	}
	passingScore := cur.PassingScore
	if passingScore < 0 || passingScore > 100 {
		passingScore = 75
	}
	if req.PassingScore != nil {
		passingScore = *req.PassingScore
	}
	if passingScore < 0 || passingScore > 100 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "passing_score must be between 0 and 100"}})
		return
	}
	showDiscussion := cur.ShowDiscussionToStudents
	if req.ShowDiscussionToStudents != nil {
		showDiscussion = *req.ShowDiscussionToStudents
	}
	status := cur.Status
	if req.Status != nil {
		status = strings.TrimSpace(*req.Status)
	}
	if status != "draft" && status != "published" && status != "archived" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid status"}})
		return
	}

	var sessID *string
	if cur.SessionID != nil {
		sessID = cur.SessionID
	}
	if req.SessionID != nil {
		s := strings.TrimSpace(*req.SessionID)
		if s == "" {
			sessID = nil
		} else {
			sessID = &s
		}
	}

	it, ok, err := h.ex.Update(c.Request.Context(), c.Param("id"), examrepo.UpdateInput{
		SessionID:                sessID,
		Title:                    title,
		StartsAt:                 startsAt,
		EndsAt:                   endsAt,
		DurationMinutes:          dur,
		ShuffleQuestions:         shQ,
		ShuffleOptions:           shO,
		ScoringMode:              scoringMode,
		MaxAttempts:              maxAttempts,
		PassingScore:             passingScore,
		ShowDiscussionToStudents: showDiscussion,
		Status:                   status,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it})
}

func (h *ExamsHandler) Delete(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	cur, ok, err := h.ex.Get(c.Request.Context(), c.Param("id"))
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
		if !ok || cur.TeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	ok, err = h.ex.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true}})
}

type createTokenReq struct {
	ValidFrom *string `json:"valid_from"` // RFC3339 (optional)
	ValidTo   *string `json:"valid_to"`   // RFC3339 (optional)
	Length    int     `json:"length"`     // optional (default 6)
}

func (h *ExamsHandler) CreateToken(c *gin.Context) {
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

	var req createTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	var vf *time.Time
	var vt *time.Time
	if req.ValidFrom != nil && strings.TrimSpace(*req.ValidFrom) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ValidFrom))
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid valid_from (RFC3339)"}})
			return
		}
		vf = &t
	}
	if req.ValidTo != nil && strings.TrimSpace(*req.ValidTo) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ValidTo))
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid valid_to (RFC3339)"}})
			return
		}
		vt = &t
	}
	if vf != nil && vt != nil && !vt.After(*vf) {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "valid_to must be after valid_from"}})
		return
	}

	it, err := h.ex.CreateToken(c.Request.Context(), examrepo.CreateTokenInput{
		ExamID:          examID,
		ValidFrom:       vf,
		ValidTo:         vt,
		CreatedByUserID: userID,
		Length:          req.Length,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *ExamsHandler) ListTokens(c *gin.Context) {
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

	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	items, total, err := h.ex.ListTokens(c.Request.Context(), examID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items, "meta": gin.H{"limit": limit, "offset": offset, "total": total}})
}

type patchTokenReq struct {
	IsActive *bool `json:"is_active"`
}

func (h *ExamsHandler) PatchToken(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	// Authorize via exam ownership
	const q = `
SELECT et.exam_id::text, e.teacher_id::text
FROM exam_tokens et
JOIN exams e ON e.id = et.exam_id
WHERE et.id = $1
LIMIT 1`
	var examID string
	var teacherID string
	if err := h.ex.Pool().QueryRow(c.Request.Context(), q, c.Param("id")).Scan(&examID, &teacherID); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || teacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	var req patchTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if req.IsActive == nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "is_active required"}})
		return
	}

	it, ok, err := h.ex.SetTokenActive(c.Request.Context(), c.Param("id"), *req.IsActive)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": it, "meta": gin.H{"exam_id": examID}})
}

func (h *ExamsHandler) DeleteToken(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	// Authorize via exam ownership
	const q = `
SELECT et.exam_id::text, e.teacher_id::text
FROM exam_tokens et
JOIN exams e ON e.id = et.exam_id
WHERE et.id = $1
LIMIT 1`
	var examID string
	var teacherID string
	if err := h.ex.Pool().QueryRow(c.Request.Context(), q, c.Param("id")).Scan(&examID, &teacherID); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.ex.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || teacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	ok, err := h.ex.DeleteToken(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	c.JSON(200, gin.H{"data": gin.H{"ok": true}, "meta": gin.H{"exam_id": examID}})
}

func (h *ExamsHandler) DeactivateAllTokens(c *gin.Context) {
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

	affected, err := h.ex.DeactivateAllTokens(c.Request.Context(), examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"ok": true, "deactivated": affected}})
}

type rotateTokenReq struct {
	ValidFrom        *string `json:"valid_from"` // RFC3339 (optional)
	ValidTo          *string `json:"valid_to"`   // RFC3339 (optional)
	Length           int     `json:"length"`     // optional (default 6)
	DeactivateOthers *bool   `json:"deactivate_others"`
}

func (h *ExamsHandler) RotateToken(c *gin.Context) {
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

	var req rotateTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	var vf *time.Time
	var vt *time.Time
	if req.ValidFrom != nil && strings.TrimSpace(*req.ValidFrom) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ValidFrom))
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid valid_from (RFC3339)"}})
			return
		}
		vf = &t
	}
	if req.ValidTo != nil && strings.TrimSpace(*req.ValidTo) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ValidTo))
		if err != nil {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid valid_to (RFC3339)"}})
			return
		}
		vt = &t
	}
	if vf != nil && vt != nil && !vt.After(*vf) {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "valid_to must be after valid_from"}})
		return
	}

	deactivate := true
	if req.DeactivateOthers != nil {
		deactivate = *req.DeactivateOthers
	}

	it, err := h.ex.RotateToken(c.Request.Context(), examrepo.RotateTokenInput{
		ExamID:           examID,
		ValidFrom:        vf,
		ValidTo:          vt,
		CreatedByUserID:  userID,
		Length:           req.Length,
		DeactivateOthers: deactivate,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *ExamsHandler) GetTargets(c *gin.Context) {
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

	items, err := h.ex.ListTargets(c.Request.Context(), examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type replaceTargetsReq struct {
	LevelIDs   []string `json:"level_ids"`
	GroupIDs   []string `json:"group_ids"`
	StudentIDs []string `json:"student_ids"`
}

func (h *ExamsHandler) PutTargets(c *gin.Context) {
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

	var req replaceTargetsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	levelIDs := uniqTrim(req.LevelIDs)
	groupIDs := uniqTrim(req.GroupIDs)
	studentIDs := uniqTrim(req.StudentIDs)
	if len(levelIDs)+len(groupIDs)+len(studentIDs) == 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "at least one target id required"}})
		return
	}

	if role == "teacher" {
		tid, _, _ := h.ex.TeacherIDByUserID(c.Request.Context(), userID) // guaranteed to succeed because of check above
		for _, lid := range levelIDs {
			ok, err := h.teacherLevels.Has(c.Request.Context(), tid, lid)
			if err != nil {
				c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
				return
			}
			if !ok {
				c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "not assigned to level " + lid}})
				return
			}
		}
		for _, gid := range groupIDs {
			ok, err := h.teacherGroups.Has(c.Request.Context(), tid, gid)
			if err != nil {
				c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
				return
			}
			if !ok {
				c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "not assigned to group " + gid}})
				return
			}
		}
	}

	items, err := h.ex.ReplaceTargets(c.Request.Context(), examrepo.ReplaceTargetsInput{
		ExamID:     examID,
		LevelIDs:   levelIDs,
		GroupIDs:   groupIDs,
		StudentIDs: studentIDs,
	})
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid level_id/group_id/student_id"}})
			return
		}
		if pgerr.Code(err) == pgerr.CodeUniqueViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "duplicate target"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

func (h *ExamsHandler) GetQuestionSets(c *gin.Context) {
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

	items, err := h.ex.ListQuestionSets(c.Request.Context(), examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type replaceQuestionSetsReq struct {
	Items []examrepo.ExamQuestionSet `json:"items"`
}

func (h *ExamsHandler) PutQuestionSets(c *gin.Context) {
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

	teacherID := ""
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
		teacherID = tid
	}

	var req replaceQuestionSetsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "items required"}})
		return
	}

	items := make([]examrepo.ExamQuestionSet, 0, len(req.Items))
	seen := map[string]bool{}
	for _, it := range req.Items {
		it.QuestionSetID = strings.TrimSpace(it.QuestionSetID)
		if it.QuestionSetID == "" {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "question_set_id required"}})
			return
		}
		if seen[it.QuestionSetID] {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "duplicate question_set_id"}})
			return
		}
		seen[it.QuestionSetID] = true
		if it.NumQuestions != nil && *it.NumQuestions <= 0 {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "num_questions must be > 0 or null"}})
			return
		}
		items = append(items, it)
	}

	// Teacher can only attach their own question sets.
	if role == "teacher" {
		const q = `
SELECT COUNT(*)
FROM question_sets
WHERE id::text = ANY($1::text[])
  AND owner_teacher_id = $2::uuid`
		ids := make([]string, 0, len(items))
		for _, it := range items {
			ids = append(ids, it.QuestionSetID)
		}
		var allowed int
		if err := h.ex.Pool().QueryRow(c.Request.Context(), q, ids, teacherID).Scan(&allowed); err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if allowed != len(ids) {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "one or more question_set not owned by teacher"}})
			return
		}
	}

	// Enforce subject consistency (prevents attaching a bank soal from a different subject).
	{
		const q = `
SELECT COUNT(*)
FROM question_sets
WHERE id::text = ANY($1::text[])
  AND subject_id = $2::uuid`
		ids := make([]string, 0, len(items))
		for _, it := range items {
			ids = append(ids, it.QuestionSetID)
		}
		var okCount int
		if err := h.ex.Pool().QueryRow(c.Request.Context(), q, ids, exam.SubjectID).Scan(&okCount); err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if okCount != len(ids) {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "one or more question_set subject_id does not match exam subject_id"}})
			return
		}
	}

	out, err := h.ex.ReplaceQuestionSets(c.Request.Context(), examrepo.ReplaceQuestionSetsInput{
		ExamID: examID,
		Items:  items,
	})
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid question_set_id"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": out})
}

func uniqTrim(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it == "" || seen[it] {
			continue
		}
		seen[it] = true
		out = append(out, it)
	}
	return out
}
