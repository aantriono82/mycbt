package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeacherGroup struct {
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
}

type TeacherGroupsRepo struct{ pool *pgxpool.Pool }

func NewTeacherGroups(pool *pgxpool.Pool) *TeacherGroupsRepo {
	return &TeacherGroupsRepo{pool: pool}
}

func (r *TeacherGroupsRepo) ListByTeacherID(ctx context.Context, teacherID string) ([]TeacherGroup, error) {
	const q = `
SELECT g.id, g.name
FROM teacher_groups tg
JOIN groups g ON g.id = tg.group_id
WHERE tg.teacher_id = $1
ORDER BY g.name ASC`

	rows, err := r.pool.Query(ctx, q, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list teacher groups: %w", err)
	}
	defer rows.Close()

	out := []TeacherGroup{}
	for rows.Next() {
		var it TeacherGroup
		if err := rows.Scan(&it.GroupID, &it.Name); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *TeacherGroupsRepo) Has(ctx context.Context, teacherID, groupID string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM teacher_groups WHERE teacher_id = $1 AND group_id = $2)`
	var ok bool
	if err := r.pool.QueryRow(ctx, q, teacherID, groupID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check teacher group: %w", err)
	}
	return ok, nil
}

func (r *TeacherGroupsRepo) Replace(ctx context.Context, teacherID string, groupIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM teacher_groups WHERE teacher_id = $1`, teacherID); err != nil {
		return fmt.Errorf("clear: %w", err)
	}

	if len(groupIDs) > 0 {
		const ins = `INSERT INTO teacher_groups (teacher_id, group_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
		for _, gid := range groupIDs {
			if gid == "" {
				continue
			}
			if _, err := tx.Exec(ctx, ins, teacherID, gid); err != nil {
				return fmt.Errorf("insert: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
