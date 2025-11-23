-- Create 5 schemas for domain organization
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS blockchain;
CREATE SCHEMA IF NOT EXISTS events;
CREATE SCHEMA IF NOT EXISTS access;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set search path
ALTER DATABASE api_gateway SET search_path TO auth, blockchain, events, access, audit, public;

