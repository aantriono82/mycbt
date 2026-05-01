package examrepo

import (
	"context"
	"testing"
	"time"

	"atigacbt/backend/internal/testutil/pgtest"
)

func mustInsertUser(t *testing.T, ctx context.Context, h *pgtest.Harness, username, role string) string {
	t.Helper()
	var id string
	err := h.Pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, role, name, email, is_active)
		VALUES ($1, 'hash', $2, $3, $4, true)
		RETURNING id::text
	`, username, role, username, username+"@example.com").Scan(&id)
	if err != nil {
		t.Fatalf("insert user %s: %v", username, err)
	}
	return id
}

func mustInsertTeacher(t *testing.T, ctx context.Context, h *pgtest.Harness, userID string) string {
	t.Helper()
	var id string
	err := h.Pool.QueryRow(ctx, `INSERT INTO teachers (user_id, nip) VALUES ($1, gen_random_uuid()::text) RETURNING id::text`, userID).Scan(&id)
	if err != nil {
		t.Fatalf("insert teacher: %v", err)
	}
	return id
}

func mustInsertStudent(t *testing.T, ctx context.Context, h *pgtest.Harness, userID, programID, levelID, groupID string) string {
	t.Helper()
	var id string
	err := h.Pool.QueryRow(ctx, `
		INSERT INTO students (user_id, nis, program_id, level_id, group_id)
		VALUES ($1, gen_random_uuid()::text, $2, $3, $4)
		RETURNING id::text
	`, userID, programID, levelID, groupID).Scan(&id)
	if err != nil {
		t.Fatalf("insert student: %v", err)
	}
	return id
}

func mustInsertProgram(t *testing.T, ctx context.Context, h *pgtest.Harness) string {
	t.Helper()
	var id string
	if err := h.Pool.QueryRow(ctx, `INSERT INTO programs (code, name) VALUES ('PRG1', 'Program 1') RETURNING id::text`).Scan(&id); err != nil {
		t.Fatalf("insert program: %v", err)
	}
	return id
}

func mustInsertLevel(t *testing.T, ctx context.Context, h *pgtest.Harness) string {
	t.Helper()
	var id string
	if err := h.Pool.QueryRow(ctx, `INSERT INTO levels (name) VALUES ('Level 1') RETURNING id::text`).Scan(&id); err != nil {
		t.Fatalf("insert level: %v", err)
	}
	return id
}

func mustInsertGroup(t *testing.T, ctx context.Context, h *pgtest.Harness) string {
	t.Helper()
	var id string
	if err := h.Pool.QueryRow(ctx, `INSERT INTO groups (name) VALUES ('Group 1') RETURNING id::text`).Scan(&id); err != nil {
		t.Fatalf("insert group: %v", err)
	}
	return id
}

func mustInsertSubject(t *testing.T, ctx context.Context, h *pgtest.Harness) string {
	t.Helper()
	var id string
	if err := h.Pool.QueryRow(ctx, `INSERT INTO subjects (code, name) VALUES ('MATH', 'Math') RETURNING id::text`).Scan(&id); err != nil {
		t.Fatalf("insert subject: %v", err)
	}
	return id
}

func TestExamRepo_CreateGetListAndTokenFlow(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()

	teacherUserID := mustInsertUser(t, ctx, h, "teacher1", "teacher")
	teacherID := mustInsertTeacher(t, ctx, h, teacherUserID)
	subjectID := mustInsertSubject(t, ctx, h)
	adminUserID := mustInsertUser(t, ctx, h, "admin1", "admin")

	repo := New(h.Pool)
	startsAt := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	endsAt := startsAt.Add(2 * time.Hour)
	duration := 120

	exam, err := repo.Create(ctx, CreateInput{
		SubjectID:        subjectID,
		TeacherID:        teacherID,
		Title:            "Midterm Math",
		StartsAt:         startsAt,
		EndsAt:           endsAt,
		DurationMinutes:  &duration,
		ShuffleQuestions: true,
		ShuffleOptions:   true,
		ScoringMode:      "partial",
		MaxAttempts:      2,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	got, ok, err := repo.Get(ctx, exam.ID)
	if err != nil || !ok {
		t.Fatalf("Get error=%v ok=%v", err, ok)
	}
	if got.Title != "Midterm Math" || got.TeacherID != teacherID || got.SubjectID != subjectID {
		t.Fatalf("unexpected exam: %+v", got)
	}

	items, total, err := repo.List(ctx, "admin", "", ListFilter{Status: "", Q: "Midterm", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one exam, total=%d len=%d", total, len(items))
	}

	token, err := repo.CreateToken(ctx, CreateTokenInput{
		ExamID:          exam.ID,
		CreatedByUserID: adminUserID,
		Length:          3,
	})
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}
	if len(token.Token) != 4 {
		t.Fatalf("expected normalized token length 4, got %q", token.Token)
	}

	tokens, totalTokens, err := repo.ListTokens(ctx, exam.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListTokens error: %v", err)
	}
	if totalTokens != 1 || len(tokens) != 1 {
		t.Fatalf("expected one token, total=%d len=%d", totalTokens, len(tokens))
	}

	updated, ok, err := repo.SetTokenActive(ctx, token.ID, false)
	if err != nil || !ok {
		t.Fatalf("SetTokenActive error=%v ok=%v", err, ok)
	}
	if updated.IsActive {
		t.Fatalf("expected token inactive after update: %+v", updated)
	}

	rotated, err := repo.RotateToken(ctx, RotateTokenInput{
		ExamID:           exam.ID,
		CreatedByUserID:  adminUserID,
		Length:           20,
		DeactivateOthers: true,
	})
	if err != nil {
		t.Fatalf("RotateToken error: %v", err)
	}
	if len(rotated.Token) != 12 {
		t.Fatalf("expected normalized rotated token length 12, got %q", rotated.Token)
	}
}

func TestExamRepo_SessionLookups(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()

	teacherUserID := mustInsertUser(t, ctx, h, "teacher2", "teacher")
	teacherID := mustInsertTeacher(t, ctx, h, teacherUserID)
	subjectID := mustInsertSubject(t, ctx, h)
	programID := mustInsertProgram(t, ctx, h)
	levelID := mustInsertLevel(t, ctx, h)
	groupID := mustInsertGroup(t, ctx, h)
	studentUserID := mustInsertUser(t, ctx, h, "student1", "student")
	studentID := mustInsertStudent(t, ctx, h, studentUserID, programID, levelID, groupID)

	repo := New(h.Pool)
	exam, err := repo.Create(ctx, CreateInput{
		SubjectID:   subjectID,
		TeacherID:   teacherID,
		Title:       "Session Exam",
		StartsAt:    time.Now().UTC(),
		EndsAt:      time.Now().UTC().Add(time.Hour),
		ScoringMode: "partial",
		MaxAttempts: 1,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	var sessionID string
	err = h.Pool.QueryRow(ctx, `
		INSERT INTO exam_sessions (exam_id, student_id, status, attempt_no)
		VALUES ($1, $2, 'in_progress', 1)
		RETURNING id::text
	`, exam.ID, studentID).Scan(&sessionID)
	if err != nil {
		t.Fatalf("insert exam session: %v", err)
	}

	gotExamID, ok, err := repo.SessionExamID(ctx, sessionID)
	if err != nil || !ok {
		t.Fatalf("SessionExamID error=%v ok=%v", err, ok)
	}
	if gotExamID != exam.ID {
		t.Fatalf("expected exam id %q, got %q", exam.ID, gotExamID)
	}

	gotTeacherID, ok, err := repo.TeacherIDByUserID(ctx, teacherUserID)
	if err != nil || !ok {
		t.Fatalf("TeacherIDByUserID error=%v ok=%v", err, ok)
	}
	if gotTeacherID != teacherID {
		t.Fatalf("expected teacher id %q, got %q", teacherID, gotTeacherID)
	}
}
