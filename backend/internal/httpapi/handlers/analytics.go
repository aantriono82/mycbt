package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsHandler struct {
	pool *pgxpool.Pool
}

func NewAnalyticsHandler(pool *pgxpool.Pool) *AnalyticsHandler {
	return &AnalyticsHandler{pool: pool}
}

type TrendPoint struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type SubjectStat struct {
	SubjectName string  `json:"subject_name"`
	AvgScore    float64 `json:"avg_score"`
	ExamCount   int     `json:"exam_count"`
}

type GroupStat struct {
	GroupName string  `json:"group_name"`
	AvgScore  float64 `json:"avg_score"`
}

func (h *AnalyticsHandler) Dashboard(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Global Stats
	var totalExams, totalParticipants int
	var globalAvg float64

	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exams`).Scan(&totalExams)
	h.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT student_id) FROM exam_sessions WHERE status = 'submitted'`).Scan(&totalParticipants)
	h.pool.QueryRow(ctx, `SELECT COALESCE(AVG(score), 0) FROM exam_sessions WHERE status = 'submitted'`).Scan(&globalAvg)

	// 2. Trend (Last 10 Exams)
	trends, err := h.getPerformanceTrend(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Subject Performance
	subjects, err := h.getSubjectPerformance(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Group Performance
	groups, err := h.getGroupPerformance(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"summary": gin.H{
				"total_exams":        totalExams,
				"total_participants": totalParticipants,
				"average_score":      globalAvg,
			},
			"trends":   trends,
			"subjects": subjects,
			"groups":   groups,
		},
	})
}

func (h *AnalyticsHandler) getPerformanceTrend(ctx context.Context) ([]TrendPoint, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT e.title, COALESCE(AVG(es.score), 0) as avg_score
		FROM exam_sessions es
		JOIN exams e ON e.id = es.exam_id
		WHERE es.status = 'submitted'
		GROUP BY e.id, e.title, e.starts_at
		ORDER BY e.starts_at DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TrendPoint
	for rows.Next() {
		var p TrendPoint
		if err := rows.Scan(&p.Label, &p.Score); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	// Reverse to show chronologically on chart (left to right)
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}
	return points, nil
}

func (h *AnalyticsHandler) getSubjectPerformance(ctx context.Context) ([]SubjectStat, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT s.name, COALESCE(AVG(es.score), 0), COUNT(DISTINCT e.id)
		FROM exam_sessions es
		JOIN exams e ON e.id = es.exam_id
		JOIN subjects s ON s.id = e.subject_id
		WHERE es.status = 'submitted'
		GROUP BY s.name
		ORDER BY AVG(es.score) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []SubjectStat
	for rows.Next() {
		var s SubjectStat
		if err := rows.Scan(&s.SubjectName, &s.AvgScore, &s.ExamCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (h *AnalyticsHandler) getGroupPerformance(ctx context.Context) ([]GroupStat, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT g.name, COALESCE(AVG(es.score), 0)
		FROM exam_sessions es
		JOIN students st ON st.user_id = es.student_id
		JOIN groups g ON g.id = st.group_id
		WHERE es.status = 'submitted'
		GROUP BY g.name
		ORDER BY AVG(es.score) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []GroupStat
	for rows.Next() {
		var s GroupStat
		if err := rows.Scan(&s.GroupName, &s.AvgScore); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
