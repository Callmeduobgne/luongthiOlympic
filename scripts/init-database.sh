#!/bin/bash

# Copyright (c) 2025 IBN Network
# Script to initialize database with migrations and seed admin user

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Configuration
DB_CONTAINER="ibn-postgres"
DB_USER="gateway"
DB_NAME="ibn_gateway"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@ibn.vn}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin123!@#}"

echo -e "${GREEN}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     IBN Network - Database Initialization Script          ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if PostgreSQL is running
log_info "Checking PostgreSQL container..."
if ! docker ps | grep -q "${DB_CONTAINER}"; then
    log_error "PostgreSQL container '${DB_CONTAINER}' is not running!"
    exit 1
fi

# Wait for PostgreSQL to be ready
log_info "Waiting for PostgreSQL to be ready..."
until docker exec ${DB_CONTAINER} pg_isready -U ${DB_USER} -d ${DB_NAME} > /dev/null 2>&1; do
    sleep 1
done
log_success "PostgreSQL is ready"

# Create public schema tables (Backend uses public schema)
# Update: Switched to using migrations source of truth. Manual table creation removed.

# Run auth schema migrations
log_info "Running auth schema migrations..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

for migration in ${PROJECT_ROOT}/backend/migrations/*.up.sql; do
    filename=$(basename "$migration")
    log_info "Running migration: $filename"
    docker exec -i ${DB_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} < "$migration" 2>/dev/null || true
done

log_success "Migrations completed"

# Check if admin user exists
log_info "Checking for existing admin user..."
ADMIN_EXISTS=$(docker exec ${DB_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -tAc "SELECT COUNT(*) FROM auth.users WHERE email='${ADMIN_EMAIL}';")

if [ "$ADMIN_EXISTS" -eq "0" ]; then
    log_info "Creating admin user via API..."
    
    # Wait for backend to be ready
    BACKEND_URL="http://localhost:9900"
    until curl -s "${BACKEND_URL}/health" > /dev/null 2>&1; do
        log_info "Waiting for backend..."
        sleep 2
    done
    
    # Register admin user
    RESPONSE=$(curl -s -X POST "${BACKEND_URL}/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${ADMIN_EMAIL}\",
            \"password\": \"${ADMIN_PASSWORD}\",
            \"full_name\": \"System Administrator\",
            \"role\": \"admin\"
        }")
    
    if echo "$RESPONSE" | grep -q '"id"'; then
        log_success "Admin user created: ${ADMIN_EMAIL}"
    else
        log_warn "Could not create admin user via API: $RESPONSE"
    fi
else
    log_info "Admin user already exists: ${ADMIN_EMAIL}"
fi

# Summary
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         DATABASE INITIALIZATION COMPLETE!                 ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Admin Credentials:${NC}"
echo -e "  Email:    ${ADMIN_EMAIL}"
echo -e "  Password: ${ADMIN_PASSWORD}"
echo ""
echo -e "${YELLOW}Test login:${NC}"
echo "  curl -X POST http://localhost:9900/api/v1/auth/login \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"email\": \"${ADMIN_EMAIL}\", \"password\": \"${ADMIN_PASSWORD}\"}'"
echo ""
