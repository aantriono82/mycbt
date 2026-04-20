package userrepo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetRepo struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepo(pool *pgxpool.Pool) *PasswordResetRepo {
	return &PasswordResetRepo{pool: pool}
}

func (r *PasswordResetRepo) Create(ctx context.Context, email, token string, expiresAt time.Time) error {
	// Delete any existing tokens for this email first
	_, err := r.pool.Exec(ctx, "DELETE FROM password_reset_tokens WHERE email = $1", email)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, "INSERT INTO password_reset_tokens (email, token, expires_at) VALUES ($1, $2, $3)", email, token, expiresAt)
	return err
}

func (r *PasswordResetRepo) GetByToken(ctx context.Context, token string) (string, bool, error) {
	var email string
	var expiresAt time.Time
	err := r.pool.QueryRow(ctx, "SELECT email, expires_at FROM password_reset_tokens WHERE token = $1", token).Scan(&email, &expiresAt)
	if err != nil {
		return "", false, nil
	}

	if time.Now().UTC().After(expiresAt) {
		return "", false, nil
	}

	return email, true, nil
}

func (r *PasswordResetRepo) Delete(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM password_reset_tokens WHERE token = $1", token)
	return err
}
