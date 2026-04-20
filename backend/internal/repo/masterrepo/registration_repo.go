package masterrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistrationRequest struct {
	ID           string     `json:"id"`
	Role         string     `json:"role"`
	Username     string     `json:"username"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Status       string     `json:"status"`
	Note         string     `json:"note"`
	PasswordHash string     `json:"-"`
	DecidedAt    *time.Time `json:"decided_at,omitempty"`

	// Optional payload for approval-to-create.
	NIS          string     `json:"nis"`
	NIP          string     `json:"nip"`
	ProgramCode  string     `json:"program_code"`
	LevelName    string     `json:"level_name"`
	GroupName    string     `json:"group_name"`
	MapelCodes   string     `json:"mapel_codes"`
	GoogleID     string     `json:"google_id"`
	GooglePicture string    `json:"google_picture"`
	Phone        string     `json:"phone"`
	Jenjang      string     `json:"jenjang"`
	Gender       string     `json:"gender"`
	BirthDate    *time.Time `json:"birth_date"`
	SchoolName   string     `json:"school_name"`
	AcademicYear string     `json:"academic_year"`
	NISN         string     `json:"nisn"`
	NISSekolah   string     `json:"nis_sekolah"`
}

type RegistrationRepo struct{ pool *pgxpool.Pool }

func NewRegistrations(pool *pgxpool.Pool) *RegistrationRepo { return &RegistrationRepo{pool: pool} }

func (r *RegistrationRepo) ListPending(ctx context.Context) ([]RegistrationRequest, error) {
	const q = `
	SELECT id, role, username, name, COALESCE(email,''), status, COALESCE(note,''),
	       COALESCE(nis,''), COALESCE(nip,''), COALESCE(program_code,''), COALESCE(level_name,''), COALESCE(group_name,''), COALESCE(mapel_codes,''),
	       COALESCE(google_id,''), COALESCE(google_picture,''), COALESCE(phone,''), COALESCE(jenjang,''), COALESCE(gender,''), birth_date,
	       COALESCE(school_name,''), COALESCE(academic_year,''), COALESCE(nisn,''), COALESCE(nis_sekolah,'')
	FROM registration_requests
	WHERE status = 'pending'
	ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list pending registrations: %w", err)
	}
	defer rows.Close()

	out := []RegistrationRequest{}
	for rows.Next() {
		var it RegistrationRequest
		if err := rows.Scan(
			&it.ID, &it.Role, &it.Username, &it.Name, &it.Email, &it.Status, &it.Note,
			&it.NIS, &it.NIP, &it.ProgramCode, &it.LevelName, &it.GroupName, &it.MapelCodes,
			&it.GoogleID, &it.GooglePicture, &it.Phone, &it.Jenjang, &it.Gender, &it.BirthDate,
			&it.SchoolName, &it.AcademicYear, &it.NISN, &it.NISSekolah,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *RegistrationRepo) List(ctx context.Context, status string, role string, query string, limit int, offset int) ([]RegistrationRequest, int, error) {
	const base = `
FROM registration_requests
WHERE ($1 = '' OR status = $1)
  AND ($2 = '' OR role = $2)
  AND ($3 = '' OR username ILIKE '%'||$3||'%' OR name ILIKE '%'||$3||'%' OR COALESCE(email,'') ILIKE '%'||$3||'%' OR COALESCE(nis,'') ILIKE '%'||$3||'%' OR COALESCE(nip,'') ILIKE '%'||$3||'%')`

	const q = `
SELECT id, role, username, name, COALESCE(email,''), status, COALESCE(note,''), decided_at,
       COALESCE(nis,''), COALESCE(nip,''), COALESCE(program_code,''), COALESCE(level_name,''), COALESCE(group_name,''), COALESCE(mapel_codes,''),
       COALESCE(google_id,''), COALESCE(google_picture,''), COALESCE(phone,''), COALESCE(jenjang,''), COALESCE(gender,''), birth_date,
       COALESCE(school_name,''), COALESCE(academic_year,''), COALESCE(nisn,''), COALESCE(nis_sekolah,'')
` + base + `
ORDER BY created_at DESC
LIMIT $4 OFFSET $5`

	rows, err := r.pool.Query(ctx, q, status, role, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list registrations: %w", err)
	}
	defer rows.Close()

	out := []RegistrationRequest{}
	for rows.Next() {
		var it RegistrationRequest
		if err := rows.Scan(
			&it.ID, &it.Role, &it.Username, &it.Name, &it.Email, &it.Status, &it.Note, &it.DecidedAt,
			&it.NIS, &it.NIP, &it.ProgramCode, &it.LevelName, &it.GroupName, &it.MapelCodes,
			&it.GoogleID, &it.GooglePicture, &it.Phone, &it.Jenjang, &it.Gender, &it.BirthDate,
			&it.SchoolName, &it.AcademicYear, &it.NISN, &it.NISSekolah,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	const qc = `SELECT COUNT(*) ` + base
	var total int
	if err := r.pool.QueryRow(ctx, qc, status, role, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count registrations: %w", err)
	}

	return out, total, nil
}

func (r *RegistrationRepo) Get(ctx context.Context, id string) (RegistrationRequest, bool, error) {
	const q = `
SELECT id, role, username, name, COALESCE(email,''), status, COALESCE(note,''), decided_at,
       COALESCE(password_hash,''),
       COALESCE(nis,''), COALESCE(nip,''), COALESCE(program_code,''), COALESCE(level_name,''), COALESCE(group_name,''), COALESCE(mapel_codes,''),
       COALESCE(google_id,''), COALESCE(google_picture,''), COALESCE(phone,''), COALESCE(jenjang,''), COALESCE(gender,''), birth_date,
       COALESCE(school_name,''), COALESCE(academic_year,''), COALESCE(nisn,''), COALESCE(nis_sekolah,'')
FROM registration_requests
WHERE id = $1
LIMIT 1`
	var it RegistrationRequest
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&it.ID, &it.Role, &it.Username, &it.Name, &it.Email, &it.Status, &it.Note, &it.DecidedAt,
		&it.PasswordHash,
		&it.NIS, &it.NIP, &it.ProgramCode, &it.LevelName, &it.GroupName, &it.MapelCodes,
		&it.GoogleID, &it.GooglePicture, &it.Phone, &it.Jenjang, &it.Gender, &it.BirthDate,
		&it.SchoolName, &it.AcademicYear, &it.NISN, &it.NISSekolah,
	)
	if err != nil {
		if isNoRows(err) {
			return RegistrationRequest{}, false, nil
		}
		return RegistrationRequest{}, false, fmt.Errorf("get registration: %w", err)
	}
	return it, true, nil
}

func (r *RegistrationRepo) Create(ctx context.Context, req RegistrationRequest) (string, error) {
	const q = `
INSERT INTO registration_requests (role, username, name, email, password_hash, nis, nip, program_code, level_name, group_name, mapel_codes, google_id, google_picture, phone, jenjang, gender, birth_date, school_name, academic_year, nisn, nis_sekolah, status)
VALUES ($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),NULLIF($11,''),NULLIF($12,''),NULLIF($13,''),NULLIF($14,''),NULLIF($15,''),NULLIF($16,''),$17,NULLIF($18,''),NULLIF($19,''),NULLIF($20,''),NULLIF($21,''),'pending')
RETURNING id`
	var id string
	if err := r.pool.QueryRow(
		ctx,
		q,
		req.Role, req.Username, req.Name, req.Email, req.PasswordHash, req.NIS, req.NIP, req.ProgramCode, req.LevelName, req.GroupName, req.MapelCodes, req.GoogleID, req.GooglePicture,
		req.Phone, req.Jenjang, req.Gender, req.BirthDate, req.SchoolName, req.AcademicYear, req.NISN, req.NISSekolah,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("create registration: %w", err)
	}
	return id, nil
}

func (r *RegistrationRepo) UpdatePending(ctx context.Context, req RegistrationRequest) (RegistrationRequest, bool, error) {
	const q = `
UPDATE registration_requests
SET role = $2,
    username = $3,
    name = $4,
    email = NULLIF($5,''),
    nis = NULLIF($6,''),
    nip = NULLIF($7,''),
    program_code = NULLIF($8,''),
    level_name = NULLIF($9,''),
    group_name = NULLIF($10,''),
    mapel_codes = NULLIF($11,'')
WHERE id = $1 AND status = 'pending'
RETURNING id, role, username, name, COALESCE(email,''), status, COALESCE(note,''), decided_at,
          COALESCE(password_hash,''),
          COALESCE(nis,''), COALESCE(nip,''), COALESCE(program_code,''), COALESCE(level_name,''), COALESCE(group_name,''), COALESCE(mapel_codes,''),
          COALESCE(google_id,''), COALESCE(google_picture,''), COALESCE(phone,''), COALESCE(jenjang,''), COALESCE(gender,''), birth_date,
          COALESCE(school_name,''), COALESCE(academic_year,''), COALESCE(nisn,''), COALESCE(nis_sekolah,'')`

	var it RegistrationRequest
	err := r.pool.QueryRow(
		ctx,
		q,
		req.ID,
		req.Role,
		req.Username,
		req.Name,
		req.Email,
		req.NIS,
		req.NIP,
		req.ProgramCode,
		req.LevelName,
		req.GroupName,
		req.MapelCodes,
	).Scan(
		&it.ID, &it.Role, &it.Username, &it.Name, &it.Email, &it.Status, &it.Note, &it.DecidedAt,
		&it.PasswordHash,
		&it.NIS, &it.NIP, &it.ProgramCode, &it.LevelName, &it.GroupName, &it.MapelCodes,
		&it.GoogleID, &it.GooglePicture, &it.Phone, &it.Jenjang, &it.Gender, &it.BirthDate,
		&it.SchoolName, &it.AcademicYear, &it.NISN, &it.NISSekolah,
	)
	if err != nil {
		if isNoRows(err) {
			return RegistrationRequest{}, false, nil
		}
		return RegistrationRequest{}, false, fmt.Errorf("update registration: %w", err)
	}
	return it, true, nil
}

func (r *RegistrationRepo) Decide(ctx context.Context, id string, status string, note string) error {
	const q = `
	UPDATE registration_requests
	SET status = $2, note = NULLIF($3,''), decided_at = now()
	WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, status, note)
	if err != nil {
		return fmt.Errorf("decide registration: %w", err)
	}
	return nil
}

func (r *RegistrationRepo) DecidePending(ctx context.Context, id string, status string, note string) (bool, error) {
	const q = `
UPDATE registration_requests
SET status = $2, note = NULLIF($3,''), decided_at = now()
WHERE id = $1 AND status = 'pending'`
	ct, err := r.pool.Exec(ctx, q, id, status, note)
	if err != nil {
		return false, fmt.Errorf("decide pending registration: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

func (r *RegistrationRepo) GetByGoogleID(ctx context.Context, googleID string) (RegistrationRequest, bool, error) {
	const q = `
SELECT id, role, username, name, COALESCE(email,''), status, COALESCE(note,''), decided_at,
       COALESCE(password_hash,''),
       COALESCE(nis,''), COALESCE(nip,''), COALESCE(program_code,''), COALESCE(level_name,''), COALESCE(group_name,''), COALESCE(mapel_codes,''),
       COALESCE(google_id,''), COALESCE(google_picture,''), COALESCE(phone,''), COALESCE(jenjang,''), COALESCE(gender,''), birth_date,
       COALESCE(school_name,''), COALESCE(academic_year,''), COALESCE(nisn,''), COALESCE(nis_sekolah,'')
FROM registration_requests
WHERE google_id = $1
ORDER BY created_at DESC
LIMIT 1`
	var it RegistrationRequest
	err := r.pool.QueryRow(ctx, q, googleID).Scan(
		&it.ID, &it.Role, &it.Username, &it.Name, &it.Email, &it.Status, &it.Note, &it.DecidedAt,
		&it.PasswordHash,
		&it.NIS, &it.NIP, &it.ProgramCode, &it.LevelName, &it.GroupName, &it.MapelCodes,
		&it.GoogleID, &it.GooglePicture, &it.Phone, &it.Jenjang, &it.Gender, &it.BirthDate,
		&it.SchoolName, &it.AcademicYear, &it.NISN, &it.NISSekolah,
	)
	if err != nil {
		if isNoRows(err) {
			return RegistrationRequest{}, false, nil
		}
		return RegistrationRequest{}, false, fmt.Errorf("get registration by google id: %w", err)
	}
	return it, true, nil
}

func (r *RegistrationRepo) GetByEmail(ctx context.Context, email string) (RegistrationRequest, bool, error) {
	const q = `
SELECT id, role, username, name, COALESCE(email,''), status, COALESCE(note,''), decided_at,
       COALESCE(password_hash,''),
       COALESCE(nis,''), COALESCE(nip,''), COALESCE(program_code,''), COALESCE(level_name,''), COALESCE(group_name,''), COALESCE(mapel_codes,''),
       COALESCE(google_id,''), COALESCE(google_picture,''), COALESCE(phone,''), COALESCE(jenjang,''), COALESCE(gender,''), birth_date,
       COALESCE(school_name,''), COALESCE(academic_year,''), COALESCE(nisn,''), COALESCE(nis_sekolah,'')
FROM registration_requests
WHERE email = $1
ORDER BY created_at DESC
LIMIT 1`
	var it RegistrationRequest
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&it.ID, &it.Role, &it.Username, &it.Name, &it.Email, &it.Status, &it.Note, &it.DecidedAt,
		&it.PasswordHash,
		&it.NIS, &it.NIP, &it.ProgramCode, &it.LevelName, &it.GroupName, &it.MapelCodes,
		&it.GoogleID, &it.GooglePicture, &it.Phone, &it.Jenjang, &it.Gender, &it.BirthDate,
		&it.SchoolName, &it.AcademicYear, &it.NISN, &it.NISSekolah,
	)
	if err != nil {
		if isNoRows(err) {
			return RegistrationRequest{}, false, nil
		}
		return RegistrationRequest{}, false, fmt.Errorf("get registration by email: %w", err)
	}
	return it, true, nil
}
