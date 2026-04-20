package studentexamrepo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type AttendanceSubmission struct {
	ID                  string   `json:"id"`
	ExamID              string   `json:"exam_id"`
	StudentID           string   `json:"student_id"`
	Note                string   `json:"note,omitempty"`
	ClientIP            string   `json:"client_ip,omitempty"`
	AttendedAt          string   `json:"attended_at"`
	Lat                 *float64 `json:"lat,omitempty"`
	Lon                 *float64 `json:"lon,omitempty"`
	Accuracy            *float64 `json:"accuracy,omitempty"`
	IsQR                bool     `json:"is_qr"`
	AttendanceSessionID *string  `json:"attendance_session_id,omitempty"`
}

type AttendanceHistoryItem struct {
	ID          string `json:"id"`
	ExamID      string `json:"exam_id"`
	ExamTitle   string `json:"exam_title"`
	SubjectName string `json:"subject"`
	StartsAt    string `json:"starts_at"`
	EndsAt      string `json:"ends_at"`
	Note        string `json:"note,omitempty"`
	ClientIP    string `json:"client_ip,omitempty"`
	AttendedAt  string `json:"attended_at"`
}

type ListAttendanceHistoryFilter struct {
	Q      string
	Limit  int
	Offset int
}

func (r *Repo) EnsureStudentCanAttendExam(ctx context.Context, examID, studentID, levelID, groupID string) (bool, error) {
	const q = `
SELECT 1
FROM exams e
WHERE e.id = $1
  AND e.status = 'published'
  AND EXISTS (
    SELECT 1
    FROM exam_targets t
    WHERE t.exam_id = e.id
      AND (
        (t.student_id IS NOT NULL AND t.student_id::text = $2)
        OR (t.level_id IS NOT NULL AND $3 <> '' AND t.level_id::text = $3)
        OR (t.group_id IS NOT NULL AND $4 <> '' AND t.group_id::text = $4)
      )
  )
LIMIT 1`

	var one int
	if err := r.pool.QueryRow(ctx, q, examID, studentID, levelID, groupID).Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("ensure student can attend exam: %w", err)
	}
	return true, nil
}

func (r *Repo) UpsertAttendance(ctx context.Context, examID, studentID, note string, clientIP net.IP, nowUTC time.Time, opts ...AttendanceOption) (AttendanceSubmission, error) {
	// Default values
	var lat, lon, accuracy *float64
	var isQR bool
	var sessionID *string

	for _, opt := range opts {
		if opt.Lat != nil {
			lat = opt.Lat
		}
		if opt.Lon != nil {
			lon = opt.Lon
		}
		if opt.Accuracy != nil {
			accuracy = opt.Accuracy
		}
		if opt.IsQR {
			isQR = true
		}
		if opt.SessionID != "" {
			sessionID = &opt.SessionID
		}
	}

	const q = `
INSERT INTO student_attendance (
    exam_id, student_id, note, client_ip, attended_at, 
    lat, lon, accuracy, is_qr, attendance_session_id,
    created_at, updated_at
)
VALUES ($1, $2, NULLIF($3,''), NULLIF($4,'')::inet, $5, $6, $7, $8, $9, $10, now(), now())
ON CONFLICT (exam_id, student_id)
DO UPDATE SET
  note = EXCLUDED.note,
  client_ip = EXCLUDED.client_ip,
  attended_at = EXCLUDED.attended_at,
  lat = COALESCE(EXCLUDED.lat, student_attendance.lat),
  lon = COALESCE(EXCLUDED.lon, student_attendance.lon),
  accuracy = COALESCE(EXCLUDED.accuracy, student_attendance.accuracy),
  is_qr = EXCLUDED.is_qr,
  attendance_session_id = COALESCE(EXCLUDED.attendance_session_id, student_attendance.attendance_session_id),
  updated_at = now()
RETURNING id::text, exam_id::text, student_id::text, COALESCE(note,''),
          COALESCE(client_ip::text,''),
          to_char(attended_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
          lat, lon, accuracy, is_qr, attendance_session_id::text`

	var it AttendanceSubmission
	ipStr := ""
	if clientIP != nil {
		ipStr = clientIP.String()
	}
	if err := r.pool.QueryRow(ctx, q, examID, studentID, strings.TrimSpace(note), ipStr, nowUTC, lat, lon, accuracy, isQR, sessionID).Scan(
		&it.ID,
		&it.ExamID,
		&it.StudentID,
		&it.Note,
		&it.ClientIP,
		&it.AttendedAt,
		&it.Lat,
		&it.Lon,
		&it.Accuracy,
		&it.IsQR,
		&it.AttendanceSessionID,
	); err != nil {
		return AttendanceSubmission{}, fmt.Errorf("upsert attendance: %w", err)
	}
	return it, nil
}

type AttendanceOption struct {
	Lat       *float64
	Lon       *float64
	Accuracy  *float64
	IsQR      bool
	SessionID string
}

func (r *Repo) ListAttendanceHistory(
	ctx context.Context,
	studentID string,
	f ListAttendanceHistoryFilter,
) ([]AttendanceHistoryItem, int, error) {
	const base = `
FROM student_attendance a
JOIN exams e ON e.id = a.exam_id
JOIN subjects s ON s.id = e.subject_id
WHERE a.student_id = $1
  AND (
    $2 = ''
    OR e.title ILIKE '%'||$2||'%'
    OR s.name ILIKE '%'||$2||'%'
    OR COALESCE(a.note,'') ILIKE '%'||$2||'%'
  )`

	rows, err := r.pool.Query(ctx, `
SELECT a.id::text,
       e.id::text,
       e.title,
       s.name,
       to_char(e.starts_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       to_char(e.ends_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),
       COALESCE(a.note,''),
       COALESCE(a.client_ip::text,''),
       to_char(a.attended_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
`+base+`
ORDER BY a.attended_at DESC, a.created_at DESC
LIMIT $3 OFFSET $4`,
		studentID,
		strings.TrimSpace(f.Q),
		f.Limit,
		f.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list attendance history: %w", err)
	}
	defer rows.Close()

	out := []AttendanceHistoryItem{}
	for rows.Next() {
		var it AttendanceHistoryItem
		if err := rows.Scan(
			&it.ID,
			&it.ExamID,
			&it.ExamTitle,
			&it.SubjectName,
			&it.StartsAt,
			&it.EndsAt,
			&it.Note,
			&it.ClientIP,
			&it.AttendedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan attendance history: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base, studentID, strings.TrimSpace(f.Q)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count attendance history: %w", err)
	}

	return out, total, nil
}
