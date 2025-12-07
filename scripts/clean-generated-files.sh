#!/bin/bash

# Copyright (c) 2025 IBN Network
# Script to clean generated artifacts and docker resources

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CORE_DIR="${PROJECT_ROOT}/core"

# Functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo -e "${GREEN}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     IBN Network - Clean Up Script                         ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Ask for confirmation
echo -e "${YELLOW}WARNING: This script will remove:${NC}"
echo -e "  • All generated crypto materials (certificates)"
echo -e "  • Genesis blocks and channel artifacts"
echo -e "  • Docker containers and volumes (optional)"
echo ""
read -p "Are you sure you want to proceed? (y/N): " confirm

if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    log_info "Aborted."
    exit 0
fi

# Stop containers
log_info "Stopping Docker containers..."
if [ -f "${PROJECT_ROOT}/docker-compose.yml" ]; then
    cd "${PROJECT_ROOT}"
    docker compose down --volumes --remove-orphans || true
else
    log_warn "docker-compose.yml not found, skipping 'docker compose down'"
fi

# Remove generated files
log_info "Removing generated files..."

# Crypto materials
if [ -d "${CORE_DIR}/organizations" ]; then
    sudo rm -rf "${CORE_DIR}/organizations"
    log_success "Removed ${CORE_DIR}/organizations"
fi

# System Genesis Block
if [ -d "${CORE_DIR}/system-genesis-block" ]; then
    sudo rm -rf "${CORE_DIR}/system-genesis-block"
    log_success "Removed ${CORE_DIR}/system-genesis-block"
fi

# Channel Artifacts
if [ -d "${CORE_DIR}/channel-artifacts" ]; then
    sudo rm -rf "${CORE_DIR}/channel-artifacts"
    log_success "Removed ${CORE_DIR}/channel-artifacts"
fi

# Fabric CA server config (if any)
if [ -d "${CORE_DIR}/fabric-ca" ]; then
    sudo rm -rf "${CORE_DIR}/fabric-ca"
    log_success "Removed ${CORE_DIR}/fabric-ca"
fi

# Remove Docker volumes (redundant if 'down -v' worked, but good to be sure)
# Only remove volumes associated with this project if not removed by compose
log_info "Cleaning up leftover volumes..."
docker volume prune -f > /dev/null 2>&1

# Prune networks
docker network prune -f > /dev/null 2>&1

echo ""
echo -e "${GREEN}Clean up complete!${NC}"
echo "You can now run ./scripts/generate-crypto.sh to regenerate materials."
