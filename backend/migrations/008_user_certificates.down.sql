-- Rollback migration for user certificates
DROP TRIGGER IF EXISTS ensure_single_active_cert_trigger ON auth.user_certificates;
DROP FUNCTION IF EXISTS auth.ensure_single_active_cert();
DROP TRIGGER IF EXISTS update_user_certificates_updated_at ON auth.user_certificates;
DROP TABLE IF EXISTS auth.user_certificates;

