package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sourceExam struct {
	ID               string
	SubjectID        string
	TeacherID        string
	ShuffleQuestions bool
	ShuffleOptions   bool
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func loadSourceExam(ctx context.Context, tx pgx.Tx, studentID, explicitExamID string) (sourceExam, error) {
	if explicitExamID != "" {
		var ex sourceExam
		err := tx.QueryRow(ctx, `
SELECT id::text, subject_id::text, teacher_id::text, shuffle_questions, shuffle_options
FROM exams
WHERE id = $1
  AND EXISTS (SELECT 1 FROM exam_question_sets eqs WHERE eqs.exam_id = exams.id)
LIMIT 1`, explicitExamID).Scan(&ex.ID, &ex.SubjectID, &ex.TeacherID, &ex.ShuffleQuestions, &ex.ShuffleOptions)
		if err == nil {
			return ex, nil
		}
		if err != nil && err != pgx.ErrNoRows {
			return sourceExam{}, fmt.Errorf("load explicit source exam: %w", err)
		}
	}

	var ex sourceExam
	err := tx.QueryRow(ctx, `
SELECT e.id::text, e.subject_id::text, e.teacher_id::text, e.shuffle_questions, e.shuffle_options
FROM exams e
WHERE e.status = 'published'
  AND EXISTS (SELECT 1 FROM exam_question_sets eqs WHERE eqs.exam_id = e.id)
  AND EXISTS (SELECT 1 FROM exam_targets t WHERE t.exam_id = e.id AND t.student_id = $1::uuid)
ORDER BY e.updated_at DESC
LIMIT 1`, studentID).Scan(&ex.ID, &ex.SubjectID, &ex.TeacherID, &ex.ShuffleQuestions, &ex.ShuffleOptions)
	if err == nil {
		return ex, nil
	}
	if err != nil && err != pgx.ErrNoRows {
		return sourceExam{}, fmt.Errorf("load targeted source exam: %w", err)
	}

	err = tx.QueryRow(ctx, `
SELECT e.id::text, e.subject_id::text, e.teacher_id::text, e.shuffle_questions, e.shuffle_options
FROM exams e
WHERE e.status = 'published'
  AND EXISTS (SELECT 1 FROM exam_question_sets eqs WHERE eqs.exam_id = e.id)
ORDER BY e.updated_at DESC
LIMIT 1`).Scan(&ex.ID, &ex.SubjectID, &ex.TeacherID, &ex.ShuffleQuestions, &ex.ShuffleOptions)
	if err != nil {
		if err == pgx.ErrNoRows {
			return sourceExam{}, fmt.Errorf("no published exam with question sets found")
		}
		return sourceExam{}, fmt.Errorf("load fallback source exam: %w", err)
	}
	return ex, nil
}

func main() {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		fmt.Println("DATABASE_URL is required")
		os.Exit(1)
	}

	username := getenv("FIXTURE_USERNAME", "siswa1")
	explicitExamID := strings.TrimSpace(os.Getenv("FIXTURE_SOURCE_EXAM_ID"))
	token := getenv("FIXTURE_TOKEN", "LOADTEST123")
	titlePrefix := getenv("FIXTURE_TITLE_PREFIX", "LOADTEST ACTIVE")
	durationMin := getenvInt("FIXTURE_DURATION_MINUTES", 60)
	windowMin := getenvInt("FIXTURE_WINDOW_MINUTES", 120)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Printf("open pool error: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		fmt.Printf("begin tx error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var studentID string
	var userID string
	if err := tx.QueryRow(ctx, `
SELECT s.id::text, u.id::text
FROM students s
JOIN users u ON u.id = s.user_id
WHERE u.username = $1
LIMIT 1`, username).Scan(&studentID, &userID); err != nil {
		fmt.Printf("student lookup error (%s): %v\n", username, err)
		os.Exit(1)
	}

	src, err := loadSourceExam(ctx, tx, studentID, explicitExamID)
	if err != nil {
		fmt.Printf("source exam error: %v\n", err)
		os.Exit(1)
	}

	var creatorUserID string
	if err := tx.QueryRow(ctx, `
SELECT id::text
FROM users
WHERE role = 'admin'
ORDER BY created_at ASC
LIMIT 1`).Scan(&creatorUserID); err != nil {
		if err == pgx.ErrNoRows {
			creatorUserID = userID
		} else {
			fmt.Printf("creator user lookup error: %v\n", err)
			os.Exit(1)
		}
	}

	now := time.Now().UTC()
	startsAt := now.Add(-10 * time.Minute)
	endsAt := now.Add(time.Duration(windowMin) * time.Minute)
	title := fmt.Sprintf("%s %s", titlePrefix, now.Format("20060102-150405"))

	var newExamID string
	err = tx.QueryRow(ctx, `
INSERT INTO exams (
  subject_id, teacher_id, title, starts_at, ends_at, duration_minutes,
  shuffle_questions, shuffle_options, status, created_at, updated_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, 'published', now(), now()
)
RETURNING id::text`, src.SubjectID, src.TeacherID, title, startsAt, endsAt, durationMin, src.ShuffleQuestions, src.ShuffleOptions).Scan(&newExamID)
	if err != nil {
		fmt.Printf("insert new exam error: %v\n", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_question_sets (exam_id, question_set_id, num_questions)
SELECT $1::uuid, question_set_id, num_questions
FROM exam_question_sets
WHERE exam_id = $2::uuid`, newExamID, src.ID); err != nil {
		fmt.Printf("copy exam_question_sets error: %v\n", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_targets (exam_id, student_id)
VALUES ($1::uuid, $2::uuid)`, newExamID, studentID); err != nil {
		fmt.Printf("insert exam target error: %v\n", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO exam_tokens (
  exam_id, token, valid_from, valid_to, is_active, created_by_user_id, created_at
) VALUES (
  $1::uuid, $2, now() - interval '1 minute', now() + interval '12 hours', true, $3::uuid, now()
)`, newExamID, token, creatorUserID); err != nil {
		fmt.Printf("insert exam token error: %v\n", err)
		os.Exit(1)
	}

	var sampleQuestionID string
	if err := tx.QueryRow(ctx, `
SELECT q.id::text
FROM exam_question_sets eqs
JOIN questions q ON q.question_set_id = eqs.question_set_id
WHERE eqs.exam_id = $1::uuid
ORDER BY q.order_no ASC, q.id ASC
LIMIT 1`, newExamID).Scan(&sampleQuestionID); err != nil {
		fmt.Printf("sample question lookup error: %v\n", err)
		os.Exit(1)
	}

	var activeSessions int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*)
FROM exam_sessions
WHERE student_id = $1::uuid
  AND status = 'in_progress'`, studentID).Scan(&activeSessions); err != nil {
		fmt.Printf("active session count error: %v\n", err)
		os.Exit(1)
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Printf("commit error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("fixture_created=true\n")
	fmt.Printf("fixture_username=%s\n", username)
	fmt.Printf("source_exam_id=%s\n", src.ID)
	fmt.Printf("fixture_exam_id=%s\n", newExamID)
	fmt.Printf("fixture_token=%s\n", token)
	fmt.Printf("fixture_sample_question_id=%s\n", sampleQuestionID)
	fmt.Printf("fixture_active_sessions_for_student=%d\n", activeSessions)
	fmt.Printf("fixture_starts_at_utc=%s\n", startsAt.Format(time.RFC3339))
	fmt.Printf("fixture_ends_at_utc=%s\n", endsAt.Format(time.RFC3339))
	fmt.Println("next_step_join=POST /api/v1/student/exams/{fixture_exam_id}/join body:{\"token\":\"" + token + "\"}")
}
