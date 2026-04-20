package masterrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	SettingsKeySchoolIdentity = "school_identity"
	SettingsKeySystem         = "system"
	SettingsKeySMTP           = "smtp"
	SettingsKeyWhatsApp       = "whatsapp"
)

type SchoolIdentity struct {
	SchoolName    string `json:"school_name"`
	Address       string `json:"address,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Email         string `json:"email,omitempty"`
	Website       string `json:"website,omitempty"`
	PrincipalName string `json:"principal_name,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
}

type SystemSettings struct {
	Timezone            string `json:"timezone"`
	TokenRequired       bool   `json:"token_required"`
	AllowResetLogin     bool   `json:"allow_reset_login"`
	MaxActiveSessions   int    `json:"max_active_sessions"`
	AttendanceRequireIP bool   `json:"attendance_require_ip"`
}

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`
	UseTLS   bool   `json:"use_tls"`
}

type WhatsAppConfig struct {
	APIURL string `json:"api_url"`
	APIKey string `json:"api_key"`
	Sender string `json:"sender"`
}

type SettingsRepo struct {
	pool *pgxpool.Pool
}

func NewSettings(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{pool: pool}
}

func defaultSchoolIdentity() SchoolIdentity {
	return SchoolIdentity{
		SchoolName: "MYCBT School",
	}
}

func defaultSystemSettings() SystemSettings {
	return SystemSettings{
		Timezone:            "Asia/Jakarta",
		TokenRequired:       true,
		AllowResetLogin:     true,
		MaxActiveSessions:   1,
		AttendanceRequireIP: false,
	}
}

func (r *SettingsRepo) GetSchoolIdentity(ctx context.Context) (SchoolIdentity, error) {
	var out SchoolIdentity
	ok, err := r.getJSON(ctx, SettingsKeySchoolIdentity, &out)
	if err != nil {
		return SchoolIdentity{}, err
	}
	if !ok {
		return defaultSchoolIdentity(), nil
	}
	if out.SchoolName == "" {
		out.SchoolName = "MYCBT School"
	}
	return out, nil
}

func (r *SettingsRepo) UpsertSchoolIdentity(ctx context.Context, v SchoolIdentity) (SchoolIdentity, error) {
	if v.SchoolName == "" {
		v.SchoolName = "MYCBT School"
	}
	if err := r.upsertJSON(ctx, SettingsKeySchoolIdentity, v); err != nil {
		return SchoolIdentity{}, err
	}
	return v, nil
}

func (r *SettingsRepo) GetSystem(ctx context.Context) (SystemSettings, error) {
	var out SystemSettings
	ok, err := r.getJSON(ctx, SettingsKeySystem, &out)
	if err != nil {
		return SystemSettings{}, err
	}
	if !ok {
		return defaultSystemSettings(), nil
	}
	if out.Timezone == "" {
		out.Timezone = "Asia/Jakarta"
	}
	if out.MaxActiveSessions < 1 {
		out.MaxActiveSessions = 1
	}
	return out, nil
}

func (r *SettingsRepo) UpsertSystem(ctx context.Context, v SystemSettings) (SystemSettings, error) {
	if v.Timezone == "" {
		v.Timezone = "Asia/Jakarta"
	}
	if v.MaxActiveSessions < 1 {
		v.MaxActiveSessions = 1
	}
	if err := r.upsertJSON(ctx, SettingsKeySystem, v); err != nil {
		return SystemSettings{}, err
	}
	return v, nil
}

func (r *SettingsRepo) GetSMTP(ctx context.Context) (SMTPConfig, error) {
	var out SMTPConfig
	ok, err := r.getJSON(ctx, SettingsKeySMTP, &out)
	if err != nil {
		return SMTPConfig{}, err
	}
	if !ok {
		return SMTPConfig{}, nil
	}
	return out, nil
}

func (r *SettingsRepo) UpsertSMTP(ctx context.Context, v SMTPConfig) (SMTPConfig, error) {
	if err := r.upsertJSON(ctx, SettingsKeySMTP, v); err != nil {
		return SMTPConfig{}, err
	}
	return v, nil
}

func (r *SettingsRepo) GetWhatsApp(ctx context.Context) (WhatsAppConfig, error) {
	var out WhatsAppConfig
	ok, err := r.getJSON(ctx, SettingsKeyWhatsApp, &out)
	if err != nil {
		return WhatsAppConfig{}, err
	}
	if !ok {
		return WhatsAppConfig{}, nil
	}
	return out, nil
}

func (r *SettingsRepo) UpsertWhatsApp(ctx context.Context, v WhatsAppConfig) (WhatsAppConfig, error) {
	if err := r.upsertJSON(ctx, SettingsKeyWhatsApp, v); err != nil {
		return WhatsAppConfig{}, err
	}
	return v, nil
}

func (r *SettingsRepo) getJSON(ctx context.Context, key string, out any) (bool, error) {
	const q = `SELECT value_json FROM app_settings WHERE key = $1 LIMIT 1`
	var raw []byte
	err := r.pool.QueryRow(ctx, q, key).Scan(&raw)
	if err != nil {
		if isNoRows(err) {
			return false, nil
		}
		return false, fmt.Errorf("get app setting %s: %w", key, err)
	}
	if len(raw) == 0 {
		return true, nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return false, fmt.Errorf("unmarshal app setting %s: %w", key, err)
	}
	return true, nil
}

func (r *SettingsRepo) upsertJSON(ctx context.Context, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal app setting %s: %w", key, err)
	}
	const q = `
INSERT INTO app_settings (key, value_json, updated_at)
VALUES ($1, $2::jsonb, now())
ON CONFLICT (key)
DO UPDATE SET value_json = EXCLUDED.value_json, updated_at = now()`
	if _, err := r.pool.Exec(ctx, q, key, string(raw)); err != nil {
		return fmt.Errorf("upsert app setting %s: %w", key, err)
	}
	return nil
}
