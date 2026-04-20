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

	if _, ok, err := users.GetByUsername(ctx, cfg.AdminUsername); err != nil {
		log.Fatalf("check user: %v", err)
	} else if ok {
		log.Printf("user %q already exists; nothing to do", cfg.AdminUsername)
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
