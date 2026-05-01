package pgerr

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestCode(t *testing.T) {
	t.Parallel()

	if got := Code(nil); got != "" {
		t.Fatalf("expected empty code for nil error, got %q", got)
	}

	pgErr := &pgconn.PgError{Code: CodeUniqueViolation}
	if got := Code(pgErr); got != CodeUniqueViolation {
		t.Fatalf("expected unique violation code, got %q", got)
	}

	wrapped := errors.New("outer: " + pgErr.Error())
	if got := Code(wrapped); got != "" {
		t.Fatalf("expected empty code for non-pg wrapped string error, got %q", got)
	}

	err := errors.Join(errors.New("x"), pgErr)
	if got := Code(err); got != CodeUniqueViolation {
		t.Fatalf("expected code from joined error, got %q", got)
	}
}
