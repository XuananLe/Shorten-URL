package utils

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsPgUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)
	if !ok {
		return false
	}
	return pgErr.Code == "23505" // Postgres unique_violation error code
}
