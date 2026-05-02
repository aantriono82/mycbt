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
	IsRead      bool   `json:"is_read"`
}

type ListStudentAnnouncementsFilter struct {
	Q          string
	Limit      int
	Offset     int
	NowUTC     time.Time
	UnreadOnly bool
}

func (r *Repo) ListStudentAnnouncements(
	ctx context.Context,
	studentID, levelID, groupID string,
	f ListStudentAnnouncementsFilter,
) ([]StudentAnnouncement, int, error) {
	const base = `
FROM announcements a
LEFT JOIN student_announcement_reads sar
  ON sar.announcement_id = a.id AND sar.student_id::text = $5
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
    $6 = ''
    OR a.title ILIKE '%'||$6||'%'
    OR a.body ILIKE '%'||$6||'%'
    OR a.category ILIKE '%'||$6||'%'
  )
  AND ($7 = false OR sar.announcement_id IS NULL)`

	rows, err := r.pool.Query(ctx, `
SELECT a.id::text,
       a.title,
       a.body,
       COALESCE(a.category, 'general'),
       to_char(a.published_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(a.expires_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       sar.announcement_id IS NOT NULL
`+base+`
ORDER BY a.published_at DESC, a.created_at DESC
LIMIT $8 OFFSET $9`,
		f.NowUTC,
		studentID,
		groupID,
		levelID,
		studentID,
		strings.TrimSpace(f.Q),
		f.UnreadOnly,
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
		if err := rows.Scan(&it.ID, &it.Title, &it.Body, &it.Category, &it.PublishedAt, &expiresAt, &it.IsRead); err != nil {
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
		studentID,
		strings.TrimSpace(f.Q),
		f.UnreadOnly,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count student announcements: %w", err)
	}

	return out, total, nil
}

func (r *Repo) MarkAnnouncementsRead(ctx context.Context, studentID string, announcementIDs []string) error {
	if strings.TrimSpace(studentID) == "" || len(announcementIDs) == 0 {
		return nil
	}

	unique := make([]string, 0, len(announcementIDs))
	seen := make(map[string]struct{}, len(announcementIDs))
	for _, id := range announcementIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return nil
	}

	_, err := r.pool.Exec(ctx, `
INSERT INTO student_announcement_reads (student_id, announcement_id, read_at)
SELECT $1::uuid, a.id, now()
FROM announcements a
WHERE a.id::text = ANY($2::text[])
ON CONFLICT (student_id, announcement_id) DO UPDATE
SET read_at = EXCLUDED.read_at
`, studentID, unique)
	if err != nil {
		return fmt.Errorf("mark announcements read: %w", err)
	}
	return nil
}
