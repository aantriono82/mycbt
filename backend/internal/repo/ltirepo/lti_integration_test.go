package ltirepo

import (
	"context"
	"testing"
	"time"

	"atigacbt/backend/internal/testutil/pgtest"
)

func ltiInsertUser(t *testing.T, ctx context.Context, h *pgtest.Harness, username, role string) string {
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

func ltiInsertTeacher(t *testing.T, ctx context.Context, h *pgtest.Harness, userID string) string {
	t.Helper()
	var id string
	err := h.Pool.QueryRow(ctx, `INSERT INTO teachers (user_id, nip) VALUES ($1, gen_random_uuid()::text) RETURNING id::text`, userID).Scan(&id)
	if err != nil {
		t.Fatalf("insert teacher: %v", err)
	}
	return id
}

func ltiInsertSubject(t *testing.T, ctx context.Context, h *pgtest.Harness) string {
	t.Helper()
	var id string
	if err := h.Pool.QueryRow(ctx, `INSERT INTO subjects (code, name) VALUES ('LTI-SUB', 'LTI Subject') RETURNING id::text`).Scan(&id); err != nil {
		t.Fatalf("insert subject: %v", err)
	}
	return id
}

func ltiInsertProgramLevelGroupStudent(t *testing.T, ctx context.Context, h *pgtest.Harness, userID string) string {
	t.Helper()
	var programID, levelID, groupID, studentID string
	if err := h.Pool.QueryRow(ctx, `INSERT INTO programs (code, name) VALUES ('LTI-PRG', 'LTI Program') RETURNING id::text`).Scan(&programID); err != nil {
		t.Fatalf("insert program: %v", err)
	}
	if err := h.Pool.QueryRow(ctx, `INSERT INTO levels (name) VALUES ('LTI Level') RETURNING id::text`).Scan(&levelID); err != nil {
		t.Fatalf("insert level: %v", err)
	}
	if err := h.Pool.QueryRow(ctx, `INSERT INTO groups (name) VALUES ('LTI Group') RETURNING id::text`).Scan(&groupID); err != nil {
		t.Fatalf("insert group: %v", err)
	}
	if err := h.Pool.QueryRow(ctx, `
		INSERT INTO students (user_id, nis, program_id, level_id, group_id)
		VALUES ($1, gen_random_uuid()::text, $2, $3, $4)
		RETURNING id::text
	`, userID, programID, levelID, groupID).Scan(&studentID); err != nil {
		t.Fatalf("insert student: %v", err)
	}
	return studentID
}

func TestLTIRepo_PlatformUserNonceAndSessionFlow(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()
	repo := New(h.Pool)

	platform, err := repo.CreatePlatform(ctx, Platform{
		Name:           "Canvas",
		Issuer:         "https://canvas.example.com",
		ClientID:       "client-1",
		DeploymentID:   "dep-1",
		OIDCAuthURL:    "https://canvas.example.com/auth",
		OIDCTokenURL:   "https://canvas.example.com/token",
		JWKSURL:        "https://canvas.example.com/jwks",
		ToolPrivateKey: "priv",
		ToolPublicKey:  "pub",
	})
	if err != nil {
		t.Fatalf("CreatePlatform error: %v", err)
	}

	gotByIssuer, ok, err := repo.GetPlatformByIssuer(ctx, "https://canvas.example.com")
	if err != nil || !ok {
		t.Fatalf("GetPlatformByIssuer error=%v ok=%v", err, ok)
	}
	if gotByIssuer.ID != platform.ID {
		t.Fatalf("unexpected platform by issuer: %+v", gotByIssuer)
	}

	gotByID, ok, err := repo.GetPlatformByID(ctx, platform.ID)
	if err != nil || !ok {
		t.Fatalf("GetPlatformByID error=%v ok=%v", err, ok)
	}
	if gotByID.ClientID != "client-1" {
		t.Fatalf("unexpected platform by id: %+v", gotByID)
	}

	items, err := repo.ListPlatforms(ctx)
	if err != nil {
		t.Fatalf("ListPlatforms error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(items))
	}

	localUserID := ltiInsertUser(t, ctx, h, "lti-user", "student")
	if err := repo.LinkUser(ctx, platform.ID, "sub-1", localUserID); err != nil {
		t.Fatalf("LinkUser error: %v", err)
	}
	gotUserID, ok, err := repo.FindUserByLTI(ctx, platform.ID, "sub-1")
	if err != nil || !ok {
		t.Fatalf("FindUserByLTI error=%v ok=%v", err, ok)
	}
	if gotUserID != localUserID {
		t.Fatalf("expected linked user %q, got %q", localUserID, gotUserID)
	}

	if err := repo.StoreNonce(ctx, "nonce-1", time.Hour); err != nil {
		t.Fatalf("StoreNonce error: %v", err)
	}
	used, err := repo.UseNonce(ctx, "nonce-1")
	if err != nil {
		t.Fatalf("UseNonce error: %v", err)
	}
	if !used {
		t.Fatal("expected nonce to be usable once")
	}
	usedAgain, err := repo.UseNonce(ctx, "nonce-1")
	if err != nil {
		t.Fatalf("UseNonce second error: %v", err)
	}
	if usedAgain {
		t.Fatal("expected nonce to be consumed")
	}

	sessionID, err := repo.CreateSession(ctx, LTISession{
		PlatformID:   platform.ID,
		LocalUserID:  localUserID,
		MessageType:  "LtiResourceLinkRequest",
		ReturnURL:    "https://canvas.example.com/return",
		Data:         "opaque",
		DeploymentID: "dep-1",
		ExpiresAt:    time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}
	session, ok, err := repo.GetSession(ctx, sessionID)
	if err != nil || !ok {
		t.Fatalf("GetSession error=%v ok=%v", err, ok)
	}
	if session.LocalUserID != localUserID || session.ReturnURL != "https://canvas.example.com/return" {
		t.Fatalf("unexpected lti session: %+v", session)
	}
}

func TestLTIRepo_AGSLaunchAndScoreTargets(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()
	repo := New(h.Pool)

	platform, err := repo.CreatePlatform(ctx, Platform{
		Name:           "Moodle",
		Issuer:         "https://moodle.example.com",
		ClientID:       "client-2",
		DeploymentID:   "dep-2",
		OIDCAuthURL:    "https://moodle.example.com/auth",
		OIDCTokenURL:   "https://moodle.example.com/token",
		JWKSURL:        "https://moodle.example.com/jwks",
		ToolPrivateKey: "priv2",
		ToolPublicKey:  "pub2",
	})
	if err != nil {
		t.Fatalf("CreatePlatform error: %v", err)
	}

	teacherUserID := ltiInsertUser(t, ctx, h, "teacher-lti", "teacher")
	teacherID := ltiInsertTeacher(t, ctx, h, teacherUserID)
	subjectID := ltiInsertSubject(t, ctx, h)
	studentUserID := ltiInsertUser(t, ctx, h, "student-lti", "student")
	studentID := ltiInsertProgramLevelGroupStudent(t, ctx, h, studentUserID)

	var examID string
	err = h.Pool.QueryRow(ctx, `
		INSERT INTO exams (subject_id, teacher_id, title, starts_at, ends_at, shuffle_questions, shuffle_options, scoring_mode, max_attempts, status)
		VALUES ($1, $2, 'LTI Exam', now(), now() + interval '1 hour', false, false, 'partial', 1, 'published')
		RETURNING id::text
	`, subjectID, teacherID).Scan(&examID)
	if err != nil {
		t.Fatalf("insert exam: %v", err)
	}

	var sessionID string
	err = h.Pool.QueryRow(ctx, `
		INSERT INTO exam_sessions (exam_id, student_id, status, finished_at, attempt_no)
		VALUES ($1, $2, 'submitted', now(), 1)
		RETURNING id::text
	`, examID, studentID).Scan(&sessionID)
	if err != nil {
		t.Fatalf("insert exam session: %v", err)
	}

	if err := repo.UpsertAGSLaunchContext(ctx, AGSLaunchContext{
		PlatformID:     platform.ID,
		DeploymentID:   "dep-2",
		ResourceLinkID: "resource-1",
		ExamID:         examID,
		LocalUserID:    studentUserID,
		LTISub:         "student-sub-1",
		LineItemURL:    "https://moodle.example.com/lineitems/1",
		LineItemsURL:   "https://moodle.example.com/lineitems",
		ScopeText:      "foo https://purl.imsglobal.org/spec/lti-ags/scope/score",
	}); err != nil {
		t.Fatalf("UpsertAGSLaunchContext error: %v", err)
	}

	targets, err := repo.ListAGSScoreTargets(ctx, examID)
	if err != nil {
		t.Fatalf("ListAGSScoreTargets error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 ags target, got %d", len(targets))
	}
	if targets[0].SessionID != sessionID || targets[0].LTISub != "student-sub-1" || targets[0].Platform.ID != platform.ID {
		t.Fatalf("unexpected ags target: %+v", targets[0])
	}
}
