package notificationsvc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"atigacbt/backend/internal/config"
	"atigacbt/backend/internal/repo/masterrepo"
)

type Service struct {
	settings    settingsProvider
	cfg         config.Config
	sendMail    func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	sendTLSMail func(host string, port int, a smtp.Auth, from string, to []string, msg []byte) error
}

func New(settings *masterrepo.SettingsRepo, cfg config.Config) *Service {
	return &Service{settings: settings, cfg: cfg, sendMail: smtp.SendMail, sendTLSMail: sendMailImplicitTLS}
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
	useTLS := cfg.UseTLS

	// Fallback to env config if not set in DB
	if host == "" {
		host = s.cfg.SMTPHost
		if s.cfg.SMTPPort != "" {
			if parsedPort, parseErr := strconv.Atoi(strings.TrimSpace(s.cfg.SMTPPort)); parseErr == nil {
				port = parsedPort
			}
		}
		user = s.cfg.SMTPUser
		password = s.cfg.SMTPPass
		from = s.cfg.SMTPFrom
	}

	host = strings.TrimSpace(host)
	from = strings.TrimSpace(from)
	if host == "" {
		return fmt.Errorf("smtp host is not configured")
	}
	if from == "" {
		return fmt.Errorf("smtp from address is not configured")
	}
	if port <= 0 {
		if useTLS {
			port = 465
		} else {
			port = 587
		}
	}

	var auth smtp.Auth
	if strings.TrimSpace(user) != "" {
		auth = smtp.PlainAuth("", user, password, host)
	}
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body))

	if useTLS {
		return s.sendTLSMail(host, port, auth, from, []string{to}, msg)
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	err = s.sendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

func sendMailImplicitTLS(host string, port int, a smtp.Auth, from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := tls.DialWithDialer(&net.Dialer{}, "tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("connect smtp tls: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Quit()

	if a != nil {
		if err := client.Auth(a); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", addr, err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write smtp message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close smtp data: %w", err)
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
