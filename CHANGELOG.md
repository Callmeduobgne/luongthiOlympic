# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-12-05

### Added
- **Go Chaincode**: Migrated from Node.js to Go chaincode for better performance and reliability.
- **CCaaS Architecture**: Implemented Chaincode as a Service to eliminate Docker-in-Docker issues.
- **Deployment Scripts**:
  - `scripts/generate-crypto.sh` - Generate crypto materials and channel artifacts from scratch.
  - `scripts/init-database.sh` - Initialize PostgreSQL database and create admin user.
  - `scripts/join-orderers.sh` - Join all orderers to channel.
  - `scripts/deploy-chaincode-ccaas.sh` - Automated CCaaS chaincode deployment.
  - `scripts/test-chaincode.sh` - Test chaincode functions.
- **SETUP.md**: Comprehensive A-Z setup guide for new deployments.
- **Admin User Setup**: Automatic admin user creation with `admin@ibn.vn` / `Admin123!@#`.

### Changed
- Chaincode now runs as external container (CCaaS) instead of peer-managed.
- Updated README.md with new Quick Start guide and script references.
- Database schema uses `public` schema for backend compatibility.

### Fixed
- Docker-in-Docker issues when deploying chaincode.
- Database migration schema mismatch (`auth` vs `public`).
- Orderer channel join process after network restart.

### Removed
- Node.js chaincode (teaTraceCC) - replaced by Go version (teaTraceCC-go).
- `scripts/deploy-chaincode-go.sh` - replaced by CCaaS deployment script.

## [1.1.0] - 2025-12-03

### Added
- API Gateway load balancing with Nginx.
- Enhanced monitoring with Prometheus exporters.
- Bug report and feature request issue templates.

### Changed
- Upgraded Hyperledger Fabric to 2.5.9.
- Improved chaincode hash verification.

## [1.0.0] - 2025-12-01

### Added
- **Blockchain Network**: Hyperledger Fabric 2.5.9 with 3 Orderers, 3 Peers, 3 CouchDBs.
- **Chaincode**: TeaTraceCC v1.1.0 with complete traceability logic and hash verification.
- **Backend**: Go API with 85+ endpoints, JWT auth, Redis caching, OPA policy.
- **Frontend**: React 19 + TypeScript UI for farmers, verifiers, and consumers.
- **API Gateway**: Nginx load balancer with Fabric Gateway SDK integration.
- **Monitoring**: Prometheus and Grafana stack.
- **Documentation**: Comprehensive README, API docs, and deployment guides.
- **CI/CD**: GitHub Actions workflow for automated testing and building.
- **Release Automation**: Scripts for release preparation.

### Security
- Implemented JWT authentication with refresh tokens.
- Enabled TLS for all Fabric network communications.
- Integrated OPA for fine-grained authorization policies.

### Changed
- Migrated from development setup to production-ready Docker Compose configuration.
- Optimized database queries with indexing.
