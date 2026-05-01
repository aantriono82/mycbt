package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"atigacbt/backend/internal/repo/studentexamrepo"
)

type AnalyticsHandler struct {
	pool    *pgxpool.Pool
	student *studentexamrepo.Repo
}

func NewAnalyticsHandler(pool *pgxpool.Pool, student *studentexamrepo.Repo) *AnalyticsHandler {
	return &AnalyticsHandler{pool: pool, student: student}
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

	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exams`).Scan(&totalExams)
	h.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT student_id) FROM exam_sessions WHERE status = 'submitted'`).Scan(&totalParticipants)

	// Dashboard aggregates are computed from the most recent submitted sessions.
	// We intentionally avoid relying on non-existent denormalized columns (score/correct_count/etc)
	// and reuse the scoring engine that backs results pages.
	type sessRow struct {
		SessionID   string
		ExamID      string
		ExamTitle   string
		StartsAt    time.Time
		SubjectName string
		GroupName   string
	}

	rows, err := h.pool.Query(ctx, `
SELECT es.id::text,
       es.exam_id::text,
       e.title,
       e.starts_at,
       COALESCE(subj.name,'') AS subject_name,
       COALESCE(g.name,'') AS group_name
FROM exam_sessions es
JOIN exams e ON e.id = es.exam_id
LEFT JOIN subjects subj ON subj.id = e.subject_id
JOIN students st ON st.id = es.student_id
JOIN groups g ON g.id = st.group_id
WHERE es.status = 'submitted'
ORDER BY e.starts_at DESC, es.finished_at DESC NULLS LAST
LIMIT 250
`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var sessions []sessRow
	for rows.Next() {
		var r sessRow
		if err := rows.Scan(&r.SessionID, &r.ExamID, &r.ExamTitle, &r.StartsAt, &r.SubjectName, &r.GroupName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sessions = append(sessions, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nowUTC := time.Now().UTC()

	type agg struct {
		sum   float64
		count int
	}
	exams := map[string]struct {
		title    string
		startsAt time.Time
		agg      agg
	}{}
	subjectAgg := map[string]struct {
		agg     agg
		examIDs map[string]struct{}
	}{}
	groupAgg := map[string]agg{}

	var globalSum float64
	var globalCount int

	for _, s := range sessions {
		sum, err := h.student.ComputeAutoScoreAny(ctx, s.SessionID, nowUTC)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		score := float64(sum.Score)

		globalSum += score
		globalCount++

		ex := exams[s.ExamID]
		ex.title = s.ExamTitle
		ex.startsAt = s.StartsAt
		ex.agg.sum += score
		ex.agg.count++
		exams[s.ExamID] = ex

		subjKey := s.SubjectName
		subj := subjectAgg[subjKey]
		if subj.examIDs == nil {
			subj.examIDs = map[string]struct{}{}
		}
		subj.agg.sum += score
		subj.agg.count++
		subj.examIDs[s.ExamID] = struct{}{}
		subjectAgg[subjKey] = subj

		ga := groupAgg[s.GroupName]
		ga.sum += score
		ga.count++
		groupAgg[s.GroupName] = ga
	}

	globalAvg := 0.0
	if globalCount > 0 {
		globalAvg = globalSum / float64(globalCount)
	}

	// Trends: average score per exam (last 10 exams by starts_at).
	type examPoint struct {
		examID   string
		title    string
		startsAt time.Time
		avg      float64
	}
	points := make([]examPoint, 0, len(exams))
	for id, e := range exams {
		avg := 0.0
		if e.agg.count > 0 {
			avg = e.agg.sum / float64(e.agg.count)
		}
		points = append(points, examPoint{examID: id, title: e.title, startsAt: e.startsAt, avg: avg})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].startsAt.After(points[j].startsAt) })
	if len(points) > 10 {
		points = points[:10]
	}
	// Make chart left->right chronological.
	sort.Slice(points, func(i, j int) bool { return points[i].startsAt.Before(points[j].startsAt) })
	trends := make([]TrendPoint, 0, len(points))
	for _, p := range points {
		trends = append(trends, TrendPoint{Label: p.title, Score: p.avg})
	}

	// Subjects: average score per subject, plus exam_count (unique exams in sample).
	type subjStatTmp struct {
		name      string
		avg       float64
		examCount int
	}
	subjStats := make([]subjStatTmp, 0, len(subjectAgg))
	for name, a := range subjectAgg {
		avg := 0.0
		if a.agg.count > 0 {
			avg = a.agg.sum / float64(a.agg.count)
		}
		subjStats = append(subjStats, subjStatTmp{name: name, avg: avg, examCount: len(a.examIDs)})
	}
	sort.Slice(subjStats, func(i, j int) bool { return subjStats[i].avg > subjStats[j].avg })
	if len(subjStats) > 10 {
		subjStats = subjStats[:10]
	}
	subjects := make([]SubjectStat, 0, len(subjStats))
	for _, s := range subjStats {
		subjects = append(subjects, SubjectStat{SubjectName: s.name, AvgScore: s.avg, ExamCount: s.examCount})
	}

	// Groups: average score per group (top 10 by average).
	type grpStatTmp struct {
		name string
		avg  float64
	}
	grpStats := make([]grpStatTmp, 0, len(groupAgg))
	for name, a := range groupAgg {
		avg := 0.0
		if a.count > 0 {
			avg = a.sum / float64(a.count)
		}
		grpStats = append(grpStats, grpStatTmp{name: name, avg: avg})
	}
	sort.Slice(grpStats, func(i, j int) bool { return grpStats[i].avg > grpStats[j].avg })
	if len(grpStats) > 10 {
		grpStats = grpStats[:10]
	}
	groups := make([]GroupStat, 0, len(grpStats))
	for _, g := range grpStats {
		groups = append(groups, GroupStat{GroupName: g.name, AvgScore: g.avg})
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
