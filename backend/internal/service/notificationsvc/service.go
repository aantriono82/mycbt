package notificationsvc

import (
	"context"
	"fmt"
	"net/smtp"

	"atigacbt/backend/internal/config"
	"atigacbt/backend/internal/repo/masterrepo"
)

type Service struct {
	settings settingsProvider
	cfg      config.Config
	sendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

func New(settings *masterrepo.SettingsRepo, cfg config.Config) *Service {
	return &Service{settings: settings, cfg: cfg, sendMail: smtp.SendMail}
}

type settingsProvider interface {
	GetSMTP(ctx context.Context) (masterrepo.SMTPConfig, error)
	GetWhatsApp(ctx context.Context) (masterrepo.WhatsAppConfig, error)
}

func (s *Service) SendEmail(ctx context.Context, to string, subject string, body string) error {
	cfg, err := s.settings.GetSMTP(ctx)
	if err != nil {
		return fmt.Errorf("get smtp config: %w", err)
	}

	host := cfg.Host
	port := cfg.Port
	user := cfg.User
	password := cfg.Password
	from := cfg.From

	// Fallback to env config if not set in DB
	if host == "" {
		host = s.cfg.SMTPHost
		if s.cfg.SMTPPort != "" {
			fmt.Sscanf(s.cfg.SMTPPort, "%d", &port)
		}
		user = s.cfg.SMTPUser
		password = s.cfg.SMTPPass
		from = s.cfg.SMTPFrom
	}

	if host == "" {
		return fmt.Errorf("smtp host is not configured")
	}

	auth := smtp.PlainAuth("", user, password, host)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body))

	addr := fmt.Sprintf("%s:%d", host, port)
	err = s.sendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

func (s *Service) SendWhatsApp(ctx context.Context, to string, message string) error {
	cfg, err := s.settings.GetWhatsApp(ctx)
	if err != nil {
		return fmt.Errorf("get whatsapp config: %w", err)
	}
	if cfg.APIURL == "" {
		return fmt.Errorf("whatsapp api url is not configured")
	}

	// For now, this is a mock implementation of a generic HTTP-based WhatsApp API.
	// In a real scenario, this would use the configured APIURL and APIKey.
	// Example: http.Post(cfg.APIURL, "application/json", body)

	fmt.Printf("[MOCK WA] Sending to %s: %s\n", to, message)
	return nil
}
