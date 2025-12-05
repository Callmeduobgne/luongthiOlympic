#!/bin/bash

# Copyright (c) 2025 IBN Network
# Script to generate crypto materials and channel artifacts from scratch

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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
log_step() { echo -e "\n${CYAN}========== $1 ==========${NC}"; }

# Check dependencies
check_dependencies() {
    log_step "Checking dependencies"
    
    command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed."; exit 1; }
    docker info >/dev/null 2>&1 || { log_error "Docker daemon is not running."; exit 1; }
    
    # Pull fabric-tools image if needed
    if ! docker images | grep -q "hyperledger/fabric-tools.*2.5.9"; then
        log_info "Pulling hyperledger/fabric-tools:2.5.9..."
        docker pull hyperledger/fabric-tools:2.5.9
    fi
    
    log_success "All dependencies OK"
}

# Check if crypto already exists
check_existing() {
    log_step "Checking existing crypto materials"
    
    if [ -d "${CORE_DIR}/organizations/ordererOrganizations" ] && [ -d "${CORE_DIR}/organizations/peerOrganizations" ]; then
        log_warn "Crypto materials already exist!"
        read -p "Do you want to regenerate? This will DELETE all existing crypto! (y/N): " confirm
        if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
            log_info "Aborted. Keeping existing crypto materials."
            exit 0
        fi
        
        log_info "Removing existing crypto materials..."
        rm -rf "${CORE_DIR}/organizations/ordererOrganizations"
        rm -rf "${CORE_DIR}/organizations/peerOrganizations"
    fi
}

# Generate crypto materials
generate_crypto() {
    log_step "Generating crypto materials with cryptogen"
    
    # Check if crypto-config.yaml exists
    if [ ! -f "${CORE_DIR}/crypto-config.yaml" ]; then
        log_error "crypto-config.yaml not found at ${CORE_DIR}/crypto-config.yaml"
        exit 1
    fi
    
    # Create organizations directory
    mkdir -p "${CORE_DIR}/organizations"
    
    # Run cryptogen
    docker run --rm \
        -v "${CORE_DIR}:/fabric" \
        -w /fabric \
        hyperledger/fabric-tools:2.5.9 \
        cryptogen generate --config=/fabric/crypto-config.yaml --output=/fabric/organizations
    
    log_success "Crypto materials generated"
    log_info "Output: ${CORE_DIR}/organizations/"
}

# Create config.yaml for MSPs (NodeOUs)
create_msp_configs() {
    log_step "Creating MSP config.yaml files"
    
    # Org1 MSP config
    cat > "${CORE_DIR}/organizations/peerOrganizations/org1.ibn.vn/msp/config.yaml" << 'EOF'
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: orderer
EOF
    
    # Copy config.yaml to all peer and user MSP directories
    for peer_dir in "${CORE_DIR}/organizations/peerOrganizations/org1.ibn.vn/peers/"*/msp; do
        cp "${CORE_DIR}/organizations/peerOrganizations/org1.ibn.vn/msp/config.yaml" "$peer_dir/"
    done
    
    for user_dir in "${CORE_DIR}/organizations/peerOrganizations/org1.ibn.vn/users/"*/msp; do
        cp "${CORE_DIR}/organizations/peerOrganizations/org1.ibn.vn/msp/config.yaml" "$user_dir/"
    done
    
    # Orderer MSP config
    cat > "${CORE_DIR}/organizations/ordererOrganizations/ibn.vn/msp/config.yaml" << 'EOF'
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: orderer
EOF
    
    # Copy config.yaml to all orderer MSP directories
    for orderer_dir in "${CORE_DIR}/organizations/ordererOrganizations/ibn.vn/orderers/"*/msp; do
        cp "${CORE_DIR}/organizations/ordererOrganizations/ibn.vn/msp/config.yaml" "$orderer_dir/"
    done
    
    for user_dir in "${CORE_DIR}/organizations/ordererOrganizations/ibn.vn/users/"*/msp; do
        cp "${CORE_DIR}/organizations/ordererOrganizations/ibn.vn/msp/config.yaml" "$user_dir/"
    done
    
    log_success "MSP config.yaml files created"
}

# Generate genesis block
generate_genesis_block() {
    log_step "Generating genesis block"
    
    # Check if configtx.yaml exists
    if [ ! -f "${CORE_DIR}/configtx/configtx.yaml" ]; then
        log_error "configtx.yaml not found at ${CORE_DIR}/configtx/configtx.yaml"
        exit 1
    fi
    
    # Create output directory
    mkdir -p "${CORE_DIR}/system-genesis-block"
    
    # Generate genesis block
    docker run --rm \
        -v "${CORE_DIR}:/fabric" \
        -w /fabric/configtx \
        hyperledger/fabric-tools:2.5.9 \
        configtxgen -profile RaftOrdererGenesis \
            -channelID system-channel \
            -outputBlock /fabric/system-genesis-block/genesis.block
    
    log_success "Genesis block created: ${CORE_DIR}/system-genesis-block/genesis.block"
}

# Generate channel artifacts
generate_channel_artifacts() {
    log_step "Generating channel artifacts"
    
    # Create output directory
    mkdir -p "${CORE_DIR}/channel-artifacts"
    
    # Generate channel transaction
    log_info "Creating channel transaction..."
    docker run --rm \
        -v "${CORE_DIR}:/fabric" \
        -w /fabric/configtx \
        hyperledger/fabric-tools:2.5.9 \
        configtxgen -profile ThreePeersChannel \
            -channelID ibnchannel \
            -outputCreateChannelTx /fabric/channel-artifacts/ibnchannel.tx
    
    # Generate anchor peer update
    log_info "Creating anchor peer update..."
    docker run --rm \
        -v "${CORE_DIR}:/fabric" \
        -w /fabric/configtx \
        hyperledger/fabric-tools:2.5.9 \
        configtxgen -profile ThreePeersChannel \
            -channelID ibnchannel \
            -outputAnchorPeersUpdate /fabric/channel-artifacts/Org1MSPanchors.tx \
            -asOrg Org1MSP
    
    # Generate channel block (for osnadmin join)
    log_info "Creating channel genesis block..."
    docker run --rm \
        -v "${CORE_DIR}:/fabric" \
        -w /fabric/configtx \
        hyperledger/fabric-tools:2.5.9 \
        configtxgen -profile ThreePeersChannel \
            -channelID ibnchannel \
            -outputBlock /fabric/channel-artifacts/ibnchannel.block
    
    log_success "Channel artifacts created:"
    ls -la "${CORE_DIR}/channel-artifacts/"
}

# Set permissions
set_permissions() {
    log_step "Setting file permissions"
    
    chmod -R 755 "${CORE_DIR}/organizations/" 2>/dev/null || true
    chmod -R 755 "${CORE_DIR}/channel-artifacts/" 2>/dev/null || true
    chmod -R 755 "${CORE_DIR}/system-genesis-block/" 2>/dev/null || true
    
    log_success "Permissions set"
}

# Verify all files
verify_files() {
    log_step "Verifying generated files"
    
    local errors=0
    
    # Check genesis block
    if [ -f "${CORE_DIR}/system-genesis-block/genesis.block" ]; then
        echo -e "  ${GREEN}✅${NC} genesis.block"
    else
        echo -e "  ${RED}❌${NC} genesis.block MISSING"
        ((errors++))
    fi
    
    # Check channel block
    if [ -f "${CORE_DIR}/channel-artifacts/ibnchannel.block" ]; then
        echo -e "  ${GREEN}✅${NC} ibnchannel.block"
    else
        echo -e "  ${RED}❌${NC} ibnchannel.block MISSING"
        ((errors++))
    fi
    
    # Check orderer TLS certs
    for orderer in orderer orderer1 orderer2; do
        if [ -f "${CORE_DIR}/organizations/ordererOrganizations/ibn.vn/orderers/${orderer}.ibn.vn/tls/server.crt" ]; then
            echo -e "  ${GREEN}✅${NC} ${orderer} TLS cert"
        else
            echo -e "  ${RED}❌${NC} ${orderer} TLS cert MISSING"
            ((errors++))
        fi
    done
    
    # Check peer TLS certs
    for peer in peer0 peer1 peer2; do
        if [ -f "${CORE_DIR}/organizations/peerOrganizations/org1.ibn.vn/peers/${peer}.org1.ibn.vn/tls/server.crt" ]; then
            echo -e "  ${GREEN}✅${NC} ${peer} TLS cert"
        else
            echo -e "  ${RED}❌${NC} ${peer} TLS cert MISSING"
            ((errors++))
        fi
    done
    
    # Check Admin MSP
    if [ -d "${CORE_DIR}/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp" ]; then
        echo -e "  ${GREEN}✅${NC} Admin MSP"
    else
        echo -e "  ${RED}❌${NC} Admin MSP MISSING"
        ((errors++))
    fi
    
    if [ $errors -gt 0 ]; then
        log_error "Verification failed with $errors errors"
        exit 1
    fi
    
    log_success "All files verified!"
}

# Print summary
print_summary() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         CRYPTO MATERIALS GENERATED SUCCESSFULLY!          ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Generated files:${NC}"
    echo -e "  • ${CORE_DIR}/organizations/     - MSP certificates"
    echo -e "  • ${CORE_DIR}/system-genesis-block/genesis.block"
    echo -e "  • ${CORE_DIR}/channel-artifacts/ibnchannel.block"
    echo -e "  • ${CORE_DIR}/channel-artifacts/ibnchannel.tx"
    echo -e "  • ${CORE_DIR}/channel-artifacts/Org1MSPanchors.tx"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "  1. Start the network:  docker compose up -d"
    echo -e "  2. Wait for healthy:   sleep 60"
    echo -e "  3. Join orderers:      ./scripts/join-orderers.sh"
    echo -e "  4. Deploy chaincode:   ./scripts/deploy-chaincode-ccaas.sh"
    echo ""
}

# Main execution
main() {
    echo -e "${GREEN}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║     IBN Network - Crypto & Channel Artifacts Generator    ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    cd "${PROJECT_ROOT}"
    
    check_dependencies
    check_existing
    generate_crypto
    create_msp_configs
    generate_genesis_block
    generate_channel_artifacts
    set_permissions
    verify_files
    print_summary
}

# Run main function
main "$@"
