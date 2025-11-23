-- Add blocks table for block tracking
-- Migration: 000006_add_blocks_table

CREATE TABLE IF NOT EXISTS blocks (
    id BIGSERIAL PRIMARY KEY,
    number BIGINT NOT NULL,
    hash VARCHAR(255) UNIQUE NOT NULL,
    previous_hash VARCHAR(255),
    data_hash VARCHAR(255),
    transaction_count INT DEFAULT 0,
    channel_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(channel_name, number)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_blocks_channel ON blocks(channel_name);
CREATE INDEX IF NOT EXISTS idx_blocks_number ON blocks(number DESC);
CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON blocks(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_blocks_channel_number ON blocks(channel_name, number DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_blocks_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at
CREATE TRIGGER update_blocks_updated_at BEFORE UPDATE ON blocks
    FOR EACH ROW EXECUTE FUNCTION update_blocks_updated_at();

