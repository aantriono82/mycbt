package middleware

import "testing"

func TestShouldSkipAuditBody(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/api/v1/settings/smtp", true},
		{"/api/v1/settings/whatsapp", true},
		{"/api/v1/student/exams/123/join", true},
		{"/api/v1/student/sessions/123/verify-token", true},
		{"/api/v1/exams/123/tokens", false},
		{"/api/v1/admin/students", false},
	}
	for _, tc := range cases {
		if got := shouldSkipAuditBody(tc.path); got != tc.want {
			t.Fatalf("shouldSkipAuditBody(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

