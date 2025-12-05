-- name: CreateTransaction :one
INSERT INTO transactions (
    tx_id, channel_name, chaincode_name, function_name, args, transient_data,
    user_id, api_key_id, status, block_number, endorsing_orgs
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetTransactionByTxID :one
SELECT * FROM transactions WHERE tx_id = $1 LIMIT 1;

-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = $1 LIMIT 1;

-- name: UpdateTransactionStatus :one
UPDATE transactions
SET status = $2, block_number = $3, block_hash = $4, error_message = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE 
    ($1::text IS NULL OR $1 = '' OR channel_name = $1)
    AND ($2::text IS NULL OR $2 = '' OR chaincode_name = $2)
    AND ($3::text IS NULL OR $3 = '' OR status = $3)
    AND ($4::uuid IS NULL OR user_id = $4)
    AND ($5::timestamp IS NULL OR timestamp >= $5)
    AND ($6::timestamp IS NULL OR timestamp <= $6)
ORDER BY timestamp DESC
LIMIT $7 OFFSET $8;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions
WHERE 
    ($1::text IS NULL OR $1 = '' OR channel_name = $1)
    AND ($2::text IS NULL OR $2 = '' OR chaincode_name = $2)
    AND ($3::text IS NULL OR $3 = '' OR status = $3)
    AND ($4::uuid IS NULL OR user_id = $4)
    AND ($5::timestamp IS NULL OR timestamp >= $5)
    AND ($6::timestamp IS NULL OR timestamp <= $6);

-- name: AddTransactionStatusHistory :one
INSERT INTO transaction_status_history (transaction_id, status, block_number, details)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTransactionStatusHistory :many
SELECT * FROM transaction_status_history
WHERE transaction_id = $1
ORDER BY timestamp DESC;

