package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"mycbt/backend/internal/config"
	"mycbt/backend/internal/db"
)

type rowCount struct {
	Label string
	Query string
}

var countQueries = []rowCount{
	{Label: "users_admin", Query: "SELECT COUNT(*) FROM users WHERE role='admin'"},
	{Label: "users_teacher", Query: "SELECT COUNT(*) FROM users WHERE role='teacher'"},
	{Label: "users_student", Query: "SELECT COUNT(*) FROM users WHERE role='student'"},
	{Label: "teachers", Query: "SELECT COUNT(*) FROM teachers"},
	{Label: "students", Query: "SELECT COUNT(*) FROM students"},
	{Label: "exams", Query: "SELECT COUNT(*) FROM exams"},
	{Label: "question_sets", Query: "SELECT COUNT(*) FROM question_sets"},
	{Label: "questions", Query: "SELECT COUNT(*) FROM questions"},
	{Label: "exam_sessions", Query: "SELECT COUNT(*) FROM exam_sessions"},
	{Label: "exam_attempts", Query: "SELECT COUNT(*) FROM exam_attempts"},
	{Label: "announcements", Query: "SELECT COUNT(*) FROM announcements"},
	{Label: "registration_requests", Query: "SELECT COUNT(*) FROM registration_requests"},
}

var cleanupStatements = []string{
	"DELETE FROM password_reset_tokens",
	"DELETE FROM login_logs",
	"DELETE FROM audit_logs",
	"DELETE FROM student_exam_dismissals",
	"DELETE FROM student_attendance",
	"DELETE FROM attendance_sessions",
	"DELETE FROM exam_events",
	"DELETE FROM exam_attempts",
	"DELETE FROM exam_session_questions",
	"DELETE FROM exam_sessions",
	"DELETE FROM lti_ags_launches",
	"DELETE FROM lti_sessions",
	"DELETE FROM lti_users",
	"DELETE FROM lti_nonces",
	"DELETE FROM exam_tokens",
	"DELETE FROM exam_targets",
	"DELETE FROM exam_question_sets",
	"DELETE FROM exams",
	"DELETE FROM question_true_false_statements",
	"DELETE FROM question_true_false",
	"DELETE FROM question_essays",
	"DELETE FROM question_short_answers",
	"DELETE FROM question_matching_pairs",
	"DELETE FROM question_options",
	"DELETE FROM questions",
	"DELETE FROM question_sets",
	"DELETE FROM announcements",
	"DELETE FROM registration_requests",
	"DELETE FROM teacher_subjects",
	"DELETE FROM teacher_groups",
	"DELETE FROM teacher_levels",
	"DELETE FROM teachers",
	"DELETE FROM students",
	"DELETE FROM users WHERE role IN ('teacher','student')",
}

func main() {
	execute := flag.Bool("execute", false, "execute cleanup (default: dry-run only)")
	flag.Parse()

	cfg := config.Load()
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	d, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer d.Pool.Close()

	fmt.Println("== PRE-DEPLOY CLEANUP ==")
	fmt.Println("Mode:", map[bool]string{true: "EXECUTE", false: "DRY-RUN"}[*execute])
	printCounts(ctx, d)

	if !*execute {
		fmt.Println("\nDry-run complete. No data changed.")
		fmt.Println("To execute cleanup, run:")
		fmt.Println("  go run ./cmd/cleanup --execute")
		return
	}

	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for i, q := range cleanupStatements {
		if _, err := tx.Exec(ctx, q); err != nil {
			log.Fatalf("cleanup failed at step %d (%s): %v", i+1, q, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit: %v", err)
	}

	fmt.Println("\nCleanup executed successfully.")
	printCounts(ctx, d)
}

func printCounts(ctx context.Context, d *db.DB) {
	fmt.Println("\nCurrent counts:")
	for _, item := range countQueries {
		var n int64
		if err := d.Pool.QueryRow(ctx, item.Query).Scan(&n); err != nil {
			fmt.Printf("- %-24s error: %v\n", item.Label, err)
			continue
		}
		fmt.Printf("- %-24s %d\n", item.Label, n)
	}
}

