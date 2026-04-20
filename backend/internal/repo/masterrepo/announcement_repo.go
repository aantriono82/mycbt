package masterrepo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Announcement struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Body            string `json:"body"`
	Category        string `json:"category"`
	IsActive        bool   `json:"is_active"`
	PublishedAt     string `json:"published_at"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	TargetLevelID   string `json:"target_level_id,omitempty"`
	TargetGroupID   string `json:"target_group_id,omitempty"`
	TargetStudentID string `json:"target_student_id,omitempty"`
	CreatedByUserID string `json:"created_by_user_id,omitempty"`
}

type AnnouncementsRepo struct{ pool *pgxpool.Pool }

func NewAnnouncements(pool *pgxpool.Pool) *AnnouncementsRepo { return &AnnouncementsRepo{pool: pool} }

func (r *AnnouncementsRepo) List(ctx context.Context, q string, isActiveFilter string, limit, offset int) ([]Announcement, int, error) {
	const base = `
FROM announcements a
WHERE (
  $1 = ''
  OR a.title ILIKE '%'||$1||'%'
  OR a.body ILIKE '%'||$1||'%'
  OR COALESCE(a.category,'') ILIKE '%'||$1||'%'
)
AND (
  $2 = ''
  OR ($2 = 'true' AND a.is_active = true)
  OR ($2 = 'false' AND a.is_active = false)
)`

	rows, err := r.pool.Query(ctx, `
SELECT a.id::text,
       a.title,
       a.body,
       COALESCE(a.category,'general'),
       a.is_active,
       to_char(a.published_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(a.expires_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       COALESCE(a.target_level_id::text,''),
       COALESCE(a.target_group_id::text,''),
       COALESCE(a.target_student_id::text,''),
       COALESCE(a.created_by_user_id::text,'')
`+base+`
ORDER BY a.published_at DESC, a.created_at DESC
LIMIT $3 OFFSET $4`, strings.TrimSpace(q), strings.TrimSpace(isActiveFilter), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list announcements: %w", err)
	}
	defer rows.Close()

	items := []Announcement{}
	for rows.Next() {
		var it Announcement
		if err := rows.Scan(
			&it.ID,
			&it.Title,
			&it.Body,
			&it.Category,
			&it.IsActive,
			&it.PublishedAt,
			&it.ExpiresAt,
			&it.TargetLevelID,
			&it.TargetGroupID,
			&it.TargetStudentID,
			&it.CreatedByUserID,
		); err != nil {
			return nil, 0, fmt.Errorf("scan announcement: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base, strings.TrimSpace(q), strings.TrimSpace(isActiveFilter)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count announcements: %w", err)
	}

	return items, total, nil
}

func (r *AnnouncementsRepo) Get(ctx context.Context, id string) (Announcement, bool, error) {
	const q = `
SELECT a.id::text,
       a.title,
       a.body,
       COALESCE(a.category,'general'),
       a.is_active,
       to_char(a.published_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(a.expires_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       COALESCE(a.target_level_id::text,''),
       COALESCE(a.target_group_id::text,''),
       COALESCE(a.target_student_id::text,''),
       COALESCE(a.created_by_user_id::text,'')
FROM announcements a
WHERE a.id = $1
LIMIT 1`
	var it Announcement
	if err := r.pool.QueryRow(ctx, q, id).Scan(
		&it.ID,
		&it.Title,
		&it.Body,
		&it.Category,
		&it.IsActive,
		&it.PublishedAt,
		&it.ExpiresAt,
		&it.TargetLevelID,
		&it.TargetGroupID,
		&it.TargetStudentID,
		&it.CreatedByUserID,
	); err != nil {
		if isNoRows(err) {
			return Announcement{}, false, nil
		}
		return Announcement{}, false, fmt.Errorf("get announcement: %w", err)
	}
	return it, true, nil
}

func (r *AnnouncementsRepo) Create(
	ctx context.Context,
	title, body, category string,
	isActive bool,
	publishedAt *time.Time,
	expiresAt *time.Time,
	targetLevelID, targetGroupID, targetStudentID, createdByUserID string,
) (Announcement, error) {
	const q = `
INSERT INTO announcements (
  title, body, category, is_active, published_at, expires_at,
  target_level_id, target_group_id, target_student_id, created_by_user_id
)
VALUES (
  $1, $2, NULLIF($3,''), $4, COALESCE($5, now()), $6,
  NULLIF($7,'')::uuid, NULLIF($8,'')::uuid, NULLIF($9,'')::uuid, NULLIF($10,'')::uuid
)
RETURNING id::text`
	var id string
	if err := r.pool.QueryRow(
		ctx,
		q,
		title,
		body,
		strings.TrimSpace(category),
		isActive,
		publishedAt,
		expiresAt,
		strings.TrimSpace(targetLevelID),
		strings.TrimSpace(targetGroupID),
		strings.TrimSpace(targetStudentID),
		strings.TrimSpace(createdByUserID),
	).Scan(&id); err != nil {
		return Announcement{}, fmt.Errorf("create announcement: %w", err)
	}

	it, _, err := r.Get(ctx, id)
	if err != nil {
		return Announcement{}, err
	}
	return it, nil
}

func (r *AnnouncementsRepo) Update(
	ctx context.Context,
	id, title, body, category string,
	isActive bool,
	publishedAt *time.Time,
	expiresAt *time.Time,
	targetLevelID, targetGroupID, targetStudentID string,
) (Announcement, bool, error) {
	const q = `
UPDATE announcements
SET title = $2,
    body = $3,
    category = NULLIF($4,''),
    is_active = $5,
    published_at = COALESCE($6, now()),
    expires_at = $7,
    target_level_id = NULLIF($8,'')::uuid,
    target_group_id = NULLIF($9,'')::uuid,
    target_student_id = NULLIF($10,'')::uuid,
    updated_at = now()
WHERE id = $1
RETURNING id::text,
          title,
          body,
          COALESCE(category,'general'),
          is_active,
          to_char(published_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
          COALESCE(to_char(expires_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
          COALESCE(target_level_id::text,''),
          COALESCE(target_group_id::text,''),
          COALESCE(target_student_id::text,''),
          COALESCE(created_by_user_id::text,'')`
	var it Announcement
	if err := r.pool.QueryRow(
		ctx,
		q,
		id,
		title,
		body,
		strings.TrimSpace(category),
		isActive,
		publishedAt,
		expiresAt,
		strings.TrimSpace(targetLevelID),
		strings.TrimSpace(targetGroupID),
		strings.TrimSpace(targetStudentID),
	).Scan(
		&it.ID,
		&it.Title,
		&it.Body,
		&it.Category,
		&it.IsActive,
		&it.PublishedAt,
		&it.ExpiresAt,
		&it.TargetLevelID,
		&it.TargetGroupID,
		&it.TargetStudentID,
		&it.CreatedByUserID,
	); err != nil {
		if isNoRows(err) {
			return Announcement{}, false, nil
		}
		return Announcement{}, false, fmt.Errorf("update announcement: %w", err)
	}
	return it, true, nil
}

func (r *AnnouncementsRepo) Delete(ctx context.Context, id string) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM announcements WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete announcement: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}
