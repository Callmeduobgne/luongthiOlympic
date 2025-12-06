#!/bin/bash

# Copyright (c) 2025 IBN Network
# Script to deploy teaTraceCC-go as Chaincode as a Service (CCaaS)
# This method avoids Docker-in-Docker issues completely

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

# Configuration
CHAINCODE_NAME="teaTraceCC"
CHAINCODE_VERSION="${CHAINCODE_VERSION:-1.2.0}"
CHAINCODE_LABEL="${CHAINCODE_NAME}_${CHAINCODE_VERSION}"
CHANNEL_NAME="ibnchannel"
CHAINCODE_DIR="${PROJECT_ROOT}/teaTraceCC-go"
CHAINCODE_CONTAINER_NAME="teatracecc-go"
CHAINCODE_IMAGE="teatracecc-go:${CHAINCODE_VERSION}"
CHAINCODE_ADDRESS="${CHAINCODE_CONTAINER_NAME}:9999"
NETWORK_NAME="ibn-network"

ADMIN_SERVICE_URL="${ADMIN_SERVICE_URL:-http://localhost:9902}"
ADMIN_API_KEY="${ADMIN_API_KEY:-admin-service-secret-key-change-in-production-min-32-chars}"

# Functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "\n${CYAN}========== $1 ==========${NC}"; }

check_dependencies() {
    log_step "Checking dependencies"
    
    command -v docker >/dev/null 2>&1 || { log_error "Docker is required but not installed."; exit 1; }
    command -v curl >/dev/null 2>&1 || { log_error "curl is required but not installed."; exit 1; }
    
    # Check Docker running
    docker info >/dev/null 2>&1 || { log_error "Docker daemon is not running."; exit 1; }
    
    log_success "All dependencies OK"
}

check_network() {
    log_step "Checking Fabric network"
    
    # Check if network exists
    if ! docker network ls | grep -q "${NETWORK_NAME}"; then
        log_error "Docker network '${NETWORK_NAME}' not found. Start the network first."
        exit 1
    fi
    
    # Check if peer is running
    if ! docker ps | grep -q "peer0.org1.ibn.vn"; then
        log_error "Peer container not running. Start the network first."
        exit 1
    fi
    
    # Check if admin-service is running
    if ! curl -s "${ADMIN_SERVICE_URL}/health" >/dev/null 2>&1; then
        log_warn "Admin service might not be running at ${ADMIN_SERVICE_URL}"
    fi
    
    log_success "Fabric network is running"
}

build_chaincode_image() {
    log_step "Building Go chaincode Docker image"
    
    if [ ! -d "${CHAINCODE_DIR}" ]; then
        log_error "Chaincode directory not found: ${CHAINCODE_DIR}"
        exit 1
    fi
    
    cd "${CHAINCODE_DIR}"
    
    log_info "Building ${CHAINCODE_IMAGE}..."
    docker build -t "${CHAINCODE_IMAGE}" .
    
    log_success "Image ${CHAINCODE_IMAGE} built successfully"
    cd "${PROJECT_ROOT}"
}

create_ccaas_package() {
    log_step "Creating CCaaS package"
    
    # Create temp directory
    rm -rf /tmp/ccaas-package
    mkdir -p /tmp/ccaas-package
    
    # Create connection.json
    cat > /tmp/ccaas-package/connection.json << EOF
{
  "address": "${CHAINCODE_ADDRESS}",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF
    
    # Create metadata.json - type must be "ccaas" for CCaaS
    cat > /tmp/ccaas-package/metadata.json << EOF
{
  "type": "ccaas",
  "label": "${CHAINCODE_LABEL}"
}
EOF
    
    # Create code.tar.gz containing connection.json
    cd /tmp/ccaas-package
    tar czf code.tar.gz connection.json
    
    # Create final package
    tar czf "${CHAINCODE_NAME}-ccaas.tar.gz" code.tar.gz metadata.json
    
    cd "${PROJECT_ROOT}"
    
    log_success "CCaaS package created: /tmp/ccaas-package/${CHAINCODE_NAME}-ccaas.tar.gz"
    log_info "Package contents:"
    tar -tzf "/tmp/ccaas-package/${CHAINCODE_NAME}-ccaas.tar.gz"
}

install_chaincode() {
    log_step "Installing chaincode"
    
    # Copy package to admin-service
    log_info "Copying package to admin-service container..."
    docker cp "/tmp/ccaas-package/${CHAINCODE_NAME}-ccaas.tar.gz" admin-service:/tmp/
    
    # Install via API
    log_info "Installing chaincode via admin-service API..."
    INSTALL_RESPONSE=$(curl -s -X POST "${ADMIN_SERVICE_URL}/api/v1/chaincode/install" \
        -H "X-API-Key: ${ADMIN_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{\"packagePath\": \"/tmp/${CHAINCODE_NAME}-ccaas.tar.gz\", \"targets\": [\"peer0.org1.ibn.vn\"]}")
    
    log_info "Install response: ${INSTALL_RESPONSE}"
    
    # Extract package ID
    if echo "${INSTALL_RESPONSE}" | grep -q '"success":true'; then
        PACKAGE_ID=$(echo "${INSTALL_RESPONSE}" | grep -o '"packageId":"[^"]*' | cut -d'"' -f4)
    else
        # Check if already installed
        ERROR_MSG=$(echo "${INSTALL_RESPONSE}" | grep -o '"detail":"[^"]*' | cut -d'"' -f4)
        if echo "${ERROR_MSG}" | grep -q "already successfully installed"; then
            # Extract package ID from error message
            PACKAGE_ID=$(echo "${ERROR_MSG}" | grep -o "${CHAINCODE_LABEL}:[a-f0-9]*")
        else
            log_error "Installation failed: ${INSTALL_RESPONSE}"
            exit 1
        fi
    fi
    
    if [ -z "${PACKAGE_ID}" ]; then
        log_error "Could not extract package ID"
        exit 1
    fi
    
    # Save package ID
    echo "${PACKAGE_ID}" > "${CHAINCODE_DIR}/PACKAGE_ID.txt"
    
    log_success "Chaincode installed with Package ID: ${PACKAGE_ID}"
}

get_current_sequence() {
    log_info "Checking current chaincode sequence..."
    
    QUERY_RESULT=$(docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
        -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
        hyperledger/fabric-tools:2.5.9 \
        peer lifecycle chaincode querycommitted --channelID "${CHANNEL_NAME}" --name "${CHAINCODE_NAME}" 2>/dev/null || echo "")
    
    if echo "${QUERY_RESULT}" | grep -q "Sequence:"; then
        CURRENT_SEQ=$(echo "${QUERY_RESULT}" | grep -o "Sequence: [0-9]*" | cut -d' ' -f2)
        SEQUENCE=$((CURRENT_SEQ + 1))
        log_info "Current sequence: ${CURRENT_SEQ}, will use: ${SEQUENCE}"
    else
        SEQUENCE=1
        log_info "No committed chaincode found, will use sequence: 1"
    fi
}

approve_chaincode() {
    log_step "Approving chaincode for Org1MSP"
    
    log_info "Approving chaincode definition..."
    docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
        -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
        hyperledger/fabric-tools:2.5.9 \
        peer lifecycle chaincode approveformyorg \
            --channelID "${CHANNEL_NAME}" \
            --name "${CHAINCODE_NAME}" \
            --version "${CHAINCODE_VERSION}" \
            --package-id "${PACKAGE_ID}" \
            --sequence "${SEQUENCE}" \
            --tls \
            --cafile /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/ca.crt \
            --orderer orderer.ibn.vn:7050 \
            --ordererTLSHostnameOverride orderer.ibn.vn
    
    log_success "Chaincode approved"
}

commit_chaincode() {
    log_step "Committing chaincode definition"
    
    log_info "Committing chaincode definition to channel..."
    docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
        -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
        hyperledger/fabric-tools:2.5.9 \
        peer lifecycle chaincode commit \
            --channelID "${CHANNEL_NAME}" \
            --name "${CHAINCODE_NAME}" \
            --version "${CHAINCODE_VERSION}" \
            --sequence "${SEQUENCE}" \
            --tls \
            --cafile /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/ca.crt \
            --orderer orderer.ibn.vn:7050 \
            --ordererTLSHostnameOverride orderer.ibn.vn \
            --peerAddresses peer0.org1.ibn.vn:7051 \
            --tlsRootCertFiles /fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
    
    log_success "Chaincode committed"
}

start_chaincode_container() {
    log_step "Starting chaincode container"
    
    # Stop and remove existing container
    docker stop "${CHAINCODE_CONTAINER_NAME}" 2>/dev/null || true
    docker rm "${CHAINCODE_CONTAINER_NAME}" 2>/dev/null || true
    
    log_info "Starting chaincode container with ID: ${PACKAGE_ID}"
    
    docker run -d \
        --name "${CHAINCODE_CONTAINER_NAME}" \
        --network "${NETWORK_NAME}" \
        --restart unless-stopped \
        -e CHAINCODE_ID="${PACKAGE_ID}" \
        -e CORE_CHAINCODE_ID_NAME="${PACKAGE_ID}" \
        -e CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999 \
        "${CHAINCODE_IMAGE}" \
        ./teaTraceCC
    
    # Wait and check
    sleep 3
    
    if docker ps | grep -q "${CHAINCODE_CONTAINER_NAME}"; then
        log_success "Chaincode container is running"
        docker logs "${CHAINCODE_CONTAINER_NAME}" --tail 5
    else
        log_error "Chaincode container failed to start"
        docker logs "${CHAINCODE_CONTAINER_NAME}"
        exit 1
    fi
}

verify_deployment() {
    log_step "Verifying deployment"
    
    log_info "Querying committed chaincode..."
    docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
        -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
        hyperledger/fabric-tools:2.5.9 \
        peer lifecycle chaincode querycommitted --channelID "${CHANNEL_NAME}" --name "${CHAINCODE_NAME}"
    
    log_success "Deployment verified"
}

print_summary() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║           CHAINCODE DEPLOYMENT COMPLETE!                  ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Chaincode Details:${NC}"
    echo -e "  Name:        ${CHAINCODE_NAME}"
    echo -e "  Version:     ${CHAINCODE_VERSION}"
    echo -e "  Sequence:    ${SEQUENCE}"
    echo -e "  Channel:     ${CHANNEL_NAME}"
    echo -e "  Package ID:  ${PACKAGE_ID}"
    echo ""
    echo -e "${CYAN}Container Details:${NC}"
    echo -e "  Name:        ${CHAINCODE_CONTAINER_NAME}"
    echo -e "  Image:       ${CHAINCODE_IMAGE}"
    echo -e "  Address:     ${CHAINCODE_ADDRESS}"
    echo -e "  Network:     ${NETWORK_NAME}"
    echo ""
    echo -e "${GREEN}The chaincode is now running as CCaaS (Chaincode as a Service)${NC}"
    echo -e "${GREEN}No Docker-in-Docker issues!${NC}"
    echo ""
    echo -e "${YELLOW}To test the chaincode, run:${NC}"
    echo -e "  ./scripts/test-chaincode.sh"
    echo ""
}

# Main execution
main() {
    echo -e "${GREEN}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║     IBN Network - Chaincode CCaaS Deployment Script       ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    check_dependencies
    check_network
    build_chaincode_image
    create_ccaas_package
    install_chaincode
    get_current_sequence
    approve_chaincode
    commit_chaincode
    start_chaincode_container
    verify_deployment
    print_summary
}

# Run main function
main "$@"
