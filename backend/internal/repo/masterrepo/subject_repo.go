package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Subject struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type SubjectsRepo struct{ pool *pgxpool.Pool }

func NewSubjects(pool *pgxpool.Pool) *SubjectsRepo { return &SubjectsRepo{pool: pool} }

func (r *SubjectsRepo) List(ctx context.Context) ([]Subject, error) {
	const q = `SELECT id, COALESCE(code,''), name FROM subjects ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	out := []Subject{}
	for rows.Next() {
		var it Subject
		if err := rows.Scan(&it.ID, &it.Code, &it.Name); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *SubjectsRepo) ListForTeacherUserID(ctx context.Context, userID string) (items []Subject, ok bool, err error) {
	// Ensure teacher exists even if they have zero subject mappings.
	var teacherID string
	if err := r.pool.QueryRow(ctx, `SELECT id::text FROM teachers WHERE user_id = $1 LIMIT 1`, userID).Scan(&teacherID); err != nil {
		if isNoRows(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("teacher lookup: %w", err)
	}

	const q = `
SELECT s.id, COALESCE(s.code,''), s.name
FROM teacher_subjects ts
JOIN subjects s ON s.id = ts.subject_id
WHERE ts.teacher_id = $1
ORDER BY s.name ASC`
	rows, err := r.pool.Query(ctx, q, teacherID)
	if err != nil {
		return nil, true, fmt.Errorf("list teacher subjects: %w", err)
	}
	defer rows.Close()

	out := []Subject{}
	for rows.Next() {
		var it Subject
		if err := rows.Scan(&it.ID, &it.Code, &it.Name); err != nil {
			return nil, true, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, true, err
	}
	return out, true, nil
}

func (r *SubjectsRepo) Get(ctx context.Context, id string) (Subject, bool, error) {
	const q = `SELECT id, COALESCE(code,''), name FROM subjects WHERE id = $1 LIMIT 1`
	var it Subject
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.Code, &it.Name)
	if err != nil {
		if isNoRows(err) {
			return Subject{}, false, nil
		}
		return Subject{}, false, fmt.Errorf("get subject: %w", err)
	}
	return it, true, nil
}

func (r *SubjectsRepo) Create(ctx context.Context, code, name string) (Subject, error) {
	const q = `INSERT INTO subjects (code, name) VALUES (NULLIF($1,''), $2) RETURNING id`
	var id string
	if err := r.pool.QueryRow(ctx, q, code, name).Scan(&id); err != nil {
		return Subject{}, fmt.Errorf("create subject: %w", err)
	}
	return Subject{ID: id, Code: code, Name: name}, nil
}

func (r *SubjectsRepo) Update(ctx context.Context, id, code, name string) (Subject, bool, error) {
	const q = `
UPDATE subjects
SET code = NULLIF($2,''), name = $3, updated_at = now()
WHERE id = $1
RETURNING id, COALESCE(code,''), name`
	var it Subject
	err := r.pool.QueryRow(ctx, q, id, code, name).Scan(&it.ID, &it.Code, &it.Name)
	if err != nil {
		if isNoRows(err) {
			return Subject{}, false, nil
		}
		return Subject{}, false, fmt.Errorf("update subject: %w", err)
	}
	return it, true, nil
}

func (r *SubjectsRepo) Delete(ctx context.Context, id string) (bool, error) {
	const q = `DELETE FROM subjects WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete subject: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

func (r *SubjectsRepo) MissingIDs(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Build a simple IN list with ANY($1).
	const q = `SELECT id::text FROM subjects WHERE id = ANY($1::uuid[])`
	rows, err := r.pool.Query(ctx, q, ids)
	if err != nil {
		return nil, fmt.Errorf("subjects missing ids query: %w", err)
	}
	defer rows.Close()

	found := map[string]struct{}{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		found[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	missing := make([]string, 0)
	for _, id := range ids {
		if _, ok := found[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing, nil
}
