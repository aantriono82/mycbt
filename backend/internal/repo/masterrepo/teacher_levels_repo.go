package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeacherLevel struct {
	LevelID string `json:"level_id"`
	Name    string `json:"name"`
}

type TeacherLevelsRepo struct{ pool *pgxpool.Pool }

func NewTeacherLevels(pool *pgxpool.Pool) *TeacherLevelsRepo {
	return &TeacherLevelsRepo{pool: pool}
}

func (r *TeacherLevelsRepo) ListByTeacherID(ctx context.Context, teacherID string) ([]TeacherLevel, error) {
	const q = `
SELECT l.id, l.name
FROM teacher_levels tl
JOIN levels l ON l.id = tl.level_id
WHERE tl.teacher_id = $1
ORDER BY l.name ASC`

	rows, err := r.pool.Query(ctx, q, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list teacher levels: %w", err)
	}
	defer rows.Close()

	out := []TeacherLevel{}
	for rows.Next() {
		var it TeacherLevel
		if err := rows.Scan(&it.LevelID, &it.Name); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *TeacherLevelsRepo) Has(ctx context.Context, teacherID, levelID string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM teacher_levels WHERE teacher_id = $1 AND level_id = $2)`
	var ok bool
	if err := r.pool.QueryRow(ctx, q, teacherID, levelID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check teacher level: %w", err)
	}
	return ok, nil
}

func (r *TeacherLevelsRepo) Replace(ctx context.Context, teacherID string, levelIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM teacher_levels WHERE teacher_id = $1`, teacherID); err != nil {
		return fmt.Errorf("clear: %w", err)
	}

	if len(levelIDs) > 0 {
		const ins = `INSERT INTO teacher_levels (teacher_id, level_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
		for _, lid := range levelIDs {
			if lid == "" {
				continue
			}
			if _, err := tx.Exec(ctx, ins, teacherID, lid); err != nil {
				return fmt.Errorf("insert: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
