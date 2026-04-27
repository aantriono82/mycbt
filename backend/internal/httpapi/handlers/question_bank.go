package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/httpapi/params"
	"mycbt/backend/internal/httpapi/pgerr"
	"mycbt/backend/internal/repo/masterrepo"
	"mycbt/backend/internal/repo/questionbankrepo"
)

type QuestionBankHandler struct {
	qb          *questionbankrepo.Repo
	teacherSubs *masterrepo.TeacherSubjectsRepo
}

func NewQuestionBankHandler(qb *questionbankrepo.Repo, teacherSubs *masterrepo.TeacherSubjectsRepo) *QuestionBankHandler {
	return &QuestionBankHandler{qb: qb, teacherSubs: teacherSubs}
}

func (h *QuestionBankHandler) ListSets(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	subjectID := params.StringQueryTrim(c, "subject_id")
	q := params.StringQueryTrim(c, "q")
	limit := params.IntQuery(c, "limit", 50, 1, 200)
	offset := params.IntQuery(c, "offset", 0, 0, 1_000_000)

	teacherID := ""
	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
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

	items, total, err := h.qb.ListSets(c.Request.Context(), role, teacherID, subjectID, q, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items, "meta": gin.H{"subject_id": subjectID, "q": q, "limit": limit, "offset": offset, "total": total}})
}

type createSetReq struct {
	SubjectID string `json:"subject_id"`
	Title     string `json:"title"`
	Jenjang   string `json:"jenjang"`
	LevelID   string `json:"level_id"`
}

func (h *QuestionBankHandler) CreateSet(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	var req createSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	if strings.TrimSpace(req.SubjectID) == "" || strings.TrimSpace(req.Title) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "subject_id and title required"}})
		return
	}

	ownerTeacherID := ""
	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "teacher not registered"}})
			return
		}
		ownerTeacherID = tid

		allowed, err := h.teacherSubs.Has(c.Request.Context(), ownerTeacherID, strings.TrimSpace(req.SubjectID))
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !allowed {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "subject not assigned to teacher"}})
			return
		}
	}

	it, err := h.qb.CreateSet(c.Request.Context(), strings.TrimSpace(req.SubjectID), ownerTeacherID, strings.TrimSpace(req.Title), strings.TrimSpace(req.Jenjang), strings.TrimSpace(req.LevelID))
	if err != nil {
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid subject_id"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *QuestionBankHandler) GetSet(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	it, ok, err := h.qb.GetSet(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || it.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	c.JSON(200, gin.H{"data": it})
}

type patchSetReq struct {
	Title   string `json:"title"`
	Status  string `json:"status"` // draft | published
	Jenjang string `json:"jenjang"`
	LevelID string `json:"level_id"`
}

func (h *QuestionBankHandler) PatchSet(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	var req patchSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Title) == "" || (req.Status != "draft" && req.Status != "published") {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "title and valid status required"}})
		return
	}

	cur, ok, err := h.qb.GetSet(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || cur.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	it, ok, err := h.qb.UpdateSet(c.Request.Context(), c.Param("id"), strings.TrimSpace(req.Title), req.Status, strings.TrimSpace(req.Jenjang), strings.TrimSpace(req.LevelID))
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

func (h *QuestionBankHandler) DeleteSet(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	cur, ok, err := h.qb.GetSet(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || cur.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	ok, err = h.qb.DeleteSet(c.Request.Context(), c.Param("id"))
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

func (h *QuestionBankHandler) ListQuestions(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	setID := c.Param("id")
	set, ok, err := h.qb.GetSet(c.Request.Context(), setID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || set.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	items, err := h.qb.ListQuestions(c.Request.Context(), setID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(200, gin.H{"data": items})
}

type createQuestionReq struct {
	Type       string                            `json:"type"`
	Stem       string                            `json:"stem"`
	OrderNo    int                               `json:"order_no"`
	Weight     *int                              `json:"weight"`
	Options    []questionbankrepo.QuestionOption `json:"options"`
	Pairs      []questionbankrepo.MatchingPair   `json:"pairs"`
	Answers    []questionbankrepo.ShortAnswer    `json:"answers"`
	Correct    *bool                             `json:"correct"`
	RubricText string                            `json:"rubric_text"`
	MaxScore   *int                              `json:"max_score"`
	Statements []questionbankrepo.TFStatement    `json:"statements"`
}

func (h *QuestionBankHandler) CreateQuestion(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	setID := c.Param("id")
	set, ok, err := h.qb.GetSet(c.Request.Context(), setID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || set.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	var req createQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.Stem) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "type and stem required"}})
		return
	}

	in, err := validateAndBuildCreateQuestionInput(req)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": err.Error()}})
		return
	}

	it, err := h.qb.CreateQuestion(c.Request.Context(), setID, in)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	c.JSON(201, gin.H{"data": it})
}

func (h *QuestionBankHandler) DeleteQuestion(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	// Authorize via set ownership
	// (cheap join: find set owner from question id)
	const q = `
SELECT qs.id, COALESCE(qs.owner_teacher_id::text,'')
FROM questions qu
JOIN question_sets qs ON qs.id = qu.question_set_id
WHERE qu.id = $1
LIMIT 1`
	var setID string
	var ownerTeacherID string
	if err := h.qb.Pool().QueryRow(c.Request.Context(), q, c.Param("id")).Scan(&setID, &ownerTeacherID); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
			return
		}
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || ownerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	ok, err := h.qb.DeleteQuestion(c.Request.Context(), c.Param("id"))
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

func (h *QuestionBankHandler) GetQuestion(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	it, ok, err := h.qb.GetQuestion(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	set, ok, err := h.qb.GetSet(c.Request.Context(), it.QuestionSetID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || set.OwnerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	c.JSON(200, gin.H{"data": it, "meta": gin.H{"question_set_id": it.QuestionSetID}})
}

type patchQuestionReq struct {
	Type       *string                            `json:"type"`
	Stem       *string                            `json:"stem"`
	OrderNo    *int                               `json:"order_no"`
	Weight     *int                               `json:"weight"`
	Options    *[]questionbankrepo.QuestionOption `json:"options"`
	Pairs      *[]questionbankrepo.MatchingPair   `json:"pairs"`
	Answers    *[]questionbankrepo.ShortAnswer    `json:"answers"`
	Correct    *bool                              `json:"correct"`
	RubricText *string                            `json:"rubric_text"`
	MaxScore   nullableInt                        `json:"max_score"`
	Statements *[]questionbankrepo.TFStatement    `json:"statements"`
}

func (h *QuestionBankHandler) PatchQuestion(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userID := middleware.GetUserID(c)

	// Authorize via set ownership
	const q = `
SELECT qs.id, COALESCE(qs.owner_teacher_id::text,'')
FROM questions qu
JOIN question_sets qs ON qs.id = qu.question_set_id
WHERE qu.id = $1
LIMIT 1`
	var setID string
	var ownerTeacherID string
	if err := h.qb.Pool().QueryRow(c.Request.Context(), q, c.Param("id")).Scan(&setID, &ownerTeacherID); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
			return
		}
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	if role == "teacher" {
		tid, ok, err := h.qb.TeacherIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
			return
		}
		if !ok || ownerTeacherID != tid {
			c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "forbidden"}})
			return
		}
	}

	cur, ok, err := h.qb.GetQuestion(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}

	var req patchQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}

	next := cur
	if req.Type != nil {
		next.Type = strings.TrimSpace(*req.Type)
	}
	if req.Stem != nil {
		next.Stem = strings.TrimSpace(*req.Stem)
	}
	if req.OrderNo != nil {
		next.OrderNo = *req.OrderNo
	}
	if req.Weight != nil {
		if *req.Weight <= 0 {
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "weight must be > 0"}})
			return
		}
		next.Weight = *req.Weight
	}
	if req.Options != nil {
		next.Options = *req.Options
	}
	if req.Pairs != nil {
		next.MatchingPairs = *req.Pairs
	}
	if req.Answers != nil {
		next.ShortAnswers = *req.Answers
	}
	if req.Correct != nil {
		next.TrueFalse = &questionbankrepo.TrueFalse{Correct: *req.Correct}
	}
	if req.Statements != nil {
		next.Statements = *req.Statements
	}
	if req.RubricText != nil || req.MaxScore.Set {
		es := &questionbankrepo.Essay{}
		if cur.Essay != nil {
			*es = *cur.Essay
		}
		if req.RubricText != nil {
			es.RubricText = strings.TrimSpace(*req.RubricText)
		}
		if req.MaxScore.Set {
			es.MaxScore = req.MaxScore.Value
		}
		next.Essay = es
	}

	// If type changes, require payload for the new type (avoid silent invalid transformations).
	if req.Type != nil && cur.Type != next.Type {
		switch next.Type {
		case "mc_single", "mc_multiple":
			if req.Options == nil {
				c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "options required when changing type to mc_*"}})
				return
			}
		case "matching":
			if req.Pairs == nil {
				c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "pairs required when changing type to matching"}})
				return
			}
		case "short_answer":
			if req.Answers == nil {
				c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "answers required when changing type to short_answer"}})
				return
			}
		case "true_false":
			if req.Correct == nil {
				c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "correct required when changing type to true_false"}})
				return
			}
		case "essay":
			// No required payload
		default:
			c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid type"}})
			return
		}
	}

	// Validate final shape.
	cReq := createQuestionReq{
		Type:       next.Type,
		Stem:       next.Stem,
		OrderNo:    next.OrderNo,
		Weight:     &next.Weight,
		Options:    next.Options,
		Pairs:      next.MatchingPairs,
		Answers:    next.ShortAnswers,
		Correct:    nil,
		RubricText: "",
		MaxScore:   nil,
		Statements: next.Statements,
	}
	if next.TrueFalse != nil {
		cReq.Correct = &next.TrueFalse.Correct
	}
	if next.Essay != nil {
		cReq.RubricText = next.Essay.RubricText
		cReq.MaxScore = next.Essay.MaxScore
	}
	in, err := validateAndBuildCreateQuestionInput(cReq)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": err.Error()}})
		return
	}

	updated, ok, err := h.qb.UpdateQuestion(c.Request.Context(), c.Param("id"), questionbankrepo.UpdateQuestionInput{
		Type:          in.Type,
		Stem:          in.Stem,
		OrderNo:       in.OrderNo,
		Weight:        in.Weight,
		Options:       in.Options,
		MatchingPairs: in.MatchingPairs,
		ShortAnswers:  in.ShortAnswers,
		TrueFalse:     in.TrueFalse,
		Essay:         in.Essay,
		Statements:    in.Statements,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "not found"}})
		return
	}
	c.JSON(200, gin.H{"data": updated, "meta": gin.H{"question_set_id": setID}})
}

type nullableInt struct {
	Set   bool
	Value *int
}

func (n *nullableInt) UnmarshalJSON(b []byte) error {
	n.Set = true
	if string(b) == "null" {
		n.Value = nil
		return nil
	}
	var v int
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	n.Value = &v
	return nil
}

func validateAndBuildCreateQuestionInput(req createQuestionReq) (questionbankrepo.CreateQuestionInput, error) {
	qType := strings.TrimSpace(req.Type)
	stem := strings.TrimSpace(req.Stem)
	if qType == "" || stem == "" {
		return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("type and stem required")
	}

	switch qType {
	case "mc_single", "mc_multiple", "matching", "short_answer", "essay", "true_false":
	default:
		return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("invalid type")
	}

	in := questionbankrepo.CreateQuestionInput{
		Type:    qType,
		Stem:    stem,
		OrderNo: req.OrderNo,
		Weight:  1,
	}
	if req.Weight != nil {
		if *req.Weight <= 0 {
			return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("weight must be > 0")
		}
		in.Weight = *req.Weight
	}

	switch qType {
	case "mc_single", "mc_multiple":
		opts := make([]questionbankrepo.QuestionOption, 0, len(req.Options))
		seen := map[string]bool{}
		correct := 0
		for _, o := range req.Options {
			o.Label = strings.TrimSpace(o.Label)
			o.Content = strings.TrimSpace(o.Content)
			if o.Label == "" || o.Content == "" {
				return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("option label/content required")
			}
			key := strings.ToUpper(o.Label)
			if seen[key] {
				return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("duplicate option label")
			}
			seen[key] = true
			o.Label = key
			if o.IsCorrect {
				correct++
			}
			opts = append(opts, questionbankrepo.QuestionOption{Label: o.Label, Content: o.Content, IsCorrect: o.IsCorrect})
		}
		if len(opts) < 2 {
			return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("options min 2")
		}
		if qType == "mc_single" && correct != 1 {
			return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("mc_single must have exactly 1 correct option")
		}
		if qType == "mc_multiple" && correct < 1 {
			return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("mc_multiple must have at least 1 correct option")
		}
		in.Options = opts

	case "matching":
		pairs := make([]questionbankrepo.MatchingPair, 0, len(req.Pairs))
		for i, p := range req.Pairs {
			p.LeftContent = strings.TrimSpace(p.LeftContent)
			p.RightContent = strings.TrimSpace(p.RightContent)
			if p.LeftContent == "" || p.RightContent == "" {
				return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("pair left/right required")
			}
			orderNo := p.OrderNo
			if orderNo == 0 {
				orderNo = i + 1
			}
			pairs = append(pairs, questionbankrepo.MatchingPair{LeftContent: p.LeftContent, RightContent: p.RightContent, OrderNo: orderNo})
		}
		if len(pairs) < 2 {
			return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("pairs min 2")
		}
		in.MatchingPairs = pairs

	case "short_answer":
		answers := make([]questionbankrepo.ShortAnswer, 0, len(req.Answers))
		for i, a := range req.Answers {
			a.AnswerText = strings.TrimSpace(a.AnswerText)
			if a.AnswerText == "" {
				return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("answer_text required")
			}
			orderNo := a.OrderNo
			if orderNo == 0 {
				orderNo = i + 1
			}
			answers = append(answers, questionbankrepo.ShortAnswer{AnswerText: a.AnswerText, OrderNo: orderNo})
		}
		if len(answers) < 1 {
			return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("answers min 1")
		}
		in.ShortAnswers = answers

	case "true_false":
		if len(req.Statements) > 0 {
			stats := make([]questionbankrepo.TFStatement, 0, len(req.Statements))
			for i, s := range req.Statements {
				s.Content = strings.TrimSpace(s.Content)
				if s.Content == "" {
					return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("statement content required")
				}
				ord := s.OrderNo
				if ord == 0 {
					ord = i + 1
				}
				stats = append(stats, questionbankrepo.TFStatement{Content: s.Content, Correct: s.Correct, OrderNo: ord})
			}
			in.Statements = stats
			// Fallback correct to first statement if available
			if len(stats) > 0 {
				in.TrueFalse = &questionbankrepo.TrueFalse{Correct: stats[0].Correct}
			}
		} else {
			if req.Correct == nil {
				return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("correct or statements required for true_false")
			}
			in.TrueFalse = &questionbankrepo.TrueFalse{Correct: *req.Correct}
		}

	case "essay":
		es := &questionbankrepo.Essay{}
		es.RubricText = strings.TrimSpace(req.RubricText)
		es.MaxScore = req.MaxScore
		if es.MaxScore != nil && *es.MaxScore < 0 {
			return questionbankrepo.CreateQuestionInput{}, fmt.Errorf("max_score must be >= 0")
		}
		in.Essay = es
	}

	return in, nil
}
func (h *QuestionBankHandler) CloneSet(c *gin.Context) {
	role := middleware.GetUserRole(c)
	if role != "admin" {
		c.JSON(403, gin.H{"error": gin.H{"code": "forbidden", "message": "only admin can clone sets to other teachers"}})
		return
	}

	var req struct {
		TeacherID string `json:"teacher_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid json"}})
		return
	}
	if strings.TrimSpace(req.TeacherID) == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "teacher_id required"}})
		return
	}

	newID, err := h.qb.CloneSet(c.Request.Context(), c.Param("id"), req.TeacherID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "question set not found"}})
			return
		}
		if pgerr.Code(err) == pgerr.CodeForeignKeyViolation {
			c.JSON(409, gin.H{"error": gin.H{"code": "conflict", "message": "invalid teacher_id"}})
			return
		}
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}

	c.JSON(201, gin.H{"data": gin.H{"id": newID}})
}
