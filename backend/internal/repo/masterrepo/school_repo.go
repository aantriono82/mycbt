package masterrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type School struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	LogoURL       string    `json:"logo_url"`
	Address       string    `json:"address"`
	Phone         string    `json:"phone"`
	Email         string    `json:"email"`
	Website       string    `json:"website"`
	PrincipalName string    `json:"principal_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SchoolRepo struct {
	pool *pgxpool.Pool
}

func NewSchool(pool *pgxpool.Pool) *SchoolRepo {
	return &SchoolRepo{pool: pool}
}

func (r *SchoolRepo) GetByID(ctx context.Context, id string) (School, bool, error) {
	const q = `
SELECT id, name, COALESCE(logo_url,''), COALESCE(address,''), COALESCE(phone,''), COALESCE(email,''), COALESCE(website,''), COALESCE(principal_name,''), created_at, updated_at
FROM schools
WHERE id = $1
LIMIT 1`

	var s School
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.Name, &s.LogoURL, &s.Address, &s.Phone, &s.Email, &s.Website, &s.PrincipalName, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return School{}, false, nil
		}
		return School{}, false, fmt.Errorf("get school by id: %w", err)
	}
	return s, true, nil
}

func (r *SchoolRepo) List(ctx context.Context) ([]School, error) {
	const q = `SELECT id, name, COALESCE(logo_url,''), created_at, updated_at FROM schools ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []School
	for rows.Next() {
		var s School
		if err := rows.Scan(&s.ID, &s.Name, &s.LogoURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *SchoolRepo) Create(ctx context.Context, s School) (string, error) {
	const q = `
INSERT INTO schools (name, logo_url, address, phone, email, website, principal_name)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id`
	var id string
	err := r.pool.QueryRow(ctx, q, s.Name, s.LogoURL, s.Address, s.Phone, s.Email, s.Website, s.PrincipalName).Scan(&id)
	return id, err
}

func (r *SchoolRepo) Update(ctx context.Context, s School) error {
	const q = `
UPDATE schools
SET name = $1, logo_url = $2, address = $3, phone = $4, email = $5, website = $6, principal_name = $7, updated_at = now()
WHERE id = $8`
	_, err := r.pool.Exec(ctx, q, s.Name, s.LogoURL, s.Address, s.Phone, s.Email, s.Website, s.PrincipalName, s.ID)
	return err
}

func (r *SchoolRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM schools WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}
