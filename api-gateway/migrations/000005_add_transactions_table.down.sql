-- Rollback transactions table
DROP INDEX IF EXISTS idx_transaction_status_history_timestamp;
DROP INDEX IF EXISTS idx_transaction_status_history_transaction_id;
DROP INDEX IF EXISTS idx_transactions_block_number;
DROP INDEX IF EXISTS idx_transactions_timestamp;
DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_chaincode;
DROP INDEX IF EXISTS idx_transactions_channel;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_transactions_tx_id;

DROP TABLE IF EXISTS transaction_status_history;
DROP TABLE IF EXISTS transactions;



