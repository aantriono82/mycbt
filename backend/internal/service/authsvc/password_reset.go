package authsvc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/notificationsvc"
)

type PasswordResetService struct {
	users     *userrepo.Repo
	tokens    *userrepo.PasswordResetRepo
	notif     *notificationsvc.Service
	appURL    string
}

func NewPasswordResetService(users *userrepo.Repo, tokens *userrepo.PasswordResetRepo, notif *notificationsvc.Service, appURL string) *PasswordResetService {
	return &PasswordResetService{
		users:  users,
		tokens: tokens,
		notif:  notif,
		appURL: appURL,
	}
}

func (s *PasswordResetService) ForgotPassword(ctx context.Context, email string) error {
	user, ok, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if !ok {
		// We return nil to avoid revealing if an email exists
		return nil
	}

	token := s.generateToken(32)
	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	if err := s.tokens.Create(ctx, email, token, expiresAt); err != nil {
		return fmt.Errorf("create reset token: %w", err)
	}

	resetURL := fmt.Sprintf("%s/#/auth/reset-password?token=%s", s.appURL, token)
	subject := "Reset Kata Sandi Atiga CBT"
	body := fmt.Sprintf(`
		<h3>Permintaan Reset Kata Sandi</h3>
		<p>Halo %s,</p>
		<p>Kami menerima permintaan untuk mengatur ulang kata sandi akun Anda. Silakan klik tautan di bawah ini:</p>
		<p><a href="%s">%s</a></p>
		<p>Tautan ini akan kedaluwarsa dalam 1 jam.</p>
		<p>Jika Anda tidak merasa melakukan permintaan ini, silakan abaikan email ini.</p>
	`, user.Name, resetURL, resetURL)

	return s.notif.SendEmail(ctx, email, subject, body)
}

func (s *PasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	email, ok, err := s.tokens.GetByToken(ctx, token)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("token tidak valid atau sudah kedaluwarsa")
	}

	user, ok, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("user tidak ditemukan")
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePassword(ctx, user.ID, hash); err != nil {
		return err
	}

	// Delete token after use
	_ = s.tokens.Delete(ctx, token)

	return nil
}

func (s *PasswordResetService) generateToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
