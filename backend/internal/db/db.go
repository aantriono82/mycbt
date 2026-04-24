package db

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	// Pool tuning (env-overridable):
	// - DB_MAX_CONNS (default: max(16, NumCPU*8))
	// - DB_MIN_CONNS (default: min(4, DB_MAX_CONNS))
	// - DB_MAX_CONN_LIFETIME_MINUTES (default: 30)
	// - DB_MAX_CONN_IDLE_MINUTES (default: 5)
	// - DB_HEALTHCHECK_SECONDS (default: 30)
	maxConns := int32(getEnvInt("DB_MAX_CONNS", max(16, runtime.NumCPU()*8)))
	minConns := int32(getEnvInt("DB_MIN_CONNS", min(4, int(maxConns))))
	if minConns > maxConns {
		minConns = maxConns
	}
	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = time.Duration(getEnvInt("DB_MAX_CONN_LIFETIME_MINUTES", 30)) * time.Minute
	cfg.MaxConnIdleTime = time.Duration(getEnvInt("DB_MAX_CONN_IDLE_MINUTES", 5)) * time.Minute
	cfg.HealthCheckPeriod = time.Duration(getEnvInt("DB_HEALTHCHECK_SECONDS", 30)) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
