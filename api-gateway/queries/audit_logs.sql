-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    user_id, 
    api_key_id, 
    action, 
    resource_type, 
    resource_id, 
    tx_id, 
    status, 
    details, 
    ip_address, 
    user_agent
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetAuditLog :one
SELECT * FROM audit_logs WHERE id = $1 LIMIT 1;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByAction :many
SELECT * FROM audit_logs
WHERE action = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByTxID :many
SELECT * FROM audit_logs
WHERE tx_id = $1
ORDER BY created_at DESC;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs;

-- name: CountAuditLogsByUser :one
SELECT COUNT(*) FROM audit_logs WHERE user_id = $1;

-- name: GetAuditLogsByDateRange :many
SELECT * FROM audit_logs
WHERE created_at >= $1 AND created_at <= $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteOldAuditLogs :exec
DELETE FROM audit_logs
WHERE created_at < $1;

