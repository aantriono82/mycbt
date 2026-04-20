package studentexamrepo

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type StudentAnnouncement struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Category    string `json:"category"`
	PublishedAt string `json:"published_at"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

type ListStudentAnnouncementsFilter struct {
	Q      string
	Limit  int
	Offset int
	NowUTC time.Time
}

func (r *Repo) ListStudentAnnouncements(
	ctx context.Context,
	studentID, levelID, groupID string,
	f ListStudentAnnouncementsFilter,
) ([]StudentAnnouncement, int, error) {
	const base = `
FROM announcements a
WHERE a.is_active = true
  AND a.published_at <= $1
  AND (a.expires_at IS NULL OR a.expires_at >= $1)
  AND (
    a.target_student_id IS NULL
    OR a.target_student_id::text = $2
    OR (a.target_group_id IS NOT NULL AND $3 <> '' AND a.target_group_id::text = $3)
    OR (a.target_level_id IS NOT NULL AND $4 <> '' AND a.target_level_id::text = $4)
  )
  AND (
    $5 = ''
    OR a.title ILIKE '%'||$5||'%'
    OR a.body ILIKE '%'||$5||'%'
    OR a.category ILIKE '%'||$5||'%'
  )`

	rows, err := r.pool.Query(ctx, `
SELECT a.id::text,
       a.title,
       a.body,
       COALESCE(a.category, 'general'),
       to_char(a.published_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(a.expires_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'')
`+base+`
ORDER BY a.published_at DESC, a.created_at DESC
LIMIT $6 OFFSET $7`,
		f.NowUTC,
		studentID,
		groupID,
		levelID,
		strings.TrimSpace(f.Q),
		f.Limit,
		f.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list student announcements: %w", err)
	}
	defer rows.Close()

	out := []StudentAnnouncement{}
	for rows.Next() {
		var it StudentAnnouncement
		var expiresAt string
		if err := rows.Scan(&it.ID, &it.Title, &it.Body, &it.Category, &it.PublishedAt, &expiresAt); err != nil {
			return nil, 0, fmt.Errorf("scan announcement: %w", err)
		}
		if expiresAt != "" {
			it.ExpiresAt = expiresAt
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base,
		f.NowUTC,
		studentID,
		groupID,
		levelID,
		strings.TrimSpace(f.Q),
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count student announcements: %w", err)
	}

	return out, total, nil
}
