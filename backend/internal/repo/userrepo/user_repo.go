package userrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"atigacbt/backend/internal/model"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) GetByUsername(ctx context.Context, username string) (model.User, bool, error) {
	const q = `
SELECT id, username, password_hash, role, name, COALESCE(email,''), COALESCE(photo_url,''), COALESCE(google_id,''), is_active, created_at, updated_at
FROM users
WHERE username = $1
LIMIT 1`

	var u model.User
	err := r.pool.QueryRow(ctx, q, username).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.Name,
		&u.Email,
		&u.PhotoURL,
		&u.GoogleID,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return model.User{}, false, nil
		}
		return model.User{}, false, fmt.Errorf("get user by username: %w", err)
	}
	return u, true, nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (model.User, bool, error) {
	const q = `
SELECT id, username, password_hash, role, name, COALESCE(email,''), COALESCE(photo_url,''), COALESCE(google_id,''), is_active, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1`

	var u model.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.Name,
		&u.Email,
		&u.PhotoURL,
		&u.GoogleID,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return model.User{}, false, nil
		}
		return model.User{}, false, fmt.Errorf("get user by id: %w", err)
	}
	return u, true, nil
}

func (r *Repo) Create(ctx context.Context, u model.User) (string, error) {
	const q = `
INSERT INTO users (username, password_hash, role, name, email, google_id, is_active)
VALUES ($1, $2, $3, $4, NULLIF($5,''), NULLIF($6,''), $7)
RETURNING id`

	var id string
	if err := r.pool.QueryRow(ctx, q, u.Username, u.PasswordHash, u.Role, u.Name, u.Email, u.GoogleID, u.IsActive).Scan(&id); err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

func (r *Repo) GetByGoogleID(ctx context.Context, googleID string) (model.User, bool, error) {
	const q = `
SELECT id, username, password_hash, role, name, COALESCE(email,''), COALESCE(photo_url,''), COALESCE(google_id,''), is_active, created_at, updated_at
FROM users
WHERE google_id = $1
LIMIT 1`

	var u model.User
	err := r.pool.QueryRow(ctx, q, googleID).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.Name,
		&u.Email,
		&u.PhotoURL,
		&u.GoogleID,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return model.User{}, false, nil
		}
		return model.User{}, false, fmt.Errorf("get user by google id: %w", err)
	}
	return u, true, nil
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (model.User, bool, error) {
	const q = `
SELECT id, username, password_hash, role, name, COALESCE(email,''), COALESCE(photo_url,''), COALESCE(google_id,''), is_active, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1`

	var u model.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Name, &u.Email, &u.PhotoURL, &u.GoogleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return model.User{}, false, nil
		}
		return model.User{}, false, fmt.Errorf("get user by email: %w", err)
	}
	return u, true, nil
}

func (r *Repo) UpdateGoogleID(ctx context.Context, id string, googleID string) error {
	const q = `UPDATE users SET google_id = $1, updated_at = now() WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, googleID, id)
	return err
}

func (r *Repo) UpdatePhoto(ctx context.Context, id string, photoURL string) error {
	const q = `UPDATE users SET photo_url = $1, updated_at = now() WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, photoURL, id)
	return err
}
func (r *Repo) UpdateRole(ctx context.Context, id string, role string) error {
	const q = `UPDATE users SET role = $1, updated_at = now() WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, role, id)
	return err
}

func (r *Repo) UpdatePassword(ctx context.Context, id string, hash string) error {
	const q = `UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, hash, id)
	return err
}

func (r *Repo) UpdateProfile(ctx context.Context, id, name, email string) error {
	const q = `UPDATE users SET name = $1, email = NULLIF($2,''), updated_at = now() WHERE id = $3`
	_, err := r.pool.Exec(ctx, q, name, email, id)
	return err
}
