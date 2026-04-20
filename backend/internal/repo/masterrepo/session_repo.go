package masterrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type SessionsRepo struct{ pool *pgxpool.Pool }

func NewSessions(pool *pgxpool.Pool) *SessionsRepo { return &SessionsRepo{pool: pool} }

func (r *SessionsRepo) List(ctx context.Context) ([]Session, error) {
	const q = `SELECT id, name, COALESCE(to_char(start_time, 'HH24:MI'), ''), COALESCE(to_char(end_time, 'HH24:MI'), '') FROM sessions ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	out := []Session{}
	for rows.Next() {
		var it Session
		if err := rows.Scan(&it.ID, &it.Name, &it.StartTime, &it.EndTime); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *SessionsRepo) Get(ctx context.Context, id string) (Session, bool, error) {
	const q = `SELECT id, name, COALESCE(to_char(start_time, 'HH24:MI'), ''), COALESCE(to_char(end_time, 'HH24:MI'), '') FROM sessions WHERE id = $1 LIMIT 1`
	var it Session
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.Name, &it.StartTime, &it.EndTime)
	if err != nil {
		if isNoRows(err) {
			return Session{}, false, nil
		}
		return Session{}, false, fmt.Errorf("get session: %w", err)
	}
	return it, true, nil
}

func (r *SessionsRepo) Create(ctx context.Context, name, startTime, endTime string) (Session, error) {
	const q = `INSERT INTO sessions (name, start_time, end_time) VALUES ($1, NULLIF($2, '')::TIME, NULLIF($3, '')::TIME) RETURNING id`
	var id string
	if err := r.pool.QueryRow(ctx, q, name, startTime, endTime).Scan(&id); err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return Session{ID: id, Name: name, StartTime: startTime, EndTime: endTime}, nil
}

func (r *SessionsRepo) Update(ctx context.Context, id, name, startTime, endTime string) (Session, bool, error) {
	const q = `UPDATE sessions SET name = $2, start_time = NULLIF($3, '')::TIME, end_time = NULLIF($4, '')::TIME, updated_at = now() WHERE id = $1 RETURNING id, name, COALESCE(to_char(start_time, 'HH24:MI'), ''), COALESCE(to_char(end_time, 'HH24:MI'), '')`
	var it Session
	err := r.pool.QueryRow(ctx, q, id, name, startTime, endTime).Scan(&it.ID, &it.Name, &it.StartTime, &it.EndTime)
	if err != nil {
		if isNoRows(err) {
			return Session{}, false, nil
		}
		return Session{}, false, fmt.Errorf("update session: %w", err)
	}
	return it, true, nil
}

func (r *SessionsRepo) Delete(ctx context.Context, id string) (bool, error) {
	const q = `DELETE FROM sessions WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete session: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}
