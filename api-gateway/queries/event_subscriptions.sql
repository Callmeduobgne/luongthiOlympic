-- name: CreateEventSubscription :one
INSERT INTO event_subscriptions (
    user_id, api_key_id, name, type, channel_name, chaincode_name,
    event_name, webhook_url, webhook_secret, filters, active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetEventSubscriptionByID :one
SELECT * FROM event_subscriptions WHERE id = $1 LIMIT 1;

-- name: ListEventSubscriptions :many
SELECT * FROM event_subscriptions
WHERE 
    ($1::uuid IS NULL OR user_id = $1)
    AND ($2::uuid IS NULL OR api_key_id = $2)
    AND ($3::text IS NULL OR $3 = '' OR channel_name = $3)
    AND ($4::text IS NULL OR $4 = '' OR chaincode_name = $4)
    AND ($5::text IS NULL OR $5 = '' OR type = $5)
    AND ($6::bool IS NULL OR active = $6)
ORDER BY created_at DESC
LIMIT $7 OFFSET $8;

-- name: CountEventSubscriptions :one
SELECT COUNT(*) FROM event_subscriptions
WHERE 
    ($1::uuid IS NULL OR user_id = $1)
    AND ($2::uuid IS NULL OR api_key_id = $2)
    AND ($3::text IS NULL OR $3 = '' OR channel_name = $3)
    AND ($4::text IS NULL OR $4 = '' OR chaincode_name = $4)
    AND ($5::text IS NULL OR $5 = '' OR type = $5)
    AND ($6::bool IS NULL OR active = $6);

-- name: UpdateEventSubscription :one
UPDATE event_subscriptions
SET 
    name = COALESCE($2, name),
    active = COALESCE($3, active),
    webhook_url = COALESCE($4, webhook_url),
    webhook_secret = COALESCE($5, webhook_secret),
    filters = COALESCE($6, filters),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEventSubscription :exec
DELETE FROM event_subscriptions WHERE id = $1;

-- name: GetActiveSubscriptionsByChannelAndChaincode :many
SELECT * FROM event_subscriptions
WHERE active = true
    AND channel_name = $1
    AND ($2::text IS NULL OR $2 = '' OR chaincode_name = $2)
    AND ($3::text IS NULL OR $3 = '' OR event_name = $3 OR event_name IS NULL);

-- name: CreateWebhookDelivery :one
INSERT INTO webhook_deliveries (
    subscription_id, event_id, webhook_url, payload, status, attempts
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateWebhookDelivery :one
UPDATE webhook_deliveries
SET 
    status = $2,
    status_code = $3,
    response_body = $4,
    error_message = $5,
    attempts = $6,
    delivered_at = $7
WHERE id = $1
RETURNING *;

-- name: ListWebhookDeliveries :many
SELECT * FROM webhook_deliveries
WHERE subscription_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateWebSocketConnection :one
INSERT INTO websocket_connections (
    subscription_id, connection_id, user_id, ip_address, user_agent
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWebSocketConnectionByID :one
SELECT * FROM websocket_connections WHERE id = $1 LIMIT 1;

-- name: GetWebSocketConnectionByConnectionID :one
SELECT * FROM websocket_connections WHERE connection_id = $1 LIMIT 1;

-- name: UpdateWebSocketConnection :one
UPDATE websocket_connections
SET 
    disconnected_at = $2,
    last_ping_at = $3
WHERE id = $1
RETURNING *;

-- name: DeleteWebSocketConnection :exec
DELETE FROM websocket_connections WHERE id = $1;

-- name: ListWebSocketConnections :many
SELECT * FROM websocket_connections
WHERE subscription_id = $1
    AND disconnected_at IS NULL
ORDER BY connected_at DESC;


