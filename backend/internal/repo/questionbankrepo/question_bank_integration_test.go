package questionbankrepo

import (
	"context"
	"testing"

	"atigacbt/backend/internal/testutil/pgtest"
)

func qbInsertUser(t *testing.T, ctx context.Context, h *pgtest.Harness, username, role string) string {
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

func qbInsertTeacher(t *testing.T, ctx context.Context, h *pgtest.Harness, userID string) string {
	t.Helper()
	var id string
	err := h.Pool.QueryRow(ctx, `INSERT INTO teachers (user_id, nip) VALUES ($1, gen_random_uuid()::text) RETURNING id::text`, userID).Scan(&id)
	if err != nil {
		t.Fatalf("insert teacher: %v", err)
	}
	return id
}

func qbInsertSubject(t *testing.T, ctx context.Context, h *pgtest.Harness, code, name string) string {
	t.Helper()
	var id string
	err := h.Pool.QueryRow(ctx, `INSERT INTO subjects (code, name) VALUES ($1, $2) RETURNING id::text`, code, name).Scan(&id)
	if err != nil {
		t.Fatalf("insert subject: %v", err)
	}
	return id
}

func TestQuestionBankRepo_SetLifecycleAndTeacherLookup(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()

	teacherUserID := qbInsertUser(t, ctx, h, "teacher-qb", "teacher")
	teacherID := qbInsertTeacher(t, ctx, h, teacherUserID)
	subjectID := qbInsertSubject(t, ctx, h, "BIO", "Biology")
	repo := New(h.Pool)

	gotTeacherID, ok, err := repo.TeacherIDByUserID(ctx, teacherUserID)
	if err != nil || !ok {
		t.Fatalf("TeacherIDByUserID error=%v ok=%v", err, ok)
	}
	if gotTeacherID != teacherID {
		t.Fatalf("expected teacher id %q, got %q", teacherID, gotTeacherID)
	}

	set, err := repo.CreateSet(ctx, subjectID, teacherID, "Set A", "SMA", "")
	if err != nil {
		t.Fatalf("CreateSet error: %v", err)
	}
	if set.Status != "draft" || set.Title != "Set A" {
		t.Fatalf("unexpected created set: %+v", set)
	}

	got, ok, err := repo.GetSet(ctx, set.ID)
	if err != nil || !ok {
		t.Fatalf("GetSet error=%v ok=%v", err, ok)
	}
	if got.SubjectID != subjectID || got.OwnerTeacherID != teacherID {
		t.Fatalf("unexpected fetched set: %+v", got)
	}

	items, total, err := repo.ListSets(ctx, "teacher", teacherID, subjectID, "Set", 10, 0)
	if err != nil {
		t.Fatalf("ListSets error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one set, total=%d len=%d", total, len(items))
	}

	updated, ok, err := repo.UpdateSet(ctx, set.ID, "Set A Updated", "published", "SMA", "")
	if err != nil || !ok {
		t.Fatalf("UpdateSet error=%v ok=%v", err, ok)
	}
	if updated.Title != "Set A Updated" || updated.Status != "published" {
		t.Fatalf("unexpected updated set: %+v", updated)
	}

	deleted, err := repo.DeleteSet(ctx, set.ID)
	if err != nil {
		t.Fatalf("DeleteSet error: %v", err)
	}
	if !deleted {
		t.Fatal("expected set to be deleted")
	}
}

func TestQuestionBankRepo_QuestionPayloadRoundTrip(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()

	teacherUserID := qbInsertUser(t, ctx, h, "teacher-qb2", "teacher")
	teacherID := qbInsertTeacher(t, ctx, h, teacherUserID)
	subjectID := qbInsertSubject(t, ctx, h, "PHY", "Physics")
	repo := New(h.Pool)

	set, err := repo.CreateSet(ctx, subjectID, teacherID, "Set B", "SMA", "")
	if err != nil {
		t.Fatalf("CreateSet error: %v", err)
	}

	mcQuestion, err := repo.CreateQuestion(ctx, set.ID, CreateQuestionInput{
		Type:    "mc_single",
		Stem:    "2 + 2 = ?",
		OrderNo: 1,
		Weight:  2,
		Options: []QuestionOption{
			{Label: "A", Content: "3", IsCorrect: false},
			{Label: "B", Content: "4", IsCorrect: true},
		},
	})
	if err != nil {
		t.Fatalf("CreateQuestion mc error: %v", err)
	}
	if len(mcQuestion.Options) != 2 {
		t.Fatalf("expected 2 options on create, got %+v", mcQuestion)
	}

	essayMax := 50
	essayQuestion, err := repo.CreateQuestion(ctx, set.ID, CreateQuestionInput{
		Type:    "essay",
		Stem:    "Explain gravity",
		OrderNo: 2,
		Weight:  3,
		Essay: &Essay{
			RubricText: "Clarity and accuracy",
			MaxScore:   &essayMax,
		},
	})
	if err != nil {
		t.Fatalf("CreateQuestion essay error: %v", err)
	}
	if essayQuestion.Essay == nil || essayQuestion.Essay.MaxScore == nil || *essayQuestion.Essay.MaxScore != 50 {
		t.Fatalf("unexpected essay payload on create: %+v", essayQuestion)
	}

	tfQuestion, err := repo.CreateQuestion(ctx, set.ID, CreateQuestionInput{
		Type:    "true_false",
		Stem:    "Mark statements",
		OrderNo: 3,
		Weight:  1,
		TrueFalse: &TrueFalse{
			Correct: true,
		},
		Statements: []TFStatement{
			{Content: "Earth is round", Correct: true, OrderNo: 1},
			{Content: "Sun rises from west", Correct: false, OrderNo: 2},
		},
	})
	if err != nil {
		t.Fatalf("CreateQuestion true_false error: %v", err)
	}
	if len(tfQuestion.Statements) != 2 {
		t.Fatalf("expected 2 statements on create, got %+v", tfQuestion)
	}

	gotMC, ok, err := repo.GetQuestion(ctx, mcQuestion.ID)
	if err != nil || !ok {
		t.Fatalf("GetQuestion mc error=%v ok=%v", err, ok)
	}
	if gotMC.Weight != 2 || len(gotMC.Options) != 2 || gotMC.Options[1].Label != "B" || !gotMC.Options[1].IsCorrect {
		t.Fatalf("unexpected fetched mc question: %+v", gotMC)
	}

	gotEssay, ok, err := repo.GetQuestion(ctx, essayQuestion.ID)
	if err != nil || !ok {
		t.Fatalf("GetQuestion essay error=%v ok=%v", err, ok)
	}
	if gotEssay.Essay == nil || gotEssay.Essay.RubricText != "Clarity and accuracy" {
		t.Fatalf("unexpected fetched essay question: %+v", gotEssay)
	}

	gotTF, ok, err := repo.GetQuestion(ctx, tfQuestion.ID)
	if err != nil || !ok {
		t.Fatalf("GetQuestion tf error=%v ok=%v", err, ok)
	}
	if gotTF.TrueFalse == nil || !gotTF.TrueFalse.Correct || len(gotTF.Statements) != 2 {
		t.Fatalf("unexpected fetched true_false question: %+v", gotTF)
	}

	updatedEssayMax := 75
	updatedQuestion, ok, err := repo.UpdateQuestion(ctx, essayQuestion.ID, UpdateQuestionInput{
		Type:    "essay",
		Stem:    "Explain gravity in detail",
		OrderNo: 5,
		Weight:  4,
		Essay: &Essay{
			RubricText: "More detail",
			MaxScore:   &updatedEssayMax,
		},
	})
	if err != nil || !ok {
		t.Fatalf("UpdateQuestion error=%v ok=%v", err, ok)
	}
	if updatedQuestion.OrderNo != 5 || updatedQuestion.Weight != 4 || updatedQuestion.Essay == nil || *updatedQuestion.Essay.MaxScore != 75 {
		t.Fatalf("unexpected updated question: %+v", updatedQuestion)
	}

	listed, err := repo.ListQuestions(ctx, set.ID)
	if err != nil {
		t.Fatalf("ListQuestions error: %v", err)
	}
	if len(listed) != 3 {
		t.Fatalf("expected 3 questions, got %d", len(listed))
	}

	deleted, err := repo.DeleteQuestion(ctx, mcQuestion.ID)
	if err != nil {
		t.Fatalf("DeleteQuestion error: %v", err)
	}
	if !deleted {
		t.Fatal("expected question to be deleted")
	}
}
