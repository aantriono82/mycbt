package migrate

import (
	"context"
	"testing"
)

func TestUp_InvalidDatabaseURL(t *testing.T) {
	t.Parallel()

	err := Up(context.Background(), "://bad-dsn", "/definitely/missing")
	if err == nil {
		t.Fatal("expected migrate.Up to fail for invalid database URL")
	}
}
