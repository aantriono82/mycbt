package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Program struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProgramsRepo struct{ pool *pgxpool.Pool }

func NewPrograms(pool *pgxpool.Pool) *ProgramsRepo { return &ProgramsRepo{pool: pool} }

func (r *ProgramsRepo) List(ctx context.Context) ([]Program, error) {
	const q = `SELECT id, COALESCE(code,''), name FROM programs ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	defer rows.Close()

	out := []Program{}
	for rows.Next() {
		var p Program
		if err := rows.Scan(&p.ID, &p.Code, &p.Name); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProgramsRepo) Get(ctx context.Context, id string) (Program, bool, error) {
	const q = `SELECT id, COALESCE(code,''), name FROM programs WHERE id = $1 LIMIT 1`
	var p Program
	err := r.pool.QueryRow(ctx, q, id).Scan(&p.ID, &p.Code, &p.Name)
	if err != nil {
		if isNoRows(err) {
			return Program{}, false, nil
		}
		return Program{}, false, fmt.Errorf("get program: %w", err)
	}
	return p, true, nil
}

func (r *ProgramsRepo) Create(ctx context.Context, code, name string) (Program, error) {
	const q = `INSERT INTO programs (code, name) VALUES (NULLIF($1,''), $2) RETURNING id`
	var id string
	if err := r.pool.QueryRow(ctx, q, code, name).Scan(&id); err != nil {
		return Program{}, fmt.Errorf("create program: %w", err)
	}
	return Program{ID: id, Code: code, Name: name}, nil
}

func (r *ProgramsRepo) Update(ctx context.Context, id, code, name string) (Program, bool, error) {
	const q = `
UPDATE programs
SET code = NULLIF($2,''), name = $3, updated_at = now()
WHERE id = $1
RETURNING id, COALESCE(code,''), name`
	var p Program
	err := r.pool.QueryRow(ctx, q, id, code, name).Scan(&p.ID, &p.Code, &p.Name)
	if err != nil {
		if isNoRows(err) {
			return Program{}, false, nil
		}
		return Program{}, false, fmt.Errorf("update program: %w", err)
	}
	return p, true, nil
}

func (r *ProgramsRepo) Delete(ctx context.Context, id string) (bool, error) {
	const q = `DELETE FROM programs WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete program: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}
