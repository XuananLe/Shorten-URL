// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const batchInsertURLs = `-- name: BatchInsertURLs :exec
INSERT INTO urls (shortened, original, clicks, created_at, expired_at, user_id)
SELECT unnest($1::text[]), 
       unnest($2::text[]), 
       unnest($3::bigint[]), 
       unnest($4::timestamptz[]), 
       unnest($5::timestamptz[]), 
       unnest($6::uuid[])
ON CONFLICT (shortened, user_id) DO NOTHING
`

type BatchInsertURLsParams struct {
	Column1 []string
	Column2 []string
	Column3 []int64
	Column4 []pgtype.Timestamptz
	Column5 []pgtype.Timestamptz
	Column6 []pgtype.UUID
}

func (q *Queries) BatchInsertURLs(ctx context.Context, arg BatchInsertURLsParams) error {
	_, err := q.db.Exec(ctx, batchInsertURLs,
		arg.Column1,
		arg.Column2,
		arg.Column3,
		arg.Column4,
		arg.Column5,
		arg.Column6,
	)
	return err
}

const deleteExpiredURLs = `-- name: DeleteExpiredURLs :exec
DELETE FROM urls 
WHERE expired_at < CURRENT_TIMESTAMP
`

func (q *Queries) DeleteExpiredURLs(ctx context.Context) error {
	_, err := q.db.Exec(ctx, deleteExpiredURLs)
	return err
}

const deleteURL = `-- name: DeleteURL :exec
DELETE FROM urls 
WHERE shortened = $1
`

func (q *Queries) DeleteURL(ctx context.Context, shortened string) error {
	_, err := q.db.Exec(ctx, deleteURL, shortened)
	return err
}

const getClicks = `-- name: GetClicks :one
SELECT clicks 
FROM urls 
WHERE shortened = $1
`

func (q *Queries) GetClicks(ctx context.Context, shortened string) (pgtype.Int8, error) {
	row := q.db.QueryRow(ctx, getClicks, shortened)
	var clicks pgtype.Int8
	err := row.Scan(&clicks)
	return clicks, err
}

const getExpiredURLs = `-- name: GetExpiredURLs :many
SELECT shortened, original, clicks, created_at, expired_at
FROM urls 
WHERE expired_at < CURRENT_TIMESTAMP
`

type GetExpiredURLsRow struct {
	Shortened string
	Original  string
	Clicks    pgtype.Int8
	CreatedAt pgtype.Timestamptz
	ExpiredAt pgtype.Timestamptz
}

func (q *Queries) GetExpiredURLs(ctx context.Context) ([]GetExpiredURLsRow, error) {
	rows, err := q.db.Query(ctx, getExpiredURLs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetExpiredURLsRow
	for rows.Next() {
		var i GetExpiredURLsRow
		if err := rows.Scan(
			&i.Shortened,
			&i.Original,
			&i.Clicks,
			&i.CreatedAt,
			&i.ExpiredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOriginated = `-- name: GetOriginated :one
SELECT shortened, original, clicks, created_at, expired_at, user_id
FROM urls 
WHERE shortened = $1
`

func (q *Queries) GetOriginated(ctx context.Context, shortened string) (Url, error) {
	row := q.db.QueryRow(ctx, getOriginated, shortened)
	var i Url
	err := row.Scan(
		&i.Shortened,
		&i.Original,
		&i.Clicks,
		&i.CreatedAt,
		&i.ExpiredAt,
		&i.UserID,
	)
	return i, err
}

const getURLsByUser = `-- name: GetURLsByUser :many
SELECT shortened, original, clicks, created_at, expired_at
FROM urls 
WHERE user_id = $1
ORDER BY created_at DESC
`

type GetURLsByUserRow struct {
	Shortened string
	Original  string
	Clicks    pgtype.Int8
	CreatedAt pgtype.Timestamptz
	ExpiredAt pgtype.Timestamptz
}

func (q *Queries) GetURLsByUser(ctx context.Context, userID pgtype.UUID) ([]GetURLsByUserRow, error) {
	rows, err := q.db.Query(ctx, getURLsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetURLsByUserRow
	for rows.Next() {
		var i GetURLsByUserRow
		if err := rows.Scan(
			&i.Shortened,
			&i.Original,
			&i.Clicks,
			&i.CreatedAt,
			&i.ExpiredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const incrementClicks = `-- name: IncrementClicks :exec
UPDATE urls 
SET clicks = clicks + 1 
WHERE shortened = $1
`

func (q *Queries) IncrementClicks(ctx context.Context, shortened string) error {
	_, err := q.db.Exec(ctx, incrementClicks, shortened)
	return err
}

const insertURL = `-- name: InsertURL :one
INSERT INTO urls (shortened, original, clicks, created_at, expired_at, user_id)
VALUES ($1, $2, 0, DEFAULT, DEFAULT, $3)
RETURNING shortened, original, clicks, created_at, expired_at, user_id
`

type InsertURLParams struct {
	Shortened string
	Original  string
	UserID    pgtype.UUID
}

func (q *Queries) InsertURL(ctx context.Context, arg InsertURLParams) (Url, error) {
	row := q.db.QueryRow(ctx, insertURL, arg.Shortened, arg.Original, arg.UserID)
	var i Url
	err := row.Scan(
		&i.Shortened,
		&i.Original,
		&i.Clicks,
		&i.CreatedAt,
		&i.ExpiredAt,
		&i.UserID,
	)
	return i, err
}

const insertUser = `-- name: InsertUser :exec
INSERT INTO users (user_id) VALUES ($1)
`

func (q *Queries) InsertUser(ctx context.Context, userID pgtype.UUID) error {
	_, err := q.db.Exec(ctx, insertUser, userID)
	return err
}

const isURLExpired = `-- name: IsURLExpired :one
SELECT CASE WHEN expired_at < CURRENT_TIMESTAMP THEN TRUE ELSE FALSE END AS is_expired
FROM urls
WHERE shortened = $1
`

func (q *Queries) IsURLExpired(ctx context.Context, shortened string) (bool, error) {
	row := q.db.QueryRow(ctx, isURLExpired, shortened)
	var is_expired bool
	err := row.Scan(&is_expired)
	return is_expired, err
}

const searchByOriginalURL = `-- name: SearchByOriginalURL :many
SELECT shortened, original, clicks, created_at, expired_at
FROM urls 
WHERE original LIKE '%' || $1 || '%'
`

type SearchByOriginalURLRow struct {
	Shortened string
	Original  string
	Clicks    pgtype.Int8
	CreatedAt pgtype.Timestamptz
	ExpiredAt pgtype.Timestamptz
}

func (q *Queries) SearchByOriginalURL(ctx context.Context, dollar_1 pgtype.Text) ([]SearchByOriginalURLRow, error) {
	rows, err := q.db.Query(ctx, searchByOriginalURL, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchByOriginalURLRow
	for rows.Next() {
		var i SearchByOriginalURLRow
		if err := rows.Scan(
			&i.Shortened,
			&i.Original,
			&i.Clicks,
			&i.CreatedAt,
			&i.ExpiredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateExpirationDate = `-- name: UpdateExpirationDate :exec
UPDATE urls
SET expired_at = $2
WHERE shortened = $1
`

type UpdateExpirationDateParams struct {
	Shortened string
	ExpiredAt pgtype.Timestamptz
}

func (q *Queries) UpdateExpirationDate(ctx context.Context, arg UpdateExpirationDateParams) error {
	_, err := q.db.Exec(ctx, updateExpirationDate, arg.Shortened, arg.ExpiredAt)
	return err
}

const updateOriginalURL = `-- name: UpdateOriginalURL :exec
UPDATE urls
SET original = $2
WHERE shortened = $1
`

type UpdateOriginalURLParams struct {
	Shortened string
	Original  string
}

func (q *Queries) UpdateOriginalURL(ctx context.Context, arg UpdateOriginalURLParams) error {
	_, err := q.db.Exec(ctx, updateOriginalURL, arg.Shortened, arg.Original)
	return err
}

const updateURL = `-- name: UpdateURL :exec
UPDATE urls
SET clicks = $2
WHERE shortened = $1
`

type UpdateURLParams struct {
	Shortened string
	Clicks    pgtype.Int8
}

func (q *Queries) UpdateURL(ctx context.Context, arg UpdateURLParams) error {
	_, err := q.db.Exec(ctx, updateURL, arg.Shortened, arg.Clicks)
	return err
}
