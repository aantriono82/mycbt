package questionbankrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Pool() *pgxpool.Pool { return r.pool }

func (r *Repo) TeacherIDByUserID(ctx context.Context, userID string) (string, bool, error) {
	const q = `SELECT id FROM teachers WHERE user_id = $1 LIMIT 1`
	var id string
	err := r.pool.QueryRow(ctx, q, userID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false, nil
		}
		return "", false, fmt.Errorf("teacher lookup: %w", err)
	}
	return id, true, nil
}

type QuestionSet struct {
	ID             string `json:"id"`
	SubjectID      string `json:"subject_id"`
	OwnerTeacherID string `json:"owner_teacher_id"`
	Title          string `json:"title"`
	Status         string `json:"status"`
	Jenjang        string `json:"jenjang"`
	LevelID        string `json:"level_id"`
}

func (r *Repo) ListSets(ctx context.Context, role string, teacherID string, subjectID string, q string, limit, offset int) ([]QuestionSet, int, error) {
	const base = `
FROM question_sets
WHERE ($1 = '' OR subject_id::text = $1)
  AND ($2 = '' OR title ILIKE '%'||$2||'%')
  AND ($3 = '' OR owner_teacher_id::text = $3)`

	rows, err := r.pool.Query(ctx, `
SELECT id, subject_id::text, COALESCE(owner_teacher_id::text,''), title, status, COALESCE(jenjang,''), COALESCE(level_id::text,'')
`+base+`
ORDER BY created_at DESC
LIMIT $4 OFFSET $5`, subjectID, q, teacherID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list sets: %w", err)
	}
	defer rows.Close()

	out := []QuestionSet{}
	for rows.Next() {
		var it QuestionSet
		if err := rows.Scan(&it.ID, &it.SubjectID, &it.OwnerTeacherID, &it.Title, &it.Status, &it.Jenjang, &it.LevelID); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base, subjectID, q, teacherID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sets: %w", err)
	}

	return out, total, nil
}

func (r *Repo) CreateSet(ctx context.Context, subjectID, ownerTeacherID, title, jenjang, levelID string) (QuestionSet, error) {
	const q = `
INSERT INTO question_sets (subject_id, owner_teacher_id, title, jenjang, level_id)
VALUES ($1::uuid, NULLIF($2,'')::uuid, $3, NULLIF($4,''), NULLIF($5,'')::uuid)
RETURNING id, subject_id::text, COALESCE(owner_teacher_id::text,''), title, status, COALESCE(jenjang,''), COALESCE(level_id::text,'')`
	var it QuestionSet
	if err := r.pool.QueryRow(ctx, q, subjectID, ownerTeacherID, title, jenjang, levelID).Scan(&it.ID, &it.SubjectID, &it.OwnerTeacherID, &it.Title, &it.Status, &it.Jenjang, &it.LevelID); err != nil {
		return QuestionSet{}, fmt.Errorf("create set: %w", err)
	}
	return it, nil
}

func (r *Repo) GetSet(ctx context.Context, id string) (QuestionSet, bool, error) {
	const q = `SELECT id, subject_id::text, COALESCE(owner_teacher_id::text,''), title, status, COALESCE(jenjang,''), COALESCE(level_id::text,'') FROM question_sets WHERE id = $1 LIMIT 1`
	var it QuestionSet
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.SubjectID, &it.OwnerTeacherID, &it.Title, &it.Status, &it.Jenjang, &it.LevelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return QuestionSet{}, false, nil
		}
		return QuestionSet{}, false, fmt.Errorf("get set: %w", err)
	}
	return it, true, nil
}

func (r *Repo) UpdateSet(ctx context.Context, id string, title string, status string, jenjang string, levelID string) (QuestionSet, bool, error) {
	const q = `
UPDATE question_sets
SET title = $2, status = $3, jenjang = $4, level_id = NULLIF($5,'')::uuid, updated_at = now()
WHERE id = $1
RETURNING id, subject_id::text, COALESCE(owner_teacher_id::text,''), title, status, COALESCE(jenjang,''), COALESCE(level_id::text,'')`
	var it QuestionSet
	err := r.pool.QueryRow(ctx, q, id, title, status, jenjang, levelID).Scan(&it.ID, &it.SubjectID, &it.OwnerTeacherID, &it.Title, &it.Status, &it.Jenjang, &it.LevelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return QuestionSet{}, false, nil
		}
		return QuestionSet{}, false, fmt.Errorf("update set: %w", err)
	}
	return it, true, nil
}

func (r *Repo) DeleteSet(ctx context.Context, id string) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM question_sets WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete set: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

type Question struct {
	ID            string           `json:"id"`
	QuestionSetID string           `json:"question_set_id"`
	Type          string           `json:"type"` // mc_single, mc_multiple, matching, short_answer, essay, true_false
	Stem          string           `json:"stem"`
	Explanation   string           `json:"explanation,omitempty"`
	OrderNo       int              `json:"order_no"`
	Weight        int              `json:"weight"`
	Options       []QuestionOption `json:"options,omitempty"`
	MatchingPairs []MatchingPair   `json:"pairs,omitempty"`
	ShortAnswers  []ShortAnswer    `json:"answers,omitempty"`
	TrueFalse     *TrueFalse       `json:"true_false,omitempty"`
	Essay         *Essay           `json:"essay,omitempty"`
	Statements    []TFStatement    `json:"statements,omitempty"` // Added for multiple T/F
}

type TFStatement struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Correct bool   `json:"correct"`
	OrderNo int    `json:"order_no"`
}

type QuestionOption struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"`
}

type MatchingPair struct {
	ID           string `json:"id"`
	LeftContent  string `json:"left_content"`
	RightContent string `json:"right_content"`
	OrderNo      int    `json:"order_no"`
}

type ShortAnswer struct {
	ID         string `json:"id"`
	AnswerText string `json:"answer_text"`
	OrderNo    int    `json:"order_no"`
}

type TrueFalse struct {
	Correct bool `json:"correct"`
}

type Essay struct {
	RubricText string `json:"rubric_text,omitempty"`
	MaxScore   *int   `json:"max_score,omitempty"`
}

func (r *Repo) ListQuestions(ctx context.Context, setID string) ([]Question, error) {
	// Fetch questions first
	rows, err := r.pool.Query(ctx, `SELECT id, question_set_id::text, type, stem, COALESCE(explanation,''), order_no, COALESCE(weight, 1) FROM questions WHERE question_set_id = $1 ORDER BY order_no ASC, created_at ASC`, setID)
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	defer rows.Close()

	out := []Question{}
	// Map question id to its index in `out` (don't store pointers to slice elements;
	// appends can reallocate and invalidate them, causing payloads to "not stick").
	byID := map[string]int{}
	for rows.Next() {
		var it Question
		if err := rows.Scan(&it.ID, &it.QuestionSetID, &it.Type, &it.Stem, &it.Explanation, &it.OrderNo, &it.Weight); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		it.Options = []QuestionOption{}
		it.MatchingPairs = []MatchingPair{}
		it.ShortAnswers = []ShortAnswer{}
		it.Statements = []TFStatement{}
		out = append(out, it)
		byID[it.ID] = len(out) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch options
	orows, err := r.pool.Query(ctx, `SELECT id, question_id::text, label, content, is_correct FROM question_options WHERE question_id IN (SELECT id FROM questions WHERE question_set_id = $1) ORDER BY label ASC`, setID)
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	defer orows.Close()
	for orows.Next() {
		var oid, qid, label, content string
		var isCorrect bool
		if err := orows.Scan(&oid, &qid, &label, &content, &isCorrect); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		if idx, ok := byID[qid]; ok {
			out[idx].Options = append(out[idx].Options, QuestionOption{ID: oid, Label: label, Content: content, IsCorrect: isCorrect})
		}
	}
	if err := orows.Err(); err != nil {
		return nil, err
	}

	// Fetch matching pairs
	prows, err := r.pool.Query(ctx, `
SELECT id, question_id::text, left_content, right_content, order_no
FROM question_matching_pairs
WHERE question_id IN (SELECT id FROM questions WHERE question_set_id = $1)
ORDER BY order_no ASC, id ASC`, setID)
	if err != nil {
		return nil, fmt.Errorf("list pairs: %w", err)
	}
	defer prows.Close()
	for prows.Next() {
		var pid, qid, l, rgt string
		var orderNo int
		if err := prows.Scan(&pid, &qid, &l, &rgt, &orderNo); err != nil {
			return nil, fmt.Errorf("scan pair: %w", err)
		}
		if idx, ok := byID[qid]; ok {
			out[idx].MatchingPairs = append(out[idx].MatchingPairs, MatchingPair{ID: pid, LeftContent: l, RightContent: rgt, OrderNo: orderNo})
		}
	}
	if err := prows.Err(); err != nil {
		return nil, err
	}

	// Fetch short answers
	arows, err := r.pool.Query(ctx, `
SELECT id, question_id::text, answer_text, order_no
FROM question_short_answers
WHERE question_id IN (SELECT id FROM questions WHERE question_set_id = $1)
ORDER BY order_no ASC, id ASC`, setID)
	if err != nil {
		return nil, fmt.Errorf("list answers: %w", err)
	}
	defer arows.Close()
	for arows.Next() {
		var aid, qid, ans string
		var orderNo int
		if err := arows.Scan(&aid, &qid, &ans, &orderNo); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		if idx, ok := byID[qid]; ok {
			out[idx].ShortAnswers = append(out[idx].ShortAnswers, ShortAnswer{ID: aid, AnswerText: ans, OrderNo: orderNo})
		}
	}
	if err := arows.Err(); err != nil {
		return nil, err
	}

	// Fetch true/false
	tfrows, err := r.pool.Query(ctx, `
SELECT question_id::text, correct
FROM question_true_false
WHERE question_id IN (SELECT id FROM questions WHERE question_set_id = $1)`, setID)
	if err != nil {
		return nil, fmt.Errorf("list true_false: %w", err)
	}
	defer tfrows.Close()
	for tfrows.Next() {
		var qid string
		var correct bool
		if err := tfrows.Scan(&qid, &correct); err != nil {
			return nil, fmt.Errorf("scan true_false: %w", err)
		}
		if idx, ok := byID[qid]; ok {
			out[idx].TrueFalse = &TrueFalse{Correct: correct}
		}
	}
	if err := tfrows.Err(); err != nil {
		return nil, err
	}

	// Fetch essays (rubric/max_score)
	erows, err := r.pool.Query(ctx, `
SELECT question_id::text, COALESCE(rubric_text,''), max_score
FROM question_essays
WHERE question_id IN (SELECT id FROM questions WHERE question_set_id = $1)`, setID)
	if err != nil {
		return nil, fmt.Errorf("list essays: %w", err)
	}
	defer erows.Close()
	for erows.Next() {
		var qid string
		var rubric string
		var maxScore *int
		if err := erows.Scan(&qid, &rubric, &maxScore); err != nil {
			return nil, fmt.Errorf("scan essay: %w", err)
		}
		if idx, ok := byID[qid]; ok {
			es := &Essay{MaxScore: maxScore}
			if strings.TrimSpace(rubric) != "" {
				es.RubricText = rubric
			}
			out[idx].Essay = es
		}
	}
	if err := erows.Err(); err != nil {
		return nil, err
	}

	// Fetch true/false statements (ANBK Style)
	strows, err := r.pool.Query(ctx, `
SELECT id, question_id::text, content, correct, order_no
FROM question_true_false_statements
WHERE question_id IN (SELECT id FROM questions WHERE question_set_id = $1)
ORDER BY order_no ASC, id ASC`, setID)
	if err != nil {
		return nil, fmt.Errorf("list statements: %w", err)
	}
	defer strows.Close()
	for strows.Next() {
		var st TFStatement
		var qid string
		if err := strows.Scan(&st.ID, &qid, &st.Content, &st.Correct, &st.OrderNo); err != nil {
			return nil, fmt.Errorf("scan statement: %w", err)
		}
		if idx, ok := byID[qid]; ok {
			out[idx].Statements = append(out[idx].Statements, st)
		}
	}
	if err := strows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

type CreateQuestionInput struct {
	Type          string
	Stem          string
	Explanation   string
	OrderNo       int
	Weight        int
	Options       []QuestionOption
	MatchingPairs []MatchingPair
	ShortAnswers  []ShortAnswer
	TrueFalse     *TrueFalse
	Essay         *Essay
	Statements    []TFStatement
}

func (r *Repo) CreateQuestion(ctx context.Context, setID string, in CreateQuestionInput) (Question, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Question{}, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var qID string
	err = tx.QueryRow(ctx, `INSERT INTO questions (question_set_id, type, stem, explanation, order_no, weight) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, setID, in.Type, in.Stem, in.Explanation, in.OrderNo, in.Weight).Scan(&qID)
	if err != nil {
		if strings.Contains(err.Error(), `column "explanation"`) {
			if err2 := tx.QueryRow(ctx, `INSERT INTO questions (question_set_id, type, stem, order_no, weight) VALUES ($1,$2,$3,$4,$5) RETURNING id`, setID, in.Type, in.Stem, in.OrderNo, in.Weight).Scan(&qID); err2 != nil {
				return Question{}, fmt.Errorf("insert question fallback: %w", err2)
			}
		} else {
			return Question{}, fmt.Errorf("insert question: %w", err)
		}
	}

	out, err := insertQuestionTypePayload(ctx, tx, qID, in)
	if err != nil {
		return Question{}, err
	}
	out.ID = qID
	out.QuestionSetID = setID

	if err := tx.Commit(ctx); err != nil {
		return Question{}, fmt.Errorf("commit: %w", err)
	}

	return out, nil
}

func (r *Repo) CreateQuestionsBulk(ctx context.Context, setID string, items []CreateQuestionInput) ([]Question, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	out := make([]Question, 0, len(items))
	for _, in := range items {
		var qID string
		err = tx.QueryRow(ctx, `INSERT INTO questions (question_set_id, type, stem, explanation, order_no, weight) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, setID, in.Type, in.Stem, in.Explanation, in.OrderNo, in.Weight).Scan(&qID)
		if err != nil {
			if strings.Contains(err.Error(), `column "explanation"`) {
				if err2 := tx.QueryRow(ctx, `INSERT INTO questions (question_set_id, type, stem, order_no, weight) VALUES ($1,$2,$3,$4,$5) RETURNING id`, setID, in.Type, in.Stem, in.OrderNo, in.Weight).Scan(&qID); err2 != nil {
					return nil, fmt.Errorf("insert question fallback: %w", err2)
				}
			} else {
				return nil, fmt.Errorf("insert question: %w", err)
			}
		}

		q, err := insertQuestionTypePayload(ctx, tx, qID, in)
		if err != nil {
			return nil, err
		}
		q.ID = qID
		q.QuestionSetID = setID
		out = append(out, q)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return out, nil
}

func (r *Repo) DeleteQuestion(ctx context.Context, id string) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM questions WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete question: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

func (r *Repo) GetQuestion(ctx context.Context, id string) (Question, bool, error) {
	const q = `SELECT id, question_set_id::text, type, stem, COALESCE(explanation,''), order_no, COALESCE(weight, 1) FROM questions WHERE id = $1 LIMIT 1`
	var it Question
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.QuestionSetID, &it.Type, &it.Stem, &it.Explanation, &it.OrderNo, &it.Weight)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Question{}, false, nil
		}
		return Question{}, false, fmt.Errorf("get question: %w", err)
	}
	it.Options = []QuestionOption{}
	it.MatchingPairs = []MatchingPair{}
	it.ShortAnswers = []ShortAnswer{}

	// Options
	orows, err := r.pool.Query(ctx, `SELECT id, label, content, is_correct FROM question_options WHERE question_id = $1 ORDER BY label ASC`, id)
	if err != nil {
		return Question{}, false, fmt.Errorf("get options: %w", err)
	}
	defer orows.Close()
	for orows.Next() {
		var o QuestionOption
		if err := orows.Scan(&o.ID, &o.Label, &o.Content, &o.IsCorrect); err != nil {
			return Question{}, false, fmt.Errorf("scan option: %w", err)
		}
		it.Options = append(it.Options, o)
	}
	if err := orows.Err(); err != nil {
		return Question{}, false, err
	}

	// Matching
	prows, err := r.pool.Query(ctx, `SELECT id, left_content, right_content, order_no FROM question_matching_pairs WHERE question_id = $1 ORDER BY order_no ASC, id ASC`, id)
	if err != nil {
		return Question{}, false, fmt.Errorf("get pairs: %w", err)
	}
	defer prows.Close()
	for prows.Next() {
		var p MatchingPair
		if err := prows.Scan(&p.ID, &p.LeftContent, &p.RightContent, &p.OrderNo); err != nil {
			return Question{}, false, fmt.Errorf("scan pair: %w", err)
		}
		it.MatchingPairs = append(it.MatchingPairs, p)
	}
	if err := prows.Err(); err != nil {
		return Question{}, false, err
	}

	// Short answers
	arows, err := r.pool.Query(ctx, `SELECT id, answer_text, order_no FROM question_short_answers WHERE question_id = $1 ORDER BY order_no ASC, id ASC`, id)
	if err != nil {
		return Question{}, false, fmt.Errorf("get answers: %w", err)
	}
	defer arows.Close()
	for arows.Next() {
		var a ShortAnswer
		if err := arows.Scan(&a.ID, &a.AnswerText, &a.OrderNo); err != nil {
			return Question{}, false, fmt.Errorf("scan answer: %w", err)
		}
		it.ShortAnswers = append(it.ShortAnswers, a)
	}
	if err := arows.Err(); err != nil {
		return Question{}, false, err
	}

	// True/false
	var correct bool
	if err := r.pool.QueryRow(ctx, `SELECT correct FROM question_true_false WHERE question_id = $1`, id).Scan(&correct); err == nil {
		it.TrueFalse = &TrueFalse{Correct: correct}
	} else if err != pgx.ErrNoRows {
		return Question{}, false, fmt.Errorf("get true_false: %w", err)
	}

	// Essay
	var rubric string
	var maxScore *int
	if err := r.pool.QueryRow(ctx, `SELECT COALESCE(rubric_text,''), max_score FROM question_essays WHERE question_id = $1`, id).Scan(&rubric, &maxScore); err == nil {
		es := &Essay{MaxScore: maxScore}
		if rubric != "" {
			es.RubricText = rubric
		}
		it.Essay = es
	} else if err != pgx.ErrNoRows {
		return Question{}, false, fmt.Errorf("get essay: %w", err)
	}

	// Statements
	it.Statements = []TFStatement{}
	strows, err := r.pool.Query(ctx, `SELECT id, content, correct, order_no FROM question_true_false_statements WHERE question_id = $1 ORDER BY order_no ASC`, id)
	if err == nil {
		defer strows.Close()
		for strows.Next() {
			var st TFStatement
			if err := strows.Scan(&st.ID, &st.Content, &st.Correct, &st.OrderNo); err != nil {
				return Question{}, false, fmt.Errorf("scan statement: %w", err)
			}
			it.Statements = append(it.Statements, st)
		}
	}

	return it, true, nil
}

type UpdateQuestionInput struct {
	Type          string
	Stem          string
	Explanation   string
	OrderNo       int
	Weight        int
	Options       []QuestionOption
	MatchingPairs []MatchingPair
	ShortAnswers  []ShortAnswer
	TrueFalse     *TrueFalse
	Essay         *Essay
	Statements    []TFStatement
}

func (r *Repo) UpdateQuestion(ctx context.Context, id string, in UpdateQuestionInput) (Question, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Question{}, false, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var setID string
	err = tx.QueryRow(ctx, `
UPDATE questions
SET type = $2, stem = $3, explanation = $4, order_no = $5, weight = $6, updated_at = now()
WHERE id = $1
RETURNING question_set_id::text`, id, in.Type, in.Stem, in.Explanation, in.OrderNo, in.Weight).Scan(&setID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Question{}, false, nil
		}
		return Question{}, false, fmt.Errorf("update question: %w", err)
	}

	// Clear existing payload (so type can change cleanly)
	if _, err := tx.Exec(ctx, `DELETE FROM question_options WHERE question_id = $1`, id); err != nil {
		return Question{}, false, fmt.Errorf("clear options: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM question_matching_pairs WHERE question_id = $1`, id); err != nil {
		return Question{}, false, fmt.Errorf("clear pairs: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM question_short_answers WHERE question_id = $1`, id); err != nil {
		return Question{}, false, fmt.Errorf("clear answers: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM question_true_false WHERE question_id = $1`, id); err != nil {
		return Question{}, false, fmt.Errorf("clear true_false: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM question_essays WHERE question_id = $1`, id); err != nil {
		return Question{}, false, fmt.Errorf("clear essays: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM question_true_false_statements WHERE question_id = $1`, id); err != nil {
		return Question{}, false, fmt.Errorf("clear statements: %w", err)
	}

	out, err := insertQuestionTypePayload(ctx, tx, id, CreateQuestionInput{
		Type:          in.Type,
		Stem:          in.Stem,
		Explanation:   in.Explanation,
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
		return Question{}, false, err
	}
	out.ID = id
	out.QuestionSetID = setID

	if err := tx.Commit(ctx); err != nil {
		return Question{}, false, fmt.Errorf("commit: %w", err)
	}
	return out, true, nil
}

func insertQuestionTypePayload(ctx context.Context, tx pgx.Tx, questionID string, in CreateQuestionInput) (Question, error) {
	out := Question{
		Type:          in.Type,
		Stem:          in.Stem,
		Explanation:   in.Explanation,
		OrderNo:       in.OrderNo,
		Weight:        in.Weight,
		Options:       []QuestionOption{},
		MatchingPairs: []MatchingPair{},
		ShortAnswers:  []ShortAnswer{},
	}

	switch in.Type {
	case "mc_single", "mc_multiple":
		for _, opt := range in.Options {
			var oid string
			if err := tx.QueryRow(ctx, `INSERT INTO question_options (question_id, label, content, is_correct) VALUES ($1,$2,$3,$4) RETURNING id`, questionID, opt.Label, opt.Content, opt.IsCorrect).Scan(&oid); err != nil {
				return Question{}, fmt.Errorf("insert option: %w", err)
			}
			opt.ID = oid
			out.Options = append(out.Options, opt)
		}
	case "matching":
		for _, p := range in.MatchingPairs {
			var pid string
			if err := tx.QueryRow(ctx, `INSERT INTO question_matching_pairs (question_id, left_content, right_content, order_no) VALUES ($1,$2,$3,$4) RETURNING id`, questionID, p.LeftContent, p.RightContent, p.OrderNo).Scan(&pid); err != nil {
				return Question{}, fmt.Errorf("insert pair: %w", err)
			}
			p.ID = pid
			out.MatchingPairs = append(out.MatchingPairs, p)
		}
	case "short_answer":
		for _, a := range in.ShortAnswers {
			var aid string
			if err := tx.QueryRow(ctx, `INSERT INTO question_short_answers (question_id, answer_text, order_no) VALUES ($1,$2,$3) RETURNING id`, questionID, a.AnswerText, a.OrderNo).Scan(&aid); err != nil {
				return Question{}, fmt.Errorf("insert answer: %w", err)
			}
			a.ID = aid
			out.ShortAnswers = append(out.ShortAnswers, a)
		}
	case "true_false":
		if in.TrueFalse != nil {
			if _, err := tx.Exec(ctx, `INSERT INTO question_true_false (question_id, correct) VALUES ($1,$2)`, questionID, in.TrueFalse.Correct); err != nil {
				return Question{}, fmt.Errorf("insert true_false: %w", err)
			}
			out.TrueFalse = &TrueFalse{Correct: in.TrueFalse.Correct}
		}
		// Always handle statements if present
		for _, st := range in.Statements {
			var sid string
			if err := tx.QueryRow(ctx, `INSERT INTO question_true_false_statements (question_id, content, correct, order_no) VALUES ($1,$2,$3,$4) RETURNING id`, questionID, st.Content, st.Correct, st.OrderNo).Scan(&sid); err != nil {
				return Question{}, fmt.Errorf("insert statement: %w", err)
			}
			st.ID = sid
			out.Statements = append(out.Statements, st)
		}
	case "essay":
		if in.Essay != nil {
			rubric := strings.TrimSpace(in.Essay.RubricText)
			if _, err := tx.Exec(ctx, `INSERT INTO question_essays (question_id, rubric_text, max_score) VALUES ($1, NULLIF($2,''), $3)`, questionID, rubric, in.Essay.MaxScore); err != nil {
				return Question{}, fmt.Errorf("insert essay: %w", err)
			}
			out.Essay = &Essay{RubricText: rubric, MaxScore: in.Essay.MaxScore}
		} else {
			if _, err := tx.Exec(ctx, `INSERT INTO question_essays (question_id, rubric_text, max_score) VALUES ($1, NULL, NULL)`, questionID); err != nil {
				return Question{}, fmt.Errorf("insert essay: %w", err)
			}
			out.Essay = &Essay{}
		}
	default:
		// Don't insert anything for unknown type; let handler validation catch it.
	}

	return out, nil
}
func (r *Repo) CloneSet(ctx context.Context, id string, newOwnerTeacherID string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Get original set
	var s QuestionSet
	err = tx.QueryRow(ctx, `SELECT subject_id, title, jenjang, level_id FROM question_sets WHERE id = $1`, id).Scan(&s.SubjectID, &s.Title, &s.Jenjang, &s.LevelID)
	if err != nil {
		return "", fmt.Errorf("get original set: %w", err)
	}

	// 2. Create new set
	var newID string
	const qInsertSet = `
INSERT INTO question_sets (subject_id, owner_teacher_id, title, jenjang, level_id, status)
VALUES ($1, NULLIF($2,'')::uuid, $3, NULLIF($4,''), NULLIF($5,'')::uuid, 'draft')
RETURNING id`
	err = tx.QueryRow(ctx, qInsertSet, s.SubjectID, newOwnerTeacherID, s.Title+" (Copy)", s.Jenjang, s.LevelID).Scan(&newID)
	if err != nil {
		return "", fmt.Errorf("insert new set: %w", err)
	}

	// 3. Clone Questions
	rows, err := tx.Query(ctx, `SELECT id, type, stem, COALESCE(explanation,''), order_no, COALESCE(weight, 1) FROM questions WHERE question_set_id = $1`, id)
	if err != nil {
		return "", fmt.Errorf("list original questions: %w", err)
	}
	defer rows.Close()

	type origQ struct {
		id          string
		qType       string
		stem        string
		explanation string
		orderNo     int
		weight      int
	}
	var originals []origQ
	for rows.Next() {
		var it origQ
		if err := rows.Scan(&it.id, &it.qType, &it.stem, &it.explanation, &it.orderNo, &it.weight); err != nil {
			return "", err
		}
		originals = append(originals, it)
	}
	rows.Close()

	type qMap struct {
		oldID string
		newID string
		qType string
	}
	var qs []qMap
	for _, o := range originals {
		var newQID string
		if err := tx.QueryRow(ctx, `INSERT INTO questions (question_set_id, type, stem, explanation, order_no, weight) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, newID, o.qType, o.stem, o.explanation, o.orderNo, o.weight).Scan(&newQID); err != nil {
			return "", err
		}
		qs = append(qs, qMap{o.id, newQID, o.qType})
	}

	// 4. Clone sub-tables for each question
	for _, q := range qs {
		switch q.qType {
		case "mc_single", "mc_multiple":
			_, err = tx.Exec(ctx, `INSERT INTO question_options (question_id, label, content, is_correct) SELECT $1, label, content, is_correct FROM question_options WHERE question_id = $2`, q.newID, q.oldID)
		case "matching":
			_, err = tx.Exec(ctx, `INSERT INTO question_matching_pairs (question_id, left_content, right_content, order_no) SELECT $1, left_content, right_content, order_no FROM question_matching_pairs WHERE question_id = $2`, q.newID, q.oldID)
		case "short_answer":
			_, err = tx.Exec(ctx, `INSERT INTO question_short_answers (question_id, answer_text, order_no) SELECT $1, answer_text, order_no FROM question_short_answers WHERE question_id = $2`, q.newID, q.oldID)
		case "true_false":
			_, err = tx.Exec(ctx, `INSERT INTO question_true_false (question_id, correct) SELECT $1, correct FROM question_true_false WHERE question_id = $2`, q.newID, q.oldID)
			if err == nil {
				_, err = tx.Exec(ctx, `INSERT INTO question_true_false_statements (question_id, content, correct, order_no) SELECT $1, content, correct, order_no FROM question_true_false_statements WHERE question_id = $2`, q.newID, q.oldID)
			}
		case "essay":
			_, err = tx.Exec(ctx, `INSERT INTO question_essays (question_id, rubric_text, max_score) SELECT $1, rubric_text, max_score FROM question_essays WHERE question_id = $2`, q.newID, q.oldID)
		}
		if err != nil {
			return "", fmt.Errorf("clone sub-tables for %s: %w", q.qType, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return newID, nil
}
