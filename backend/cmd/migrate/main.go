package main

import (
	"context"
	"fmt"
	"log"

	"atigacbt/backend/internal/config"
	"atigacbt/backend/internal/migrate"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	migrationsDir := config.ResolveAppPath("migrations")

	if err := migrate.Up(context.Background(), cfg.DatabaseURL, migrationsDir); err != nil {
		log.Fatalf("migrate up: %v", err)
	}

	fmt.Println("migrations applied")
}
