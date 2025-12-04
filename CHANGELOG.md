# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-12-04

### Fixed

- **TLS Certificate Mismatch:** Fixed critical issue where genesis block contained outdated TLS certificates, causing orderer TLS handshake failures and network panic
  - Automatically remove genesis block when regenerating crypto material
  - Ensure genesis block is always regenerated after crypto material changes
  - Prevents "certificate signed by unknown authority" errors

- **TLS Client Certificate Authentication:** Fixed "Server TLS handshake failed with error remote error: tls: bad certificate"
  - Added `ORDERER_GENERAL_TLS_CLIENTROOTCAS` with peer TLS CA certificates
  - Set `ORDERER_GENERAL_TLS_CLIENTAUTHREQUIRED=false` to allow peer connections
  - Orderers now trust peer TLS certificates for delivery service
  - Fixed TLS CA cross-references between orderers and peers in docker-compose.yml
  - Mounted `tlscacerts/` directories to container paths for both orderers and peers

- **Chaincode Lifecycle Queries:** Fixed permission errors when querying installed chaincode and checking commit readiness
  - Use Admin MSP (`/tmp/admin_msp`) for all lifecycle queries (queryinstalled, checkcommitreadiness, querycommitted)
  - Automatically copy Admin MSP to all peer containers before lifecycle operations
  - Improved package ID extraction from multiple output formats

- **Peer Channel Join:** Fixed "The identity does not contain OU [ADMIN]" error when joining peers to channel
  - Updated `join_fabric_peers` to use Admin MSP instead of Peer MSP
  - Added step to copy Admin MSP to all peers before join operation
  - Changed `CORE_PEER_MSPCONFIGPATH` from `/etc/hyperledger/fabric/msp` to `/tmp/admin_msp`

- **Chaincode Build Process:** Enhanced Node.js/TypeScript chaincode build process
  - Ensure `package.json` is copied to `dist/` directory
  - Regenerate `package-lock.json` in `dist/` with production dependencies only
  - Fixes "npm ci can only install packages when package.json and package-lock.json are in sync" error

- **Script Path Resolution:** Fixed chaincode path resolution in deployment script
  - Corrected default chaincode path from `../teaTraceCC` to `teaTraceCC`
  - Improved path resolution logic to handle different script execution contexts

- **Script Integer Expression Errors:** Fixed "integer expression expected" errors in wait_for_fabric_network
  - Added proper handling for multiline grep output
  - Set default values for empty variables
  - Suppressed stderr for grep commands

### Changed

- **Setup Script:** Enhanced `scripts/setup.sh` with better error handling and validation
  - Added automatic cleanup of genesis block when crypto material is regenerated
  - Improved Admin MSP preparation for lifecycle operations
  - Better package ID extraction with multiple fallback methods
  - Added `fix_tls_ca_references()` function to copy TLS CAs cross-organization
  - Added `verify_tls_certificates()` function to validate certificates after generation
  - Added `wait_for_fabric_network()` function to ensure network readiness before bootstrap
  - Enhanced `cleanup_environment()` to force remove Docker volumes

- **Docker Compose:** Updated TLS configuration for all Fabric nodes
  - Orderers: Added peer TLS CA to `ORDERER_GENERAL_TLS_CLIENTROOTCAS`
  - Peers: Added orderer TLS CA to `CORE_PEER_TLS_CLIENTROOTCAS_FILES`
  - Mounted `tlscacerts/` directories as separate volumes for certificate access

- **View Logs Option:** Changed option 7 to show only error logs instead of full logs
  - Filters for ERROR, WARN, PANIC, FAILED keywords
  - Highlights errors with color for better visibility
  - Reduces noise from normal operation logs

### Added

- **Create Default Admin:** Added option 8 to create default admin user for DEV/DEMO
  - Creates admin user in both blockchain and database
  - Email: admin@ibn.vn, Password: Admin@123456
  - Only for development/staging environments, not production

### Security

- **Certificate Management:** Improved certificate lifecycle management
  - Genesis block is automatically invalidated when certificates change
  - Prevents using stale certificates that could cause security issues
  - TLS CA certificates are now properly shared between orderers and peers
  - Mutual TLS authentication configured correctly for all Fabric components

## [1.0.0] - 2025-12-01

### Added

- **Blockchain Network:** Hyperledger Fabric 2.5.9 with 3 Orderers, 3 Peers, 3 CouchDBs
  - Raft consensus for high availability
  - TLS encryption for all network communications
  - Multi-organization support structure

- **Chaincode:** TeaTraceCC v1.1.0 with complete traceability logic and hash verification
  - Full lifecycle tracking (harvest, processing, certification, distribution)
  - Hash-based integrity verification
  - Batch and package management
  - QR code generation support

- **Backend:** Go API with 85+ endpoints, JWT auth, Redis caching, OPA policy
  - RESTful API design
  - Multi-layer caching (L1 Memory + L2 Redis)
  - Domain-Driven Design architecture
  - Comprehensive error handling

- **Frontend:** React 19 + TypeScript UI for farmers, verifiers, and consumers
  - Modern UI with Tailwind CSS
  - Real-time data synchronization
  - Responsive design
  - Type-safe with TypeScript

- **API Gateway:** Nginx load balancer with Fabric Gateway SDK integration
  - Load balancing across multiple gateway instances
  - Health checks and failover
  - Request routing and rate limiting

- **Monitoring:** Prometheus and Grafana stack
  - Metrics collection for all services
  - Custom dashboards for network monitoring
  - Alert rules for critical events

- **Documentation:** Comprehensive README, API docs, and deployment guides
  - Setup guide with step-by-step instructions
  - API reference documentation
  - Architecture design documents
  - Troubleshooting guides

- **CI/CD:** GitHub Actions workflow for automated testing and building
  - Automated tests on pull requests
  - Docker image building
  - Release automation

- **Release Automation:** Scripts for release preparation
  - Version bumping
  - Changelog generation
  - Release notes creation

### Security

- Implemented JWT authentication with refresh tokens
- Enabled TLS for all Fabric network communications
- Integrated OPA for fine-grained authorization policies
- API key authentication for service-to-service communication
- Input validation at all layers
- Audit logging for compliance

### Changed

- Migrated from development setup to production-ready Docker Compose configuration
- Optimized database queries with indexing
- Improved error handling and logging
- Enhanced caching strategy for better performance

---

## Version History

- **1.0.1** (2025-12-04): Critical fixes for TLS certificates and chaincode deployment
- **1.0.0** (2025-12-01): Initial production release

## Types of Changes

- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` for vulnerability fixes
