-- Blockchain schema tables

-- Transactions table
CREATE TABLE blockchain.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id),
    tx_id VARCHAR(255) UNIQUE NOT NULL, -- Fabric transaction ID
    channel_name VARCHAR(100) NOT NULL,
    chaincode_name VARCHAR(100) NOT NULL,
    function_name VARCHAR(100) NOT NULL,
    args JSONB,
    transient_data JSONB,
    endorsing_orgs TEXT[],
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, submitted, committed, failed
    block_number BIGINT,
    error_message TEXT,
    submitted_at TIMESTAMPTZ,
    committed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Transaction status history table
CREATE TABLE blockchain.transaction_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES blockchain.transactions(id) ON DELETE CASCADE,
    previous_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for blockchain schema
CREATE INDEX idx_transactions_user_id ON blockchain.transactions(user_id);
CREATE INDEX idx_transactions_tx_id ON blockchain.transactions(tx_id);
CREATE INDEX idx_transactions_channel ON blockchain.transactions(channel_name);
CREATE INDEX idx_transactions_chaincode ON blockchain.transactions(chaincode_name);
CREATE INDEX idx_transactions_status ON blockchain.transactions(status);
CREATE INDEX idx_transactions_created_at ON blockchain.transactions(created_at DESC);

-- Composite index for common queries
CREATE INDEX idx_transactions_user_status ON blockchain.transactions(user_id, status, created_at DESC);
CREATE INDEX idx_transactions_channel_chaincode ON blockchain.transactions(channel_name, chaincode_name, created_at DESC);

CREATE INDEX idx_transaction_history_transaction_id ON blockchain.transaction_status_history(transaction_id);
CREATE INDEX idx_transaction_history_created_at ON blockchain.transaction_status_history(created_at DESC);

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON blockchain.transactions
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Trigger to create status history on status change
CREATE OR REPLACE FUNCTION blockchain.create_status_history()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO blockchain.transaction_status_history (
            transaction_id, previous_status, new_status, error_message, metadata
        ) VALUES (
            NEW.id, OLD.status, NEW.status, NEW.error_message, 
            jsonb_build_object('block_number', NEW.block_number)
        );
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER transactions_status_history AFTER UPDATE ON blockchain.transactions
    FOR EACH ROW EXECUTE FUNCTION blockchain.create_status_history();

