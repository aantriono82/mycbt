package ltirepo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Platform struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Issuer         string    `json:"issuer"`
	ClientID       string    `json:"client_id"`
	DeploymentID   string    `json:"deployment_id"`
	OIDCAuthURL    string    `json:"oidc_auth_url"`
	OIDCTokenURL   string    `json:"oidc_token_url"`
	JWKSURL        string    `json:"jwks_url"`
	ToolPrivateKey string    `json:"tool_private_key"`
	ToolPublicKey  string    `json:"tool_public_key"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) CreatePlatform(ctx context.Context, p Platform) (Platform, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO lti_platforms (name, issuer, client_id, deployment_id, oidc_auth_url, oidc_token_url, jwks_url, tool_private_key, tool_public_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`, p.Name, p.Issuer, p.ClientID, p.DeploymentID, p.OIDCAuthURL, p.OIDCTokenURL, p.JWKSURL, p.ToolPrivateKey, p.ToolPublicKey).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repo) GetPlatformByIssuer(ctx context.Context, issuer string) (Platform, bool, error) {
	var p Platform
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, issuer, client_id, deployment_id, oidc_auth_url, oidc_token_url, jwks_url, tool_private_key, tool_public_key, created_at, updated_at
		FROM lti_platforms
		WHERE issuer = $1
	`, issuer).Scan(&p.ID, &p.Name, &p.Issuer, &p.ClientID, &p.DeploymentID, &p.OIDCAuthURL, &p.OIDCTokenURL, &p.JWKSURL, &p.ToolPrivateKey, &p.ToolPublicKey, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return p, false, nil
	}
	return p, true, err
}

func (r *Repo) GetPlatformByID(ctx context.Context, id string) (Platform, bool, error) {
	var p Platform
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, issuer, client_id, deployment_id, oidc_auth_url, oidc_token_url, jwks_url, tool_private_key, tool_public_key, created_at, updated_at
		FROM lti_platforms
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Issuer, &p.ClientID, &p.DeploymentID, &p.OIDCAuthURL, &p.OIDCTokenURL, &p.JWKSURL, &p.ToolPrivateKey, &p.ToolPublicKey, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return p, false, nil
	}
	return p, true, err
}

func (r *Repo) ListPlatforms(ctx context.Context) ([]Platform, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, issuer, client_id, created_at FROM lti_platforms ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Platform
	for rows.Next() {
		var p Platform
		if err := rows.Scan(&p.ID, &p.Name, &p.Issuer, &p.ClientID, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *Repo) FindUserByLTI(ctx context.Context, platformID, ltiSub string) (string, bool, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		SELECT local_user_id FROM lti_users WHERE platform_id = $1 AND lti_sub = $2
	`, platformID, ltiSub).Scan(&userID)
	if err == pgx.ErrNoRows {
		return "", false, nil
	}
	return userID, true, err
}

func (r *Repo) LinkUser(ctx context.Context, platformID, ltiSub, localUserID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO lti_users (platform_id, lti_sub, local_user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (platform_id, lti_sub) DO UPDATE SET local_user_id = $3
	`, platformID, ltiSub, localUserID)
	return err
}

func (r *Repo) StoreNonce(ctx context.Context, nonce string, ttl time.Duration) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO lti_nonces (nonce, expires_at) VALUES ($1, $2)`, nonce, time.Now().Add(ttl))
	return err
}

func (r *Repo) UseNonce(ctx context.Context, nonce string) (bool, error) {
	res, err := r.pool.Exec(ctx, `DELETE FROM lti_nonces WHERE nonce = $1 AND expires_at > now()`, nonce)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() > 0, nil
}

type LTISession struct {
	ID           string
	PlatformID   string
	LocalUserID  string
	MessageType  string
	ReturnURL    string
	Data         string
	DeploymentID string
	ExpiresAt    time.Time
}

func (r *Repo) CreateSession(ctx context.Context, s LTISession) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO lti_sessions (platform_id, local_user_id, message_type, return_url, data, deployment_id, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text
	`, s.PlatformID, s.LocalUserID, s.MessageType, s.ReturnURL, s.Data, s.DeploymentID, s.ExpiresAt).Scan(&id)
	return id, err
}

func (r *Repo) GetSession(ctx context.Context, id string) (LTISession, bool, error) {
	var s LTISession
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, platform_id::text, local_user_id::text, message_type, COALESCE(return_url,''), COALESCE(data,''), COALESCE(deployment_id,''), expires_at
		FROM lti_sessions
		WHERE id = $1 AND expires_at > now()
	`, id).Scan(&s.ID, &s.PlatformID, &s.LocalUserID, &s.MessageType, &s.ReturnURL, &s.Data, &s.DeploymentID, &s.ExpiresAt)
	if err == pgx.ErrNoRows {
		return s, false, nil
	}
	return s, true, err
}

type AGSLaunchContext struct {
	PlatformID     string
	DeploymentID   string
	ResourceLinkID string
	ExamID         string
	LocalUserID    string
	LTISub         string
	LineItemURL    string
	LineItemsURL   string
	ScopeText      string
}

func (r *Repo) UpsertAGSLaunchContext(ctx context.Context, launch AGSLaunchContext) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO lti_ags_launches (
			platform_id,
			deployment_id,
			resource_link_id,
			exam_id,
			local_user_id,
			lti_sub,
			lineitem_url,
			lineitems_url,
			scope_text
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (platform_id, deployment_id, resource_link_id, local_user_id)
		DO UPDATE SET
			exam_id = EXCLUDED.exam_id,
			lti_sub = EXCLUDED.lti_sub,
			lineitem_url = EXCLUDED.lineitem_url,
			lineitems_url = EXCLUDED.lineitems_url,
			scope_text = EXCLUDED.scope_text,
			updated_at = now()
	`, launch.PlatformID, launch.DeploymentID, launch.ResourceLinkID, launch.ExamID, launch.LocalUserID, launch.LTISub, launch.LineItemURL, launch.LineItemsURL, launch.ScopeText)
	if err != nil {
		return fmt.Errorf("upsert ags launch context: %w", err)
	}
	return nil
}

type AGSScoreTarget struct {
	SessionID     string
	SessionStatus string
	FinishedAt    time.Time
	LTISub        string
	LineItemURL   string
	ScopeText     string
	Platform      Platform
}

func (r *Repo) ListAGSScoreTargets(ctx context.Context, examID string) ([]AGSScoreTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT es.id::text,
		       es.status,
		       COALESCE(es.finished_at, es.updated_at),
		       lac.lti_sub,
		       lac.lineitem_url,
		       COALESCE(lac.scope_text, ''),
		       p.id::text,
		       p.name,
		       p.issuer,
		       p.client_id,
		       p.deployment_id,
		       p.oidc_auth_url,
		       p.oidc_token_url,
		       p.jwks_url,
		       p.tool_private_key,
		       p.tool_public_key,
		       p.created_at,
		       p.updated_at
		FROM exam_sessions es
		JOIN students st ON st.id = es.student_id
		JOIN lti_ags_launches lac
		  ON lac.exam_id = es.exam_id
		 AND lac.local_user_id = st.user_id
		JOIN lti_platforms p ON p.id = lac.platform_id
		WHERE es.exam_id = $1
		  AND es.status IN ('submitted', 'forced')
		  AND COALESCE(lac.lineitem_url, '') <> ''
		ORDER BY es.finished_at DESC NULLS LAST, es.updated_at DESC
	`, examID)
	if err != nil {
		return nil, fmt.Errorf("list ags score targets: %w", err)
	}
	defer rows.Close()

	out := make([]AGSScoreTarget, 0)
	for rows.Next() {
		var item AGSScoreTarget
		if err := rows.Scan(
			&item.SessionID,
			&item.SessionStatus,
			&item.FinishedAt,
			&item.LTISub,
			&item.LineItemURL,
			&item.ScopeText,
			&item.Platform.ID,
			&item.Platform.Name,
			&item.Platform.Issuer,
			&item.Platform.ClientID,
			&item.Platform.DeploymentID,
			&item.Platform.OIDCAuthURL,
			&item.Platform.OIDCTokenURL,
			&item.Platform.JWKSURL,
			&item.Platform.ToolPrivateKey,
			&item.Platform.ToolPublicKey,
			&item.Platform.CreatedAt,
			&item.Platform.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ags score target: %w", err)
		}
		item.ScopeText = strings.TrimSpace(item.ScopeText)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ags score targets: %w", err)
	}
	return out, nil
}
