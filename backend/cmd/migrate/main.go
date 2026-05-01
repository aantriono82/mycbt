package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"atigacbt/backend/internal/config"
	"atigacbt/backend/internal/migrate"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}

	// This command is expected to run from `backend/`.
	migrationsDir := filepath.Join(root, "migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		log.Fatalf("migrations dir not found at %s (run this from backend/)", migrationsDir)
	}

	if err := migrate.Up(context.Background(), cfg.DatabaseURL, migrationsDir); err != nil {
		log.Fatalf("migrate up: %v", err)
	}

	fmt.Println("migrations applied")
}
