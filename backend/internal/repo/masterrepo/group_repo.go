package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GroupsRepo struct{ pool *pgxpool.Pool }

func NewGroups(pool *pgxpool.Pool) *GroupsRepo { return &GroupsRepo{pool: pool} }

func (r *GroupsRepo) List(ctx context.Context) ([]Group, error) {
	const q = `SELECT id, name FROM groups ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	out := []Group{}
	for rows.Next() {
		var it Group
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *GroupsRepo) ListForTeacherUserID(ctx context.Context, userID string) ([]Group, bool, error) {
	const q = `
SELECT g.id::text, g.name
FROM groups g
JOIN teacher_groups tg ON tg.group_id = g.id
JOIN teachers t ON t.id = tg.teacher_id
WHERE t.user_id = $1
ORDER BY g.name ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, false, fmt.Errorf("list groups for teacher: %w", err)
	}
	defer rows.Close()

	out := []Group{}
	for rows.Next() {
		var it Group
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			return nil, false, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, true, rows.Err()
}

func (r *GroupsRepo) Get(ctx context.Context, id string) (Group, bool, error) {
	const q = `SELECT id, name FROM groups WHERE id = $1 LIMIT 1`
	var it Group
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.Name)
	if err != nil {
		if isNoRows(err) {
			return Group{}, false, nil
		}
		return Group{}, false, fmt.Errorf("get group: %w", err)
	}
	return it, true, nil
}

func (r *GroupsRepo) Create(ctx context.Context, name string) (Group, error) {
	const q = `INSERT INTO groups (name) VALUES ($1) RETURNING id`
	var id string
	if err := r.pool.QueryRow(ctx, q, name).Scan(&id); err != nil {
		return Group{}, fmt.Errorf("create group: %w", err)
	}
	return Group{ID: id, Name: name}, nil
}

func (r *GroupsRepo) Update(ctx context.Context, id, name string) (Group, bool, error) {
	const q = `UPDATE groups SET name = $2, updated_at = now() WHERE id = $1 RETURNING id, name`
	var it Group
	err := r.pool.QueryRow(ctx, q, id, name).Scan(&it.ID, &it.Name)
	if err != nil {
		if isNoRows(err) {
			return Group{}, false, nil
		}
		return Group{}, false, fmt.Errorf("update group: %w", err)
	}
	return it, true, nil
}

func (r *GroupsRepo) Delete(ctx context.Context, id string) (bool, error) {
	const q = `DELETE FROM groups WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete group: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}
