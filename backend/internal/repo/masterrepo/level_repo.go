package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Level struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Kelas *int   `json:"kelas"`
}

type LevelsRepo struct{ pool *pgxpool.Pool }

func NewLevels(pool *pgxpool.Pool) *LevelsRepo { return &LevelsRepo{pool: pool} }

func (r *LevelsRepo) List(ctx context.Context) ([]Level, error) {
	const q = `SELECT id, name, kelas FROM levels ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list levels: %w", err)
	}
	defer rows.Close()

	out := []Level{}
	for rows.Next() {
		var it Level
		if err := rows.Scan(&it.ID, &it.Name, &it.Kelas); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *LevelsRepo) ListForTeacherUserID(ctx context.Context, userID string) ([]Level, bool, error) {
	const q = `
SELECT l.id::text, l.name, l.kelas
FROM levels l
JOIN teacher_levels tl ON tl.level_id = l.id
JOIN teachers t ON t.id = tl.teacher_id
WHERE t.user_id = $1
ORDER BY l.name ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, false, fmt.Errorf("list levels for teacher: %w", err)
	}
	defer rows.Close()

	out := []Level{}
	for rows.Next() {
		var it Level
		if err := rows.Scan(&it.ID, &it.Name, &it.Kelas); err != nil {
			return nil, false, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, true, rows.Err()
}

func (r *LevelsRepo) Get(ctx context.Context, id string) (Level, bool, error) {
	const q = `SELECT id, name, kelas FROM levels WHERE id = $1 LIMIT 1`
	var it Level
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.Name, &it.Kelas)
	if err != nil {
		if isNoRows(err) {
			return Level{}, false, nil
		}
		return Level{}, false, fmt.Errorf("get level: %w", err)
	}
	return it, true, nil
}

func (r *LevelsRepo) Create(ctx context.Context, name string, kelas *int) (Level, error) {
	const q = `INSERT INTO levels (name, kelas) VALUES ($1, $2) RETURNING id`
	var id string
	if err := r.pool.QueryRow(ctx, q, name, kelas).Scan(&id); err != nil {
		return Level{}, fmt.Errorf("create level: %w", err)
	}
	return Level{ID: id, Name: name, Kelas: kelas}, nil
}

func (r *LevelsRepo) Update(ctx context.Context, id, name string, kelas *int) (Level, bool, error) {
	const q = `UPDATE levels SET name = $2, kelas = $3, updated_at = now() WHERE id = $1 RETURNING id, name, kelas`
	var it Level
	err := r.pool.QueryRow(ctx, q, id, name, kelas).Scan(&it.ID, &it.Name, &it.Kelas)
	if err != nil {
		if isNoRows(err) {
			return Level{}, false, nil
		}
		return Level{}, false, fmt.Errorf("update level: %w", err)
	}
	return it, true, nil
}

func (r *LevelsRepo) Delete(ctx context.Context, id string) (bool, error) {
	const q = `DELETE FROM levels WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete level: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}
