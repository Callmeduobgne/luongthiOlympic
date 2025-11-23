-- Drop blockchain tables
DROP TRIGGER IF EXISTS transactions_status_history ON blockchain.transactions;
DROP FUNCTION IF EXISTS blockchain.create_status_history();
DROP TRIGGER IF EXISTS update_transactions_updated_at ON blockchain.transactions;

DROP TABLE IF EXISTS blockchain.transaction_status_history;
DROP TABLE IF EXISTS blockchain.transactions;

