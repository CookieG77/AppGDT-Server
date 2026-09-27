// Package repository contains the data access layer, which runs the SQL
// queries and translates database errors into domain errors.
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error code for a unique constraint violation.
// See https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgUniqueViolation = "23505"

// isUniqueViolation reports whether err is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
