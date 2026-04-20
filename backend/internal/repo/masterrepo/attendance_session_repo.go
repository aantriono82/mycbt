package masterrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendanceSession struct {
	ID           string    `json:"id"`
	ExamID       string    `json:"exam_id"`
	ExamTitle    string    `json:"exam_title,omitempty"`
	Token        string    `json:"token"`
	Lat          *float64  `json:"lat"`
	Lon          *float64  `json:"lon"`
	RadiusMeters int       `json:"radius_meters"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type AttendanceSessionsRepo struct {
	pool *pgxpool.Pool
}

func NewAttendanceSessions(pool *pgxpool.Pool) *AttendanceSessionsRepo {
	return &AttendanceSessionsRepo{pool: pool}
}

func (r *AttendanceSessionsRepo) Create(ctx context.Context, examID, token string, lat, lon *float64, radius int, expiresAt time.Time) (AttendanceSession, error) {
	const q = `
INSERT INTO attendance_sessions (exam_id, token, lat, lon, radius_meters, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

	var id string
	var createdAt time.Time
	if err := r.pool.QueryRow(ctx, q, examID, token, lat, lon, radius, expiresAt).Scan(&id, &createdAt); err != nil {
		return AttendanceSession{}, fmt.Errorf("create attendance session: %w", err)
	}

	return AttendanceSession{
		ID:           id,
		ExamID:       examID,
		Token:        token,
		Lat:          lat,
		Lon:          lon,
		RadiusMeters: radius,
		ExpiresAt:    expiresAt,
		CreatedAt:    createdAt,
	}, nil
}

func (r *AttendanceSessionsRepo) GetByToken(ctx context.Context, token string) (AttendanceSession, bool, error) {
	const q = `
SELECT id, exam_id, token, lat, lon, radius_meters, expires_at, created_at
FROM attendance_sessions
WHERE token = $1`

	var it AttendanceSession
	if err := r.pool.QueryRow(ctx, q, token).Scan(
		&it.ID, &it.ExamID, &it.Token, &it.Lat, &it.Lon, &it.RadiusMeters, &it.ExpiresAt, &it.CreatedAt,
	); err != nil {
		if isNoRows(err) {
			return AttendanceSession{}, false, nil
		}
		return AttendanceSession{}, false, fmt.Errorf("get attendance session by token: %w", err)
	}
	return it, true, nil
}

func (r *AttendanceSessionsRepo) ListActiveByExam(ctx context.Context, examID string) ([]AttendanceSession, error) {
	const q = `
SELECT id, exam_id, token, lat, lon, radius_meters, expires_at, created_at
FROM attendance_sessions
WHERE exam_id = $1 AND expires_at > now()
ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, examID)
	if err != nil {
		return nil, fmt.Errorf("list active attendance sessions: %w", err)
	}
	defer rows.Close()

	var out []AttendanceSession
	for rows.Next() {
		var it AttendanceSession
		if err := rows.Scan(
			&it.ID, &it.ExamID, &it.Token, &it.Lat, &it.Lon, &it.RadiusMeters, &it.ExpiresAt, &it.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan attendance session: %w", err)
		}
		out = append(out, it)
	}
	return out, nil
}

func (r *AttendanceSessionsRepo) DeleteExpired(ctx context.Context) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM attendance_sessions WHERE expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
