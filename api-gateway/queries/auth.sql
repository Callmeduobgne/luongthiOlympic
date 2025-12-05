-- Refresh Token queries only (API Key queries are in api_keys.sql)
-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, is_revoked, created_at, revoked_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token_hash, expires_at, is_revoked, created_at, revoked_at 
FROM refresh_tokens 
WHERE token_hash = $1 AND is_revoked = false 
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET is_revoked = true, revoked_at = NOW() WHERE token_hash = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET is_revoked = true, revoked_at = NOW() WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < NOW();

