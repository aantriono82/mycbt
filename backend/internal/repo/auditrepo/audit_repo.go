package auditrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

type CreateLogInput struct {
	RequestID  string
	UserID     string
	Role       string
	Method     string
	Path       string
	Query      string
	StatusCode int
	IP         string
	UserAgent  string
	Payload    map[string]any
}

func (r *Repo) Create(ctx context.Context, in CreateLogInput) error {
	var payloadRaw []byte
	if in.Payload != nil {
		raw, err := json.Marshal(in.Payload)
		if err != nil {
			return fmt.Errorf("marshal audit payload: %w", err)
		}
		payloadRaw = raw
	}

	const q = `
INSERT INTO audit_logs (
  request_id, user_id, role, method, path, query, status_code, ip, user_agent, payload_json
)
VALUES (
  $1, NULLIF($2,'')::uuid, NULLIF($3,''), $4, $5, NULLIF($6,''), $7, NULLIF($8,''), NULLIF($9,''), $10::jsonb
)`

	payloadStr := ""
	if len(payloadRaw) > 0 {
		payloadStr = string(payloadRaw)
	}
	if _, err := r.pool.Exec(
		ctx, q,
		in.RequestID,
		in.UserID,
		in.Role,
		in.Method,
		in.Path,
		in.Query,
		in.StatusCode,
		in.IP,
		in.UserAgent,
		payloadStr,
	); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}
