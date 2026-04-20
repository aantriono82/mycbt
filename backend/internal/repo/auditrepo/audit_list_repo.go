package auditrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AuditLog struct {
	ID         string                 `json:"id"`
	RequestID  string                 `json:"request_id"`
	UserID     *string                `json:"user_id"`
	Username   string                 `json:"username"`
	Name       string                 `json:"name"`
	Role       *string                `json:"role"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Query      *string                `json:"query"`
	StatusCode int                    `json:"status_code"`
	IP         *string                `json:"ip"`
	UserAgent  *string                `json:"user_agent"`
	Payload    map[string]any         `json:"payload_json"`
	CreatedAt  time.Time              `json:"created_at"`
}

type ListFilter struct {
	Q       string
	Role    string
	Method  string
	Path    string
	Status  *int
	From    *time.Time
	To      *time.Time
	Limit   int
	Offset  int
}

func (r *Repo) List(ctx context.Context, f ListFilter) ([]AuditLog, int, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := make([]string, 0, 8)
	args := make([]any, 0, 12)
	argN := 1

	if f.Q != "" {
		// Search in username + path + request_id.
		where = append(where, fmt.Sprintf("(COALESCE(u.username,'') ILIKE $%d OR a.path ILIKE $%d OR a.request_id ILIKE $%d)", argN, argN, argN))
		args = append(args, "%"+f.Q+"%")
		argN++
	}
	if f.Role != "" && f.Role != "all" {
		where = append(where, fmt.Sprintf("COALESCE(a.role,'') = $%d", argN))
		args = append(args, f.Role)
		argN++
	}
	if f.Method != "" && f.Method != "all" {
		where = append(where, fmt.Sprintf("a.method = $%d", argN))
		args = append(args, f.Method)
		argN++
	}
	if f.Path != "" {
		where = append(where, fmt.Sprintf("a.path ILIKE $%d", argN))
		args = append(args, "%"+f.Path+"%")
		argN++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("a.status_code = $%d", argN))
		args = append(args, *f.Status)
		argN++
	}
	if f.From != nil {
		where = append(where, fmt.Sprintf("a.created_at >= $%d", argN))
		args = append(args, *f.From)
		argN++
	}
	if f.To != nil {
		where = append(where, fmt.Sprintf("a.created_at <= $%d", argN))
		args = append(args, *f.To)
		argN++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + joinAND(where)
	}

	qc := `
SELECT COUNT(1)
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
` + whereSQL

	var total int
	if err := r.pool.QueryRow(ctx, qc, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	q := `
SELECT
  a.id, a.request_id, a.user_id, COALESCE(u.username,''), COALESCE(u.name,''),
  a.role, a.method, a.path, a.query, a.status_code, a.ip, a.user_agent,
  COALESCE(a.payload_json, '{}'::jsonb), a.created_at
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
` + whereSQL + `
ORDER BY a.created_at DESC
LIMIT $` + fmt.Sprint(argN) + ` OFFSET $` + fmt.Sprint(argN+1) + `
`

	argsList := append(append([]any{}, args...), f.Limit, f.Offset)
	rows, err := r.pool.Query(ctx, q, argsList...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	out := make([]AuditLog, 0, f.Limit)
	for rows.Next() {
		var it AuditLog
		var payloadRaw []byte
		if err := rows.Scan(
			&it.ID, &it.RequestID, &it.UserID, &it.Username, &it.Name,
			&it.Role, &it.Method, &it.Path, &it.Query, &it.StatusCode, &it.IP, &it.UserAgent,
			&payloadRaw, &it.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		it.Payload = map[string]any{}
		_ = json.Unmarshal(payloadRaw, &it.Payload)
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list audit logs rows: %w", err)
	}

	return out, total, nil
}

func (r *Repo) DeleteByID(ctx context.Context, id string) error {
	const q = `DELETE FROM audit_logs WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, id); err != nil {
		return fmt.Errorf("delete audit log: %w", err)
	}
	return nil
}

func (r *Repo) ClearAll(ctx context.Context) (int64, error) {
	const q = `DELETE FROM audit_logs`
	tag, err := r.pool.Exec(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("clear audit logs: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *Repo) PruneOlderThan(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		days = 30
	}
	const q = `DELETE FROM audit_logs WHERE created_at < (now() - ($1::int || ' days')::interval)`
	tag, err := r.pool.Exec(ctx, q, days)
	if err != nil {
		return 0, fmt.Errorf("prune audit logs: %w", err)
	}
	return tag.RowsAffected(), nil
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
