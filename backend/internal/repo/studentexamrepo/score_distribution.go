package studentexamrepo

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type ScoreDistributionBin struct {
	Min     int     `json:"min"`
	Max     int     `json:"max"`
	Label   string  `json:"label"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

type ExamScoreDistribution struct {
	TotalSessions   int                    `json:"total_sessions"`
	SubmittedCount  int                    `json:"submitted_count"`
	ExpiredCount    int                    `json:"expired_count"`
	MinScore        int                    `json:"min_score"`
	MaxScore        int                    `json:"max_score"`
	AverageScore    float64                `json:"average_score"`
	MedianScore     float64                `json:"median_score"`
	DistributionBin []ScoreDistributionBin `json:"distribution_bins"`
}

func (r *Repo) GetExamScoreDistribution(ctx context.Context, examID string, nowUTC time.Time) (ExamScoreDistribution, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id::text,
       s.status,
       COALESCE(s.attempt_no, ROW_NUMBER() OVER (PARTITION BY s.student_id ORDER BY s.started_at ASC, s.id ASC)) AS attempt_number,
       st.id::text,
       to_char(s.started_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"') AS started_at,
       COALESCE(to_char(s.finished_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'') AS finished_at
FROM exam_sessions s
JOIN students st ON st.id = s.student_id
WHERE s.exam_id = $1
  AND s.status <> 'in_progress'`, examID)
	if err != nil {
		return ExamScoreDistribution{}, fmt.Errorf("list sessions for score distribution: %w", err)
	}
	defer rows.Close()

	allRows := make([]ExamSessionRow, 0)
	for rows.Next() {
		var it ExamSessionRow
		if scanErr := rows.Scan(&it.SessionID, &it.Status, &it.AttemptNumber, &it.StudentID, &it.StartedAt, &it.FinishedAt); scanErr != nil {
			return ExamScoreDistribution{}, fmt.Errorf("scan session for score distribution: %w", scanErr)
		}
		allRows = append(allRows, it)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return ExamScoreDistribution{}, rowsErr
	}

	bins := make([]ScoreDistributionBin, 10)
	for i := range bins {
		min := i * 10
		max := min + 9
		if i == 9 {
			max = 100
		}
		bins[i] = ScoreDistributionBin{
			Min:   min,
			Max:   max,
			Label: fmt.Sprintf("%d-%d", min, max),
		}
	}

	if len(allRows) == 0 {
		return ExamScoreDistribution{
			TotalSessions:   0,
			SubmittedCount:  0,
			ExpiredCount:    0,
			MinScore:        0,
			MaxScore:        0,
			AverageScore:    0,
			MedianScore:     0,
			DistributionBin: bins,
		}, nil
	}

	for i := range allRows {
		sum, sErr := r.ComputeAutoScoreAny(ctx, allRows[i].SessionID, nowUTC)
		if sErr != nil {
			return ExamScoreDistribution{}, fmt.Errorf("compute score for distribution: %w", sErr)
		}
		allRows[i].Score = sum.Score
	}

	sessions := selectBestSessionsByStudent(allRows)
	submitted := 0
	expired := 0
	scores := make([]int, 0, len(sessions))
	totalScore := 0
	for _, s := range sessions {
		if s.Status == "submitted" || s.Status == "forced" {
			submitted++
		}
		if s.Status == "expired" {
			expired++
		}
		score := s.Score
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		scores = append(scores, score)
		totalScore += score

		idx := score / 10
		if score == 100 {
			idx = 9
		}
		bins[idx].Count++
	}

	sort.Ints(scores)
	minScore := scores[0]
	maxScore := scores[len(scores)-1]
	avg := round2(float64(totalScore) / float64(len(scores)))

	median := 0.0
	n := len(scores)
	if n%2 == 1 {
		median = float64(scores[n/2])
	} else {
		median = round2((float64(scores[(n/2)-1]) + float64(scores[n/2])) / 2.0)
	}

	for i := range bins {
		bins[i].Percent = round2((float64(bins[i].Count) / float64(len(scores))) * 100.0)
	}

	return ExamScoreDistribution{
		TotalSessions:   len(sessions),
		SubmittedCount:  submitted,
		ExpiredCount:    expired,
		MinScore:        minScore,
		MaxScore:        maxScore,
		AverageScore:    avg,
		MedianScore:     median,
		DistributionBin: bins,
	}, nil
}
