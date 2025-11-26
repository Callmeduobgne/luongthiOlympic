# Changelog

All notable changes to the IBN Network project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-11-22

### Added
- **Initial Release** - IBN Network (ICTU Blockchain Network)
- **Frontend Layer**
  - React 19.2.0 với TypeScript 5.9.3
  - Vite 7.2.2 build tool
  - Tailwind CSS 3.4.18 cho styling
  - Zustand 5.0.8 cho state management
  - TanStack Query 5.90.8 cho data fetching
  - React Router DOM 7.9.5 cho routing
  - React Hook Form 7.66.0 cho form handling
  - Zod 4.1.12 cho schema validation
  - Axios 1.13.2 cho HTTP client

- **Backend Layer**
  - Go 1.24.6 backend API
  - Chi Router v5.2.3
  - PostgreSQL 16 với pgx/v5 5.7.6
  - Redis 9.16.0 cho caching
  - JWT authentication (golang-jwt/v5)
  - Zap 1.27.0 cho structured logging
  - Multi-layer caching (L1 Memory + L2 Redis)

- **API Gateway Layer**
  - Go 1.23.5 API Gateway
  - Fabric Gateway SDK v1.4.0
  - 50+ REST API endpoints
  - Transaction management
  - Event system với WebSocket support
  - Block explorer
  - Chaincode lifecycle management
  - Audit logging
  - Advanced metrics & monitoring
  - Circuit breaker pattern
  - Prometheus metrics
  - OpenTelemetry tracing

- **Network Layer (Blockchain)**
  - Hyperledger Fabric 2.5.9
  - Raft Consensus (3 orderers)
  - 3 Peer nodes (Org1MSP)
  - 3 CouchDB instances cho state database
  - teaTraceCC chaincode v1.0 (Sequence 6)
  - Channel: ibnchannel

- **Chaincode (teaTraceCC)**
  - `createBatch` - Tạo lô trà mới
  - `verifyBatch` - Xác minh hash của lô trà
  - `getBatchInfo` - Query thông tin lô trà
  - `updateBatchStatus` - Cập nhật trạng thái
  - MSP-based authorization (Farmer, Verifier, Admin)
  - SHA-256 hash verification

- **Documentation**
  - README.md với tổng quan dự án
  - Kiến trúc hệ thống 4 tầng
  - Công nghệ sử dụng (100% Open Source)
  - Hướng dẫn Quick Start
  - API documentation
  - Deployment guides

- **License**
  - Apache License 2.0 (OSI-approved)
  - License header trong tất cả 263 file mã nguồn
  - LICENSE file ở root directory
  - LICENSE_NOTICE.md với mục đích sử dụng license

- **Infrastructure**
  - Docker Compose setup
  - Dockerfiles cho tất cả services
  - Makefiles cho build automation
  - Health checks
  - Graceful shutdown

### Changed
- Updated MSP configuration trong chaincode để match với network (tất cả roles = Org1MSP)
- Fixed hash verification logic trong chaincode (hash input trước khi so sánh)
- Improved error handling trong frontend services
- Wrapped console.log trong development mode checks

### Security
- JWT authentication với refresh tokens
- API key support
- TLS encryption cho Fabric connections
- MSP-based identity management
- Rate limiting
- Audit logging
- Input validation
- Parameterized queries (sqlc)

### Performance
- Multi-layer caching (L1 Memory 5-15min, L2 Redis 30min-1h)
- Connection pooling (pgx/v5, pool size 5-25)
- Optimized bundle size với Vite
- Background workers cho async processing

---

## [1.1.0] - 2025-11-24

### Added
- **QR Code Generation System**
  - QR code generation for batches: `GET /api/v1/qrcode/batches/{batchId}`
  - QR code generation for packages: `GET /api/v1/qrcode/packages/{packageId}`
  - QR code from transaction ID: `GET /api/v1/qrcode/transactions/{txId}`
  - Base64 data URI support: `/base64` endpoints
  - QR code data structure: `/data` endpoints
  - Frontend component `QRCodeDisplay` for display and download
  
- **Product Verification by Hash**
  - Endpoint: `POST /api/v1/teatrace/verify-by-hash`
  - Verify products by hash/blockhash/transaction ID
  - Multi-layer caching (L1: 5min, L2: 24h)
  - Rate limiting (10 requests/minute/IP)
  - Public endpoint (no authentication required)
  
- **NFC Support**
  - NFC payload generation: `GET /api/v1/nfc/packages/{packageId}`
  - NDEF format support for NFC tags
  
- **Dashboard Improvements**
  - Real-time WebSocket dashboard handler
  - Enhanced WebSocket authentication in API Gateway
  - Improved dashboard service with better data fetching
  
- **Chaincode Enhancements**
  - Tea package model with validation
  - Package creation: `createPackage` function
  - Enhanced validation utilities
  - Support for package lifecycle tracking

- **Database Optimizations**
  - Migration 016: Hash verification indexes
  - Composite index for teaTraceCC transactions
  - Optimized queries for hash verification

- **Infrastructure**
  - Fabric connection manager with health checks
  - gRPC credentials management
  - Improved error handling and logging

### Changed
- **Frontend Architecture Fix**
  - Frontend now calls Backend (port 9090) instead of Gateway directly
  - Updated Vite proxy configuration to point to backend
  - Fixed API config for proper service communication
  
- **API Gateway**
  - Enhanced WebSocket authentication flow
  - Improved nginx configuration for WebSocket support
  - Better error handling in dashboard WebSocket

- **Backend**
  - Enhanced Gateway client with better error handling
  - Improved authentication middleware
  - Better service initialization and dependency injection

### Fixed
- Frontend-Backend communication architecture
- WebSocket authentication flow
- API endpoint routing
- Service dependency injection

### Security
- Rate limiting for verification endpoints
- Enhanced WebSocket authentication
- Improved error messages (no sensitive data leakage)

### Performance
- Multi-layer caching for verification (L1: 5min, L2: 24h)
- Database indexes for hash queries
- Optimized chaincode queries

---

## [Unreleased]

### Planned
- Network Discovery service
- Channel Operations
- ACL System enhancements
- Performance optimizations
- Additional test coverage

---

[1.1.0]: https://github.com/Callmeduobgne/luongthiOlympic/releases/tag/v1.1.0
[1.0.0]: https://github.com/Callmeduobgne/luongthiOlympic/releases/tag/v1.0.0


