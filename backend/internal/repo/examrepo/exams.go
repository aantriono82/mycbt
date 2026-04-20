package examrepo

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

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

type Exam struct {
	ID               string  `json:"id"`
	SubjectID        string  `json:"subject_id"`
	SubjectName      string  `json:"subject_name,omitempty"`
	TeacherID        string  `json:"teacher_id"`
	TeacherName      string  `json:"teacher_name,omitempty"`
	SessionID        *string `json:"session_id,omitempty"`
	SessionName      string  `json:"session_name,omitempty"`
	SessionStartTime string  `json:"session_start_time,omitempty"`
	SessionEndTime   string  `json:"session_end_time,omitempty"`
	Title            string  `json:"title"`
	StartsAt         string  `json:"starts_at"` // RFC3339
	EndsAt           string  `json:"ends_at"`   // RFC3339
	DurationMinutes  *int    `json:"duration_minutes,omitempty"`
	ShuffleQuestions bool    `json:"shuffle_questions"`
	ShuffleOptions   bool    `json:"shuffle_options"`
	Status           string  `json:"status"`
}

type ListFilter struct {
	Status string
	Q      string
	Limit  int
	Offset int
}

func (r *Repo) List(ctx context.Context, role, teacherID string, f ListFilter) ([]Exam, int, error) {
	const base = `
FROM exams e
LEFT JOIN subjects s ON e.subject_id = s.id
LEFT JOIN teachers t ON e.teacher_id = t.id
LEFT JOIN sessions sess ON e.session_id = sess.id
WHERE ($1 = '' OR e.status = $1)
  AND ($2 = '' OR e.title ILIKE '%'||$2||'%')
  AND ($3 = '' OR e.teacher_id::text = $3)`

	rows, err := r.pool.Query(ctx, `
SELECT e.id::text, e.subject_id::text, e.teacher_id::text, e.title,
       to_char(e.starts_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       to_char(e.ends_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       e.duration_minutes, e.shuffle_questions, e.shuffle_options, e.status,
       COALESCE(s.name,''), COALESCE(sess.id::text,''), COALESCE(sess.name,''),
       COALESCE(to_char(sess.start_time, 'HH24:MI'), ''), COALESCE(to_char(sess.end_time, 'HH24:MI'), '')
`+base+`
ORDER BY e.starts_at DESC, e.created_at DESC
LIMIT $4 OFFSET $5`, strings.TrimSpace(f.Status), strings.TrimSpace(f.Q), teacherID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list exams: %w", err)
	}
	defer rows.Close()

	out := []Exam{}
	for rows.Next() {
		var it Exam
		var sessID, sessName, sessStart, sessEnd string
		if err := rows.Scan(&it.ID, &it.SubjectID, &it.TeacherID, &it.Title, &it.StartsAt, &it.EndsAt, &it.DurationMinutes, &it.ShuffleQuestions, &it.ShuffleOptions, &it.Status, &it.SubjectName, &sessID, &sessName, &sessStart, &sessEnd); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		if sessID != "" {
			it.SessionID = &sessID
			it.SessionName = sessName
			it.SessionStartTime = sessStart
			it.SessionEndTime = sessEnd
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base, strings.TrimSpace(f.Status), strings.TrimSpace(f.Q), teacherID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count exams: %w", err)
	}

	_ = role // currently only affects teacherID param in handler
	return out, total, nil
}

type CreateInput struct {
	SubjectID        string
	TeacherID        string
	SessionID        *string
	Title            string
	StartsAt         time.Time
	EndsAt           time.Time
	DurationMinutes  *int
	ShuffleQuestions bool
	ShuffleOptions   bool
}

func (r *Repo) Create(ctx context.Context, in CreateInput) (Exam, error) {
	const q = `
INSERT INTO exams (subject_id, teacher_id, session_id, title, starts_at, ends_at, duration_minutes, shuffle_questions, shuffle_options)
VALUES ($1::uuid,$2::uuid,$3::uuid,$4,$5,$6,$7,$8,$9)
RETURNING id::text, subject_id::text, teacher_id::text, title,
       to_char(starts_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       to_char(ends_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       duration_minutes, shuffle_questions, shuffle_options, status, session_id::text`
	var it Exam
	var sessID *string
	if err := r.pool.QueryRow(ctx, q, in.SubjectID, in.TeacherID, in.SessionID, in.Title, in.StartsAt, in.EndsAt, in.DurationMinutes, in.ShuffleQuestions, in.ShuffleOptions).
		Scan(&it.ID, &it.SubjectID, &it.TeacherID, &it.Title, &it.StartsAt, &it.EndsAt, &it.DurationMinutes, &it.ShuffleQuestions, &it.ShuffleOptions, &it.Status, &sessID); err != nil {
		return Exam{}, fmt.Errorf("create exam: %w", err)
	}
	it.SessionID = sessID
	return it, nil
}

func (r *Repo) Get(ctx context.Context, id string) (Exam, bool, error) {
	const q = `
SELECT e.id::text, e.subject_id::text, e.teacher_id::text, e.title,
       to_char(e.starts_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       to_char(e.ends_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       e.duration_minutes, e.shuffle_questions, e.shuffle_options, e.status,
       COALESCE(sess.id::text,''), COALESCE(sess.name,''),
       COALESCE(to_char(sess.start_time, 'HH24:MI'), ''), COALESCE(to_char(sess.end_time, 'HH24:MI'), '')
FROM exams e
LEFT JOIN sessions sess ON e.session_id = sess.id
WHERE e.id = $1
LIMIT 1`
	var it Exam
	var sessID, sessName, sessStart, sessEnd string
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&it.ID, &it.SubjectID, &it.TeacherID, &it.Title, &it.StartsAt, &it.EndsAt, &it.DurationMinutes, &it.ShuffleQuestions, &it.ShuffleOptions, &it.Status, &sessID, &sessName, &sessStart, &sessEnd)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Exam{}, false, nil
		}
		return Exam{}, false, fmt.Errorf("get exam: %w", err)
	}
	if sessID != "" {
		it.SessionID = &sessID
		it.SessionName = sessName
		it.SessionStartTime = sessStart
		it.SessionEndTime = sessEnd
	}
	return it, true, nil
}

type UpdateInput struct {
	SessionID        *string
	Title            string
	StartsAt         time.Time
	EndsAt           time.Time
	DurationMinutes  *int
	ShuffleQuestions bool
	ShuffleOptions   bool
	Status           string
}

func (r *Repo) Update(ctx context.Context, id string, in UpdateInput) (Exam, bool, error) {
	const q = `
UPDATE exams
SET title = $2,
    starts_at = $3,
    ends_at = $4,
    duration_minutes = $5,
    shuffle_questions = $6,
    shuffle_options = $7,
    status = $8,
    session_id = $9::uuid,
    updated_at = now()
WHERE id = $1
RETURNING id::text, subject_id::text, teacher_id::text, title,
       to_char(starts_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       to_char(ends_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       duration_minutes, shuffle_questions, shuffle_options, status, session_id::text`
	var it Exam
	var sessID *string
	err := r.pool.QueryRow(ctx, q, id, in.Title, in.StartsAt, in.EndsAt, in.DurationMinutes, in.ShuffleQuestions, in.ShuffleOptions, in.Status, in.SessionID).
		Scan(&it.ID, &it.SubjectID, &it.TeacherID, &it.Title, &it.StartsAt, &it.EndsAt, &it.DurationMinutes, &it.ShuffleQuestions, &it.ShuffleOptions, &it.Status, &sessID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Exam{}, false, nil
		}
		return Exam{}, false, fmt.Errorf("update exam: %w", err)
	}
	it.SessionID = sessID
	return it, true, nil
}

func (r *Repo) Delete(ctx context.Context, id string) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM exams WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete exam: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

type ExamToken struct {
	ID        string `json:"id"`
	ExamID    string `json:"exam_id"`
	Token     string `json:"token"`
	ValidFrom string `json:"valid_from,omitempty"`
	ValidTo   string `json:"valid_to,omitempty"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func (r *Repo) ListTokens(ctx context.Context, examID string, limit, offset int) ([]ExamToken, int, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id::text, exam_id::text, token,
       COALESCE(to_char(valid_from at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS valid_from,
       COALESCE(to_char(valid_to at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS valid_to,
       is_active,
       to_char(created_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
FROM exam_tokens
WHERE exam_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`, examID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list tokens: %w", err)
	}
	defer rows.Close()

	out := []ExamToken{}
	for rows.Next() {
		var it ExamToken
		var vf, vt string
		if err := rows.Scan(&it.ID, &it.ExamID, &it.Token, &vf, &vt, &it.IsActive, &it.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan token: %w", err)
		}
		if vf != "" {
			it.ValidFrom = vf
		}
		if vt != "" {
			it.ValidTo = vt
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exam_tokens WHERE exam_id = $1`, examID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tokens: %w", err)
	}
	return out, total, nil
}

type CreateTokenInput struct {
	ExamID          string
	ValidFrom       *time.Time
	ValidTo         *time.Time
	CreatedByUserID string
	Length          int
}

func (r *Repo) CreateToken(ctx context.Context, in CreateTokenInput) (ExamToken, error) {
	length := in.Length
	if length <= 0 {
		length = 6
	}
	if length < 4 {
		length = 4
	}
	if length > 12 {
		length = 12
	}

	// Try multiple times in case of collision.
	var lastErr error
	for i := 0; i < 10; i++ {
		token, err := randToken(length)
		if err != nil {
			return ExamToken{}, err
		}

		var it ExamToken
		var vf, vt string
		err = r.pool.QueryRow(ctx, `
INSERT INTO exam_tokens (exam_id, token, valid_from, valid_to, created_by_user_id)
VALUES ($1::uuid,$2,$3,$4,$5::uuid)
RETURNING id::text, exam_id::text, token,
       COALESCE(to_char(valid_from at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       COALESCE(to_char(valid_to at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       is_active,
       to_char(created_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
`, in.ExamID, token, in.ValidFrom, in.ValidTo, in.CreatedByUserID).Scan(&it.ID, &it.ExamID, &it.Token, &vf, &vt, &it.IsActive, &it.CreatedAt)
		if err != nil {
			lastErr = err
			continue
		}
		if vf != "" {
			it.ValidFrom = vf
		}
		if vt != "" {
			it.ValidTo = vt
		}
		return it, nil
	}
	return ExamToken{}, fmt.Errorf("create token failed: %w", lastErr)
}

func (r *Repo) SetTokenActive(ctx context.Context, tokenID string, isActive bool) (ExamToken, bool, error) {
	var it ExamToken
	var vf, vt string
	err := r.pool.QueryRow(ctx, `
UPDATE exam_tokens
SET is_active = $2
WHERE id = $1
RETURNING id::text, exam_id::text, token,
       COALESCE(to_char(valid_from at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       COALESCE(to_char(valid_to at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       is_active,
       to_char(created_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
`, tokenID, isActive).Scan(&it.ID, &it.ExamID, &it.Token, &vf, &vt, &it.IsActive, &it.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ExamToken{}, false, nil
		}
		return ExamToken{}, false, fmt.Errorf("set token active: %w", err)
	}
	if vf != "" {
		it.ValidFrom = vf
	}
	if vt != "" {
		it.ValidTo = vt
	}
	return it, true, nil
}

const tokenAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // avoid 0,O,1,I

func randToken(n int) (string, error) {
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		x, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenAlphabet))))
		if err != nil {
			return "", fmt.Errorf("rand: %w", err)
		}
		sb.WriteByte(tokenAlphabet[x.Int64()])
	}
	return sb.String(), nil
}

type ExamTarget struct {
	ID        string `json:"id"`
	ExamID    string `json:"exam_id"`
	LevelID   string `json:"level_id,omitempty"`
	GroupID   string `json:"group_id,omitempty"`
	StudentID string `json:"student_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

func (r *Repo) ListTargets(ctx context.Context, examID string) ([]ExamTarget, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id::text, exam_id::text,
       COALESCE(level_id::text,''), COALESCE(group_id::text,''), COALESCE(student_id::text,''),
       to_char(created_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
FROM exam_targets
WHERE exam_id = $1
ORDER BY created_at ASC`, examID)
	if err != nil {
		return nil, fmt.Errorf("list targets: %w", err)
	}
	defer rows.Close()

	out := []ExamTarget{}
	for rows.Next() {
		var it ExamTarget
		if err := rows.Scan(&it.ID, &it.ExamID, &it.LevelID, &it.GroupID, &it.StudentID, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan target: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type ReplaceTargetsInput struct {
	ExamID     string
	LevelIDs   []string
	GroupIDs   []string
	StudentIDs []string
}

func (r *Repo) ReplaceTargets(ctx context.Context, in ReplaceTargetsInput) ([]ExamTarget, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM exam_targets WHERE exam_id = $1`, in.ExamID); err != nil {
		return nil, fmt.Errorf("clear targets: %w", err)
	}

	out := []ExamTarget{}
	for _, id := range in.LevelIDs {
		var it ExamTarget
		err := tx.QueryRow(ctx, `
INSERT INTO exam_targets (exam_id, level_id)
VALUES ($1::uuid, $2::uuid)
RETURNING id::text, exam_id::text,
       COALESCE(level_id::text,''), COALESCE(group_id::text,''), COALESCE(student_id::text,''),
       to_char(created_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
`, in.ExamID, id).Scan(&it.ID, &it.ExamID, &it.LevelID, &it.GroupID, &it.StudentID, &it.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert level target: %w", err)
		}
		out = append(out, it)
	}
	for _, id := range in.GroupIDs {
		var it ExamTarget
		err := tx.QueryRow(ctx, `
INSERT INTO exam_targets (exam_id, group_id)
VALUES ($1::uuid, $2::uuid)
RETURNING id::text, exam_id::text,
       COALESCE(level_id::text,''), COALESCE(group_id::text,''), COALESCE(student_id::text,''),
       to_char(created_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
`, in.ExamID, id).Scan(&it.ID, &it.ExamID, &it.LevelID, &it.GroupID, &it.StudentID, &it.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert group target: %w", err)
		}
		out = append(out, it)
	}
	for _, id := range in.StudentIDs {
		var it ExamTarget
		err := tx.QueryRow(ctx, `
INSERT INTO exam_targets (exam_id, student_id)
VALUES ($1::uuid, $2::uuid)
RETURNING id::text, exam_id::text,
       COALESCE(level_id::text,''), COALESCE(group_id::text,''), COALESCE(student_id::text,''),
       to_char(created_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
`, in.ExamID, id).Scan(&it.ID, &it.ExamID, &it.LevelID, &it.GroupID, &it.StudentID, &it.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert student target: %w", err)
		}
		out = append(out, it)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return out, nil
}

type ExamQuestionSet struct {
	QuestionSetID string `json:"question_set_id"`
	NumQuestions  *int   `json:"num_questions,omitempty"`
}

func (r *Repo) ListQuestionSets(ctx context.Context, examID string) ([]ExamQuestionSet, error) {
	rows, err := r.pool.Query(ctx, `
SELECT question_set_id::text, num_questions
FROM exam_question_sets
WHERE exam_id = $1
ORDER BY question_set_id::text ASC`, examID)
	if err != nil {
		return nil, fmt.Errorf("list exam question sets: %w", err)
	}
	defer rows.Close()

	out := []ExamQuestionSet{}
	for rows.Next() {
		var it ExamQuestionSet
		if err := rows.Scan(&it.QuestionSetID, &it.NumQuestions); err != nil {
			return nil, fmt.Errorf("scan exam question set: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type ReplaceQuestionSetsInput struct {
	ExamID string
	Items  []ExamQuestionSet
}

func (r *Repo) ReplaceQuestionSets(ctx context.Context, in ReplaceQuestionSetsInput) ([]ExamQuestionSet, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM exam_question_sets WHERE exam_id = $1`, in.ExamID); err != nil {
		return nil, fmt.Errorf("clear exam question sets: %w", err)
	}

	out := []ExamQuestionSet{}
	for _, it := range in.Items {
		if _, err := tx.Exec(ctx, `
INSERT INTO exam_question_sets (exam_id, question_set_id, num_questions)
VALUES ($1::uuid, $2::uuid, $3)`, in.ExamID, it.QuestionSetID, it.NumQuestions); err != nil {
			return nil, fmt.Errorf("insert exam question set: %w", err)
		}
		out = append(out, it)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return out, nil
}
