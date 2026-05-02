package handlers

import (
	"context"
	"testing"
	"time"

	"atigacbt/backend/internal/repo/studentexamrepo"
)

type mockNotificationStudentRepo struct {
	studentInfo studentexamrepo.StudentInfo
	studentOK   bool
	studentErr  error

	announcements      []studentexamrepo.StudentAnnouncement
	announcementsTotal int
	announcementsErr   error

	exams     []studentexamrepo.StudentExam
	examTotal int
	examsErr  error
}

func (m *mockNotificationStudentRepo) StudentByUserID(ctx context.Context, userID string) (studentexamrepo.StudentInfo, bool, error) {
	return m.studentInfo, m.studentOK, m.studentErr
}

func (m *mockNotificationStudentRepo) ListAvailableForStudent(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentExamsFilter) ([]studentexamrepo.StudentExam, int, error) {
	return m.exams, m.examTotal, m.examsErr
}

func (m *mockNotificationStudentRepo) ListStudentAnnouncements(ctx context.Context, studentID, levelID, groupID string, f studentexamrepo.ListStudentAnnouncementsFilter) ([]studentexamrepo.StudentAnnouncement, int, error) {
	return m.announcements, m.announcementsTotal, m.announcementsErr
}

func TestNotificationHandlerBuildSnapshot_ChangesWhenCountsOrItemsChange(t *testing.T) {
	repo := &mockNotificationStudentRepo{
		studentInfo: studentexamrepo.StudentInfo{StudentID: "stu-1", LevelID: "lvl-1", GroupID: "grp-1", IsActive: true},
		studentOK:   true,
		announcements: []studentexamrepo.StudentAnnouncement{
			{ID: "ann-1", PublishedAt: "2026-05-02T10:00:00Z"},
		},
		announcementsTotal: 1,
		exams: []studentexamrepo.StudentExam{
			{ID: "exam-1", StartsAt: "2026-05-02T11:00:00Z", EndsAt: "2026-05-02T12:00:00Z", SessionStatus: "", ActiveToken: "ABC", CanJoin: false},
		},
		examTotal: 1,
	}
	h := &NotificationHandler{st: repo}

	base, err := h.buildSnapshot(context.Background(), "user-1", time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("buildSnapshot() error = %v", err)
	}

	repo.announcementsTotal = 2
	changedTotal, err := h.buildSnapshot(context.Background(), "user-1", time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("buildSnapshot() after total change error = %v", err)
	}
	if base == changedTotal {
		t.Fatalf("expected snapshot to change when totals change")
	}

	repo.announcementsTotal = 1
	repo.exams[0].ActiveToken = "XYZ"
	changedItem, err := h.buildSnapshot(context.Background(), "user-1", time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("buildSnapshot() after item change error = %v", err)
	}
	if base == changedItem {
		t.Fatalf("expected snapshot to change when exam item changes")
	}
}

func TestNotificationHandlerBuildSnapshot_EmptyWhenStudentMissingOrInactive(t *testing.T) {
	cases := []struct {
		name string
		repo *mockNotificationStudentRepo
	}{
		{
			name: "missing student",
			repo: &mockNotificationStudentRepo{},
		},
		{
			name: "inactive student",
			repo: &mockNotificationStudentRepo{
				studentInfo: studentexamrepo.StudentInfo{StudentID: "stu-1", IsActive: false},
				studentOK:   true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &NotificationHandler{st: tc.repo}
			got, err := h.buildSnapshot(context.Background(), "user-1", time.Now().UTC())
			if err != nil {
				t.Fatalf("buildSnapshot() error = %v", err)
			}
			if got != "" {
				t.Fatalf("expected empty snapshot, got %q", got)
			}
		})
	}
}
