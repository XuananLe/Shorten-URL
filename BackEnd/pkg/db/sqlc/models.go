// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Url struct {
	Shortened string
	Original  string
	Clicks    pgtype.Int8
	CreatedAt pgtype.Timestamptz
	ExpiredAt pgtype.Timestamptz
	UserID    pgtype.UUID
}

type User struct {
	UserID pgtype.UUID
}
