package loginlogrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

type LoginLog struct {
	ID         string     `json:"id"`
	UserID     *string    `json:"user_id"`
	Username   string     `json:"username"`
	Role       string     `json:"role"`
	IP         string     `json:"ip"`
	UserAgent  string     `json:"user_agent"`
	LoggedInAt time.Time  `json:"logged_in_at"`
}

type ListFilter struct {
	Q      string
	Role   string
	IP     string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

func (r *Repo) Insert(ctx context.Context, in LoginLog) error {
	if r == nil || r.pool == nil {
		return nil
	}

	const q = `
INSERT INTO login_logs (user_id, username, role, ip, user_agent, logged_in_at)
VALUES ($1, $2, $3, $4, $5, COALESCE($6, now()))
`
	_, err := r.pool.Exec(ctx, q, in.UserID, in.Username, in.Role, in.IP, in.UserAgent, in.LoggedInAt)
	if err != nil {
		return fmt.Errorf("insert login log: %w", err)
	}
	return nil
}

func (r *Repo) List(ctx context.Context, f ListFilter) ([]LoginLog, int, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := make([]string, 0, 6)
	args := make([]any, 0, 8)
	argN := 1

	if f.Q != "" {
		where = append(where, fmt.Sprintf("username ILIKE $%d", argN))
		args = append(args, "%"+f.Q+"%")
		argN++
	}
	if f.Role != "" && f.Role != "all" {
		where = append(where, fmt.Sprintf("role = $%d", argN))
		args = append(args, f.Role)
		argN++
	}
	if f.IP != "" {
		where = append(where, fmt.Sprintf("ip = $%d", argN))
		args = append(args, f.IP)
		argN++
	}
	if f.From != nil {
		where = append(where, fmt.Sprintf("logged_in_at >= $%d", argN))
		args = append(args, *f.From)
		argN++
	}
	if f.To != nil {
		where = append(where, fmt.Sprintf("logged_in_at <= $%d", argN))
		args = append(args, *f.To)
		argN++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + joinAND(where)
	}

	qc := `SELECT COUNT(1) FROM login_logs ` + whereSQL
	var total int
	if err := r.pool.QueryRow(ctx, qc, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count login logs: %w", err)
	}

	q := `
SELECT id, user_id, username, role, ip, user_agent, logged_in_at
FROM login_logs
` + whereSQL + `
ORDER BY logged_in_at DESC
LIMIT $` + fmt.Sprint(argN) + ` OFFSET $` + fmt.Sprint(argN+1) + `
`
	argsList := append(append([]any{}, args...), f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, q, argsList...)
	if err != nil {
		return nil, 0, fmt.Errorf("list login logs: %w", err)
	}
	defer rows.Close()

	out := make([]LoginLog, 0, f.Limit)
	for rows.Next() {
		var it LoginLog
		if err := rows.Scan(&it.ID, &it.UserID, &it.Username, &it.Role, &it.IP, &it.UserAgent, &it.LoggedInAt); err != nil {
			return nil, 0, fmt.Errorf("scan login log: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list login logs rows: %w", err)
	}

	return out, total, nil
}

func (r *Repo) DeleteByID(ctx context.Context, id string) error {
	const q = `DELETE FROM login_logs WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete login log: %w", err)
	}
	return nil
}

func (r *Repo) ClearAll(ctx context.Context) (int64, error) {
	const q = `DELETE FROM login_logs`
	tag, err := r.pool.Exec(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("clear login logs: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *Repo) PruneOlderThan(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		days = 30
	}
	const q = `DELETE FROM login_logs WHERE logged_in_at < (now() - ($1::int || ' days')::interval)`
	tag, err := r.pool.Exec(ctx, q, days)
	if err != nil {
		return 0, fmt.Errorf("prune login logs: %w", err)
	}
	return tag.RowsAffected(), nil
}

// Small helper used by handler validation.
func IsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}

func joinAND(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " AND " + parts[i]
	}
	return out
}
