-- Drop audit tables
DROP FUNCTION IF EXISTS audit.create_monthly_partition();

DROP TABLE IF EXISTS audit.audit_logs_2025_02;
DROP TABLE IF EXISTS audit.audit_logs_2025_01;
DROP TABLE IF EXISTS audit.audit_logs_2024_12;
DROP TABLE IF EXISTS audit.audit_logs_2024_11;
DROP TABLE IF EXISTS audit.audit_logs;

