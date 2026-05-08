package masterrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Teacher struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	PasswordPlain string `json:"password_plain"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	NIP           string `json:"nip"`
	Jenjang       string `json:"jenjang"`
	PhotoURL      string `json:"photo_url"`
	IsActive      bool   `json:"is_active"`
	MapelSummary  string `json:"mapel_summary"`
	LevelSummary  string `json:"level_summary"`
	GroupSummary  string `json:"group_summary"`
}

type Student struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	PasswordPlain string `json:"password_plain"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	NIS           string `json:"nis"`
	ParticipantNo string `json:"participant_no"`
	Jenjang       string `json:"jenjang"`
	ProgramID     string `json:"program_id"`
	LevelID       string `json:"level_id"`
	GroupID       string `json:"group_id"`
	PhotoURL      string `json:"photo_url"`
	IsActive      bool   `json:"is_active"`
}

type TeachersRepo struct{ pool *pgxpool.Pool }
type StudentsRepo struct{ pool *pgxpool.Pool }

func NewTeachers(pool *pgxpool.Pool) *TeachersRepo { return &TeachersRepo{pool: pool} }
func NewStudents(pool *pgxpool.Pool) *StudentsRepo { return &StudentsRepo{pool: pool} }

func (r *TeachersRepo) List(ctx context.Context, query string, limit, offset int) ([]Teacher, int, error) {
	const q = `
SELECT t.id, u.id, u.username, COALESCE(u.password_plain,''), u.name, COALESCE(u.email,''), COALESCE(u.phone,''), COALESCE(t.nip,''), COALESCE(t.jenjang,''), COALESCE(u.photo_url,''), u.is_active,
       COALESCE(sub.codes, ''), COALESCE(lvl.names, ''), COALESCE(grp.names, '')
FROM teachers t
JOIN users u ON u.id = t.user_id
LEFT JOIN (
    SELECT ts.teacher_id, STRING_AGG(s.code, ', ') as codes
    FROM teacher_subjects ts
    JOIN subjects s ON s.id = ts.subject_id
    GROUP BY ts.teacher_id
) sub ON sub.teacher_id = t.id
LEFT JOIN (
    SELECT tl.teacher_id, STRING_AGG(l.name, ', ') as names
    FROM teacher_levels tl
    JOIN levels l ON l.id = tl.level_id
    GROUP BY tl.teacher_id
) lvl ON lvl.teacher_id = t.id
LEFT JOIN (
    SELECT tg.teacher_id, STRING_AGG(g.name, ', ') as names
    FROM teacher_groups tg
    JOIN groups g ON g.id = tg.group_id
    GROUP BY tg.teacher_id
) grp ON grp.teacher_id = t.id
WHERE ($1 = '' OR u.username ILIKE '%'||$1||'%' OR u.name ILIKE '%'||$1||'%' OR COALESCE(u.email,'') ILIKE '%'||$1||'%' OR COALESCE(t.nip,'') ILIKE '%'||$1||'%')
  AND u.role = 'teacher'
ORDER BY u.name ASC
LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list teachers: %w", err)
	}
	defer rows.Close()
	out := []Teacher{}
	for rows.Next() {
		var it Teacher
		if err := rows.Scan(&it.ID, &it.UserID, &it.Username, &it.PasswordPlain, &it.Name, &it.Email, &it.Phone, &it.NIP, &it.Jenjang, &it.PhotoURL, &it.IsActive, &it.MapelSummary, &it.LevelSummary, &it.GroupSummary); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	const qc = `
SELECT COUNT(*)
FROM teachers t
JOIN users u ON u.id = t.user_id
WHERE ($1 = '' OR u.username ILIKE '%'||$1||'%' OR u.name ILIKE '%'||$1||'%' OR COALESCE(u.email,'') ILIKE '%'||$1||'%' OR COALESCE(t.nip,'') ILIKE '%'||$1||'%')
  AND u.role = 'teacher'`
	if err := r.pool.QueryRow(ctx, qc, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count teachers: %w", err)
	}

	return out, total, nil
}

func (r *TeachersRepo) Get(ctx context.Context, id string) (Teacher, bool, error) {
	const q = `
SELECT t.id, u.id, u.username, COALESCE(u.password_plain,''), u.name, COALESCE(u.email,''), COALESCE(u.phone,''), COALESCE(t.nip,''), COALESCE(t.jenjang,''), COALESCE(u.photo_url,''), u.is_active,
       COALESCE(sub.codes, ''), COALESCE(lvl.names, ''), COALESCE(grp.names, '')
FROM teachers t
JOIN users u ON u.id = t.user_id
LEFT JOIN (
    SELECT ts.teacher_id, STRING_AGG(s.code, ', ') as codes
    FROM teacher_subjects ts
    JOIN subjects s ON s.id = ts.subject_id
    GROUP BY ts.teacher_id
) sub ON sub.teacher_id = t.id
LEFT JOIN (
    SELECT tl.teacher_id, STRING_AGG(l.name, ', ') as names
    FROM teacher_levels tl
    JOIN levels l ON l.id = tl.level_id
    GROUP BY tl.teacher_id
) lvl ON lvl.teacher_id = t.id
LEFT JOIN (
    SELECT tg.teacher_id, STRING_AGG(g.name, ', ') as names
    FROM teacher_groups tg
    JOIN groups g ON g.id = tg.group_id
    GROUP BY tg.teacher_id
) grp ON grp.teacher_id = t.id
WHERE t.id = $1
LIMIT 1`
	var it Teacher
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.UserID, &it.Username, &it.PasswordPlain, &it.Name, &it.Email, &it.Phone, &it.NIP, &it.Jenjang, &it.PhotoURL, &it.IsActive, &it.MapelSummary, &it.LevelSummary, &it.GroupSummary)
	if err != nil {
		if isNoRows(err) {
			return Teacher{}, false, nil
		}
		return Teacher{}, false, fmt.Errorf("get teacher: %w", err)
	}
	return it, true, nil
}

func (r *StudentsRepo) List(ctx context.Context, query string, limit, offset int) ([]Student, int, error) {
	const q = `
SELECT s.id, u.id, u.username, COALESCE(u.password_plain,''), u.name, COALESCE(u.email,''), COALESCE(u.phone,''), s.nis, COALESCE(s.participant_no,''),
       COALESCE(s.jenjang,''), COALESCE(s.program_id::text,''), COALESCE(s.level_id::text,''), COALESCE(s.group_id::text,''),
       COALESCE(u.photo_url,''), u.is_active
FROM students s
JOIN users u ON u.id = s.user_id
WHERE ($1 = '' OR u.username ILIKE '%'||$1||'%' OR u.name ILIKE '%'||$1||'%' OR COALESCE(u.email,'') ILIKE '%'||$1||'%' OR s.nis ILIKE '%'||$1||'%' OR COALESCE(s.participant_no,'') ILIKE '%'||$1||'%')
  AND u.role = 'student'
ORDER BY u.name ASC
LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list students: %w", err)
	}
	defer rows.Close()
	out := []Student{}
	for rows.Next() {
		var it Student
		if err := rows.Scan(&it.ID, &it.UserID, &it.Username, &it.PasswordPlain, &it.Name, &it.Email, &it.Phone, &it.NIS, &it.ParticipantNo, &it.Jenjang, &it.ProgramID, &it.LevelID, &it.GroupID, &it.PhotoURL, &it.IsActive); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	const qc = `
SELECT COUNT(*)
FROM students s
JOIN users u ON u.id = s.user_id
WHERE ($1 = '' OR u.username ILIKE '%'||$1||'%' OR u.name ILIKE '%'||$1||'%' OR COALESCE(u.email,'') ILIKE '%'||$1||'%' OR s.nis ILIKE '%'||$1||'%' OR COALESCE(s.participant_no,'') ILIKE '%'||$1||'%')
  AND u.role = 'student'`
	if err := r.pool.QueryRow(ctx, qc, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count students: %w", err)
	}

	return out, total, nil
}

func (r *StudentsRepo) ListForTeacherUserID(ctx context.Context, userID string, query string, limit, offset int) ([]Student, int, error) {
	const base = `
FROM students s
JOIN users u ON u.id = s.user_id
WHERE ($1 = '' OR u.username ILIKE '%'||$1||'%' OR u.name ILIKE '%'||$1||'%' OR COALESCE(u.email,'') ILIKE '%'||$1||'%' OR s.nis ILIKE '%'||$1||'%' OR COALESCE(s.participant_no,'') ILIKE '%'||$1||'%')
  AND (
      s.group_id IN (SELECT group_id FROM teacher_groups tg JOIN teachers t ON t.id = tg.teacher_id WHERE t.user_id = $2)
      OR s.level_id IN (SELECT level_id FROM teacher_levels tl JOIN teachers t ON t.id = tl.teacher_id WHERE t.user_id = $2)
  )`

	const q = `
SELECT s.id, u.id, u.username, COALESCE(u.password_plain,''), u.name, COALESCE(u.email,''), COALESCE(u.phone,''), s.nis, COALESCE(s.participant_no,''),
       COALESCE(s.jenjang,''), COALESCE(s.program_id::text,''), COALESCE(s.level_id::text,''), COALESCE(s.group_id::text,''),
       COALESCE(u.photo_url,''), u.is_active ` + base + `
ORDER BY u.name ASC
LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, q, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list students for teacher: %w", err)
	}
	defer rows.Close()

	out := []Student{}
	for rows.Next() {
		var it Student
		if err := rows.Scan(&it.ID, &it.UserID, &it.Username, &it.PasswordPlain, &it.Name, &it.Email, &it.Phone, &it.NIS, &it.ParticipantNo, &it.Jenjang, &it.ProgramID, &it.LevelID, &it.GroupID, &it.PhotoURL, &it.IsActive); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base, query, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count students for teacher: %w", err)
	}

	return out, total, nil
}

func (r *StudentsRepo) Get(ctx context.Context, id string) (Student, bool, error) {
	const q = `
SELECT s.id, u.id, u.username, COALESCE(u.password_plain,''), u.name, COALESCE(u.email,''), COALESCE(u.phone,''), s.nis, COALESCE(s.participant_no,''),
       COALESCE(s.jenjang,''), COALESCE(s.program_id::text,''), COALESCE(s.level_id::text,''), COALESCE(s.group_id::text,''),
       COALESCE(u.photo_url,''), u.is_active
FROM students s
JOIN users u ON u.id = s.user_id
WHERE s.id = $1
LIMIT 1`
	var it Student
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.UserID, &it.Username, &it.PasswordPlain, &it.Name, &it.Email, &it.Phone, &it.NIS, &it.ParticipantNo, &it.Jenjang, &it.ProgramID, &it.LevelID, &it.GroupID, &it.PhotoURL, &it.IsActive)
	if err != nil {
		if isNoRows(err) {
			return Student{}, false, nil
		}
		return Student{}, false, fmt.Errorf("get student: %w", err)
	}
	return it, true, nil
}

func (r *TeachersRepo) UpdateTeacher(ctx context.Context, teacherID, username, name, email, phone, nip, jenjang string, isActive bool, passwordHash, passwordPlain string) (Teacher, bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Teacher{}, false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var userID string
	if err := tx.QueryRow(ctx, `SELECT user_id FROM teachers WHERE id = $1`, teacherID).Scan(&userID); err != nil {
		if isNoRows(err) {
			return Teacher{}, false, nil
		}
		return Teacher{}, false, fmt.Errorf("load teacher: %w", err)
	}

	if passwordHash != "" {
		const q = `UPDATE users SET username=$2, name=$3, email=NULLIF($4,''), phone=NULLIF($5,''), is_active=$6, password_hash=$7, password_plain=NULLIF($8,''), updated_at=now() WHERE id=$1`
		if _, err := tx.Exec(ctx, q, userID, username, name, email, phone, isActive, passwordHash, passwordPlain); err != nil {
			return Teacher{}, false, fmt.Errorf("update user: %w", err)
		}
	} else {
		const q = `UPDATE users SET username=$2, name=$3, email=NULLIF($4,''), phone=NULLIF($5,''), is_active=$6, updated_at=now() WHERE id=$1`
		if _, err := tx.Exec(ctx, q, userID, username, name, email, phone, isActive); err != nil {
			return Teacher{}, false, fmt.Errorf("update user: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE teachers SET nip=NULLIF($2,''), jenjang=$3, updated_at=now() WHERE id=$1`, teacherID, nip, jenjang); err != nil {
		return Teacher{}, false, fmt.Errorf("update teacher: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Teacher{}, false, fmt.Errorf("commit: %w", err)
	}

	// Return fresh view
	it, ok, err := r.Get(ctx, teacherID)
	return it, ok, err
}

func (r *StudentsRepo) UpdateStudent(ctx context.Context, studentID, username, name, email, phone, nis, participantNo, jenjang, programID, levelID, groupID string, isActive bool, passwordHash, passwordPlain string) (Student, bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Student{}, false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var userID string
	if err := tx.QueryRow(ctx, `SELECT user_id FROM students WHERE id = $1`, studentID).Scan(&userID); err != nil {
		if isNoRows(err) {
			return Student{}, false, nil
		}
		return Student{}, false, fmt.Errorf("load student: %w", err)
	}

	if passwordHash != "" {
		const q = `UPDATE users SET username=$2, name=$3, email=NULLIF($4,''), phone=NULLIF($5,''), is_active=$6, password_hash=$7, password_plain=NULLIF($8,''), updated_at=now() WHERE id=$1`
		if _, err := tx.Exec(ctx, q, userID, username, name, email, phone, isActive, passwordHash, passwordPlain); err != nil {
			return Student{}, false, fmt.Errorf("update user: %w", err)
		}
	} else {
		const q = `UPDATE users SET username=$2, name=$3, email=NULLIF($4,''), phone=NULLIF($5,''), is_active=$6, updated_at=now() WHERE id=$1`
		if _, err := tx.Exec(ctx, q, userID, username, name, email, phone, isActive); err != nil {
			return Student{}, false, fmt.Errorf("update user: %w", err)
		}
	}

	const qStudent = `
UPDATE students
SET nis=$2,
    participant_no=NULLIF($3,''),
    jenjang=NULLIF($4,''),
    program_id=NULLIF($5,'')::uuid,
    level_id=NULLIF($6,'')::uuid,
    group_id=NULLIF($7,'')::uuid,
    updated_at=now()
WHERE id=$1`
	if _, err := tx.Exec(ctx, qStudent, studentID, nis, participantNo, jenjang, programID, levelID, groupID); err != nil {
		return Student{}, false, fmt.Errorf("update student: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Student{}, false, fmt.Errorf("commit: %w", err)
	}

	it, ok, err := r.Get(ctx, studentID)
	return it, ok, err
}

func (r *TeachersRepo) Delete(ctx context.Context, teacherID string) (bool, error) {
	// Deleting the user cascades teacher.
	const q = `DELETE FROM users WHERE id = (SELECT user_id FROM teachers WHERE id = $1)`
	ct, err := r.pool.Exec(ctx, q, teacherID)
	if err != nil {
		return false, fmt.Errorf("delete teacher: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

func (r *StudentsRepo) Delete(ctx context.Context, studentID string) (bool, error) {
	const q = `DELETE FROM users WHERE id = (SELECT user_id FROM students WHERE id = $1)`
	ct, err := r.pool.Exec(ctx, q, studentID)
	if err != nil {
		return false, fmt.Errorf("delete student: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

func (r *TeachersRepo) CreateTeacherTx(ctx context.Context, tx pgx.Tx, username, passwordHash, passwordPlain, name, email, phone, nip, jenjang, googleID string, subjectIDs []string, groupIDs []string, levelIDs []string) (teacherID, userID string, err error) {
	const insUser = `INSERT INTO users (username, password_hash, password_plain, role, name, email, phone, google_id, is_active) VALUES ($1,$2,NULLIF($3,''),'teacher',$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),true) RETURNING id`
	if err := tx.QueryRow(ctx, insUser, username, passwordHash, passwordPlain, name, email, phone, googleID).Scan(&userID); err != nil {
		return "", "", fmt.Errorf("insert user: %w", err)
	}

	const insTeacher = `INSERT INTO teachers (user_id, nip, jenjang) VALUES ($1, NULLIF($2,''), $3) RETURNING id`
	if err := tx.QueryRow(ctx, insTeacher, userID, nip, jenjang).Scan(&teacherID); err != nil {
		return "", "", fmt.Errorf("insert teacher: %w", err)
	}

	if len(subjectIDs) > 0 {
		const insMap = `INSERT INTO teacher_subjects (teacher_id, subject_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
		for _, sid := range subjectIDs {
			if _, err := tx.Exec(ctx, insMap, teacherID, sid); err != nil {
				return "", "", fmt.Errorf("insert teacher_subject: %w", err)
			}
		}
	}
	if len(groupIDs) > 0 {
		const insMap = `INSERT INTO teacher_groups (teacher_id, group_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
		for _, gid := range groupIDs {
			if _, err := tx.Exec(ctx, insMap, teacherID, gid); err != nil {
				return "", "", fmt.Errorf("insert teacher_group: %w", err)
			}
		}
	}
	if len(levelIDs) > 0 {
		const insMap = `INSERT INTO teacher_levels (teacher_id, level_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
		for _, lid := range levelIDs {
			if _, err := tx.Exec(ctx, insMap, teacherID, lid); err != nil {
				return "", "", fmt.Errorf("insert teacher_level: %w", err)
			}
		}
	}
	return teacherID, userID, nil
}

func (r *StudentsRepo) CreateStudentTx(ctx context.Context, tx pgx.Tx, username, passwordHash, passwordPlain, name, email, phone, nis, participantNo, jenjang, programID, levelID, groupID, googleID string) (studentID, userID string, err error) {
	const insUser = `INSERT INTO users (username, password_hash, password_plain, role, name, email, phone, google_id, is_active) VALUES ($1,$2,NULLIF($3,''),'student',$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),true) RETURNING id`
	if err := tx.QueryRow(ctx, insUser, username, passwordHash, passwordPlain, name, email, phone, googleID).Scan(&userID); err != nil {
		return "", "", fmt.Errorf("insert user: %w", err)
	}

	const insStudent = `
INSERT INTO students (user_id, nis, participant_no, jenjang, program_id, level_id, group_id)
VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,'')::uuid, NULLIF($6,'')::uuid, NULLIF($7,'')::uuid)
RETURNING id`
	if err := tx.QueryRow(ctx, insStudent, userID, nis, participantNo, jenjang, programID, levelID, groupID).Scan(&studentID); err != nil {
		return "", "", fmt.Errorf("insert student: %w", err)
	}

	return studentID, userID, nil
}

// CreateTeacher creates user(role=teacher) + teacher in one transaction.
func (r *TeachersRepo) CreateTeacher(ctx context.Context, username, passwordHash, passwordPlain, name, email, phone, nip, jenjang, googleID string, subjectIDs []string, groupIDs []string, levelIDs []string) (teacherID, userID string, err error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	teacherID, userID, err = r.CreateTeacherTx(ctx, tx, username, passwordHash, passwordPlain, name, email, phone, nip, jenjang, googleID, subjectIDs, groupIDs, levelIDs)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("commit: %w", err)
	}
	return teacherID, userID, nil
}

// CreateStudent creates user(role=student) + student in one transaction.
func (r *StudentsRepo) CreateStudent(ctx context.Context, username, passwordHash, passwordPlain, name, email, phone, nis, participantNo, jenjang, programID, levelID, groupID, googleID string) (studentID, userID string, err error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	studentID, userID, err = r.CreateStudentTx(ctx, tx, username, passwordHash, passwordPlain, name, email, phone, nis, participantNo, jenjang, programID, levelID, groupID, googleID)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("commit: %w", err)
	}
	return studentID, userID, nil
}

func (r *StudentsRepo) FindByNISOrUsername(ctx context.Context, key string) (Student, bool, error) {
	const q = `
SELECT s.id, u.id, u.username, COALESCE(u.password_plain,''), u.name, COALESCE(u.email,''), COALESCE(u.phone,''), s.nis, COALESCE(s.participant_no,''),
       COALESCE(s.jenjang,''), COALESCE(s.program_id::text,''), COALESCE(s.level_id::text,''), COALESCE(s.group_id::text,''),
       COALESCE(u.photo_url,''), u.is_active
FROM students s
JOIN users u ON u.id = s.user_id
WHERE s.nis = $1 OR u.username = $1
LIMIT 1`
	var it Student
	err := r.pool.QueryRow(ctx, q, key).Scan(&it.ID, &it.UserID, &it.Username, &it.PasswordPlain, &it.Name, &it.Email, &it.Phone, &it.NIS, &it.ParticipantNo, &it.Jenjang, &it.ProgramID, &it.LevelID, &it.GroupID, &it.PhotoURL, &it.IsActive)
	if err != nil {
		if isNoRows(err) {
			return Student{}, false, nil
		}
		return Student{}, false, fmt.Errorf("find student by nis/username: %w", err)
	}
	return it, true, nil
}

func (r *StudentsRepo) ListByTarget(ctx context.Context, levelID, groupID, studentID string) ([]Student, int, error) {
	const base = `
FROM students s
JOIN users u ON s.user_id = u.id
WHERE (
      (NULLIF($1,'') IS NULL AND NULLIF($2,'') IS NULL AND NULLIF($3,'') IS NULL) -- All students if no target
   OR ($1 <> '' AND s.level_id = $1::uuid)
   OR ($2 <> '' AND s.group_id = $2::uuid)
   OR ($3 <> '' AND s.id = $3::uuid)
)
AND u.is_active = true`

	rows, err := r.pool.Query(ctx, `
SELECT s.id::text,
       s.user_id::text,
       u.username,
       COALESCE(u.password_plain,''),
       u.name,
       COALESCE(u.email,''),
       COALESCE(u.phone,''),
       s.nis,
       COALESCE(s.participant_no,''),
       COALESCE(s.jenjang,''),
       COALESCE(s.program_id::text,''),
       COALESCE(s.level_id::text,''),
       COALESCE(s.group_id::text,''),
       COALESCE(u.photo_url,''),
       u.is_active
`+base, strings.TrimSpace(levelID), strings.TrimSpace(groupID), strings.TrimSpace(studentID))
	if err != nil {
		return nil, 0, fmt.Errorf("list students by target: %w", err)
	}
	defer rows.Close()

	var items []Student
	for rows.Next() {
		var it Student
		if err := rows.Scan(
			&it.ID,
			&it.UserID,
			&it.Username,
			&it.PasswordPlain,
			&it.Name,
			&it.Email,
			&it.Phone,
			&it.NIS,
			&it.ParticipantNo,
			&it.Jenjang,
			&it.ProgramID,
			&it.LevelID,
			&it.GroupID,
			&it.PhotoURL,
			&it.IsActive,
		); err != nil {
			return nil, 0, fmt.Errorf("scan student: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+base, strings.TrimSpace(levelID), strings.TrimSpace(groupID), strings.TrimSpace(studentID)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count students by target: %w", err)
	}

	return items, total, nil
}
