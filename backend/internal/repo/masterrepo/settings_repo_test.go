package masterrepo

import "testing"

func TestDefaultSchoolIdentity(t *testing.T) {
	t.Parallel()

	got := defaultSchoolIdentity()
	if got.SchoolName != "AtigaCBT School" {
		t.Fatalf("expected default school name, got %q", got.SchoolName)
	}
}

func TestEnsureSchoolIdentityDefaults(t *testing.T) {
	t.Parallel()

	got := ensureSchoolIdentityDefaults(SchoolIdentity{})
	if got.SchoolName != "AtigaCBT School" {
		t.Fatalf("expected fallback school name, got %q", got.SchoolName)
	}

	got2 := ensureSchoolIdentityDefaults(SchoolIdentity{SchoolName: "SMAN 1"})
	if got2.SchoolName != "SMAN 1" {
		t.Fatalf("expected existing school name preserved, got %q", got2.SchoolName)
	}
}

func TestDefaultSystemSettings(t *testing.T) {
	t.Parallel()

	got := defaultSystemSettings()
	if got.Timezone != "Asia/Jakarta" {
		t.Fatalf("expected default timezone, got %q", got.Timezone)
	}
	if got.MaxActiveSessions != 1 {
		t.Fatalf("expected default max active sessions 1, got %d", got.MaxActiveSessions)
	}
	if !got.TokenRequired || !got.AllowResetLogin {
		t.Fatalf("expected token required and allow reset login defaults enabled, got %+v", got)
	}
}

func TestEnsureSystemSettingsDefaults(t *testing.T) {
	t.Parallel()

	got := ensureSystemSettingsDefaults(SystemSettings{})
	if got.Timezone != "Asia/Jakarta" {
		t.Fatalf("expected fallback timezone, got %q", got.Timezone)
	}
	if got.MaxActiveSessions != 1 {
		t.Fatalf("expected fallback max active sessions 1, got %d", got.MaxActiveSessions)
	}

	got2 := ensureSystemSettingsDefaults(SystemSettings{Timezone: "UTC", MaxActiveSessions: 3, TokenRequired: true})
	if got2.Timezone != "UTC" || got2.MaxActiveSessions != 3 {
		t.Fatalf("expected explicit values preserved, got %+v", got2)
	}
}
