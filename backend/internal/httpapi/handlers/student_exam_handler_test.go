package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/repo/masterrepo"
	"atigacbt/backend/internal/repo/studentexamrepo"
)

type mockStudentExamRepo struct {
	mu sync.Mutex

	studentInfo studentexamrepo.StudentInfo
	studentOK   bool
	studentErr  error

	examForJoin    studentexamrepo.ExamForJoin
	examForJoinErr error

	sessionByExam    map[string]studentexamrepo.Session
	sessionByExamOK  map[string]bool
	sessionByExamErr error

	createdSession studentexamrepo.Session
	createErr      error

	ensureSessionQuestionsCount int
	ensureSessionQuestionsErr   error

	sessionStateByID map[string]studentexamrepo.SessionState
	sessionStateOK   map[string]bool

	questionsBySession map[string][]studentexamrepo.StudentQuestion
	questionsErr       error

	upsertAnswerErr error
	upsertCalls     map[string]int

	submitCalls     map[string]int
	submitErrFirst  error
	submitErrSecond error

	dismissCalls map[string]int
	dismissErr   error

	verifyTokenErr  error
	activeCount     int
	activeCountErr  error
	attemptsCount   int
	attemptsErr     error
	sessionExists   bool
	sessionExistErr error
}

func (m *mockStudentExamRepo) StudentByUserID(ctx context.Context, userID string) (studentexamrepo.StudentInfo, bool, error) {
	return m.studentInfo, m.studentOK, m.studentErr
}

func (m *mockStudentExamRepo) ListAvailableForStudent(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentExamsFilter) ([]studentexamrepo.StudentExam, int, error) {
	return nil, 0, nil
}

func (m *mockStudentExamRepo) VerifyExamToken(ctx context.Context, examID, token string, nowUTC time.Time) error {
	return m.verifyTokenErr
}

func (m *mockStudentExamRepo) DismissExamCard(ctx context.Context, examID, studentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dismissCalls == nil {
		m.dismissCalls = map[string]int{}
	}
	m.dismissCalls[examID]++
	return m.dismissErr
}

func (m *mockStudentExamRepo) GetExamForStudentJoin(ctx context.Context, examID, studentID, levelID, groupID string, nowUTC time.Time, loc *time.Location) (studentexamrepo.ExamForJoin, error) {
	return m.examForJoin, m.examForJoinErr
}

func (m *mockStudentExamRepo) GetSessionByExamStudent(ctx context.Context, examID, studentID string) (studentexamrepo.Session, bool, error) {
	if m.sessionByExamErr != nil {
		return studentexamrepo.Session{}, false, m.sessionByExamErr
	}
	if m.sessionByExamOK[examID] {
		return m.sessionByExam[examID], true, nil
	}
	return studentexamrepo.Session{}, false, nil
}

func (m *mockStudentExamRepo) CountInProgressSessionsByStudent(ctx context.Context, studentID string) (int, error) {
	return m.activeCount, m.activeCountErr
}

func (m *mockStudentExamRepo) CountAttemptsByExamStudent(ctx context.Context, examID, studentID string) (int, error) {
	if m.attemptsErr != nil {
		return 0, m.attemptsErr
	}
	if m.attemptsCount != 0 {
		return m.attemptsCount, nil
	}
	if m.sessionByExamOK[examID] {
		return 1, nil
	}
	return 0, nil
}

func (m *mockStudentExamRepo) CreateSessionAttempt(ctx context.Context, examID, studentID string, clientIP net.IP, userAgent string) (studentexamrepo.Session, error) {
	return m.createdSession, m.createErr
}

func (m *mockStudentExamRepo) GetOrCreateSession(ctx context.Context, examID, studentID string, clientIP net.IP, userAgent string) (studentexamrepo.Session, error) {
	return m.createdSession, m.createErr
}

func (m *mockStudentExamRepo) EnsureSessionQuestions(ctx context.Context, sessionID, examID string, shuffleQuestions bool) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureSessionQuestionsCount++
	return 10, m.ensureSessionQuestionsErr
}

func (m *mockStudentExamRepo) GetSessionState(ctx context.Context, sessionID, studentID string, nowUTC time.Time) (studentexamrepo.SessionState, bool, error) {
	st, ok := m.sessionStateByID[sessionID]
	return st, m.sessionStateOK[sessionID] && ok, nil
}

func (m *mockStudentExamRepo) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	return m.sessionExists, m.sessionExistErr
}

func (m *mockStudentExamRepo) ListSessionQuestions(ctx context.Context, sessionID, studentID string, shuffleOptions bool) ([]studentexamrepo.StudentQuestion, error) {
	return m.questionsBySession[sessionID], m.questionsErr
}

func (m *mockStudentExamRepo) ListSessionAnswers(ctx context.Context, sessionID, studentID string) ([]studentexamrepo.SessionAnswer, error) {
	return nil, nil
}

func (m *mockStudentExamRepo) UpsertAnswer(ctx context.Context, sessionID, studentID, questionID string, answerJSON json.RawMessage, nowUTC time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.upsertCalls == nil {
		m.upsertCalls = map[string]int{}
	}
	m.upsertCalls[sessionID+":"+questionID]++
	return m.upsertAnswerErr
}

func (m *mockStudentExamRepo) SubmitSession(ctx context.Context, sessionID, studentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.submitCalls == nil {
		m.submitCalls = map[string]int{}
	}
	m.submitCalls[sessionID]++
	if m.submitCalls[sessionID] == 1 {
		return m.submitErrFirst
	}
	return m.submitErrSecond
}

func (m *mockStudentExamRepo) ComputeAutoScore(ctx context.Context, sessionID, studentID string, nowUTC time.Time) (studentexamrepo.AutoScoreSummary, error) {
	return studentexamrepo.AutoScoreSummary{TotalQuestions: 10, CorrectCount: 8, Score: 80}, nil
}

func (m *mockStudentExamRepo) Heartbeat(ctx context.Context, sessionID, studentID string, payload json.RawMessage) error {
	return nil
}

func (m *mockStudentExamRepo) ListStudentResults(ctx context.Context, studentID string, f studentexamrepo.ListStudentResultsFilter) ([]studentexamrepo.StudentResultSummary, int, error) {
	return nil, 0, nil
}

func (m *mockStudentExamRepo) ListStudentAnnouncements(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentAnnouncementsFilter) ([]studentexamrepo.StudentAnnouncement, int, error) {
	return nil, 0, nil
}

func (m *mockStudentExamRepo) EnsureStudentCanAttendExam(ctx context.Context, examID, studentID, levelID, groupID string) (bool, error) {
	return true, nil
}

func (m *mockStudentExamRepo) UpsertAttendance(ctx context.Context, examID, studentID, note string, clientIP net.IP, nowUTC time.Time, opts ...studentexamrepo.AttendanceOption) (studentexamrepo.AttendanceSubmission, error) {
	return studentexamrepo.AttendanceSubmission{}, nil
}

func (m *mockStudentExamRepo) ListAttendanceHistory(ctx context.Context, studentID string, f studentexamrepo.ListAttendanceHistoryFilter) ([]studentexamrepo.AttendanceHistoryItem, int, error) {
	return nil, 0, nil
}

type mockSettingsRepo struct {
	sys masterrepo.SystemSettings
}

func (m *mockSettingsRepo) GetSystem(ctx context.Context) (masterrepo.SystemSettings, error) {
	return m.sys, nil
}

func withAuthUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("user_role", "student")
		c.Next()
	}
}

func TestStudentExamHandler_Join_CreateSessionAndConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockStudentExamRepo{
		studentInfo:     studentexamrepo.StudentInfo{StudentID: "stu-1", LevelID: "lvl-1", GroupID: "grp-1", IsActive: true},
		studentOK:       true,
		examForJoin:     studentexamrepo.ExamForJoin{ID: "exam-1", Title: "Math", ShuffleQuestions: true},
		sessionByExam:   map[string]studentexamrepo.Session{},
		sessionByExamOK: map[string]bool{},
		createdSession: studentexamrepo.Session{
			ID:        "sess-1",
			ExamID:    "exam-1",
			StudentID: "stu-1",
			Status:    "in_progress",
			StartedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta", TokenRequired: false, MaxActiveSessions: 1}})

	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/student/exams/:id/join", h.Join)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/student/exams/exam-1/join", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"session_id":"sess-1"`) {
		t.Fatalf("expected session_id in response, body=%s", rec.Body.String())
	}

	// Existing finished session should return 409.
	mockRepo.sessionByExam["exam-1"] = studentexamrepo.Session{ID: "sess-old", ExamID: "exam-1", StudentID: "stu-1", Status: "submitted"}
	mockRepo.sessionByExamOK["exam-1"] = true
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/student/exams/exam-1/join", strings.NewReader(`{}`))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for finished existing session, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestStudentExamHandler_GetQuestions_NoAnswerKeyAndSessionVariation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	state1 := studentexamrepo.SessionState{}
	state1.Session = studentexamrepo.Session{ID: "sess-a", ExamID: "exam-1", StudentID: "stu-1", Status: "in_progress"}
	state1.Exam.ShuffleOptions = true
	state2 := studentexamrepo.SessionState{}
	state2.Session = studentexamrepo.Session{ID: "sess-b", ExamID: "exam-1", StudentID: "stu-1", Status: "in_progress"}
	state2.Exam.ShuffleOptions = true

	mockRepo := &mockStudentExamRepo{
		studentInfo: studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: true},
		studentOK:   true,
		sessionStateByID: map[string]studentexamrepo.SessionState{
			"sess-a": state1,
			"sess-b": state2,
		},
		sessionStateOK: map[string]bool{"sess-a": true, "sess-b": true},
		questionsBySession: map[string][]studentexamrepo.StudentQuestion{
			"sess-a": {
				{OrderNo: 1, ID: "q1", Type: "mc_single", Stem: "A"},
				{OrderNo: 2, ID: "q2", Type: "mc_single", Stem: "B"},
			},
			"sess-b": {
				{OrderNo: 1, ID: "q2", Type: "mc_single", Stem: "B"},
				{OrderNo: 2, ID: "q1", Type: "mc_single", Stem: "A"},
			},
		},
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})

	r := gin.New()
	r.Use(withAuthUser())
	r.GET("/api/v1/student/sessions/:id/questions", h.GetQuestions)

	recA := httptest.NewRecorder()
	r.ServeHTTP(recA, httptest.NewRequest(http.MethodGet, "/api/v1/student/sessions/sess-a/questions", nil))
	if recA.Code != http.StatusOK {
		t.Fatalf("sess-a expected 200 got %d body=%s", recA.Code, recA.Body.String())
	}
	recB := httptest.NewRecorder()
	r.ServeHTTP(recB, httptest.NewRequest(http.MethodGet, "/api/v1/student/sessions/sess-b/questions", nil))
	if recB.Code != http.StatusOK {
		t.Fatalf("sess-b expected 200 got %d body=%s", recB.Code, recB.Body.String())
	}

	bodyA := recA.Body.String()
	bodyB := recB.Body.String()
	if bodyA == bodyB {
		t.Fatalf("expected different order payload between sessions")
	}
	if strings.Contains(bodyA, "correct_answer") || strings.Contains(bodyA, "answer_key") {
		t.Fatalf("response should not contain answer key fields: %s", bodyA)
	}
	if strings.Contains(bodyB, "correct_answer") || strings.Contains(bodyB, "answer_key") {
		t.Fatalf("response should not contain answer key fields: %s", bodyB)
	}
}

func TestStudentExamHandler_DismissExamCard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockStudentExamRepo{
		studentInfo: studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: true},
		studentOK:   true,
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})

	r := gin.New()
	r.Use(withAuthUser())
	r.DELETE("/api/v1/student/exams/:id/dismiss", h.DismissExamCard)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/student/exams/exam-1/dismiss", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rec.Code, rec.Body.String())
	}

	mockRepo.dismissErr = studentexamrepo.ErrExamNotDismissible
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodDelete, "/api/v1/student/exams/exam-2/dismiss", nil))
	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected 409 got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestStudentExamHandler_SaveAnswer_ValidAndUpsertAndForbiddenishCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockStudentExamRepo{
		studentInfo: studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: true},
		studentOK:   true,
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})
	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/student/sessions/:id/answers", h.SaveAnswer)

	body := `{"question_id":"q-1","answer_json":{"selected":"A"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-1/answers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rec.Code, rec.Body.String())
	}

	// Submit same question again: should still be OK (upsert behavior at repo layer).
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-1/answers", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected second upsert 200 got %d body=%s", rec2.Code, rec2.Body.String())
	}
	if got := mockRepo.upsertCalls["sess-1:q-1"]; got != 2 {
		t.Fatalf("expected upsert called twice, got %d", got)
	}

	// Session not active (includes expired/finalized) maps to 409 in current handler.
	mockRepo.upsertAnswerErr = studentexamrepo.ErrSessionNotActive
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-1/answers", bytes.NewBufferString(body))
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusConflict {
		t.Fatalf("expected 409 for inactive session got %d body=%s", rec3.Code, rec3.Body.String())
	}
}

func TestStudentExamHandler_Submit_IdempotentSecondCallConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockStudentExamRepo{
		studentInfo:     studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: true},
		studentOK:       true,
		submitErrFirst:  nil,
		submitErrSecond: studentexamrepo.ErrSessionNotActive,
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})
	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/student/sessions/:id/submit", h.Submit)

	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-9/submit", nil))
	if rec1.Code != http.StatusOK {
		t.Fatalf("first submit expected 200 got %d body=%s", rec1.Code, rec1.Body.String())
	}

	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-9/submit", nil))
	if rec2.Code != http.StatusConflict {
		t.Fatalf("second submit expected 409 got %d body=%s", rec2.Code, rec2.Body.String())
	}
	if got := mockRepo.submitCalls["sess-9"]; got != 2 {
		t.Fatalf("expected submit called twice, got %d", got)
	}
}

func TestStudentExamHandler_CompatStartQuestionsAnswersFinish(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	state := studentexamrepo.SessionState{}
	state.Session = studentexamrepo.Session{
		ID:         "sess-compat",
		ExamID:     "exam-compat",
		StudentID:  "stu-1",
		Status:     "in_progress",
		StartedAt:  now.Add(-5 * time.Minute).Format(time.RFC3339),
		FinishedAt: now.Format(time.RFC3339),
	}
	state.DeadlineAt = now.Add(55 * time.Minute).Format(time.RFC3339)
	state.Exam.ShuffleOptions = true

	mockRepo := &mockStudentExamRepo{
		studentInfo:     studentexamrepo.StudentInfo{StudentID: "stu-1", LevelID: "lvl-1", GroupID: "grp-1", IsActive: true},
		studentOK:       true,
		examForJoin:     studentexamrepo.ExamForJoin{ID: "exam-compat", Title: "Math", ShuffleQuestions: true},
		sessionByExam:   map[string]studentexamrepo.Session{},
		sessionByExamOK: map[string]bool{},
		createdSession: studentexamrepo.Session{
			ID:        "sess-compat",
			ExamID:    "exam-compat",
			StudentID: "stu-1",
			Status:    "in_progress",
			StartedAt: now.Format(time.RFC3339),
		},
		sessionStateByID: map[string]studentexamrepo.SessionState{"sess-compat": state},
		sessionStateOK:   map[string]bool{"sess-compat": true},
		questionsBySession: map[string][]studentexamrepo.StudentQuestion{
			"sess-compat": {
				{OrderNo: 1, ID: "q1", Type: "mc_single", Stem: "A"},
				{OrderNo: 2, ID: "q2", Type: "mc_single", Stem: "B"},
				{OrderNo: 3, ID: "q3", Type: "mc_single", Stem: "C"},
			},
		},
		submitErrFirst:  nil,
		submitErrSecond: studentexamrepo.ErrSessionNotActive,
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})

	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/exams/:id/start", h.StartCompat)
	r.GET("/api/v1/exams/:id/questions", h.GetQuestionsByExamCompat)
	r.POST("/api/v1/sessions/:session_id/answers", h.SaveAnswerCompat)
	r.POST("/api/v1/sessions/:session_id/finish", h.FinishCompat)

	// start -> creates session and returns token/expired_at
	recStart := httptest.NewRecorder()
	r.ServeHTTP(recStart, httptest.NewRequest(http.MethodPost, "/api/v1/exams/exam-compat/start", strings.NewReader(`{"user_id":1}`)))
	if recStart.Code != http.StatusOK {
		t.Fatalf("start expected 200 got %d body=%s", recStart.Code, recStart.Body.String())
	}
	if !strings.Contains(recStart.Body.String(), `"session_token":"sess-compat"`) {
		t.Fatalf("expected session_token in start response, body=%s", recStart.Body.String())
	}
	if !strings.Contains(recStart.Body.String(), `"expired_at"`) {
		t.Fatalf("expected expired_at in start response, body=%s", recStart.Body.String())
	}

	// active session exists -> 409
	mockRepo.sessionByExam["exam-compat"] = studentexamrepo.Session{ID: "sess-compat", ExamID: "exam-compat", StudentID: "stu-1", Status: "in_progress"}
	mockRepo.sessionByExamOK["exam-compat"] = true
	recStart2 := httptest.NewRecorder()
	r.ServeHTTP(recStart2, httptest.NewRequest(http.MethodPost, "/api/v1/exams/exam-compat/start", strings.NewReader(`{"user_id":1}`)))
	if recStart2.Code != http.StatusConflict {
		t.Fatalf("start-active expected 409 got %d body=%s", recStart2.Code, recStart2.Body.String())
	}

	// questions + pagination and no answer key leakage
	recQ := httptest.NewRecorder()
	r.ServeHTTP(recQ, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-compat/questions?limit=2&offset=1", nil))
	if recQ.Code != http.StatusOK {
		t.Fatalf("questions expected 200 got %d body=%s", recQ.Code, recQ.Body.String())
	}
	if strings.Contains(recQ.Body.String(), "correct_answer") || strings.Contains(recQ.Body.String(), "answer_key") {
		t.Fatalf("questions response leaks answer key fields: %s", recQ.Body.String())
	}
	if !strings.Contains(recQ.Body.String(), `"total":3`) || !strings.Contains(recQ.Body.String(), `"limit":2`) {
		t.Fatalf("questions pagination meta missing: %s", recQ.Body.String())
	}

	// answers valid
	recA := httptest.NewRecorder()
	reqA := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/sess-compat/answers", strings.NewReader(`{"question_id":"q1","answer_json":{"selected":"A"}}`))
	reqA.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusOK {
		t.Fatalf("answer expected 200 got %d body=%s", recA.Code, recA.Body.String())
	}

	// answers to unauthorized session -> 403 (compat mapping)
	mockRepo.upsertAnswerErr = studentexamrepo.ErrQuestionNotInSession
	recA2 := httptest.NewRecorder()
	reqA2 := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/sess-compat/answers", strings.NewReader(`{"question_id":"q1","answer_json":{"selected":"A"}}`))
	reqA2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recA2, reqA2)
	if recA2.Code != http.StatusForbidden {
		t.Fatalf("answer forbidden expected 403 got %d body=%s", recA2.Code, recA2.Body.String())
	}

	// answers after expiry -> 422 exam time expired
	mockRepo.upsertAnswerErr = studentexamrepo.ErrSessionNotActive
	recA3 := httptest.NewRecorder()
	reqA3 := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/sess-compat/answers", strings.NewReader(`{"question_id":"q1","answer_json":{"selected":"A"}}`))
	reqA3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recA3, reqA3)
	if recA3.Code != http.StatusUnprocessableEntity {
		t.Fatalf("answer expired expected 422 got %d body=%s", recA3.Code, recA3.Body.String())
	}
	if !strings.Contains(recA3.Body.String(), "exam time expired") {
		t.Fatalf("expected exam time expired message, body=%s", recA3.Body.String())
	}

	// reset error for finish flow
	mockRepo.upsertAnswerErr = nil

	// finish first call -> 200 with score
	recF1 := httptest.NewRecorder()
	r.ServeHTTP(recF1, httptest.NewRequest(http.MethodPost, "/api/v1/sessions/sess-compat/finish", nil))
	if recF1.Code != http.StatusOK {
		t.Fatalf("finish-1 expected 200 got %d body=%s", recF1.Code, recF1.Body.String())
	}
	if !strings.Contains(recF1.Body.String(), `"score":80`) {
		t.Fatalf("finish-1 expected score in body=%s", recF1.Body.String())
	}

	// finish second call -> conflict (idempotent/race-safe)
	recF2 := httptest.NewRecorder()
	r.ServeHTTP(recF2, httptest.NewRequest(http.MethodPost, "/api/v1/sessions/sess-compat/finish", nil))
	if recF2.Code != http.StatusConflict {
		t.Fatalf("finish-2 expected 409 got %d body=%s", recF2.Code, recF2.Body.String())
	}
}

func TestStudentExamHandler_Join_TokenValidationAndMaxActiveSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockStudentExamRepo{
		studentInfo:     studentexamrepo.StudentInfo{StudentID: "stu-1", LevelID: "lvl-1", GroupID: "grp-1", IsActive: true},
		studentOK:       true,
		examForJoin:     studentexamrepo.ExamForJoin{ID: "exam-1", Title: "Math", ShuffleQuestions: true, MaxAttempts: 2},
		sessionByExam:   map[string]studentexamrepo.Session{},
		sessionByExamOK: map[string]bool{},
		createdSession: studentexamrepo.Session{
			ID:        "sess-1",
			ExamID:    "exam-1",
			StudentID: "stu-1",
			Status:    "in_progress",
		},
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta", TokenRequired: true, MaxActiveSessions: 1}})

	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/student/exams/:id/join", h.Join)

	recMissing := httptest.NewRecorder()
	reqMissing := httptest.NewRequest(http.MethodPost, "/api/v1/student/exams/exam-1/join", strings.NewReader(`{}`))
	reqMissing.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recMissing, reqMissing)
	if recMissing.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing token, got %d body=%s", recMissing.Code, recMissing.Body.String())
	}

	mockRepo.verifyTokenErr = studentexamrepo.ErrTokenExpired
	recExpired := httptest.NewRecorder()
	reqExpired := httptest.NewRequest(http.MethodPost, "/api/v1/student/exams/exam-1/join", strings.NewReader(`{"token":"ABC123"}`))
	reqExpired.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recExpired, reqExpired)
	if recExpired.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for expired token, got %d body=%s", recExpired.Code, recExpired.Body.String())
	}

	mockRepo.verifyTokenErr = nil
	mockRepo.activeCount = 1
	recMax := httptest.NewRecorder()
	reqMax := httptest.NewRequest(http.MethodPost, "/api/v1/student/exams/exam-1/join", strings.NewReader(`{"token":"ABC123"}`))
	reqMax.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recMax, reqMax)
	if recMax.Code != http.StatusConflict {
		t.Fatalf("expected 409 for max active sessions, got %d body=%s", recMax.Code, recMax.Body.String())
	}
}

func TestStudentExamHandler_GetSession_NotOwnedOrMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockStudentExamRepo{
		studentInfo:      studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: true},
		studentOK:        true,
		sessionStateByID: map[string]studentexamrepo.SessionState{},
		sessionStateOK:   map[string]bool{},
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})

	r := gin.New()
	r.Use(withAuthUser())
	r.GET("/api/v1/student/sessions/:id", h.GetSession)

	mockRepo.sessionExists = true
	recForbidden := httptest.NewRecorder()
	r.ServeHTTP(recForbidden, httptest.NewRequest(http.MethodGet, "/api/v1/student/sessions/sess-owned-by-other", nil))
	if recForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when session exists but not owned, got %d body=%s", recForbidden.Code, recForbidden.Body.String())
	}

	mockRepo.sessionExists = false
	recMissing := httptest.NewRecorder()
	r.ServeHTTP(recMissing, httptest.NewRequest(http.MethodGet, "/api/v1/student/sessions/sess-missing", nil))
	if recMissing.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when session missing, got %d body=%s", recMissing.Code, recMissing.Body.String())
	}
}

func TestStudentExamHandler_VerifyToken_StateAndTokenErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	state := studentexamrepo.SessionState{}
	state.Session = studentexamrepo.Session{ID: "sess-1", ExamID: "exam-1", StudentID: "stu-1", Status: "in_progress"}

	mockRepo := &mockStudentExamRepo{
		studentInfo:      studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: true},
		studentOK:        true,
		sessionStateByID: map[string]studentexamrepo.SessionState{"sess-1": state},
		sessionStateOK:   map[string]bool{"sess-1": true},
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})

	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/student/sessions/:id/verify-token", h.VerifyToken)

	mockRepo.verifyTokenErr = studentexamrepo.ErrTokenInactive
	recInactive := httptest.NewRecorder()
	reqInactive := httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-1/verify-token", strings.NewReader(`{"token":"ABC123"}`))
	reqInactive.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recInactive, reqInactive)
	if recInactive.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for inactive token, got %d body=%s", recInactive.Code, recInactive.Body.String())
	}

	mockRepo.verifyTokenErr = nil
	mockRepo.sessionStateByID["sess-1"] = studentexamrepo.SessionState{Session: studentexamrepo.Session{ID: "sess-1", ExamID: "exam-1", StudentID: "stu-1", Status: "submitted"}}
	recState := httptest.NewRecorder()
	reqState := httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-1/verify-token", strings.NewReader(`{"token":"ABC123"}`))
	reqState.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recState, reqState)
	if recState.Code != http.StatusConflict {
		t.Fatalf("expected 409 for inactive session state, got %d body=%s", recState.Code, recState.Body.String())
	}
}

func TestStudentExamHandler_SaveAnswer_InvalidAnswerJSONMapsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockStudentExamRepo{
		studentInfo:     studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: true},
		studentOK:       true,
		upsertAnswerErr: errors.New("invalid answer_json: malformed payload"),
	}
	h := NewStudentExamHandler(mockRepo, &mockSettingsRepo{sys: masterrepo.SystemSettings{Timezone: "Asia/Jakarta"}})
	r := gin.New()
	r.Use(withAuthUser())
	r.POST("/api/v1/student/sessions/:id/answers", h.SaveAnswer)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/student/sessions/sess-1/answers", bytes.NewBufferString(`{"question_id":"q-1","answer_json":{"selected":"A"}}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid answer_json, got %d body=%s", rec.Code, rec.Body.String())
	}
}
