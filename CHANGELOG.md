# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
