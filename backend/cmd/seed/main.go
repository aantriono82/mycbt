package main

import (
	"context"
	"log"

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

	u, ok, err := users.GetByUsername(ctx, cfg.AdminUsername)
	if err != nil {
		log.Fatalf("check user: %v", err)
	} else if ok {
		log.Printf("user %q already exists; resetting password", cfg.AdminUsername)
		hash, _ := authsvc.HashPassword(cfg.AdminPassword)
		if err := users.UpdatePassword(ctx, u.ID, hash); err != nil {
			log.Fatalf("reset password: %v", err)
		}
		return
	}

	hash, err := authsvc.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Fatalf("hash password: %v", err)
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
		log.Fatalf("create admin: %v", err)
	}

	log.Printf("seeded admin user id=%s username=%s", id, cfg.AdminUsername)
}
