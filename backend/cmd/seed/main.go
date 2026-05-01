package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"mycbt/backend/internal/config"
	"mycbt/backend/internal/db"
	"mycbt/backend/internal/model"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	d, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer d.Pool.Close()

	users := userrepo.New(d.Pool)
	if err := ensureAdmin(ctx, users, cfg); err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	if err := seedMasterData(ctx, d.Pool); err != nil {
		log.Fatalf("seed master data: %v", err)
	}
	log.Printf("seed selesai: master data masing-masing 10 record")
}

func ensureAdmin(ctx context.Context, users *userrepo.Repo, cfg config.Config) error {
	u, ok, err := users.GetByUsername(ctx, cfg.AdminUsername)
	if err != nil {
		return fmt.Errorf("check user: %w", err)
	}
	hash, err := authsvc.HashPassword(cfg.AdminPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if ok {
		log.Printf("user %q already exists; resetting password", cfg.AdminUsername)
		return users.UpdatePassword(ctx, u.ID, hash)
	}

	id, err := users.Create(ctx, model.User{
		Username:     cfg.AdminUsername,
		PasswordHash: hash,
		Role:         "admin",
		Name:         cfg.AdminName,
		Email:        cfg.AdminEmail,
		IsActive:     true,
	})
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	log.Printf("seeded admin user id=%s username=%s", id, cfg.AdminUsername)
	return nil
}

func seedMasterData(ctx context.Context, pool *pgxpool.Pool) error {
	teacherHash, err := authsvc.HashPassword("guru12345")
	if err != nil {
		return err
	}
	studentHash, err := authsvc.HashPassword("siswa12345")
	if err != nil {
		return err
	}

	programIDs := make([]string, 0, 10)
	levelIDs := make([]string, 0, 10)
	groupIDs := make([]string, 0, 10)
	subjectIDs := make([]string, 0, 10)
	teacherIDs := make([]string, 0, 10)
	studentIDs := make([]string, 0, 10)

	for i := 1; i <= 10; i++ {
		var id string
		code := fmt.Sprintf("PRG%02d", i)
		name := fmt.Sprintf("Program %02d", i)
		if err := pool.QueryRow(ctx, `
			INSERT INTO programs (code, name)
			VALUES ($1, $2)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, updated_at = now()
			RETURNING id
		`, code, name).Scan(&id); err != nil {
			return fmt.Errorf("upsert program %d: %w", i, err)
		}
		programIDs = append(programIDs, id)
	}

	for i := 1; i <= 10; i++ {
		var id string
		name := fmt.Sprintf("Level %02d", i)
		kelas := int16(i)
		if err := pool.QueryRow(ctx, `
			INSERT INTO levels (name, kelas)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET kelas = EXCLUDED.kelas, updated_at = now()
			RETURNING id
		`, name, kelas).Scan(&id); err != nil {
			return fmt.Errorf("upsert level %d: %w", i, err)
		}
		levelIDs = append(levelIDs, id)
	}

	for i := 1; i <= 10; i++ {
		var id string
		name := fmt.Sprintf("Kelas %02d", i)
		if err := pool.QueryRow(ctx, `
			INSERT INTO groups (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET updated_at = now()
			RETURNING id
		`, name).Scan(&id); err != nil {
			return fmt.Errorf("upsert group %d: %w", i, err)
		}
		groupIDs = append(groupIDs, id)
	}

	for i := 1; i <= 10; i++ {
		var id string
		code := fmt.Sprintf("SUB%02d", i)
		name := fmt.Sprintf("Mata Pelajaran %02d", i)
		if err := pool.QueryRow(ctx, `
			INSERT INTO subjects (code, name)
			VALUES ($1, $2)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, updated_at = now()
			RETURNING id
		`, code, name).Scan(&id); err != nil {
			return fmt.Errorf("upsert subject %d: %w", i, err)
		}
		subjectIDs = append(subjectIDs, id)
	}

	for i := 1; i <= 10; i++ {
		username := fmt.Sprintf("guru%02d", i)
		name := fmt.Sprintf("Guru %02d", i)
		email := fmt.Sprintf("guru%02d@example.com", i)
		nip := fmt.Sprintf("198700%04d", i)
		phone := fmt.Sprintf("08123000%04d", i)

		var userID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO users (username, password_hash, role, name, email, phone, is_active)
			VALUES ($1, $2, 'teacher', $3, $4, $5, true)
			ON CONFLICT (username) DO UPDATE SET
				password_hash = EXCLUDED.password_hash,
				role = EXCLUDED.role,
				name = EXCLUDED.name,
				email = EXCLUDED.email,
				phone = EXCLUDED.phone,
				is_active = true,
				updated_at = now()
			RETURNING id
		`, username, teacherHash, name, email, phone).Scan(&userID); err != nil {
			return fmt.Errorf("upsert teacher user %d: %w", i, err)
		}

		var teacherID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO teachers (user_id, nip)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET nip = EXCLUDED.nip, updated_at = now()
			RETURNING id
		`, userID, nip).Scan(&teacherID); err != nil {
			return fmt.Errorf("upsert teacher %d: %w", i, err)
		}
		teacherIDs = append(teacherIDs, teacherID)

		if _, err := pool.Exec(ctx, `
			INSERT INTO teacher_subjects (teacher_id, subject_id)
			VALUES ($1, $2)
			ON CONFLICT (teacher_id, subject_id) DO NOTHING
		`, teacherID, subjectIDs[(i-1)%len(subjectIDs)]); err != nil {
			return fmt.Errorf("attach teacher subject %d: %w", i, err)
		}
	}

	for i := 1; i <= 10; i++ {
		username := fmt.Sprintf("siswa%02d", i)
		name := fmt.Sprintf("Siswa %02d", i)
		email := fmt.Sprintf("siswa%02d@example.com", i)
		nis := fmt.Sprintf("2026%04d", i)
		phone := fmt.Sprintf("08213000%04d", i)
		jenjang := fmt.Sprintf("Jenjang %02d", i)

		var userID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO users (username, password_hash, role, name, email, phone, is_active)
			VALUES ($1, $2, 'student', $3, $4, $5, true)
			ON CONFLICT (username) DO UPDATE SET
				password_hash = EXCLUDED.password_hash,
				role = EXCLUDED.role,
				name = EXCLUDED.name,
				email = EXCLUDED.email,
				phone = EXCLUDED.phone,
				is_active = true,
				updated_at = now()
			RETURNING id
		`, username, studentHash, name, email, phone).Scan(&userID); err != nil {
			return fmt.Errorf("upsert student user %d: %w", i, err)
		}

		var studentID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO students (user_id, nis, program_id, level_id, group_id, jenjang)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id) DO UPDATE SET
				nis = EXCLUDED.nis,
				program_id = EXCLUDED.program_id,
				level_id = EXCLUDED.level_id,
				group_id = EXCLUDED.group_id,
				jenjang = EXCLUDED.jenjang,
				updated_at = now()
			RETURNING id
		`, userID, nis, programIDs[(i-1)%len(programIDs)], levelIDs[(i-1)%len(levelIDs)], groupIDs[(i-1)%len(groupIDs)], jenjang).Scan(&studentID); err != nil {
			return fmt.Errorf("upsert student %d: %w", i, err)
		}
		studentIDs = append(studentIDs, studentID)
	}

	if _, err := pool.Exec(ctx, `DELETE FROM announcements WHERE title LIKE '[SEED] %'`); err != nil {
		return fmt.Errorf("cleanup announcements: %w", err)
	}
	for i := 1; i <= 10; i++ {
		title := fmt.Sprintf("[SEED] Pengumuman %02d", i)
		body := fmt.Sprintf("Ini adalah pengumuman dummy nomor %02d untuk data seed.", i)
		publishedAt := time.Now().Add(time.Duration(-i) * time.Hour)
		if _, err := pool.Exec(ctx, `
			INSERT INTO announcements (
				title, body, category, is_active, published_at, target_level_id, target_group_id, target_student_id, created_by_user_id
			)
			VALUES ($1, $2, 'informasi', true, $3, $4, $5, $6, NULL)
		`, title, body, publishedAt, levelIDs[(i-1)%len(levelIDs)], groupIDs[(i-1)%len(groupIDs)], studentIDs[(i-1)%len(studentIDs)]); err != nil {
			return fmt.Errorf("insert announcement %d: %w", i, err)
		}
	}

	return nil
}
