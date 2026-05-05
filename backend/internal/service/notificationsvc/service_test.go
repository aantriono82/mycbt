package notificationsvc

import (
	"context"
	"errors"
	"net/smtp"
	"strings"
	"testing"

	"atigacbt/backend/internal/config"
	"atigacbt/backend/internal/repo/masterrepo"
)

type stubSettingsProvider struct {
	smtpCfg masterrepo.SMTPConfig
	smtpErr error
	waCfg   masterrepo.WhatsAppConfig
	waErr   error
}

func (s stubSettingsProvider) GetSMTP(_ context.Context) (masterrepo.SMTPConfig, error) {
	return s.smtpCfg, s.smtpErr
}

func (s stubSettingsProvider) GetWhatsApp(_ context.Context) (masterrepo.WhatsAppConfig, error) {
	return s.waCfg, s.waErr
}

func TestSendEmail_UsesDBConfigAndBuildsMessage(t *testing.T) {
	t.Parallel()

	var gotAddr, gotFrom string
	var gotTo []string
	var gotMsg []byte

	svc := &Service{
		settings: stubSettingsProvider{
			smtpCfg: masterrepo.SMTPConfig{
				Host: "smtp.example.com",
				Port: 2525,
				User: "mailer",
				From: "noreply@example.com",
			},
		},
		sendMail: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			gotAddr = addr
			gotFrom = from
			gotTo = to
			gotMsg = msg
			return nil
		},
	}

	err := svc.SendEmail(context.Background(), "user@example.com", "Hello", "<b>Body</b>")
	if err != nil {
		t.Fatalf("SendEmail error: %v", err)
	}
	if gotAddr != "smtp.example.com:2525" || gotFrom != "noreply@example.com" || len(gotTo) != 1 || gotTo[0] != "user@example.com" {
		t.Fatalf("unexpected send args: addr=%q from=%q to=%v", gotAddr, gotFrom, gotTo)
	}
	msg := string(gotMsg)
	if !strings.Contains(msg, "Subject: Hello") || !strings.Contains(msg, "Content-Type: text/html") || !strings.Contains(msg, "<b>Body</b>") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestSendEmail_FallsBackToEnvConfig(t *testing.T) {
	t.Parallel()

	var gotAddr string
	svc := &Service{
		settings: stubSettingsProvider{smtpCfg: masterrepo.SMTPConfig{}},
		cfg: config.Config{
			SMTPHost: "env-smtp.example.com",
			SMTPPort: "1025",
			SMTPUser: "env-user",
			SMTPPass: "env-pass",
			SMTPFrom: "env@example.com",
		},
		sendMail: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			gotAddr = addr
			return nil
		},
	}

	if err := svc.SendEmail(context.Background(), "user@example.com", "Hello", "Body"); err != nil {
		t.Fatalf("SendEmail error: %v", err)
	}
	if gotAddr != "env-smtp.example.com:1025" {
		t.Fatalf("expected env fallback addr, got %q", gotAddr)
	}
}

func TestSendEmail_UsesImplicitTLSWhenConfigured(t *testing.T) {
	t.Parallel()

	var gotHost string
	var gotPort int
	var gotFrom string
	svc := &Service{
		settings: stubSettingsProvider{
			smtpCfg: masterrepo.SMTPConfig{
				Host:   "smtp.example.com",
				User:   "mailer",
				From:   "noreply@example.com",
				UseTLS: true,
			},
		},
		sendTLSMail: func(host string, port int, a smtp.Auth, from string, to []string, msg []byte) error {
			gotHost = host
			gotPort = port
			gotFrom = from
			return nil
		},
	}

	if err := svc.SendEmail(context.Background(), "user@example.com", "Hello", "Body"); err != nil {
		t.Fatalf("SendEmail error: %v", err)
	}
	if gotHost != "smtp.example.com" || gotPort != 465 || gotFrom != "noreply@example.com" {
		t.Fatalf("unexpected TLS send args: host=%q port=%d from=%q", gotHost, gotPort, gotFrom)
	}
}

func TestSendEmail_PropagatesConfigAndTransportErrors(t *testing.T) {
	t.Parallel()

	svc := &Service{
		settings: stubSettingsProvider{smtpErr: errors.New("db down")},
		sendMail: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return nil
		},
	}
	if err := svc.SendEmail(context.Background(), "user@example.com", "Hello", "Body"); err == nil {
		t.Fatal("expected settings error")
	}

	svc2 := &Service{
		settings: stubSettingsProvider{smtpCfg: masterrepo.SMTPConfig{}},
		cfg:      config.Config{},
		sendMail: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return nil
		},
	}
	if err := svc2.SendEmail(context.Background(), "user@example.com", "Hello", "Body"); err == nil {
		t.Fatal("expected missing host error")
	}

	svcMissingFrom := &Service{
		settings: stubSettingsProvider{smtpCfg: masterrepo.SMTPConfig{Host: "smtp.example.com", Port: 25}},
		sendMail: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return nil
		},
	}
	if err := svcMissingFrom.SendEmail(context.Background(), "user@example.com", "Hello", "Body"); err == nil {
		t.Fatal("expected missing from error")
	}

	svc3 := &Service{
		settings: stubSettingsProvider{smtpCfg: masterrepo.SMTPConfig{Host: "smtp.example.com", Port: 25, From: "from@example.com"}},
		sendMail: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return errors.New("send failed")
		},
	}
	if err := svc3.SendEmail(context.Background(), "user@example.com", "Hello", "Body"); err == nil {
		t.Fatal("expected sendMail error")
	}
}

func TestSendWhatsApp_ValidatesConfig(t *testing.T) {
	t.Parallel()

	svc := &Service{settings: stubSettingsProvider{waErr: errors.New("db down")}}
	if err := svc.SendWhatsApp(context.Background(), "081", "hello"); err == nil {
		t.Fatal("expected whatsapp settings error")
	}

	svc2 := &Service{settings: stubSettingsProvider{waCfg: masterrepo.WhatsAppConfig{}}}
	if err := svc2.SendWhatsApp(context.Background(), "081", "hello"); err == nil {
		t.Fatal("expected missing api url error")
	}

	svc3 := &Service{settings: stubSettingsProvider{waCfg: masterrepo.WhatsAppConfig{APIURL: "https://wa.example.com"}}}
	if err := svc3.SendWhatsApp(context.Background(), "081", "hello"); err != nil {
		t.Fatalf("expected mock whatsapp success, got %v", err)
	}
}
