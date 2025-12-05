-- Drop approval workflow tables

-- Drop triggers first
DROP TRIGGER IF EXISTS approval_vote_status_update ON blockchain.approval_votes;
DROP TRIGGER IF EXISTS update_approval_policies_updated_at ON blockchain.approval_policies;
DROP TRIGGER IF EXISTS update_approval_requests_updated_at ON blockchain.approval_requests;

-- Drop functions
DROP FUNCTION IF EXISTS blockchain.update_approval_status();
DROP FUNCTION IF EXISTS blockchain.check_approval_status(UUID);

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS blockchain.approval_votes CASCADE;
DROP TABLE IF EXISTS blockchain.approval_requests CASCADE;
DROP TABLE IF EXISTS blockchain.approval_policies CASCADE;

