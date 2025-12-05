-- Create admin user: callmeduongne
-- Password: callmenam1
-- This script creates an admin user in the auth.users table

-- Insert admin user
-- Password hash generated using bcrypt with cost 10 for password: callmenam1
INSERT INTO public.users (email, username, password_hash, role, msp_id, is_active) 
VALUES (
    'callmeduongne@ibn.vn',
    'callmeduongne',
    '$2b$12$CklpHxvC6YRONPlrjGPnQOiE1gt8XEqGcDX2KpIi.dKWJWqnRCuFO',  -- bcrypt hash for: callmenam1
    'admin',
    'Org1MSP',
    TRUE
) ON CONFLICT (email) DO UPDATE SET
    username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;
