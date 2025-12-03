-- Drop version management tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_version_release_notes_updated_at ON blockchain.version_release_notes;
DROP TRIGGER IF EXISTS update_version_tags_updated_at ON blockchain.version_tags;

-- Drop functions
DROP FUNCTION IF EXISTS blockchain.get_latest_version(VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS blockchain.compare_semantic_versions(VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS blockchain.parse_semantic_version(VARCHAR);

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS blockchain.version_comparisons CASCADE;
DROP TABLE IF EXISTS blockchain.version_release_notes CASCADE;
DROP TABLE IF EXISTS blockchain.version_dependencies CASCADE;
DROP TABLE IF EXISTS blockchain.version_tags CASCADE;

