package masterrepo

import (
	"context"
	"testing"

	"atigacbt/backend/internal/testutil/pgtest"
)

func mrInsertTeacherUser(t *testing.T, ctx context.Context, h *pgtest.Harness, username string) (string, string) {
	t.Helper()
	var userID string
	err := h.Pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, role, name, email, is_active)
		VALUES ($1, 'hash', 'teacher', $2, $3, true)
		RETURNING id::text
	`, username, username, username+"@example.com").Scan(&userID)
	if err != nil {
		t.Fatalf("insert teacher user: %v", err)
	}
	var teacherID string
	err = h.Pool.QueryRow(ctx, `INSERT INTO teachers (user_id, nip) VALUES ($1, gen_random_uuid()::text) RETURNING id::text`, userID).Scan(&teacherID)
	if err != nil {
		t.Fatalf("insert teacher: %v", err)
	}
	return userID, teacherID
}

func TestMasterRepo_CRUDAndLookups(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()

	programs := NewPrograms(h.Pool)
	subjects := NewSubjects(h.Pool)
	groups := NewGroups(h.Pool)
	levels := NewLevels(h.Pool)
	lookups := NewLookups(h.Pool)

	program, err := programs.Create(ctx, "SCI", "Science")
	if err != nil {
		t.Fatalf("program Create error: %v", err)
	}
	gotProgram, ok, err := programs.Get(ctx, program.ID)
	if err != nil || !ok {
		t.Fatalf("program Get error=%v ok=%v", err, ok)
	}
	if gotProgram.Code != "SCI" {
		t.Fatalf("unexpected program: %+v", gotProgram)
	}
	program, ok, err = programs.Update(ctx, program.ID, "", "Science Updated")
	if err != nil || !ok {
		t.Fatalf("program Update error=%v ok=%v", err, ok)
	}
	if program.Code != "" || program.Name != "Science Updated" {
		t.Fatalf("unexpected updated program: %+v", program)
	}

	subject, err := subjects.Create(ctx, "BIO", "Biology")
	if err != nil {
		t.Fatalf("subject Create error: %v", err)
	}
	gotSubject, ok, err := subjects.Get(ctx, subject.ID)
	if err != nil || !ok {
		t.Fatalf("subject Get error=%v ok=%v", err, ok)
	}
	if gotSubject.Code != "BIO" {
		t.Fatalf("unexpected subject: %+v", gotSubject)
	}
	subject, ok, err = subjects.Update(ctx, subject.ID, "BIO2", "Biology 2")
	if err != nil || !ok {
		t.Fatalf("subject Update error=%v ok=%v", err, ok)
	}
	if subject.Code != "BIO2" {
		t.Fatalf("unexpected updated subject: %+v", subject)
	}

	group, err := groups.Create(ctx, "Group A")
	if err != nil {
		t.Fatalf("group Create error: %v", err)
	}
	gotGroup, ok, err := groups.Get(ctx, group.ID)
	if err != nil || !ok {
		t.Fatalf("group Get error=%v ok=%v", err, ok)
	}
	if gotGroup.Name != "Group A" {
		t.Fatalf("unexpected group: %+v", gotGroup)
	}
	group, ok, err = groups.Update(ctx, group.ID, "Group A Updated")
	if err != nil || !ok {
		t.Fatalf("group Update error=%v ok=%v", err, ok)
	}
	if group.Name != "Group A Updated" {
		t.Fatalf("unexpected updated group: %+v", group)
	}

	kelas := 11
	level, err := levels.Create(ctx, "Level A", &kelas)
	if err != nil {
		t.Fatalf("level Create error: %v", err)
	}
	gotLevel, ok, err := levels.Get(ctx, level.ID)
	if err != nil || !ok {
		t.Fatalf("level Get error=%v ok=%v", err, ok)
	}
	if gotLevel.Kelas == nil || *gotLevel.Kelas != 11 {
		t.Fatalf("unexpected level: %+v", gotLevel)
	}
	newKelas := 12
	level, ok, err = levels.Update(ctx, level.ID, "Level A Updated", &newKelas)
	if err != nil || !ok {
		t.Fatalf("level Update error=%v ok=%v", err, ok)
	}
	if level.Kelas == nil || *level.Kelas != 12 {
		t.Fatalf("unexpected updated level: %+v", level)
	}

	programID, ok, err := lookups.ProgramIDByCode(ctx, "")
	if err != nil || ok || programID != "" {
		t.Fatalf("expected empty code lookup miss, id=%q ok=%v err=%v", programID, ok, err)
	}
	programID, ok, err = lookups.ProgramIDByCode(ctx, "SCI")
	if err != nil || !ok || programID != program.ID {
		t.Fatalf("unexpected program lookup id=%q ok=%v err=%v", programID, ok, err)
	}
	levelID, ok, err := lookups.LevelIDByName(ctx, "Level A Updated")
	if err != nil || !ok || levelID != level.ID {
		t.Fatalf("unexpected level lookup id=%q ok=%v err=%v", levelID, ok, err)
	}
	groupID, ok, err := lookups.GroupIDByName(ctx, "Group A Updated")
	if err != nil || !ok || groupID != group.ID {
		t.Fatalf("unexpected group lookup id=%q ok=%v err=%v", groupID, ok, err)
	}

	subjectIDs, err := lookups.SubjectIDsByCodes(ctx, []string{"BIO2", "MISSING"})
	if err != nil {
		t.Fatalf("SubjectIDsByCodes error: %v", err)
	}
	if len(subjectIDs) != 1 || subjectIDs[0] != subject.ID {
		t.Fatalf("unexpected subject ids: %+v", subjectIDs)
	}

	ids, missing, err := lookups.SubjectIDsByCodesStrict(ctx, []string{"BIO2", "MISS1"})
	if err != nil {
		t.Fatalf("SubjectIDsByCodesStrict error: %v", err)
	}
	if len(ids) != 1 || ids[0] != subject.ID || len(missing) != 1 || missing[0] != "MISS1" {
		t.Fatalf("unexpected strict lookup ids=%v missing=%v", ids, missing)
	}

	groupIDs, err := lookups.GroupIDsByNames(ctx, []string{"Group A Updated", "Unknown"})
	if err != nil {
		t.Fatalf("GroupIDsByNames error: %v", err)
	}
	if len(groupIDs) != 1 || groupIDs[0] != group.ID {
		t.Fatalf("unexpected group ids: %+v", groupIDs)
	}

	levelIDs, err := lookups.LevelIDsByNames(ctx, []string{"Level A Updated", "Unknown"})
	if err != nil {
		t.Fatalf("LevelIDsByNames error: %v", err)
	}
	if len(levelIDs) != 1 || levelIDs[0] != level.ID {
		t.Fatalf("unexpected level ids: %+v", levelIDs)
	}

	subjectsMissing, err := subjects.MissingIDs(ctx, []string{subject.ID, "00000000-0000-0000-0000-000000000000"})
	if err != nil {
		t.Fatalf("MissingIDs error: %v", err)
	}
	if len(subjectsMissing) != 1 {
		t.Fatalf("expected 1 missing subject id, got %+v", subjectsMissing)
	}

	if items, err := programs.List(ctx); err != nil || len(items) != 1 {
		t.Fatalf("program List err=%v items=%d", err, len(items))
	}
	if items, err := subjects.List(ctx); err != nil || len(items) != 1 {
		t.Fatalf("subject List err=%v items=%d", err, len(items))
	}
	if items, err := groups.List(ctx); err != nil || len(items) != 1 {
		t.Fatalf("group List err=%v items=%d", err, len(items))
	}
	if items, err := levels.List(ctx); err != nil || len(items) != 1 {
		t.Fatalf("level List err=%v items=%d", err, len(items))
	}
}

func TestMasterRepo_TeacherScopedLists(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()

	subjects := NewSubjects(h.Pool)
	groups := NewGroups(h.Pool)
	levels := NewLevels(h.Pool)

	userID, teacherID := mrInsertTeacherUser(t, ctx, h, "teacher-scope")

	subject, err := subjects.Create(ctx, "CHEM", "Chemistry")
	if err != nil {
		t.Fatalf("subject Create error: %v", err)
	}
	group, err := groups.Create(ctx, "Teacher Group")
	if err != nil {
		t.Fatalf("group Create error: %v", err)
	}
	kelas := 10
	level, err := levels.Create(ctx, "Teacher Level", &kelas)
	if err != nil {
		t.Fatalf("level Create error: %v", err)
	}

	if _, err := h.Pool.Exec(ctx, `INSERT INTO teacher_subjects (teacher_id, subject_id) VALUES ($1, $2)`, teacherID, subject.ID); err != nil {
		t.Fatalf("insert teacher_subjects: %v", err)
	}
	if _, err := h.Pool.Exec(ctx, `INSERT INTO teacher_groups (teacher_id, group_id) VALUES ($1, $2)`, teacherID, group.ID); err != nil {
		t.Fatalf("insert teacher_groups: %v", err)
	}
	if _, err := h.Pool.Exec(ctx, `INSERT INTO teacher_levels (teacher_id, level_id) VALUES ($1, $2)`, teacherID, level.ID); err != nil {
		t.Fatalf("insert teacher_levels: %v", err)
	}

	subjItems, ok, err := subjects.ListForTeacherUserID(ctx, userID)
	if err != nil || !ok || len(subjItems) != 1 || subjItems[0].ID != subject.ID {
		t.Fatalf("unexpected teacher subject list ok=%v err=%v items=%+v", ok, err, subjItems)
	}
	groupItems, ok, err := groups.ListForTeacherUserID(ctx, userID)
	if err != nil || !ok || len(groupItems) != 1 || groupItems[0].ID != group.ID {
		t.Fatalf("unexpected teacher group list ok=%v err=%v items=%+v", ok, err, groupItems)
	}
	levelItems, ok, err := levels.ListForTeacherUserID(ctx, userID)
	if err != nil || !ok || len(levelItems) != 1 || levelItems[0].ID != level.ID {
		t.Fatalf("unexpected teacher level list ok=%v err=%v items=%+v", ok, err, levelItems)
	}
}
