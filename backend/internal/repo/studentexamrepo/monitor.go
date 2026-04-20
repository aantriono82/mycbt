package studentexamrepo

import (
	"context"
	"fmt"
	"time"
)

type MonitorSession struct {
	SessionID string `json:"session_id"`
	ExamID    string `json:"exam_id"`

	StudentID       string `json:"student_id"`
	StudentName     string `json:"student_name"`
	StudentUsername string `json:"student_username"`
	StudentNIS      string `json:"student_nis"`

	Status     string `json:"status"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	LastSeenAt string `json:"last_seen_at,omitempty"`

	AnsweredQuestions int `json:"answered_questions"`
	TotalQuestions    int `json:"total_questions"`
	ProgressPercent   int `json:"progress_percent"`

	ConnectionStatus string `json:"connection_status"` // online|offline
	WarningCount     int    `json:"warning_count"`
}

type MonitorSessionsFilter struct {
	NowUTC       time.Time
	OnlineWindow time.Duration
	Limit        int
	Offset       int
}

func (r *Repo) ListMonitorSessions(ctx context.Context, examID string, f MonitorSessionsFilter) ([]MonitorSession, int, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id::text,
       s.exam_id::text,
       st.id::text,
       u.name,
       u.username,
       st.nis,
       s.status,
       to_char(s.started_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(s.finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at,
       COALESCE(to_char(s.last_seen_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS last_seen_at,
       (SELECT COUNT(*) FROM exam_attempts a WHERE a.exam_session_id = s.id) AS answered,
       (SELECT COUNT(*) FROM exam_session_questions q WHERE q.exam_session_id = s.id) AS total,
       (SELECT COUNT(*) FROM exam_events e WHERE e.exam_session_id = s.id AND e.type = 'focus_loss') AS warnings
FROM exam_sessions s
JOIN students st ON st.id = s.student_id
JOIN users u ON u.id = st.user_id
WHERE s.exam_id = $1
ORDER BY s.last_seen_at DESC NULLS LAST, s.started_at DESC
LIMIT $2 OFFSET $3`, examID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list monitor sessions: %w", err)
	}
	defer rows.Close()

	out := []MonitorSession{}
	for rows.Next() {
		var it MonitorSession
		var finished, lastSeen string
		var answered, total int
		if err := rows.Scan(
			&it.SessionID,
			&it.ExamID,
			&it.StudentID,
			&it.StudentName,
			&it.StudentUsername,
			&it.StudentNIS,
			&it.Status,
			&it.StartedAt,
			&finished,
			&lastSeen,
			&answered,
			&total,
			&it.WarningCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
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

		// Online heuristic: in_progress + last_seen within window.
		it.ConnectionStatus = "offline"
		if it.Status == "in_progress" && it.LastSeenAt != "" {
			if t, err := time.Parse(time.RFC3339, it.LastSeenAt); err == nil {
				if f.NowUTC.Sub(t.UTC()) <= f.OnlineWindow {
					it.ConnectionStatus = "online"
				}
			}
		}

		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exam_sessions WHERE exam_id = $1`, examID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count monitor sessions: %w", err)
	}
	return out, total, nil
}
