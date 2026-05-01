package masterrepo

import (
	"context"
	"testing"

	"atigacbt/backend/internal/testutil/pgtest"
)

func TestSettingsRepo_DefaultsAndRoundTrip(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()
	repo := NewSettings(h.Pool)

	sys, err := repo.GetSystem(ctx)
	if err != nil {
		t.Fatalf("GetSystem default error: %v", err)
	}
	if sys.Timezone != "Asia/Jakarta" || sys.MaxActiveSessions != 1 {
		t.Fatalf("unexpected default system settings: %+v", sys)
	}

	school, err := repo.GetSchoolIdentity(ctx)
	if err != nil {
		t.Fatalf("GetSchoolIdentity default error: %v", err)
	}
	if school.SchoolName != "AtigaCBT School" {
		t.Fatalf("unexpected default school identity: %+v", school)
	}

	updatedSys, err := repo.UpsertSystem(ctx, SystemSettings{Timezone: "", MaxActiveSessions: 0, TokenRequired: false})
	if err != nil {
		t.Fatalf("UpsertSystem error: %v", err)
	}
	if updatedSys.Timezone != "Asia/Jakarta" || updatedSys.MaxActiveSessions != 1 {
		t.Fatalf("expected sanitized system settings, got %+v", updatedSys)
	}

	readSys, err := repo.GetSystem(ctx)
	if err != nil {
		t.Fatalf("GetSystem readback error: %v", err)
	}
	if readSys.Timezone != "Asia/Jakarta" || readSys.MaxActiveSessions != 1 {
		t.Fatalf("unexpected system readback: %+v", readSys)
	}

	updatedSchool, err := repo.UpsertSchoolIdentity(ctx, SchoolIdentity{})
	if err != nil {
		t.Fatalf("UpsertSchoolIdentity error: %v", err)
	}
	if updatedSchool.SchoolName != "AtigaCBT School" {
		t.Fatalf("expected fallback school name, got %+v", updatedSchool)
	}

	smtpCfg, err := repo.UpsertSMTP(ctx, SMTPConfig{
		Host:     "smtp.example.com",
		Port:     2525,
		User:     "mailer",
		Password: "secret",
		From:     "noreply@example.com",
		UseTLS:   true,
	})
	if err != nil {
		t.Fatalf("UpsertSMTP error: %v", err)
	}
	if smtpCfg.Host != "smtp.example.com" || smtpCfg.Port != 2525 {
		t.Fatalf("unexpected smtp cfg: %+v", smtpCfg)
	}

	gotSMTP, err := repo.GetSMTP(ctx)
	if err != nil {
		t.Fatalf("GetSMTP error: %v", err)
	}
	if gotSMTP.From != "noreply@example.com" || !gotSMTP.UseTLS {
		t.Fatalf("unexpected smtp readback: %+v", gotSMTP)
	}

	waCfg, err := repo.UpsertWhatsApp(ctx, WhatsAppConfig{
		APIURL: "https://wa.example.com",
		APIKey: "key-1",
		Sender: "Atiga",
	})
	if err != nil {
		t.Fatalf("UpsertWhatsApp error: %v", err)
	}
	if waCfg.APIURL != "https://wa.example.com" {
		t.Fatalf("unexpected wa cfg: %+v", waCfg)
	}

	gotWA, err := repo.GetWhatsApp(ctx)
	if err != nil {
		t.Fatalf("GetWhatsApp error: %v", err)
	}
	if gotWA.Sender != "Atiga" || gotWA.APIKey != "key-1" {
		t.Fatalf("unexpected wa readback: %+v", gotWA)
	}
}
