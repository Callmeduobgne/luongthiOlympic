-- name: GetAPIKey :one
SELECT * FROM api_keys WHERE id = $1 LIMIT 1;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 LIMIT 1;

-- name: ListAPIKeys :many
SELECT * FROM api_keys
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateAPIKey :one
INSERT INTO api_keys (user_id, key_hash, name, permissions, rate_limit, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateAPIKey :one
UPDATE api_keys
SET 
    name = COALESCE(sqlc.narg('name'), name),
    permissions = COALESCE(sqlc.narg('permissions'), permissions),
    rate_limit = COALESCE(sqlc.narg('rate_limit'), rate_limit),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys WHERE id = $1;

-- name: DeleteAPIKeysByUser :exec
DELETE FROM api_keys WHERE user_id = $1;

-- name: CountAPIKeysByUser :one
SELECT COUNT(*) FROM api_keys WHERE user_id = $1;

-- name: GetActiveAPIKeys :many
SELECT * FROM api_keys
WHERE is_active = TRUE
AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY created_at DESC;

