package studentexamrepo

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ExamSessionRow struct {
	SessionID       string `json:"session_id"`
	ExamID          string `json:"exam_id"`
	Status          string `json:"status"`
	StartedAt       string `json:"started_at"`
	FinishedAt      string `json:"finished_at,omitempty"`
	DurationSeconds int    `json:"duration_seconds"`
	AttemptNumber   int    `json:"attempt_number"`
	IsBestAttempt   bool   `json:"is_best_attempt"`
	StudentID       string `json:"student_id"`
	StudentName     string `json:"student_name"`
	StudentUsername string `json:"student_username"`
	StudentNIS      string `json:"student_nis"`
	ParticipantNo   string `json:"participant_no"`
	StudentEmail    string `json:"student_email"`
	StudentPhone    string `json:"student_phone"`

	TotalQuestions      int `json:"total_questions"`
	AnsweredQuestions   int `json:"answered_questions"`
	AutoScorable        int `json:"auto_scorable_questions"`
	CorrectCount        int `json:"correct_count"`
	Score               int `json:"score"`
	ManualScoredCount   int `json:"manual_scored_count"`
	PendingGradingCount int `json:"pending_grading_count"`
}

type ListExamSessionsFilter struct {
	Q           string
	Limit       int
	Offset      int
	NowUTC      time.Time
	AllAttempts bool
}

func paginateExamSessionRows(rows []ExamSessionRow, limit, offset int) []ExamSessionRow {
	if offset >= len(rows) {
		return []ExamSessionRow{}
	}
	end := len(rows)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return rows[offset:end]
}

func (r *Repo) ListExamSessionsWithScore(ctx context.Context, examID string, f ListExamSessionsFilter) ([]ExamSessionRow, int, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id::text,
       s.exam_id::text,
       s.status,
       to_char(s.started_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(to_char(s.finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at,
       COALESCE(GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (COALESCE(s.finished_at, NOW()) - s.started_at)))::int), 0) AS duration_seconds,
       COALESCE(s.attempt_no, ROW_NUMBER() OVER (PARTITION BY s.student_id ORDER BY s.started_at ASC, s.id ASC)) AS attempt_number,
       st.id::text,
       u.name,
       u.username,
       st.nis,
       COALESCE(st.participant_no,''),
       COALESCE(u.email,''),
       COALESCE(u.phone,'')
FROM exam_sessions s
JOIN students st ON st.id = s.student_id
JOIN users u ON u.id = st.user_id
WHERE s.exam_id = $1
  AND ($2 = '' OR u.username ILIKE '%'||$2||'%' OR u.name ILIKE '%'||$2||'%' OR st.nis ILIKE '%'||$2||'%')
ORDER BY s.finished_at DESC NULLS LAST, s.started_at DESC`, examID, strings.TrimSpace(f.Q))
	if err != nil {
		return nil, 0, fmt.Errorf("list exam sessions: %w", err)
	}
	defer rows.Close()

	allRows := []ExamSessionRow{}
	for rows.Next() {
		var it ExamSessionRow
		var finished string
		if err := rows.Scan(
			&it.SessionID,
			&it.ExamID,
			&it.Status,
			&it.StartedAt,
			&finished,
			&it.DurationSeconds,
			&it.AttemptNumber,
			&it.StudentID,
			&it.StudentName,
			&it.StudentUsername,
			&it.StudentNIS,
			&it.ParticipantNo,
			&it.StudentEmail,
			&it.StudentPhone,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		if finished != "" {
			it.FinishedAt = finished
		}
		allRows = append(allRows, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for i := range allRows {
		sum, err := r.ComputeAutoScoreAny(ctx, allRows[i].SessionID, f.NowUTC)
		if err != nil {
			return nil, 0, err
		}
		allRows[i].TotalQuestions = sum.TotalQuestions
		allRows[i].AnsweredQuestions = sum.AnsweredQuestions
		allRows[i].AutoScorable = sum.AutoScorable
		allRows[i].CorrectCount = sum.CorrectCount
		allRows[i].Score = sum.Score
		allRows[i].ManualScoredCount = sum.ManualScored
		allRows[i].PendingGradingCount = sum.PendingGrading
	}

	bestRows := selectBestSessionsByStudent(allRows)
	bestBySessionID := make(map[string]struct{}, len(bestRows))
	for _, row := range bestRows {
		bestBySessionID[row.SessionID] = struct{}{}
	}
	for i := range allRows {
		_, ok := bestBySessionID[allRows[i].SessionID]
		allRows[i].IsBestAttempt = ok
	}

	if f.AllAttempts {
		return paginateExamSessionRows(allRows, f.Limit, f.Offset), len(allRows), nil
	}

	total := len(bestRows)
	return paginateExamSessionRows(bestRows, f.Limit, f.Offset), total, nil
}
