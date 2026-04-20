package studentexamrepo

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MonitorParticipant struct {
	ExamID string `json:"exam_id"`

	StudentID       string `json:"student_id"`
	StudentName     string `json:"student_name"`
	StudentUsername string `json:"student_username"`
	StudentNIS      string `json:"student_nis"`
	LevelID         string `json:"level_id,omitempty"`
	GroupID         string `json:"group_id,omitempty"`

	// Session (if joined)
	SessionID     string `json:"session_id,omitempty"`
	SessionStatus string `json:"session_status,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	FinishedAt    string `json:"finished_at,omitempty"`
	LastSeenAt    string `json:"last_seen_at,omitempty"`

	AnsweredQuestions int `json:"answered_questions"`
	TotalQuestions    int `json:"total_questions"`
	ProgressPercent   int `json:"progress_percent"`

	ConnectionStatus string `json:"connection_status"` // online|offline|blocked
}

type MonitorParticipantsFilter struct {
	Q            string
	NowUTC       time.Time
	OnlineWindow time.Duration
	Limit        int
	Offset       int
}

func (r *Repo) ListMonitorParticipants(ctx context.Context, examID string, f MonitorParticipantsFilter) ([]MonitorParticipant, int, error) {
	// Expand targets to actual students:
	// - explicit student targets
	// - students whose level_id/group_id matches targets
	//
	// We then left join to session + progress.
	const baseCTE = `
WITH targeted_students AS (
  SELECT DISTINCT s.id AS student_id
  FROM exam_targets t
  JOIN students s ON (
    (t.student_id IS NOT NULL AND t.student_id = s.id)
    OR (t.level_id IS NOT NULL AND s.level_id = t.level_id)
    OR (t.group_id IS NOT NULL AND s.group_id = t.group_id)
  )
  WHERE t.exam_id = $1
),
base AS (
  SELECT ts.student_id,
         u.name,
         u.username,
         s.nis,
         COALESCE(s.level_id::text,'') AS level_id,
         COALESCE(s.group_id::text,'') AS group_id
  FROM targeted_students ts
  JOIN students s ON s.id = ts.student_id
  JOIN users u ON u.id = s.user_id
  WHERE ($2 = '' OR u.username ILIKE '%'||$2||'%' OR u.name ILIKE '%'||$2||'%' OR s.nis ILIKE '%'||$2||'%')
)`

	rows, err := r.pool.Query(ctx, baseCTE+`
SELECT $1::text AS exam_id,
       b.student_id::text,
       b.name,
       b.username,
       b.nis,
       b.level_id,
       b.group_id,
       COALESCE(sess.id::text,'') AS session_id,
       COALESCE(sess.status,'') AS session_status,
       COALESCE(to_char(sess.started_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS started_at,
       COALESCE(to_char(sess.finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at,
       COALESCE(to_char(sess.last_seen_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS last_seen_at,
       COALESCE((SELECT COUNT(*) FROM exam_attempts a WHERE a.exam_session_id = sess.id), 0) AS answered,
       COALESCE((SELECT COUNT(*) FROM exam_session_questions q WHERE q.exam_session_id = sess.id), 0) AS total
FROM base b
LEFT JOIN exam_sessions sess
  ON sess.exam_id = $1 AND sess.student_id = b.student_id
ORDER BY b.name ASC
LIMIT $3 OFFSET $4`, examID, strings.TrimSpace(f.Q), f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list monitor participants: %w", err)
	}
	defer rows.Close()

	out := []MonitorParticipant{}
	for rows.Next() {
		var it MonitorParticipant
		var levelID, groupID string
		var sessID, sessStatus, started, finished, lastSeen string
		var answered, total int
		if err := rows.Scan(
			&it.ExamID,
			&it.StudentID,
			&it.StudentName,
			&it.StudentUsername,
			&it.StudentNIS,
			&levelID,
			&groupID,
			&sessID,
			&sessStatus,
			&started,
			&finished,
			&lastSeen,
			&answered,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		if levelID != "" {
			it.LevelID = levelID
		}
		if groupID != "" {
			it.GroupID = groupID
		}
		if sessID != "" {
			it.SessionID = sessID
		}
		if sessStatus != "" {
			it.SessionStatus = sessStatus
		}
		if started != "" {
			it.StartedAt = started
		}
		if finished != "" {
			it.FinishedAt = finished
		}
		if lastSeen != "" {
			it.LastSeenAt = lastSeen
		}
		it.AnsweredQuestions = answered
		it.TotalQuestions = total
		if total > 0 {
			p := int(float64(answered) / float64(total) * 100)
			if p < 0 {
				p = 0
			}
			if p > 100 {
				p = 100
			}
			it.ProgressPercent = p
		}

		it.ConnectionStatus = "blocked"
		if it.SessionID != "" {
			it.ConnectionStatus = "offline"
			if it.SessionStatus == "in_progress" && it.LastSeenAt != "" {
				if t, err := time.Parse(time.RFC3339, it.LastSeenAt); err == nil {
					if f.NowUTC.Sub(t.UTC()) <= f.OnlineWindow {
						it.ConnectionStatus = "online"
					}
				}
			}
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, baseCTE+`SELECT COUNT(*) FROM base`, examID, strings.TrimSpace(f.Q)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count monitor participants: %w", err)
	}

	return out, total, nil
}
