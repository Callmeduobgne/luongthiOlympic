-- Initialize database with default data

-- Create default admin user
INSERT INTO users (email, password_hash, msp_id, role) 
VALUES (
    'admin@ibn.vn',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- password: admin123
    'Org1MSP',
    'admin'
) ON CONFLICT (email) DO NOTHING;

-- Create default farmer user
INSERT INTO users (email, password_hash, msp_id, role) 
VALUES (
    'farmer@ibn.vn',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- password: admin123
    'Org1MSP',
    'farmer'
) ON CONFLICT (email) DO NOTHING;

-- Create default verifier user
INSERT INTO users (email, password_hash, msp_id, role) 
VALUES (
    'verifier@ibn.vn',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- password: admin123
    'Org2MSP',
    'verifier'
) ON CONFLICT (email) DO NOTHING;

COMMIT;

