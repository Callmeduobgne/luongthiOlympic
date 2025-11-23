-- Add transactions table for transaction tracking
-- name: CreateTransactionsTable
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tx_id VARCHAR(255) UNIQUE NOT NULL, -- Fabric transaction ID
    channel_name VARCHAR(255) NOT NULL,
    chaincode_name VARCHAR(255) NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    args JSONB,
    transient_data JSONB,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'SUBMITTED', -- SUBMITTED, VALID, INVALID, FAILED
    block_number BIGINT,
    block_hash VARCHAR(255),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    error_message TEXT,
    endorsing_orgs TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Transaction status history (audit trail)
CREATE TABLE IF NOT EXISTS transaction_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    block_number BIGINT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    details JSONB
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_transactions_tx_id ON transactions(tx_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_channel ON transactions(channel_name);
CREATE INDEX IF NOT EXISTS idx_transactions_chaincode ON transactions(chaincode_name);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_block_number ON transactions(block_number);
CREATE INDEX IF NOT EXISTS idx_transaction_status_history_transaction_id ON transaction_status_history(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_status_history_timestamp ON transaction_status_history(timestamp DESC);



