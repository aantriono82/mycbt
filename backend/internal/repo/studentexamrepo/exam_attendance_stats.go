package studentexamrepo

import (
	"context"
	"fmt"
	"strings"
)

type ExamAttendanceParticipant struct {
	StudentID      string `json:"student_id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	NIS            string `json:"nis"`
	LevelName      string `json:"level_name,omitempty"`
	GroupName      string `json:"group_name,omitempty"`
	Attended       bool   `json:"attended"`
	AttendedAt     string `json:"attended_at,omitempty"`
	AttendanceNote string `json:"attendance_note,omitempty"`
	SessionStatus  string `json:"session_status,omitempty"`
}

type ExamAttendanceFilter struct {
	Q      string
	Limit  int
	Offset int
}

func (r *Repo) ListExamAttendanceParticipants(
	ctx context.Context,
	examID string,
	f ExamAttendanceFilter,
) (items []ExamAttendanceParticipant, total, targetedTotal, attendedTotal int, err error) {
	const targetedCTE = `
WITH targeted_students AS (
  SELECT t.student_id
  FROM exam_targets t
  WHERE t.exam_id = $1 AND t.student_id IS NOT NULL
  UNION
  SELECT s.id
  FROM exam_targets t
  JOIN students s ON s.level_id = t.level_id
  WHERE t.exam_id = $1 AND t.level_id IS NOT NULL
  UNION
  SELECT s.id
  FROM exam_targets t
  JOIN students s ON s.group_id = t.group_id
  WHERE t.exam_id = $1 AND t.group_id IS NOT NULL
)`

	const fromClause = `
FROM targeted_students ts
JOIN students st ON st.id = ts.student_id
JOIN users u ON u.id = st.user_id
LEFT JOIN levels lv ON lv.id = st.level_id
LEFT JOIN groups gr ON gr.id = st.group_id
LEFT JOIN student_attendance sa ON sa.exam_id = $1 AND sa.student_id = ts.student_id
LEFT JOIN exam_sessions es ON es.exam_id = $1 AND es.student_id = ts.student_id
WHERE (
  $2 = ''
  OR u.name ILIKE '%'||$2||'%'
  OR u.username ILIKE '%'||$2||'%'
  OR st.nis ILIKE '%'||$2||'%'
)`

	rows, qerr := r.pool.Query(ctx, targetedCTE+`
SELECT ts.student_id::text,
       u.username,
       u.name,
       st.nis,
       COALESCE(lv.name,''),
       COALESCE(gr.name,''),
       (sa.id IS NOT NULL) AS attended,
       COALESCE(to_char(sa.attended_at at time zone 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
       COALESCE(sa.note,''),
       COALESCE(es.status,'')
`+fromClause+`
ORDER BY
  (sa.id IS NOT NULL) DESC,
  sa.attended_at DESC NULLS LAST,
  u.name ASC
LIMIT $3 OFFSET $4`, examID, strings.TrimSpace(f.Q), f.Limit, f.Offset)
	if qerr != nil {
		err = fmt.Errorf("list exam attendance participants: %w", qerr)
		return
	}
	defer rows.Close()

	items = []ExamAttendanceParticipant{}
	for rows.Next() {
		var it ExamAttendanceParticipant
		if scanErr := rows.Scan(
			&it.StudentID,
			&it.Username,
			&it.Name,
			&it.NIS,
			&it.LevelName,
			&it.GroupName,
			&it.Attended,
			&it.AttendedAt,
			&it.AttendanceNote,
			&it.SessionStatus,
		); scanErr != nil {
			err = fmt.Errorf("scan exam attendance participant: %w", scanErr)
			return
		}
		items = append(items, it)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		err = rowsErr
		return
	}

	if qerr = r.pool.QueryRow(ctx, targetedCTE+` SELECT COUNT(*) `+fromClause, examID, strings.TrimSpace(f.Q)).Scan(&total); qerr != nil {
		err = fmt.Errorf("count exam attendance participants: %w", qerr)
		return
	}

	if qerr = r.pool.QueryRow(ctx, targetedCTE+` SELECT COUNT(*) FROM targeted_students`, examID).Scan(&targetedTotal); qerr != nil {
		err = fmt.Errorf("count targeted students: %w", qerr)
		return
	}

	if qerr = r.pool.QueryRow(ctx, targetedCTE+`
SELECT COUNT(*)
FROM targeted_students ts
JOIN student_attendance sa ON sa.exam_id = $1 AND sa.student_id = ts.student_id`, examID).Scan(&attendedTotal); qerr != nil {
		err = fmt.Errorf("count attended students: %w", qerr)
		return
	}

	return
}
