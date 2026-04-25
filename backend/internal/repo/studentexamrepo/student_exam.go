package studentexamrepo

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sort"
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

type StudentInfo struct {
	StudentID string
	LevelID   string
	GroupID   string
	IsActive  bool
}

func (r *Repo) StudentByUserID(ctx context.Context, userID string) (StudentInfo, bool, error) {
	const q = `
SELECT s.id::text, COALESCE(s.level_id::text,''), COALESCE(s.group_id::text,''), u.is_active
FROM students s
JOIN users u ON u.id = s.user_id
WHERE s.user_id = $1
LIMIT 1`
	var it StudentInfo
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&it.StudentID, &it.LevelID, &it.GroupID, &it.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StudentInfo{}, false, nil
		}
		return StudentInfo{}, false, fmt.Errorf("student lookup: %w", err)
	}
	return it, true, nil
}

type StudentExam struct {
	ID                string `json:"id"`
	SubjectID         string `json:"subject_id"`
	SubjectName       string `json:"subject_name"`
	TeacherID         string `json:"teacher_id"`
	TeacherName       string `json:"teacher_name"`
	Title             string `json:"title"`
	StartsAt          string `json:"starts_at"`
	EndsAt            string `json:"ends_at"`
	DurationMinutes   *int   `json:"duration_minutes,omitempty"`
	Status            string `json:"status"`
	CanJoin           bool   `json:"can_join"`
	SessionID         string `json:"session_id,omitempty"`
	SessionStatus     string `json:"session_status,omitempty"`
	SessionFinished   string `json:"session_finished_at,omitempty"`
	MasterSessionID   string `json:"master_session_id,omitempty"`
	MasterSessionName string `json:"master_session_name,omitempty"`
	SessionStart      string `json:"session_start,omitempty"`
	SessionEnd        string `json:"session_end,omitempty"`
	ActiveToken       string `json:"active_token,omitempty"`
}

type ListStudentExamsFilter struct {
	Q      string
	Limit  int
	Offset int
	NowUTC time.Time
	Loc    *time.Location
}

func (r *Repo) ListAvailableForStudent(ctx context.Context, studentID, levelID, groupID string, f ListStudentExamsFilter) ([]StudentExam, int, error) {
	// List published exams where student is targeted. We include both ongoing and upcoming
	// (UI can show schedule); join validation is handled separately.
	const base = `
	FROM exams e
	JOIN subjects sub ON sub.id = e.subject_id
	JOIN teachers t ON t.id = e.teacher_id
	JOIN users tu ON tu.id = t.user_id
	LEFT JOIN sessions sess ON sess.id = e.session_id
	LEFT JOIN exam_sessions s
	  ON s.exam_id = e.id AND s.student_id::text = $2
	WHERE e.status = 'published'
	  AND ($1 = '' OR e.title ILIKE '%'||$1||'%')
  AND EXISTS (
    SELECT 1
    FROM exam_targets t
    WHERE t.exam_id = e.id
      AND (
        (t.student_id IS NOT NULL AND t.student_id::text = $2)
        OR (t.level_id IS NOT NULL AND $3 <> '' AND t.level_id::text = $3)
        OR (t.group_id IS NOT NULL AND $4 <> '' AND t.group_id::text = $4)
      )
  )
  AND NOT EXISTS (
    SELECT 1
    FROM student_exam_dismissals d
    WHERE d.exam_id = e.id
      AND d.student_id::text = $2
  )`

	rows, err := r.pool.Query(ctx, `
SELECT e.id::text, e.subject_id::text, sub.name, e.teacher_id::text, tu.name, e.title,
       to_char(e.starts_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       to_char(e.ends_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       e.duration_minutes, e.status,
       (now() >= e.starts_at AND now() <= e.ends_at AND COALESCE(s.status,'') = '') OR (now() >= e.starts_at AND now() <= e.ends_at AND s.status = 'in_progress') AS can_join,
       COALESCE(s.id::text,'') AS session_id,
       COALESCE(s.status,'') AS session_status,
       COALESCE(to_char(s.finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS session_finished_at,
       COALESCE(e.session_id::text,''), COALESCE(sess.name,''),
       COALESCE(tk.token, ''),
       COALESCE(to_char(sess.start_time, 'HH24:MI'), ''),
       COALESCE(to_char(sess.end_time, 'HH24:MI'), '')
	FROM exams e
	JOIN subjects sub ON sub.id = e.subject_id
	JOIN teachers t ON t.id = e.teacher_id
	JOIN users tu ON tu.id = t.user_id
	LEFT JOIN sessions sess ON sess.id = e.session_id
	LEFT JOIN exam_sessions s
	  ON s.exam_id = e.id AND s.student_id::text = $2
  LEFT JOIN LATERAL (
    SELECT token 
    FROM exam_tokens 
    WHERE exam_id = e.id 
      AND is_active = true 
    ORDER BY created_at DESC 
    LIMIT 1
  ) tk ON true
	WHERE e.status = 'published'
	  AND ($1 = '' OR e.title ILIKE '%'||$1||'%')
  AND EXISTS (
    SELECT 1
    FROM exam_targets t
    WHERE t.exam_id = e.id
      AND (
        (t.student_id IS NOT NULL AND t.student_id::text = $2)
        OR (t.level_id IS NOT NULL AND $3 <> '' AND t.level_id::text = $3)
        OR (t.group_id IS NOT NULL AND $4 <> '' AND t.group_id::text = $4)
      )
  )
  AND NOT EXISTS (
    SELECT 1
    FROM student_exam_dismissals d
    WHERE d.exam_id = e.id
      AND d.student_id::text = $2
  )
ORDER BY e.starts_at DESC, e.created_at DESC
LIMIT $5 OFFSET $6`, strings.TrimSpace(f.Q), studentID, levelID, groupID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list student exams: %w", err)
	}
	defer rows.Close()

	loc := f.Loc
	if loc == nil {
		loc = time.UTC
	}
	nowLocal := f.NowUTC.In(loc)
	y, m, d := nowLocal.Date()

	out := []StudentExam{}
	for rows.Next() {
		var it StudentExam
		var sessID, sessStatus, sessFinished string
		var masterSessID, masterSessName string
		var sStart, sEnd string
		if err := rows.Scan(&it.ID, &it.SubjectID, &it.SubjectName, &it.TeacherID, &it.TeacherName, &it.Title, &it.StartsAt, &it.EndsAt, &it.DurationMinutes, &it.Status, &it.CanJoin, &sessID, &sessStatus, &sessFinished, &masterSessID, &masterSessName, &it.ActiveToken, &sStart, &sEnd); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		if sessID != "" {
			it.SessionID = sessID
		}
		if sessStatus != "" {
			it.SessionStatus = sessStatus
		}
		if sessFinished != "" {
			it.SessionFinished = sessFinished
		}
		it.MasterSessionID = masterSessID
		it.MasterSessionName = masterSessName
		it.SessionStart = sStart
		it.SessionEnd = sEnd

		// Refine CanJoin with session window
		if it.CanJoin && sStart != "" && sEnd != "" {
			var h1, m1, h2, m2 int
			if _, err1 := fmt.Sscanf(sStart, "%d:%d", &h1, &m1); err1 == nil {
				if _, err2 := fmt.Sscanf(sEnd, "%d:%d", &h2, &m2); err2 == nil {
					winStart := time.Date(y, m, d, h1, m1, 0, 0, loc)
					winEnd := time.Date(y, m, d, h2, m2, 0, 0, loc)
					if nowLocal.Before(winStart) || nowLocal.After(winEnd) {
						it.CanJoin = false
					}
				}
			}
		}

		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base, strings.TrimSpace(f.Q), studentID, levelID, groupID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count student exams: %w", err)
	}
	return out, total, nil
}

var (
	ErrTokenNotFound   = errors.New("token not found")
	ErrTokenInactive   = errors.New("token inactive")
	ErrTokenNotStarted = errors.New("token not started")
	ErrTokenExpired    = errors.New("token expired")
)

func (r *Repo) VerifyExamToken(ctx context.Context, examID, token string, nowUTC time.Time) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrTokenNotFound
	}
	const q = `
SELECT is_active,
       valid_from,
       valid_to
FROM exam_tokens
WHERE exam_id = $1 AND token = $2
LIMIT 1`
	var active bool
	var validFrom *time.Time
	var validTo *time.Time
	if err := r.pool.QueryRow(ctx, q, examID, token).Scan(&active, &validFrom, &validTo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("verify token: %w", err)
	}
	if !active {
		return ErrTokenInactive
	}
	if validFrom != nil && nowUTC.Before(validFrom.UTC()) {
		return ErrTokenNotStarted
	}
	if validTo != nil && nowUTC.After(validTo.UTC()) {
		return ErrTokenExpired
	}
	return nil
}

var (
	ErrExamNotFound        = errors.New("exam not found")
	ErrExamNotJoinable     = errors.New("exam not joinable")
	ErrNotTargeted         = errors.New("student not targeted")
	ErrNoQuestionSets      = errors.New("no question sets attached")
	ErrSessionTimeMismatch = errors.New("bukan sesi anda")
)

type ExamForJoin struct {
	ID               string
	SubjectID        string
	TeacherID        string
	Title            string
	StartsAt         time.Time
	EndsAt           time.Time
	DurationMinutes  *int
	ShuffleQuestions bool
	ShuffleOptions   bool
	Status           string
	SessionID        *string
	SessionStart     *string
	SessionEnd       *string
}

func (r *Repo) GetExamForStudentJoin(ctx context.Context, examID, studentID, levelID, groupID string, nowUTC time.Time, loc *time.Location) (ExamForJoin, error) {
	// Authorization + schedule validation in one query.
	const q = `
SELECT e.id::text,
       e.subject_id::text,
       e.teacher_id::text,
       e.title,
       e.starts_at,
       e.ends_at,
       e.duration_minutes,
       e.shuffle_questions,
       e.shuffle_options,
       e.status,
       e.session_id::text,
       COALESCE(to_char(sess.start_time, 'HH24:MI'), ''),
       COALESCE(to_char(sess.end_time, 'HH24:MI'), '')
FROM exams e
LEFT JOIN sessions sess ON sess.id = e.session_id
WHERE e.id = $1
  AND e.status = 'published'
  AND EXISTS (
    SELECT 1
    FROM exam_targets t
    WHERE t.exam_id = e.id
      AND (
        (t.student_id IS NOT NULL AND t.student_id::text = $2)
        OR (t.level_id IS NOT NULL AND $3 <> '' AND t.level_id::text = $3)
        OR (t.group_id IS NOT NULL AND $4 <> '' AND t.group_id::text = $4)
      )
  )
LIMIT 1`
	var ex ExamForJoin
	if err := r.pool.QueryRow(ctx, q, examID, studentID, levelID, groupID).Scan(
		&ex.ID,
		&ex.SubjectID,
		&ex.TeacherID,
		&ex.Title,
		&ex.StartsAt,
		&ex.EndsAt,
		&ex.DurationMinutes,
		&ex.ShuffleQuestions,
		&ex.ShuffleOptions,
		&ex.Status,
		&ex.SessionID,
		&ex.SessionStart,
		&ex.SessionEnd,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Could be either not found or not targeted; callers map to a safe error.
			return ExamForJoin{}, ErrExamNotFound
		}
		return ExamForJoin{}, fmt.Errorf("get exam for join: %w", err)
	}

	if nowUTC.Before(ex.StartsAt.UTC()) || nowUTC.After(ex.EndsAt.UTC()) {
		return ExamForJoin{}, ErrExamNotJoinable
	}

	// Session validation
	if ex.SessionID != nil && ex.SessionStart != nil && *ex.SessionStart != "" && ex.SessionEnd != nil && *ex.SessionEnd != "" {
		if loc == nil {
			loc = time.UTC
		}
		// Enforce time window on the current day in the target timezone.
		var h1, m1, h2, m2 int
		if _, err := fmt.Sscanf(*ex.SessionStart, "%d:%d", &h1, &m1); err == nil {
			if _, err := fmt.Sscanf(*ex.SessionEnd, "%d:%d", &h2, &m2); err == nil {
				// Convert nowUTC to local time for date and comparison.
				nowLocal := nowUTC.In(loc)
				y, m, d := nowLocal.Date()

				sessStart := time.Date(y, m, d, h1, m1, 0, 0, loc)
				sessEnd := time.Date(y, m, d, h2, m2, 0, 0, loc)

				if nowLocal.Before(sessStart) || nowLocal.After(sessEnd) {
					return ExamForJoin{}, fmt.Errorf("%w (Sesi: %s - %s)", ErrSessionTimeMismatch, *ex.SessionStart, *ex.SessionEnd)
				}
			}
		}
	}

	return ex, nil
}

type Session struct {
	ID         string `json:"id"`
	ExamID     string `json:"exam_id"`
	StudentID  string `json:"student_id"`
	Status     string `json:"status"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	LastSeenAt string `json:"last_seen_at,omitempty"`
}

func (r *Repo) GetSessionByExamStudent(ctx context.Context, examID, studentID string) (Session, bool, error) {
	const q = `
SELECT id::text, exam_id::text, student_id::text, status,
       to_char(started_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at,
       COALESCE(to_char(last_seen_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS last_seen_at
FROM exam_sessions
WHERE exam_id = $1 AND student_id = $2
LIMIT 1`
	var it Session
	var finished, lastSeen string
	if err := r.pool.QueryRow(ctx, q, examID, studentID).Scan(&it.ID, &it.ExamID, &it.StudentID, &it.Status, &it.StartedAt, &finished, &lastSeen); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, false, nil
		}
		return Session{}, false, fmt.Errorf("get session: %w", err)
	}
	if finished != "" {
		it.FinishedAt = finished
	}
	if lastSeen != "" {
		it.LastSeenAt = lastSeen
	}
	return it, true, nil
}

func (r *Repo) CountInProgressSessionsByStudent(ctx context.Context, studentID string) (int, error) {
	const q = `
SELECT COUNT(*)
FROM exam_sessions
WHERE student_id = $1
  AND status = 'in_progress'`
	var total int
	if err := r.pool.QueryRow(ctx, q, studentID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count in-progress sessions: %w", err)
	}
	return total, nil
}

func computeDeadline(startedAt time.Time, examEndsAt time.Time, durationMinutes *int) time.Time {
	deadline := examEndsAt.UTC()
	if durationMinutes != nil {
		d := startedAt.UTC().Add(time.Duration(*durationMinutes) * time.Minute)
		if d.Before(deadline) {
			deadline = d
		}
	}
	return deadline
}

func (r *Repo) expireSessionIfNeeded(ctx context.Context, sessionID, studentID string, nowUTC time.Time) (bool, error) {
	// Returns true if the session was expired by this call.
	const q = `
SELECT s.status,
       s.started_at,
       e.ends_at,
       e.duration_minutes
FROM exam_sessions s
JOIN exams e ON e.id = s.exam_id
WHERE s.id = $1 AND s.student_id = $2
LIMIT 1`
	var status string
	var startedAt time.Time
	var examEndsAt time.Time
	var dur *int
	if err := r.pool.QueryRow(ctx, q, sessionID, studentID).Scan(&status, &startedAt, &examEndsAt, &dur); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrSessionNotFound
		}
		return false, fmt.Errorf("load session deadline: %w", err)
	}
	if status != "in_progress" {
		return false, nil
	}
	deadline := computeDeadline(startedAt, examEndsAt, dur)
	if !nowUTC.After(deadline) {
		return false, nil
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ct, err := tx.Exec(ctx, `
UPDATE exam_sessions
SET status = 'expired', finished_at = now(), updated_at = now()
WHERE id = $1 AND student_id = $2 AND status = 'in_progress'`, sessionID, studentID)
	if err != nil {
		return false, fmt.Errorf("expire session: %w", err)
	}
	if ct.RowsAffected() == 0 {
		// Race: someone else updated status.
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("commit: %w", err)
		}
		return false, nil
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_events (exam_session_id, type, payload_json)
VALUES ($1::uuid, 'expired', '{}'::jsonb)`, sessionID); err != nil {
		return false, fmt.Errorf("insert expired event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return true, nil
}

func (r *Repo) expireSessionIfNeededAny(ctx context.Context, sessionID string, nowUTC time.Time) (bool, error) {
	const q = `
SELECT s.status,
       s.started_at,
       e.ends_at,
       e.duration_minutes
FROM exam_sessions s
JOIN exams e ON e.id = s.exam_id
WHERE s.id = $1
LIMIT 1`
	var status string
	var startedAt time.Time
	var examEndsAt time.Time
	var dur *int
	if err := r.pool.QueryRow(ctx, q, sessionID).Scan(&status, &startedAt, &examEndsAt, &dur); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrSessionNotFound
		}
		return false, fmt.Errorf("load session deadline: %w", err)
	}
	if status != "in_progress" {
		return false, nil
	}
	deadline := computeDeadline(startedAt, examEndsAt, dur)
	if !nowUTC.After(deadline) {
		return false, nil
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	ct, err := tx.Exec(ctx, `
UPDATE exam_sessions
SET status = 'expired', finished_at = now(), updated_at = now()
WHERE id = $1 AND status = 'in_progress'`, sessionID)
	if err != nil {
		return false, fmt.Errorf("expire session: %w", err)
	}
	if ct.RowsAffected() == 0 {
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("commit: %w", err)
		}
		return false, nil
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_events (exam_session_id, type, payload_json)
VALUES ($1::uuid, 'expired', '{}'::jsonb)`, sessionID); err != nil {
		return false, fmt.Errorf("insert expired event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return true, nil
}

func (r *Repo) GetOrCreateSession(ctx context.Context, examID, studentID string, clientIP net.IP, userAgent string) (Session, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Session{}, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Race-safe: UNIQUE(exam_id, student_id) ensures only one session exists.
	// We avoid a read-then-insert race by using an upsert.
	const upsert = `
INSERT INTO exam_sessions (exam_id, student_id, client_ip, user_agent, last_seen_at)
VALUES ($1::uuid, $2::uuid, NULLIF($3,'')::inet, NULLIF($4,''), now())
ON CONFLICT (exam_id, student_id)
DO UPDATE SET last_seen_at = now(), updated_at = now()
RETURNING id::text, exam_id::text, student_id::text, status,
       to_char(started_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at,
       COALESCE(to_char(last_seen_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS last_seen_at`

	var it Session
	var finished, lastSeen string
	ipStr := ""
	if clientIP != nil {
		ipStr = clientIP.String()
	}
	if err := tx.QueryRow(ctx, upsert, examID, studentID, ipStr, strings.TrimSpace(userAgent)).Scan(&it.ID, &it.ExamID, &it.StudentID, &it.Status, &it.StartedAt, &finished, &lastSeen); err != nil {
		return Session{}, fmt.Errorf("upsert session: %w", err)
	}
	if finished != "" {
		it.FinishedAt = finished
	}
	if lastSeen != "" {
		it.LastSeenAt = lastSeen
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("commit: %w", err)
	}
	return it, nil
}

func (r *Repo) EnsureSessionQuestions(ctx context.Context, sessionID, examID string, shuffleQuestions bool) (int, error) {
	// This method may be called concurrently (join + reload + fetch questions).
	// We lock the session row to ensure only one assembler runs at a time.
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	{
		var one int
		if err := tx.QueryRow(ctx, `SELECT 1 FROM exam_sessions WHERE id = $1 FOR UPDATE`, sessionID).Scan(&one); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, fmt.Errorf("session not found")
			}
			return 0, fmt.Errorf("lock session: %w", err)
		}
	}

	// If already assembled, noop.
	var existing int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM exam_session_questions WHERE exam_session_id = $1`, sessionID).Scan(&existing); err != nil {
		return 0, fmt.Errorf("check assembled: %w", err)
	}
	if existing > 0 {
		if err := tx.Commit(ctx); err != nil {
			return 0, fmt.Errorf("commit: %w", err)
		}
		return existing, nil
	}

	type setRow struct {
		SetID        string
		NumQuestions *int
	}
	sets := []setRow{}
	srows, err := tx.Query(ctx, `
SELECT question_set_id::text, num_questions
FROM exam_question_sets
WHERE exam_id = $1
ORDER BY question_set_id::text ASC`, examID)
	if err != nil {
		return 0, fmt.Errorf("list exam_question_sets: %w", err)
	}
	defer srows.Close()
	for srows.Next() {
		var it setRow
		if err := srows.Scan(&it.SetID, &it.NumQuestions); err != nil {
			return 0, fmt.Errorf("scan set: %w", err)
		}
		sets = append(sets, it)
	}
	if err := srows.Err(); err != nil {
		return 0, err
	}
	if len(sets) == 0 {
		return 0, ErrNoQuestionSets
	}

	selected := []string{}
	for _, s := range sets {
		qids := []string{}
		qrows, err := tx.Query(ctx, `
SELECT id::text
FROM questions
WHERE question_set_id = $1
ORDER BY order_no ASC, id ASC`, s.SetID)
		if err != nil {
			return 0, fmt.Errorf("list questions set=%s: %w", s.SetID, err)
		}
		for qrows.Next() {
			var id string
			if err := qrows.Scan(&id); err != nil {
				qrows.Close()
				return 0, fmt.Errorf("scan question id: %w", err)
			}
			qids = append(qids, id)
		}
		if err := qrows.Err(); err != nil {
			qrows.Close()
			return 0, err
		}
		qrows.Close()

		if shuffleQuestions {
			shuffleDeterministic(qids, seed64(sessionID+":set:"+s.SetID))
		}
		if s.NumQuestions != nil && *s.NumQuestions < len(qids) {
			qids = qids[:*s.NumQuestions]
		}
		selected = append(selected, qids...)
	}

	if shuffleQuestions {
		shuffleDeterministic(selected, seed64(sessionID+":all"))
	}

	// Keep stable insert order (for debugging) even if duplicates slip in (they shouldn't).
	uniq := map[string]bool{}
	final := make([]string, 0, len(selected))
	for _, id := range selected {
		if uniq[id] {
			continue
		}
		uniq[id] = true
		final = append(final, id)
	}

	rows := [][]any{}
	for i, qid := range final {
		rows = append(rows, []any{sessionID, qid, i + 1})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"exam_session_questions"}, []string{"exam_session_id", "question_id", "order_no"}, pgx.CopyFromRows(rows)); err != nil {
		return 0, fmt.Errorf("copy session questions: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_events (exam_session_id, type, payload_json)
VALUES ($1::uuid, 'start', '{}'::jsonb)`, sessionID); err != nil {
		return 0, fmt.Errorf("insert event start: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return len(final), nil
}

type SessionState struct {
	Session Session `json:"session"`
	Exam    struct {
		ID              string `json:"id"`
		Title           string `json:"title"`
		StartsAt        string `json:"starts_at"`
		EndsAt          string `json:"ends_at"`
		DurationMinutes *int   `json:"duration_minutes,omitempty"`
		ShuffleOptions  bool   `json:"shuffle_options"`
	} `json:"exam"`
	DeadlineAt       string `json:"deadline_at"`
	RemainingSeconds int    `json:"remaining_seconds"`
}

func (r *Repo) GetSessionState(ctx context.Context, sessionID, studentID string, nowUTC time.Time) (SessionState, bool, error) {
	// Best-effort expiry: keep session status consistent if deadline has passed.
	if _, err := r.expireSessionIfNeeded(ctx, sessionID, studentID, nowUTC); err != nil && err != ErrSessionNotFound {
		return SessionState{}, false, err
	}

	const q = `
SELECT s.id::text, s.exam_id::text, s.student_id::text, s.status,
       to_char(s.started_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(s.finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at,
       COALESCE(to_char(s.last_seen_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS last_seen_at,
       e.title,
       to_char(e.starts_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       to_char(e.ends_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       e.duration_minutes,
       e.shuffle_options,
       e.ends_at
FROM exam_sessions s
JOIN exams e ON e.id = s.exam_id
WHERE s.id = $1 AND s.student_id = $2
LIMIT 1`

	var st SessionState
	var finished, lastSeen string
	var endsAt time.Time
	if err := r.pool.QueryRow(ctx, q, sessionID, studentID).Scan(
		&st.Session.ID,
		&st.Session.ExamID,
		&st.Session.StudentID,
		&st.Session.Status,
		&st.Session.StartedAt,
		&finished,
		&lastSeen,
		&st.Exam.Title,
		&st.Exam.StartsAt,
		&st.Exam.EndsAt,
		&st.Exam.DurationMinutes,
		&st.Exam.ShuffleOptions,
		&endsAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionState{}, false, nil
		}
		return SessionState{}, false, fmt.Errorf("get session state: %w", err)
	}
	st.Exam.ID = st.Session.ExamID
	if finished != "" {
		st.Session.FinishedAt = finished
	}
	if lastSeen != "" {
		st.Session.LastSeenAt = lastSeen
	}

	startedAt, err := time.Parse(time.RFC3339, st.Session.StartedAt)
	if err != nil {
		return SessionState{}, false, fmt.Errorf("parse started_at: %w", err)
	}
	deadline := computeDeadline(startedAt, endsAt, st.Exam.DurationMinutes)
	st.DeadlineAt = deadline.Format(time.RFC3339)
	rem := int(deadline.Sub(nowUTC).Seconds())
	if rem < 0 {
		rem = 0
	}
	st.RemainingSeconds = rem
	return st, true, nil
}

func (r *Repo) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	var ok bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM exam_sessions WHERE id = $1)`, sessionID).Scan(&ok); err != nil {
		return false, fmt.Errorf("session exists: %w", err)
	}
	return ok, nil
}

type StudentQuestion struct {
	OrderNo int    `json:"order_no"`
	ID      string `json:"id"`
	Type    string `json:"type"`
	Stem    string `json:"stem"`

	// mc_single/mc_multiple
	Options []StudentOption `json:"options,omitempty"`

	// matching: send left and right lists without mapping.
	MatchingLeft  []StudentMatchingItem `json:"matching_left,omitempty"`
	MatchingRight []StudentMatchingItem `json:"matching_right,omitempty"`

	// essay
	MaxScore *int `json:"max_score,omitempty"`

	// true_false: no "correct" field exposed, but we need the statements.
	Statements []StudentTFStatement `json:"statements,omitempty"`
}

type StudentTFStatement struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	OrderNo int    `json:"order_no"`
}

type StudentOption struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Content string `json:"content"`
}

type StudentMatchingItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	OrderNo int    `json:"order_no"`
}

func (r *Repo) ListSessionQuestions(ctx context.Context, sessionID, studentID string, shuffleOptions bool) ([]StudentQuestion, error) {
	// Authorize ownership via join to session.
	const q = `
SELECT sq.order_no, q.id::text, q.type, q.stem
FROM exam_session_questions sq
JOIN exam_sessions s ON s.id = sq.exam_session_id
JOIN questions q ON q.id = sq.question_id
WHERE sq.exam_session_id = $1 AND s.student_id = $2
ORDER BY sq.order_no ASC`
	rows, err := r.pool.Query(ctx, q, sessionID, studentID)
	if err != nil {
		return nil, fmt.Errorf("list session questions: %w", err)
	}
	defer rows.Close()

	out := []StudentQuestion{}
	byQID := map[string]*StudentQuestion{}
	qids := []string{}
	for rows.Next() {
		var it StudentQuestion
		if err := rows.Scan(&it.OrderNo, &it.ID, &it.Type, &it.Stem); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
		byQID[it.ID] = &out[len(out)-1]
		qids = append(qids, it.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return []StudentQuestion{}, nil
	}

	// Options for MC
	orows, err := r.pool.Query(ctx, `
SELECT id::text, question_id::text, label, content
FROM question_options
WHERE question_id::text = ANY($1::text[])
ORDER BY label ASC, id ASC`, qids)
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	defer orows.Close()
	for orows.Next() {
		var oid, qid, label, content string
		if err := orows.Scan(&oid, &qid, &label, &content); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		if q := byQID[qid]; q != nil {
			q.Options = append(q.Options, StudentOption{ID: oid, Label: label, Content: content})
		}
	}
	if err := orows.Err(); err != nil {
		return nil, err
	}

	// Matching pairs: store as left and right item lists.
	type matchPair struct {
		ID    string
		QID   string
		Left  string
		Right string
		Order int
	}
	pairsByQ := map[string][]matchPair{}
	prows, err := r.pool.Query(ctx, `
SELECT id::text, question_id::text, left_content, right_content, order_no
FROM question_matching_pairs
WHERE question_id::text = ANY($1::text[])
ORDER BY order_no ASC, id ASC`, qids)
	if err != nil {
		return nil, fmt.Errorf("list matching pairs: %w", err)
	}
	defer prows.Close()
	for prows.Next() {
		var p matchPair
		if err := prows.Scan(&p.ID, &p.QID, &p.Left, &p.Right, &p.Order); err != nil {
			return nil, fmt.Errorf("scan pair: %w", err)
		}
		pairsByQ[p.QID] = append(pairsByQ[p.QID], p)
	}
	if err := prows.Err(); err != nil {
		return nil, err
	}
	for qid, ps := range pairsByQ {
		q := byQID[qid]
		if q == nil {
			continue
		}
		left := make([]StudentMatchingItem, 0, len(ps))
		right := make([]StudentMatchingItem, 0, len(ps))
		for _, p := range ps {
			left = append(left, StudentMatchingItem{ID: p.ID + ":L", Content: p.Left, OrderNo: p.Order})
			right = append(right, StudentMatchingItem{ID: p.ID + ":R", Content: p.Right, OrderNo: p.Order})
		}
		if shuffleOptions {
			shuffleDeterministic(right, seed64(sessionID+":match:"+qid))
		} else {
			sort.SliceStable(right, func(i, j int) bool { return right[i].OrderNo < right[j].OrderNo })
		}
		q.MatchingLeft = left
		q.MatchingRight = right
	}

	// Essays: expose max_score only.
	erows, err := r.pool.Query(ctx, `
SELECT question_id::text, max_score
FROM question_essays
WHERE question_id::text = ANY($1::text[])`, qids)
	if err != nil {
		return nil, fmt.Errorf("list essays: %w", err)
	}
	defer erows.Close()
	for erows.Next() {
		var qid string
		var maxScore *int
		if err := erows.Scan(&qid, &maxScore); err != nil {
			return nil, fmt.Errorf("scan essay: %w", err)
		}
		if q := byQID[qid]; q != nil {
			q.MaxScore = maxScore
		}
	}
	if err := erows.Err(); err != nil {
		return nil, err
	}

	// Fetch true/false statements
	strows, err := r.pool.Query(ctx, `
SELECT id, question_id::text, content, order_no
FROM question_true_false_statements
WHERE question_id::text = ANY($1::text[])
ORDER BY order_no ASC, id ASC`, qids)
	if err != nil {
		return nil, fmt.Errorf("list statements: %w", err)
	}
	defer strows.Close()
	for strows.Next() {
		var st StudentTFStatement
		var qid string
		if err := strows.Scan(&st.ID, &qid, &st.Content, &st.OrderNo); err != nil {
			return nil, fmt.Errorf("scan statement: %w", err)
		}
		if q := byQID[qid]; q != nil {
			q.Statements = append(q.Statements, st)
		}
	}
	if err := strows.Err(); err != nil {
		return nil, err
	}

	// Shuffle MC options per question if requested. Must be deterministic per session+question.
	if shuffleOptions {
		for _, q := range out {
			if len(q.Options) > 1 {
				shuffleDeterministic(q.Options, seed64(sessionID+":opt:"+q.ID))
			}
		}
	}

	return out, nil
}

type SessionAnswer struct {
	QuestionID string          `json:"question_id"`
	AnswerJSON json.RawMessage `json:"answer_json"`
	AnsweredAt string          `json:"answered_at"`
}

func (r *Repo) ListSessionAnswers(ctx context.Context, sessionID, studentID string) ([]SessionAnswer, error) {
	// Authorize ownership via join to session.
	const q = `
SELECT a.question_id::text,
       a.answer_json,
       to_char(a.answered_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
FROM exam_attempts a
JOIN exam_sessions s ON s.id = a.exam_session_id
WHERE a.exam_session_id = $1 AND s.student_id = $2
ORDER BY a.answered_at DESC`
	rows, err := r.pool.Query(ctx, q, sessionID, studentID)
	if err != nil {
		return nil, fmt.Errorf("list session answers: %w", err)
	}
	defer rows.Close()

	out := []SessionAnswer{}
	for rows.Next() {
		var it SessionAnswer
		var raw []byte
		if err := rows.Scan(&it.QuestionID, &raw, &it.AnsweredAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		it.AnswerJSON = json.RawMessage(raw)
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrQuestionNotInSession = errors.New("question not in session")
	ErrSessionNotActive     = errors.New("session not active")
	ErrExamNotDismissible   = errors.New("exam not dismissible")
)

func (r *Repo) DismissExamCard(ctx context.Context, examID, studentID string) error {
	// Allow dismiss when:
	// 1) student has a completed session, or
	// 2) exam schedule has ended for a targeted exam (even if student did not join).
	var canDismiss bool
	if err := r.pool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM students st
  JOIN exams e ON e.id = $1
  WHERE st.id = $2
    AND (
      EXISTS (
        SELECT 1
        FROM exam_sessions s
        WHERE s.exam_id = e.id
          AND s.student_id = st.id
          AND s.status IN ('submitted','forced','expired')
      )
      OR (
        e.ends_at <= now()
        AND EXISTS (
          SELECT 1
          FROM exam_targets t
          WHERE t.exam_id = e.id
            AND (
              (t.student_id IS NOT NULL AND t.student_id = st.id)
              OR (t.level_id IS NOT NULL AND st.level_id IS NOT NULL AND t.level_id = st.level_id)
              OR (t.group_id IS NOT NULL AND st.group_id IS NOT NULL AND t.group_id = st.group_id)
            )
        )
      )
    )
)`, examID, studentID).Scan(&canDismiss); err != nil {
		return fmt.Errorf("check dismiss eligibility: %w", err)
	}
	if !canDismiss {
		return ErrExamNotDismissible
	}

	if _, err := r.pool.Exec(ctx, `
INSERT INTO student_exam_dismissals (student_id, exam_id)
VALUES ($1, $2)
ON CONFLICT (student_id, exam_id) DO NOTHING
`, studentID, examID); err != nil {
		return fmt.Errorf("dismiss exam card: %w", err)
	}
	return nil
}

func (r *Repo) UpsertAnswer(ctx context.Context, sessionID, studentID, questionID string, answerJSON json.RawMessage, nowUTC time.Time) error {
	if _, err := r.expireSessionIfNeeded(ctx, sessionID, studentID, nowUTC); err != nil {
		// If not found we'll still fall through to authorization query and return a consistent error.
		if err != ErrSessionNotFound {
			return err
		}
	}

	// Validate ownership + question exists in session.
	const authQ = `
SELECT s.status
FROM exam_sessions s
JOIN exam_session_questions sq ON sq.exam_session_id = s.id AND sq.question_id = $3
WHERE s.id = $1 AND s.student_id = $2
LIMIT 1`
	var status string
	if err := r.pool.QueryRow(ctx, authQ, sessionID, studentID, questionID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrQuestionNotInSession
		}
		return fmt.Errorf("authorize answer: %w", err)
	}
	if status != "in_progress" {
		return ErrSessionNotActive
	}

	if len(answerJSON) == 0 {
		return fmt.Errorf("answer_json required")
	}
	// Ensure it's valid JSON (even if schema varies by type for now).
	var tmp any
	if err := json.Unmarshal(answerJSON, &tmp); err != nil {
		return fmt.Errorf("invalid answer_json")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const up = `
INSERT INTO exam_attempts (exam_session_id, question_id, answer_json, answered_at, updated_at)
VALUES ($1::uuid, $2::uuid, $3::jsonb, $4, now())
ON CONFLICT (exam_session_id, question_id)
DO UPDATE SET answer_json = EXCLUDED.answer_json, answered_at = EXCLUDED.answered_at, updated_at = now()`
	if _, err := tx.Exec(ctx, up, sessionID, questionID, string(answerJSON), nowUTC); err != nil {
		return fmt.Errorf("upsert attempt: %w", err)
	}

	const ev = `
INSERT INTO exam_events (exam_session_id, type, payload_json)
VALUES ($1::uuid, 'answer', jsonb_build_object('question_id',$2::text))`
	if _, err := tx.Exec(ctx, ev, sessionID, questionID); err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE exam_sessions SET last_seen_at = now(), updated_at = now() WHERE id = $1`, sessionID); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *Repo) SubmitSession(ctx context.Context, sessionID, studentID string) error {
	if _, err := r.expireSessionIfNeeded(ctx, sessionID, studentID, time.Now().UTC()); err != nil {
		if err != ErrSessionNotFound {
			return err
		}
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
UPDATE exam_sessions
SET status = 'submitted', finished_at = now(), updated_at = now()
WHERE id = $1 AND student_id = $2 AND status = 'in_progress'`
	ct, err := tx.Exec(ctx, q, sessionID, studentID)
	if err != nil {
		return fmt.Errorf("submit session: %w", err)
	}
	if ct.RowsAffected() == 0 {
		// Distinguish not found vs not active.
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM exam_sessions WHERE id = $1 AND student_id = $2)`, sessionID, studentID).Scan(&exists); err != nil {
			return fmt.Errorf("check session: %w", err)
		}
		if !exists {
			return ErrSessionNotFound
		}
		return ErrSessionNotActive
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_events (exam_session_id, type, payload_json)
VALUES ($1::uuid, 'submit', '{}'::jsonb)`, sessionID); err != nil {
		return fmt.Errorf("insert submit event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if err := r.updateSessionScore(ctx, sessionID, studentID); err != nil {
		return err
	}
	return nil
}

func (r *Repo) ForceSubmitSession(ctx context.Context, sessionID string) error {
	if _, err := r.expireSessionIfNeededAny(ctx, sessionID, time.Now().UTC()); err != nil {
		if err != ErrSessionNotFound {
			return err
		}
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
UPDATE exam_sessions
SET status = 'forced', finished_at = now(), updated_at = now()
WHERE id = $1 AND status = 'in_progress'`
	ct, err := tx.Exec(ctx, q, sessionID)
	if err != nil {
		return fmt.Errorf("force submit session: %w", err)
	}
	if ct.RowsAffected() == 0 {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM exam_sessions WHERE id = $1)`, sessionID).Scan(&exists); err != nil {
			return fmt.Errorf("check session: %w", err)
		}
		if !exists {
			return ErrSessionNotFound
		}
		return ErrSessionNotActive
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_events (exam_session_id, type, payload_json)
VALUES ($1::uuid, 'force_submit', '{}'::jsonb)`, sessionID); err != nil {
		return fmt.Errorf("insert force submit event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if err := r.updateSessionScoreAny(ctx, sessionID); err != nil {
		return err
	}
	return nil
}

func (r *Repo) updateSessionScore(ctx context.Context, sessionID, studentID string) error {
	sum, err := r.ComputeAutoScore(ctx, sessionID, studentID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("compute score: %w", err)
	}
	score := sum.Score
	gradingJSON, err := json.Marshal(sum.GradingDetails)
	if err != nil {
		return fmt.Errorf("marshal grading details: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `
	UPDATE exam_sessions
	SET score = $1, grading_detail_json = $2::jsonb, updated_at = now()
	WHERE id = $3 AND student_id = $4 AND status IN ('submitted','forced')`, score, string(gradingJSON), sessionID, studentID); err != nil {
		return fmt.Errorf("save score: %w", err)
	}
	return nil
}

func (r *Repo) updateSessionScoreAny(ctx context.Context, sessionID string) error {
	sum, err := r.ComputeAutoScoreAny(ctx, sessionID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("compute score: %w", err)
	}
	score := sum.Score
	gradingJSON, err := json.Marshal(sum.GradingDetails)
	if err != nil {
		return fmt.Errorf("marshal grading details: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `
	UPDATE exam_sessions
	SET score = $1, grading_detail_json = $2::jsonb, updated_at = now()
	WHERE id = $3 AND status IN ('submitted','forced')`, score, string(gradingJSON), sessionID); err != nil {
		return fmt.Errorf("save score: %w", err)
	}
	return nil
}

func (r *Repo) Heartbeat(ctx context.Context, sessionID, studentID string, payload json.RawMessage) error {
	if _, err := r.expireSessionIfNeeded(ctx, sessionID, studentID, time.Now().UTC()); err != nil {
		if err != ErrSessionNotFound {
			return err
		}
	}

	// Minimal validation: session exists and belongs to student.
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var ok bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM exam_sessions WHERE id = $1 AND student_id = $2)`, sessionID, studentID).Scan(&ok); err != nil {
		return fmt.Errorf("check session: %w", err)
	}
	if !ok {
		return ErrSessionNotFound
	}

	if len(payload) > 0 {
		var tmp any
		if err := json.Unmarshal(payload, &tmp); err != nil {
			return fmt.Errorf("invalid payload_json")
		}
	} else {
		payload = json.RawMessage(`{}`)
	}

	if _, err := tx.Exec(ctx, `UPDATE exam_sessions SET last_seen_at = now(), updated_at = now() WHERE id = $1`, sessionID); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO exam_events (exam_session_id, type, payload_json)
VALUES ($1::uuid, 'heartbeat', $2::jsonb)`, sessionID, string(payload)); err != nil {
		return fmt.Errorf("insert heartbeat: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func seed64(s string) int64 {
	sum := sha256.Sum256([]byte(s))
	return int64(binary.LittleEndian.Uint64(sum[:8]))
}

func shuffleDeterministic[T any](items []T, seed int64) {
	rng := rand.New(rand.NewSource(seed)) // deterministic
	for i := len(items) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		items[i], items[j] = items[j], items[i]
	}
}

// For reproducibility, ensure a stable order for maps serialized as JSON.
func MarshalStableJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// Attempt to re-marshal with stable ordering if it's an object. This is best-effort.
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err == nil {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		ordered := make([]any, 0, len(keys)*2)
		for _, k := range keys {
			ordered = append(ordered, k, obj[k])
		}
		// Fallback: return normal marshal output if ordering trick fails.
		_ = ordered
	}
	return b, nil
}
