package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"atigacbt/backend/internal/repo/examrepo"
	"atigacbt/backend/internal/repo/ltirepo"
	"atigacbt/backend/internal/repo/studentexamrepo"
	"atigacbt/backend/internal/service/ltisvc"
)

type mockExamResultsExamRepo struct {
	exam          examrepo.Exam
	ok            bool
	err           error
	teacherID     string
	teacherOK     bool
	teacherErr    error
	sessionExamID string
	sessionOK     bool
	sessionErr    error
}

func (m *mockExamResultsExamRepo) Get(_ context.Context, id string) (examrepo.Exam, bool, error) {
	return m.exam, m.ok, m.err
}

func (m *mockExamResultsExamRepo) TeacherIDByUserID(_ context.Context, userID string) (string, bool, error) {
	return m.teacherID, m.teacherOK, m.teacherErr
}

func (m *mockExamResultsExamRepo) SessionExamID(_ context.Context, sessionID string) (string, bool, error) {
	return m.sessionExamID, m.sessionOK, m.sessionErr
}

type mockExamResultsLTIRepo struct {
	targets []ltirepo.AGSScoreTarget
	err     error
}

func (m *mockExamResultsLTIRepo) ListAGSScoreTargets(_ context.Context, examID string) ([]ltirepo.AGSScoreTarget, error) {
	return m.targets, m.err
}

type mockExamResultsStudentRepo struct {
	sessions          []studentexamrepo.ExamSessionRow
	sessionsTotal     int
	sessionsErr       error
	attendanceItems   []studentexamrepo.ExamAttendanceParticipant
	attendanceTotal   int
	targetedTotal     int
	attendedTotal     int
	attendanceErr     error
	scoresBySession   map[string]studentexamrepo.AutoScoreSummary
	scoreErrBySession map[string]error
	itemRows          []studentexamrepo.ExamItemAnalysisRow
	itemSessions      int
	itemErr           error
	scoreDist         studentexamrepo.ExamScoreDistribution
	scoreDistErr      error
	essayAttempts     []studentexamrepo.EssayAttempt
	essayErr          error
	saveEssayCalls    []saveEssayCall
	saveEssayErr      error
}

type saveEssayCall struct {
	SessionID  string
	QuestionID string
	Score      int
	Feedback   string
}

func (m *mockExamResultsStudentRepo) ListExamSessionsWithScore(_ context.Context, examID string, f studentexamrepo.ListExamSessionsFilter) ([]studentexamrepo.ExamSessionRow, int, error) {
	return m.sessions, m.sessionsTotal, m.sessionsErr
}

func (m *mockExamResultsStudentRepo) ListExamAttendanceParticipants(_ context.Context, examID string, f studentexamrepo.ExamAttendanceFilter) ([]studentexamrepo.ExamAttendanceParticipant, int, int, int, error) {
	return m.attendanceItems, m.attendanceTotal, m.targetedTotal, m.attendedTotal, m.attendanceErr
}

func (m *mockExamResultsStudentRepo) ComputeAutoScoreAny(_ context.Context, sessionID string, nowUTC time.Time) (studentexamrepo.AutoScoreSummary, error) {
	if err := m.scoreErrBySession[sessionID]; err != nil {
		return studentexamrepo.AutoScoreSummary{}, err
	}
	return m.scoresBySession[sessionID], nil
}

func (m *mockExamResultsStudentRepo) ListExamItemAnalysis(_ context.Context, examID string) ([]studentexamrepo.ExamItemAnalysisRow, int, error) {
	return m.itemRows, m.itemSessions, m.itemErr
}

func (m *mockExamResultsStudentRepo) GetExamScoreDistribution(_ context.Context, examID string, nowUTC time.Time) (studentexamrepo.ExamScoreDistribution, error) {
	return m.scoreDist, m.scoreDistErr
}

func (m *mockExamResultsStudentRepo) ListEssayAttempts(_ context.Context, sessionID string) ([]studentexamrepo.EssayAttempt, error) {
	return m.essayAttempts, m.essayErr
}

func (m *mockExamResultsStudentRepo) SaveManualScoring(_ context.Context, sessionID, questionID string, score int, feedback string) error {
	m.saveEssayCalls = append(m.saveEssayCalls, saveEssayCall{SessionID: sessionID, QuestionID: questionID, Score: score, Feedback: feedback})
	return m.saveEssayErr
}

type mockAGSPublisher struct {
	errBySession map[string]error
	calls        []ltisvc.PublishScoreInput
}

func (m *mockAGSPublisher) PublishScore(_ context.Context, in ltisvc.PublishScoreInput, scopeText string) error {
	m.calls = append(m.calls, in)
	if err := m.errBySession[in.LTISub]; err != nil {
		return err
	}
	return nil
}

type noopNotificationSender struct{}

func (noopNotificationSender) SendEmail(_ context.Context, to string, subject string, body string) error {
	return nil
}

func (noopNotificationSender) SendWhatsApp(_ context.Context, to string, message string) error {
	return nil
}

type mockNotificationSender struct {
	emailErr   error
	waErr      error
	emailCalls []string
	waCalls    []string
}

func (m *mockNotificationSender) SendEmail(_ context.Context, to string, subject string, body string) error {
	m.emailCalls = append(m.emailCalls, to)
	return m.emailErr
}

func (m *mockNotificationSender) SendWhatsApp(_ context.Context, to string, message string) error {
	m.waCalls = append(m.waCalls, to)
	return m.waErr
}

func withRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("user_role", role)
		c.Next()
	}
}

func TestExamResultsHandler_List_TeacherForbiddenForOtherTeachersExam(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam:      examrepo.Exam{ID: "exam-1", TeacherID: "teacher-owner", Title: "Math"},
			ok:        true,
			teacherID: "teacher-other",
			teacherOK: true,
		},
		st:    &mockExamResultsStudentRepo{},
		notif: noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("teacher"))
	r.GET("/api/v1/exams/:id/results", h.List)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/results", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestExamResultsHandler_Attendance_ComputesRate(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam: examrepo.Exam{ID: "exam-1", TeacherID: "teacher-1", Title: "Math"},
			ok:   true,
		},
		st: &mockExamResultsStudentRepo{
			attendanceItems: []studentexamrepo.ExamAttendanceParticipant{{StudentID: "stu-1", Name: "A", Attended: true}},
			attendanceTotal: 1,
			targetedTotal:   5,
			attendedTotal:   3,
		},
		notif: noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.GET("/api/v1/exams/:id/attendance", h.Attendance)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/attendance", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"attendance_rate_percent":60`) {
		t.Fatalf("expected attendance rate 60, body=%s", rec.Body.String())
	}
}

func TestExamResultsHandler_SyncLTIScores_CountsSyncedSkippedAndFailed(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	ltiRepo := &mockExamResultsLTIRepo{
		targets: []ltirepo.AGSScoreTarget{
			{SessionID: "sess-1", LTISub: "sub-1", LineItemURL: "https://lms/items/1", ScopeText: ltisvc.AGSScoreScope, Platform: ltirepo.Platform{ClientID: "cid"}},
			{SessionID: "sess-2", LTISub: "sub-2", LineItemURL: "https://lms/items/2", ScopeText: ltisvc.AGSScoreScope, Platform: ltirepo.Platform{ClientID: "cid"}},
			{SessionID: "sess-3", LTISub: "sub-3", LineItemURL: "https://lms/items/3", ScopeText: ltisvc.AGSScoreScope, Platform: ltirepo.Platform{ClientID: "cid"}},
		},
	}
	stRepo := &mockExamResultsStudentRepo{
		scoresBySession: map[string]studentexamrepo.AutoScoreSummary{
			"sess-1": {Score: 88},
			"sess-2": {Score: 75, PendingGrading: 1},
		},
		scoreErrBySession: map[string]error{
			"sess-3": errors.New("compute failed"),
		},
	}
	ags := &mockAGSPublisher{errBySession: map[string]error{}}
	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam: examrepo.Exam{ID: "exam-1", Title: "Math"},
			ok:   true,
		},
		lti:    ltiRepo,
		ltiAGS: ags,
		st:     stRepo,
		notif:  noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.POST("/api/v1/exams/:id/lti/sync-scores", h.SyncLTIScores)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/exams/exam-1/lti/sync-scores", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"synced_count":1`) || !strings.Contains(rec.Body.String(), `"skipped_count":1`) || !strings.Contains(rec.Body.String(), `"failed_count":1`) {
		t.Fatalf("unexpected sync summary body=%s", rec.Body.String())
	}
	if len(ags.calls) != 1 {
		t.Fatalf("expected one publish call, got %d", len(ags.calls))
	}
}

func TestExamResultsHandler_ItemSuggestions_BuildsPriorities(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam: examrepo.Exam{ID: "exam-1", Title: "Math"},
			ok:   true,
		},
		st: &mockExamResultsStudentRepo{
			itemRows: []studentexamrepo.ExamItemAnalysisRow{
				{
					QuestionID:        "q1",
					AutoScorable:      true,
					DiscriminationIdx: -5,
					PValuePercent:     20,
					OptionStats: []studentexamrepo.ExamItemOptionStat{
						{Label: "A", IsCorrect: true, SelectedCount: 2},
						{Label: "B", IsCorrect: false, SelectedCount: 8},
					},
				},
				{
					QuestionID:        "q2",
					AutoScorable:      true,
					DiscriminationIdx: 10,
					PValuePercent:     95,
					OptionStats: []studentexamrepo.ExamItemOptionStat{
						{Label: "A", IsCorrect: true, SelectedCount: 10},
						{Label: "B", IsCorrect: false, SelectedCount: 1},
					},
				},
				{
					QuestionID:        "q3",
					AutoScorable:      false,
					DiscriminationIdx: -10,
					PValuePercent:     10,
				},
			},
		},
		notif: noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.GET("/api/v1/exams/:id/item-analysis/suggestions", h.ItemSuggestions)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/item-analysis/suggestions", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []ItemSuggestion `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(resp.Data))
	}
	if resp.Data[0].QuestionID != "q1" || resp.Data[0].Priority != "high" {
		t.Fatalf("expected q1 high priority, got %+v", resp.Data[0])
	}
	if resp.Data[1].QuestionID != "q2" {
		t.Fatalf("expected q2 suggestion, got %+v", resp.Data[1])
	}
}

func TestExamResultsHandler_ScoreDistribution_ReturnsData(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam: examrepo.Exam{ID: "exam-1", Title: "Math"},
			ok:   true,
		},
		st: &mockExamResultsStudentRepo{
			scoreDist: studentexamrepo.ExamScoreDistribution{
				TotalSessions:  4,
				SubmittedCount: 3,
				ExpiredCount:   1,
				AverageScore:   77.5,
			},
		},
		notif: noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.GET("/api/v1/exams/:id/score-distribution", h.ScoreDistribution)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/score-distribution", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"average_score":77.5`) {
		t.Fatalf("expected average score in body=%s", rec.Body.String())
	}
}

func TestExamResultsHandler_ExportsReturnFiles(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	stRepo := &mockExamResultsStudentRepo{
		sessions: []studentexamrepo.ExamSessionRow{
			{
				SessionID:           "sess-1",
				StudentName:         "Alice",
				StudentUsername:     "alice",
				StudentNIS:          "001",
				Status:              "submitted",
				StartedAt:           "2026-05-01T00:00:00Z",
				FinishedAt:          "2026-05-01T01:00:00Z",
				AnsweredQuestions:   8,
				TotalQuestions:      10,
				AutoScorable:        8,
				CorrectCount:        7,
				Score:               88,
				DurationSeconds:     3600,
				PendingGradingCount: 0,
			},
		},
		sessionsTotal: 1,
		scoreDist: studentexamrepo.ExamScoreDistribution{
			TotalSessions:   1,
			SubmittedCount:  1,
			ExpiredCount:    0,
			AverageScore:    88,
			MedianScore:     88,
			MinScore:        88,
			MaxScore:        88,
			DistributionBin: []studentexamrepo.ScoreDistributionBin{{Label: "80-89", Count: 1, Percent: 100}},
		},
		itemRows: []studentexamrepo.ExamItemAnalysisRow{
			{
				QuestionID:         "q1",
				OrderNo:            1,
				QuestionType:       "mc_single",
				Stem:               "2+2?",
				Participants:       1,
				AnsweredCount:      1,
				CorrectCount:       1,
				PValuePercent:      100,
				DifficultyLabel:    "mudah",
				AutoScorable:       true,
				DiscriminationIdx:  0,
				DiscriminationNote: "n/a",
			},
		},
		itemSessions: 1,
	}
	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam: examrepo.Exam{ID: "exam-1", Title: "Math"},
			ok:   true,
		},
		st:    stRepo,
		notif: noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.GET("/api/v1/exams/:id/export", h.Export)
	r.GET("/api/v1/exams/:id/export.pdf", h.ExportPDF)
	r.GET("/api/v1/exams/:id/item-analysis/export", h.ExportItemAnalysis)

	recXLSX := httptest.NewRecorder()
	r.ServeHTTP(recXLSX, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/export", nil))
	if recXLSX.Code != http.StatusOK {
		t.Fatalf("xlsx export expected 200, got %d body=%s", recXLSX.Code, recXLSX.Body.String())
	}
	if ct := recXLSX.Header().Get("Content-Type"); !strings.Contains(ct, "spreadsheetml.sheet") {
		t.Fatalf("unexpected xlsx content-type %q", ct)
	}
	if body := recXLSX.Body.Bytes(); len(body) < 2 || string(body[:2]) != "PK" {
		t.Fatalf("expected zip/xlsx payload, first bytes=%q", string(body))
	}

	recPDF := httptest.NewRecorder()
	r.ServeHTTP(recPDF, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/export.pdf", nil))
	if recPDF.Code != http.StatusOK {
		t.Fatalf("pdf export expected 200, got %d body=%s", recPDF.Code, recPDF.Body.String())
	}
	if ct := recPDF.Header().Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		t.Fatalf("unexpected pdf content-type %q", ct)
	}
	if body := recPDF.Body.Bytes(); len(body) < 4 || string(body[:4]) != "%PDF" {
		t.Fatalf("expected pdf payload, first bytes=%q", string(body))
	}

	recItem := httptest.NewRecorder()
	r.ServeHTTP(recItem, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/item-analysis/export", nil))
	if recItem.Code != http.StatusOK {
		t.Fatalf("item export expected 200, got %d body=%s", recItem.Code, recItem.Body.String())
	}
	if ct := recItem.Header().Get("Content-Type"); !strings.Contains(ct, "spreadsheetml.sheet") {
		t.Fatalf("unexpected item export content-type %q", ct)
	}
}

func TestExamResultsHandler_ListEssays_And_SaveEssayScore(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	stRepo := &mockExamResultsStudentRepo{
		essayAttempts: []studentexamrepo.EssayAttempt{
			{QuestionID: "q1", QuestionStem: "Explain", QuestionOrder: 1, AnswerText: "Answer", MaxScore: 100},
		},
	}
	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam:          examrepo.Exam{ID: "exam-1", Title: "Math"},
			ok:            true,
			sessionExamID: "exam-1",
			sessionOK:     true,
		},
		st:    stRepo,
		notif: noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.GET("/api/v1/exams/:id/sessions/:sessionId/essays", h.ListEssays)
	r.POST("/api/v1/exams/:id/sessions/:sessionId/essays/score", h.SaveEssayScore)

	recList := httptest.NewRecorder()
	r.ServeHTTP(recList, httptest.NewRequest(http.MethodGet, "/api/v1/exams/exam-1/sessions/sess-1/essays", nil))
	if recList.Code != http.StatusOK {
		t.Fatalf("list essays expected 200, got %d body=%s", recList.Code, recList.Body.String())
	}
	if !strings.Contains(recList.Body.String(), `"question_id":"q1"`) {
		t.Fatalf("expected essay attempt in body=%s", recList.Body.String())
	}

	recSave := httptest.NewRecorder()
	reqSave := httptest.NewRequest(http.MethodPost, "/api/v1/exams/exam-1/sessions/sess-1/essays/score", strings.NewReader(`{"question_id":"q1","score":90,"feedback":"good"}`))
	reqSave.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recSave, reqSave)
	if recSave.Code != http.StatusOK {
		t.Fatalf("save essay expected 200, got %d body=%s", recSave.Code, recSave.Body.String())
	}
	if len(stRepo.saveEssayCalls) != 1 {
		t.Fatalf("expected one save essay call, got %d", len(stRepo.saveEssayCalls))
	}
	if stRepo.saveEssayCalls[0].QuestionID != "q1" || stRepo.saveEssayCalls[0].Score != 90 {
		t.Fatalf("unexpected save essay call %+v", stRepo.saveEssayCalls[0])
	}
}

func TestExamResultsHandler_SaveEssayScore_RejectsWrongExamSessionAndBadBody(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam:          examrepo.Exam{ID: "exam-1", Title: "Math"},
			ok:            true,
			sessionExamID: "exam-2",
			sessionOK:     true,
		},
		st:    &mockExamResultsStudentRepo{},
		notif: noopNotificationSender{},
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.POST("/api/v1/exams/:id/sessions/:sessionId/essays/score", h.SaveEssayScore)

	recForbidden := httptest.NewRecorder()
	reqForbidden := httptest.NewRequest(http.MethodPost, "/api/v1/exams/exam-1/sessions/sess-1/essays/score", strings.NewReader(`{"question_id":"q1","score":90}`))
	reqForbidden.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recForbidden, reqForbidden)
	if recForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", recForbidden.Code, recForbidden.Body.String())
	}

	h.ex.(*mockExamResultsExamRepo).sessionExamID = "exam-1"
	recBad := httptest.NewRecorder()
	reqBad := httptest.NewRequest(http.MethodPost, "/api/v1/exams/exam-1/sessions/sess-1/essays/score", strings.NewReader(`{"score":90}`))
	reqBad.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", recBad.Code, recBad.Body.String())
	}
}

func TestExamResultsHandler_BlastResults_SendsSelectedChannelsForFinalizedSessions(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	notif := &mockNotificationSender{waErr: errors.New("wa down")}
	h := &ExamResultsHandler{
		ex: &mockExamResultsExamRepo{
			exam: examrepo.Exam{ID: "exam-1", Title: "Math"},
			ok:   true,
		},
		st: &mockExamResultsStudentRepo{
			sessions: []studentexamrepo.ExamSessionRow{
				{Status: "submitted", StudentName: "Alice", StudentEmail: "alice@example.com", StudentPhone: "081", Score: 88, CorrectCount: 7, TotalQuestions: 10},
				{Status: "forced", StudentName: "Bob", StudentEmail: "bob@example.com", StudentPhone: "082", Score: 70, CorrectCount: 6, TotalQuestions: 10},
				{Status: "in_progress", StudentName: "Eve", StudentEmail: "eve@example.com", StudentPhone: "083", Score: 0, CorrectCount: 0, TotalQuestions: 10},
			},
			sessionsTotal: 3,
		},
		notif: notif,
	}

	r := gin.New()
	r.Use(withRole("admin"))
	r.POST("/api/v1/exams/:id/results/blast", h.BlastResults)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/exams/exam-1/results/blast", strings.NewReader(`{"channels":["email","whatsapp"]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if len(notif.emailCalls) != 2 {
		t.Fatalf("expected 2 email calls, got %d", len(notif.emailCalls))
	}
	if len(notif.waCalls) != 2 {
		t.Fatalf("expected 2 wa calls, got %d", len(notif.waCalls))
	}
	if !strings.Contains(rec.Body.String(), `"sent_count":2`) || !strings.Contains(rec.Body.String(), `"failed_count":2`) {
		t.Fatalf("unexpected blast summary body=%s", rec.Body.String())
	}
}
