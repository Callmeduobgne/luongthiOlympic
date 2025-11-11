#!/bin/bash

# Cleanup script để dọn dẹp toàn bộ Fabric network
# Sử dụng cẩn thận - script này sẽ xóa tất cả data!

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Confirm
warning "⚠️  WARNING: This will delete ALL Fabric network data!"
warning "This includes:"
warning "  - All Docker containers"
warning "  - All Docker volumes"
warning "  - All Docker networks"
warning "  - Channel artifacts"
warning "  - Genesis blocks"
read -p "Are you sure you want to continue? (yes/no): " confirm

if [ "${confirm}" != "yes" ]; then
    log "Cleanup cancelled"
    exit 0
fi

log "Starting cleanup..."

# Stop và remove containers
log "Stopping and removing containers..."
cd /home/exp2/ibn/core/docker 2>/dev/null || true
docker-compose down -v 2>/dev/null || true

# Remove Fabric-related containers
log "Removing Fabric containers..."
docker ps -a --filter "name=orderer" --filter "name=peer" --filter "name=couchdb" --format "{{.ID}}" | xargs -r docker rm -f

# Remove Fabric-related volumes
log "Removing Fabric volumes..."
docker volume ls --format "{{.Name}}" | grep -E "(orderer|peer|couchdb|ibn|fabric)" | xargs -r docker volume rm || true

# Remove Fabric networks
log "Removing Fabric networks..."
docker network ls --format "{{.Name}}" | grep -E "(fabric|ibn)" | xargs -r docker network rm || true

# Remove channel artifacts
log "Removing channel artifacts..."
find /home/exp2/ibn -type d -name "channel-artifacts" -exec rm -rf {} + 2>/dev/null || true
find /home/exp2/ibn -type f -name "*.block" -path "*/channel-artifacts/*" -delete 2>/dev/null || true
find /home/exp2/ibn -type f -name "*.tx" -path "*/channel-artifacts/*" -delete 2>/dev/null || true

# Remove genesis blocks
log "Removing genesis blocks..."
find /home/exp2/ibn -type d -name "system-genesis-block" -exec rm -rf {} + 2>/dev/null || true
find /home/exp2/ibn -type f -name "genesis.block" -delete 2>/dev/null || true

# Remove chaincode containers
log "Removing chaincode containers..."
docker ps -a --filter "name=dev-peer" --format "{{.ID}}" | xargs -r docker rm -f || true

# Cleanup monitoring và logging
log "Stopping monitoring và logging services..."
cd /home/exp2/ibn/monitoring 2>/dev/null && docker-compose -f docker-compose-monitoring.yml down -v 2>/dev/null || true
cd /home/exp2/ibn/logging 2>/dev/null && docker-compose -f docker-compose-logging.yml down -v 2>/dev/null || true

# Remove orphaned volumes (optional - be careful!)
read -p "Remove ALL unused volumes? (yes/no): " remove_all
if [ "${remove_all}" = "yes" ]; then
    warning "Removing all unused volumes..."
    docker volume prune -f
fi

# Summary
log "Cleanup completed!"
log ""
log "Remaining containers:"
docker ps -a | wc -l
log ""
log "Remaining volumes:"
docker volume ls | wc -l
log ""
log "Remaining networks:"
docker network ls | wc -l

log "✅ Cleanup finished successfully!"

