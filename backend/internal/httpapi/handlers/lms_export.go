package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LMSExportHandler handles data portability exports for LMS interoperability.
type LMSExportHandler struct {
	pool *pgxpool.Pool
}

func NewLMSExportHandler(pool *pgxpool.Pool) *LMSExportHandler {
	return &LMSExportHandler{pool: pool}
}

// ─── Shared structs ────────────────────────────────────────────────────────────

type lmsStudentRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	NIS      string `json:"nis"`
	Jenjang  string `json:"jenjang"`
	Level    string `json:"level"`
	Group    string `json:"group"`
	Program  string `json:"program"`
	Email    string `json:"email"`
}

type lmsExamResultRow struct {
	ExamID       string  `json:"exam_id"`
	ExamTitle    string  `json:"exam_title"`
	ExamDate     string  `json:"exam_date"`
	Subject      string  `json:"subject"`
	StudentID    string  `json:"student_id"`
	StudentName  string  `json:"student_name"`
	Username     string  `json:"username"`
	NIS          string  `json:"nis"`
	Status       string  `json:"status"`
	Score        float64 `json:"score"`
	MaxScore     float64 `json:"max_score"`
	CorrectCount int     `json:"correct_count"`
	TotalItems   int     `json:"total_items"`
	StartedAt    string  `json:"started_at"`
	FinishedAt   string  `json:"finished_at"`
}

type lmsSummaryInfo struct {
	GeneratedAt   string `json:"generated_at"`
	TotalStudents int    `json:"total_students"`
	TotalExams    int    `json:"total_exams"`
	TotalSessions int    `json:"total_sessions"`
}

// ─── Summary ───────────────────────────────────────────────────────────────────

func (h *LMSExportHandler) Summary(c *gin.Context) {
	ctx := c.Request.Context()

	var totalStudents, totalExams, totalSessions int
	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'student'`).Scan(&totalStudents)
	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exams`).Scan(&totalExams)
	h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exam_sessions`).Scan(&totalSessions)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"generated_at":   time.Now().UTC().Format(time.RFC3339),
			"total_students": totalStudents,
			"total_exams":    totalExams,
			"total_sessions": totalSessions,
		},
	})
}

// ─── Exam list for picker ──────────────────────────────────────────────────────

func (h *LMSExportHandler) ListExams(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.pool.Query(ctx, `
		SELECT e.id::text, e.title, COALESCE(s.name,'') AS subject,
		       e.start_at, e.end_at
		FROM exams e
		LEFT JOIN subjects s ON s.id = e.subject_id
		ORDER BY e.start_at DESC
		LIMIT 200
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	defer rows.Close()
	type row struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Subject string `json:"subject"`
		StartAt string `json:"start_at"`
		EndAt   string `json:"end_at"`
	}
	out := []row{}
	for rows.Next() {
		var r row
		var startAt, endAt time.Time
		if err := rows.Scan(&r.ID, &r.Title, &r.Subject, &startAt, &endAt); err != nil {
			continue
		}
		r.StartAt = startAt.Format("2006-01-02 15:04")
		r.EndAt = endAt.Format("2006-01-02 15:04")
		out = append(out, r)
	}
	c.JSON(200, gin.H{"data": out})
}

// ─── Students export ───────────────────────────────────────────────────────────

func (h *LMSExportHandler) ExportStudents(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	ctx := c.Request.Context()

	rows, err := h.pool.Query(ctx, `
		SELECT u.id::text,
		       u.name,
		       u.username,
		       COALESCE(st.nis,'') AS nis,
		       COALESCE(st.jenjang,'') AS jenjang,
		       COALESCE(lv.name,'') AS level_name,
		       COALESCE(g.name,'') AS group_name,
		       COALESCE(pr.name,'') AS program_name,
		       COALESCE(u.email,'') AS email
		FROM users u
		JOIN students st ON st.user_id = u.id
		LEFT JOIN levels lv ON lv.id = st.level_id
		LEFT JOIN groups g  ON g.id  = st.group_id
		LEFT JOIN programs pr ON pr.id = st.program_id
		ORDER BY u.name
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}
	defer rows.Close()

	var data []lmsStudentRow
	for rows.Next() {
		var r lmsStudentRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Username, &r.NIS, &r.Jenjang, &r.Level, &r.Group, &r.Program, &r.Email); err != nil {
			continue
		}
		data = append(data, r)
	}

	filename := fmt.Sprintf("lms-students-%s", time.Now().Format("20060102"))
	if format == "json" {
		h.writeJSON(c, data, filename+".json")
		return
	}
	// CSV
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"id", "name", "username", "nis", "jenjang", "level", "group", "program", "email"})
	for _, r := range data {
		w.Write([]string{r.ID, r.Name, r.Username, r.NIS, r.Jenjang, r.Level, r.Group, r.Program, r.Email})
	}
	w.Flush()
	h.writeCSV(c, buf.Bytes(), filename+".csv")
}

// ─── Results export (all exams or single exam) ────────────────────────────────

func (h *LMSExportHandler) ExportResults(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	examID := c.Query("exam_id") // optional filter
	ctx := c.Request.Context()

	data, err := h.fetchResults(ctx, examID)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "internal error"}})
		return
	}

	suffix := "all"
	if examID != "" {
		suffix = examID[:8]
	}
	filename := fmt.Sprintf("lms-results-%s-%s", suffix, time.Now().Format("20060102"))

	if format == "json" {
		h.writeJSON(c, data, filename+".json")
		return
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{
		"exam_id", "exam_title", "exam_date", "subject",
		"student_id", "student_name", "username", "nis",
		"status", "score", "max_score", "correct_count", "total_items",
		"started_at", "finished_at",
	})
	for _, r := range data {
		w.Write([]string{
			r.ExamID, r.ExamTitle, r.ExamDate, r.Subject,
			r.StudentID, r.StudentName, r.Username, r.NIS,
			r.Status,
			fmt.Sprintf("%.2f", r.Score),
			fmt.Sprintf("%.2f", r.MaxScore),
			fmt.Sprintf("%d", r.CorrectCount),
			fmt.Sprintf("%d", r.TotalItems),
			r.StartedAt, r.FinishedAt,
		})
	}
	w.Flush()
	h.writeCSV(c, buf.Bytes(), filename+".csv")
}

func (h *LMSExportHandler) fetchResults(ctx context.Context, examID string) ([]lmsExamResultRow, error) {
	query := `
		SELECT e.id::text,
		       e.title,
		       to_char(e.start_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS exam_date,
		       COALESCE(subj.name,'') AS subject,
		       u.id::text AS student_id,
		       u.name AS student_name,
		       u.username,
		       COALESCE(st.nis,'') AS nis,
		       es.status,
		       COALESCE(es.score, 0) AS score,
		       100.0 AS max_score,
		       COALESCE(es.correct_count,0) AS correct_count,
		       COALESCE(es.total_questions,0) AS total_items,
		       COALESCE(to_char(es.started_at,'YYYY-MM-DD HH24:MI'),'') AS started_at,
		       COALESCE(to_char(es.finished_at,'YYYY-MM-DD HH24:MI'),'') AS finished_at
		FROM exam_sessions es
		JOIN exams e ON e.id = es.exam_id
		JOIN users u ON u.id = es.student_id
		JOIN students st ON st.user_id = u.id
		LEFT JOIN subjects subj ON subj.id = e.subject_id
	`
	args := []any{}
	if examID != "" {
		query += ` WHERE e.id = $1`
		args = append(args, examID)
	}
	query += ` ORDER BY e.start_at DESC, u.name`

	rows, err := h.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []lmsExamResultRow
	for rows.Next() {
		var r lmsExamResultRow
		if err := rows.Scan(
			&r.ExamID, &r.ExamTitle, &r.ExamDate, &r.Subject,
			&r.StudentID, &r.StudentName, &r.Username, &r.NIS,
			&r.Status, &r.Score, &r.MaxScore, &r.CorrectCount, &r.TotalItems,
			&r.StartedAt, &r.FinishedAt,
		); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

func (h *LMSExportHandler) writeCSV(c *gin.Context, data []byte, filename string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *LMSExportHandler) writeJSON(c *gin.Context, v any, filename string) {
	payload, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": "marshal error"}})
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/json", payload)
}
