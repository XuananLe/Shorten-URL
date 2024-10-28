-- name: GetOriginated :one
SELECT shortened, original, clicks, created_at, expired_at, user_id
FROM urls 
WHERE shortened = $1;

-- name: InsertURL :one
INSERT INTO urls (shortened, original, clicks, created_at, expired_at, user_id)
VALUES ($1, $2, 0, DEFAULT, DEFAULT, $3)
RETURNING *;

-- name: IncrementClicks :exec
UPDATE urls 
SET clicks = clicks + 1 
WHERE shortened = $1;

-- name: GetClicks :one
SELECT clicks 
FROM urls 
WHERE shortened = $1;


-- name: UpdateURL :exec
UPDATE urls
SET clicks = $2
WHERE shortened = $1;

-- name: InsertUser :exec
INSERT INTO users (user_id) VALUES ($1);


-- name: GetURLsByUser :many
SELECT shortened, original, clicks, created_at, expired_at
FROM urls 
WHERE user_id = $1
ORDER BY created_at DESC;


-- name: DeleteExpiredURLs :exec
DELETE FROM urls 
WHERE expired_at < CURRENT_TIMESTAMP;

-- name: UpdateExpirationDate :exec
UPDATE urls
SET expired_at = $2
WHERE shortened = $1;

-- name: GetExpiredURLs :many
SELECT shortened, original, clicks, created_at, expired_at
FROM urls 
WHERE expired_at < CURRENT_TIMESTAMP;

-- name: IsURLExpired :one
SELECT CASE WHEN expired_at < CURRENT_TIMESTAMP THEN TRUE ELSE FALSE END AS is_expired
FROM urls
WHERE shortened = $1;

-- name: UpdateOriginalURL :exec
UPDATE urls
SET original = $2
WHERE shortened = $1;

-- name: DeleteURL :exec
DELETE FROM urls 
WHERE shortened = $1;

-- name: SearchByOriginalURL :many
SELECT shortened, original, clicks, created_at, expired_at
FROM urls 
WHERE original LIKE '%' || $1 || '%';

-- 