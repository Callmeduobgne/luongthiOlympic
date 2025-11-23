-- Rollback RBAC/ABAC Schema Tables
-- Migration: 007_rbac_abac_tables

DROP TABLE IF EXISTS auth.user_permissions;
DROP TABLE IF EXISTS auth.user_roles;
DROP TABLE IF EXISTS auth.role_permissions;
DROP TABLE IF EXISTS auth.permissions;
DROP TABLE IF EXISTS auth.roles;

