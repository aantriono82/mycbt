package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeacherSubject struct {
	SubjectID string `json:"subject_id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
}

type TeacherSubjectsRepo struct{ pool *pgxpool.Pool }

func NewTeacherSubjects(pool *pgxpool.Pool) *TeacherSubjectsRepo {
	return &TeacherSubjectsRepo{pool: pool}
}

func (r *TeacherSubjectsRepo) ListByTeacherID(ctx context.Context, teacherID string) ([]TeacherSubject, error) {
	const q = `
SELECT s.id, COALESCE(s.code,''), s.name
FROM teacher_subjects ts
JOIN subjects s ON s.id = ts.subject_id
WHERE ts.teacher_id = $1
ORDER BY s.name ASC`

	rows, err := r.pool.Query(ctx, q, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list teacher subjects: %w", err)
	}
	defer rows.Close()

	out := []TeacherSubject{}
	for rows.Next() {
		var it TeacherSubject
		if err := rows.Scan(&it.SubjectID, &it.Code, &it.Name); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *TeacherSubjectsRepo) Has(ctx context.Context, teacherID, subjectID string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2)`
	var ok bool
	if err := r.pool.QueryRow(ctx, q, teacherID, subjectID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check teacher subject: %w", err)
	}
	return ok, nil
}

// Replace replaces subject mappings for a teacher in a transaction.
func (r *TeacherSubjectsRepo) Replace(ctx context.Context, teacherID string, subjectIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM teacher_subjects WHERE teacher_id = $1`, teacherID); err != nil {
		return fmt.Errorf("clear: %w", err)
	}

	if len(subjectIDs) > 0 {
		const ins = `INSERT INTO teacher_subjects (teacher_id, subject_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
		for _, sid := range subjectIDs {
			if sid == "" {
				continue
			}
			if _, err := tx.Exec(ctx, ins, teacherID, sid); err != nil {
				return fmt.Errorf("insert: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
