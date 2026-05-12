package pgtest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Harness struct {
	Pool      *pgxpool.Pool
	adminPool *pgxpool.Pool
	Schema    string
}

func Setup(t *testing.T) *Harness {
	t.Helper()

	baseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if baseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	adminCfg, err := pgxpool.ParseConfig(baseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	adminCfg.MaxConns = 2
	adminPool, err := pgxpool.NewWithConfig(ctx, adminCfg)
	if err != nil {
		t.Fatalf("open admin pool: %v", err)
	}

	schema := fmt.Sprintf("it_%d", time.Now().UnixNano())
	if _, err := adminPool.Exec(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		adminPool.Close()
		t.Fatalf("create schema: %v", err)
	}

	testCfg, err := pgxpool.ParseConfig(baseURL)
	if err != nil {
		adminPool.Close()
		t.Fatalf("parse TEST_DATABASE_URL for test pool: %v", err)
	}
	if testCfg.ConnConfig.RuntimeParams == nil {
		testCfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	testCfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	testCfg.MaxConns = 4
	pool, err := pgxpool.NewWithConfig(ctx, testCfg)
	if err != nil {
		_, _ = adminPool.Exec(ctx, `DROP SCHEMA IF EXISTS "`+schema+`" CASCADE`)
		adminPool.Close()
		t.Fatalf("open test pool: %v", err)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		pool.Close()
		_, _ = adminPool.Exec(ctx, `DROP SCHEMA IF EXISTS "`+schema+`" CASCADE`)
		adminPool.Close()
		t.Fatalf("apply migrations: %v", err)
	}

	h := &Harness{Pool: pool, adminPool: adminPool, Schema: schema}
	t.Cleanup(func() {
		pool.Close()
		dropCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = adminPool.Exec(dropCtx, `DROP SCHEMA IF EXISTS "`+schema+`" CASCADE`)
		adminPool.Close()
	})
	return h
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := migrationFiles()
	if err != nil {
		return err
	}
	for _, path := range files {
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		sqlBody := string(body)
		if strings.TrimSpace(sqlBody) == "" {
			continue
		}
		if _, err := pool.Exec(ctx, sqlBody); err != nil {
			if isExtensionCreateRace(string(body), err) {
				// Keep applying the rest of migration statements when CREATE EXTENSION
				// races across parallel integration test processes.
				rest := withoutCreateExtension(sqlBody)
				if strings.TrimSpace(rest) == "" {
					continue
				}
				if _, restErr := pool.Exec(ctx, rest); restErr == nil {
					continue
				}
				continue
			}
			return fmt.Errorf("exec migration %s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

func isExtensionCreateRace(sql string, err error) bool {
	if !strings.Contains(strings.ToUpper(sql), "CREATE EXTENSION IF NOT EXISTS") {
		return false
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && strings.EqualFold(pgErr.ConstraintName, "pg_extension_name_index")
}

func withoutCreateExtension(sql string) string {
	lines := strings.Split(sql, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		upper := strings.ToUpper(strings.TrimSpace(line))
		if strings.HasPrefix(upper, "CREATE EXTENSION IF NOT EXISTS") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func migrationFiles() ([]string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("resolve current file")
	}
	backendDir := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	matches, err := filepath.Glob(filepath.Join(backendDir, "migrations", "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(matches)
	return matches, nil
}
