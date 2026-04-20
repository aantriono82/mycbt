package pgerr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	CodeUniqueViolation     = "23505"
	CodeForeignKeyViolation = "23503"
)

func Code(err error) string {
	if err == nil {
		return ""
	}
	var e *pgconn.PgError
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}
