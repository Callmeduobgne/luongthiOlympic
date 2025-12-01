# IBN Network - ICTU Blockchain Network

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI/CD](https://github.com/Callmeduobgne/luongthiOlympic/actions/workflows/ci.yml/badge.svg)](https://github.com/Callmeduobgne/luongthiOlympic/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Callmeduobgne/luongthiOlympic)](https://goreportcard.com/report/github.com/Callmeduobgne/luongthiOlympic)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> Hệ thống truy xuất nguồn gốc sản phẩm trà dựa trên blockchain, đảm bảo tính minh bạch, bất biến và có thể kiểm chứng trong toàn bộ chuỗi cung ứng.

## 📚 Documentation
- [Contributing Guidelines](CONTRIBUTING.md)
- [Security Policy](SECURITY.md)
- [Changelog](CHANGELOG.md)
- [License](LICENSE)

## 📖 Giới Thiệu Dự Án

**IBN Network (ICTU Blockchain Network)** là một hệ thống blockchain enterprise-grade được thiết kế để giải quyết bài toán **truy xuất nguồn gốc (traceability)** cho sản phẩm trà. Hệ thống sử dụng **Hyperledger Fabric** - một nền tảng blockchain permissioned phù hợp cho các ứng dụng doanh nghiệp yêu cầu tính riêng tư, hiệu suất cao và khả năng mở rộng.

### 🎯 Mục Tiêu Dự Án

1. **Truy Xuất Nguồn Gốc Toàn Diện**
   - Theo dõi toàn bộ lifecycle của sản phẩm trà từ nông trại đến người tiêu dùng
   - Ghi lại mọi thay đổi trạng thái (harvest, processing, certification, distribution)
   - Đảm bảo tính minh bạch và có thể kiểm chứng

2. **Chống Giả Mạo & Đảm Bảo Tính Toàn Vẹn**
   - Sử dụng hash verification (SHA-256) để phát hiện thay đổi dữ liệu
   - Blockchain immutability đảm bảo dữ liệu không thể bị sửa đổi sau khi ghi
   - MSP-based authorization đảm bảo chỉ các bên được phép mới có thể thực hiện operations

3. **Tích Hợp Dễ Dàng**
   - RESTful API chuẩn cho frontend và third-party systems
   - Multi-layer caching để tối ưu hiệu suất
   - Event-driven architecture cho real-time notifications
   - **Real-time Blockchain Sync:** Tự động đồng bộ dữ liệu từ blockchain về database

4. **Enterprise-Grade Security**
   - JWT authentication + API Keys
   - TLS encryption cho tất cả blockchain connections
   - Role-based access control (RBAC)
   - Audit logging đầy đủ

## 🏗️ Kiến Trúc Hệ Thống

Hệ thống IBN Network được xây dựng theo kiến trúc **multi-layer** với sự tách biệt rõ ràng về trách nhiệm:

```
┌─────────────────────────────────────────────────────────────┐
│                    FRONTEND LAYER                           │
│  React + TypeScript + Vite + Tailwind CSS                   │
│  - User Interface (Port 9999)                               │
│  - State Management (Zustand)                               │
│  - Data Fetching (TanStack Query)                           │
└─────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST
┌─────────────────────────────────────────────────────────────┐
│                    BACKEND LAYER                            │
│  Go + Chi Router + PostgreSQL + Redis + OPA                 │
│  - Business Logic (Port 9900)                               │
│  - Authentication & Authorization                           │
│  - Caching & Metrics                                        │
│  - Policy Enforcement (OPA:8181)                            │
└─────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST
┌─────────────────────────────────────────────────────────────┐
│                  API GATEWAY LAYER                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Nginx Load Balancer (Port 9805)                    │   │
│  │  ├─ api-gateway-1 (9804)                            │   │
│  │  ├─ api-gateway-2 (9802)                            │   │
│  │  └─ api-gateway-3 (9803)                            │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Admin Service (Port 9902) - Chaincode Lifecycle    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↓ gRPC/TLS
┌─────────────────────────────────────────────────────────────┐
│                  NETWORK LAYER                              │
│  Hyperledger Fabric 2.5.9                                   │
│  - Orderer Cluster (Raft Consensus)                         │
│    • orderer.ibn.vn:7050 (Leader)                           │
│    • orderer1.ibn.vn:8050 (Follower)                        │
│    • orderer2.ibn.vn:9050 (Follower)                        │
│  - Peer Nodes + CouchDB                                     │
│    • peer0.org1.ibn.vn:7051 + couchdb0:5984                 │
│    • peer1.org1.ibn.vn:8051 + couchdb1:5985                 │
│    • peer2.org1.ibn.vn:9051 + couchdb2:5986                 │
│  - Chaincode (teaTraceCC v1.0)                              │
│  - Event Stream (Block Events) ───────────┐                 │
└───────────────────────────────────────────┼─────────────────┘
                                            │ gRPC/TLS
                                            ▼
┌─────────────────────────────────────────────────────────────┐
│              MONITORING & OBSERVABILITY                     │
│  - Prometheus (9901) - Metrics Collection                   │
│  - Grafana (9300) - Visualization                           │
│  - Alertmanager (9093) - Alert Management                   │
│  - cAdvisor (8081) - Container Metrics                      │
│  - Node Exporter (9100) - Host Metrics                      │
│  - Postgres Exporter (9187) - DB Metrics                    │
└─────────────────────────────────────────────────────────────┘
```

### 📊 Tổng Quan Các Tầng

| Tầng | Công Nghệ | Vai Trò | Port |
|------|-----------|---------|------|
| **Frontend** | React, TypeScript, Vite | User Interface | 9999 (prod), 5173 (dev) |
| **Backend** | Go, Chi, PostgreSQL, Redis | Business Logic & API | 9900 |
| **API Gateway** | Go, Fabric Gateway SDK + Nginx LB | Blockchain Proxy (Load Balanced) | 9805 (LB), 9802-9804 (instances) |
| **Admin Service** | Go, Fabric Admin SDK | Chaincode Lifecycle Management | 9902 (localhost only) |
| **Network** | Hyperledger Fabric 2.5.9 | Blockchain Network | 7050-9051 |
| **Monitoring** | Prometheus, Grafana, Alertmanager | Metrics & Visualization | 9901, 9300, 9093 |

## 🚀 Quick Start

### Yêu Cầu

-   Docker 20.10+
-   Docker Compose 1.29+
-   8GB RAM minimum
-   20GB disk space

### Khởi Động Toàn Bộ Hệ Thống

```bash
# Clone repository
cd /home/ubuntu/luongbeo

# Khởi động tất cả services (Production)
docker compose up -d

# Khởi động với monitoring stack
docker compose -f docker-compose.yml -f monitoring/docker-compose-monitoring.yml up -d
```

### Kiểm Tra Status

```bash
# Xem tất cả containers
docker compose ps

# Xem logs
docker compose logs -f

# Health check Backend
curl http://localhost:9900/health | jq '.'

# Health check API Gateway
curl http://localhost:9805/health | jq '.'

# Health check Grafana
curl http://localhost:9300/api/health
```

### Truy Cập Services

| Service | URL | Note |
|---------|-----|------|
| Frontend | http://localhost:9999 | Production build |
| Backend API | http://localhost:9900 | Requires JWT token |
| API Gateway | http://localhost:9805 | Load balanced |
| Grafana | http://localhost:9300 | admin/admin |
| Prometheus | http://localhost:9901 | Metrics |
| CouchDB | http://localhost:5984/_utils | admin/adminpw |

## 💻 Công Nghệ Sử Dụng

> **📌 Lưu ý:** Tất cả công nghệ được sử dụng trong dự án IBN Network đều là **Open Source Software (OSS)** với các license phổ biến như MIT, Apache 2.0, và BSD. Điều này đảm bảo tính minh bạch, khả năng tùy chỉnh và không có chi phí bản quyền.

### 🎨 Frontend Layer

**Công nghệ chính (100% Open Source):**

| Công Nghệ | Version | License | Trạng Thái |
|-----------|---------|---------|------------|
| **React** | 19.2.0 | MIT License | ✅ Open Source |
| **TypeScript** | 5.9.3 | Apache 2.0 | ✅ Open Source |
| **Vite** | 7.2.2 | MIT License | ✅ Open Source |
| **Tailwind CSS** | 3.4.18 | MIT License | ✅ Open Source |
| **Zustand** | 5.0.8 | MIT License | ✅ Open Source |
| **TanStack Query** | 5.90.8 | MIT License | ✅ Open Source |
| **React Router DOM** | 7.9.5 | MIT License | ✅ Open Source |
| **React Hook Form** | 7.66.0 | MIT License | ✅ Open Source |
| **Zod** | 4.1.12 | MIT License | ✅ Open Source |
| **Axios** | 1.13.2 | MIT License | ✅ Open Source |

**Đặc điểm:**
- Component-based architecture
- Type-safe với TypeScript
- Optimized bundle size với Vite
- Responsive design với Tailwind CSS
- Real-time data synchronization
- **100% Open Source** - Không có chi phí bản quyền

### 🔧 Backend Layer

**Công nghệ chính (100% Open Source):**

| Công Nghệ | Version | License | Trạng Thái |
|-----------|---------|---------|------------|
| **Go** | 1.24.6 | BSD 3-Clause | ✅ Open Source |
| **Chi Router** | v5.2.3 | MIT License | ✅ Open Source |
| **PostgreSQL** | 16 | PostgreSQL License | ✅ Open Source |
| **pgx/v5** | 5.7.6 | MIT License | ✅ Open Source |
| **Redis** | 9.16.0 | BSD 3-Clause | ✅ Open Source |
| **go-cache** | Latest | MIT License | ✅ Open Source |
| **JWT (golang-jwt)** | v5.3.0 | MIT License | ✅ Open Source |
| **Zap** | 1.27.0 | MIT License | ✅ Open Source |
| **UUID (google/uuid)** | v1.6.0 | Apache 2.0 | ✅ Open Source |

**Đặc điểm:**
- Layered architecture (Handler → Service → Repository)
- Domain-Driven Design (DDD)
- Multi-layer caching (L1 Memory + L2 Redis)
- Connection pooling (5-25 connections)
- Type-safe database queries với sqlc
- Graceful shutdown
- Health checks & metrics
- **100% Open Source** - Không có chi phí bản quyền

### 🌐 API Gateway Layer

**Công nghệ chính (100% Open Source):**

| Công Nghệ | Version | License | Trạng Thái |
|-----------|---------|---------|------------|
| **Go** | 1.23.5 | BSD 3-Clause | ✅ Open Source |
| **Fabric Gateway SDK** | v1.4.0 | Apache 2.0 | ✅ Open Source |
| **Chi Router** | v5.0.11 | MIT License | ✅ Open Source |
| **PostgreSQL** | 15 | PostgreSQL License | ✅ Open Source |
| **pgx/v5** | 5.5.4 | MIT License | ✅ Open Source |
| **Redis** | 9.4.0 | BSD 3-Clause | ✅ Open Source |
| **Circuit Breaker (gobreaker)** | Latest | MIT License | ✅ Open Source |
| **Prometheus** | Latest | Apache 2.0 | ✅ Open Source |
| **OpenTelemetry** | Latest | Apache 2.0 | ✅ Open Source |
| **WebSocket (gorilla/websocket)** | Latest | BSD 2-Clause | ✅ Open Source |

**Đặc điểm:**
- 50+ REST API endpoints
- Transaction management
- Event system với WebSocket support
- Block explorer
- Chaincode lifecycle management
- Audit logging
- Advanced metrics & monitoring
- **100% Open Source** - Không có chi phí bản quyền

### ⛓️ Network Layer (Blockchain)

**Công nghệ chính (100% Open Source):**

| Công Nghệ | Version | License | Trạng Thái |
|-----------|---------|---------|------------|
| **Hyperledger Fabric** | 2.5.9 | Apache 2.0 | ✅ Open Source |
| **Raft Consensus (etcdraft)** | Built-in | Apache 2.0 | ✅ Open Source |
| **CouchDB** | 3.3 | Apache 2.0 | ✅ Open Source |
| **Node.js** | 16+ | MIT License | ✅ Open Source |
| **TypeScript** | 5.3.3 | Apache 2.0 | ✅ Open Source |
| **Fabric Contract API** | 2.5.8 | Apache 2.0 | ✅ Open Source |

**Hyperledger Fabric:**
- **License:** Apache 2.0 (Open Source)
- **Maintained by:** Linux Foundation Hyperledger Project
- **Community:** Active open source community
- **Commercial Support:** Available từ nhiều vendors

**Cấu trúc Network:**
- **3 Orderer Nodes** - Raft consensus cluster
  - `orderer.ibn.vn:7050` (Leader)
  - `orderer1.ibn.vn:8050` (Follower)
  - `orderer2.ibn.vn:9050` (Follower)
- **3 Peer Nodes** - Endorsing peers (Org1MSP)
  - `peer0.org1.ibn.vn:7051`
  - `peer1.org1.ibn.vn:8051`
  - `peer2.org1.ibn.vn:9051`
- **3 CouchDB Instances** - State databases
  - `couchdb0:5984`
  - `couchdb1:5985`
  - `couchdb2:5986`
- **1 Channel** - `ibnchannel`
- **1 Chaincode** - `teaTraceCC v1.0` (Sequence 6)

**Chaincode Features:**
- `createBatch` - Tạo lô trà mới
- `verifyBatch` - Xác minh hash của lô trà
- `getBatchInfo` - Query thông tin lô trà
- `updateBatchStatus` - Cập nhật trạng thái
- `createPackage` - Tạo gói trà từ batch
- `verifyPackage` - Xác minh gói trà với blockhash
- `getPackageInfo` - Query thông tin gói trà
- MSP-based authorization (Farmer, Verifier, Admin)
- SHA-256 hash verification

**Open Source Standards & APIs:**
- **RESTful API** - Open standard (HTTP/JSON)
- **OpenAPI/Swagger** - API documentation standard
- **JWT (JSON Web Token)** - Open standard (RFC 7519)
- **TLS/SSL** - Open standard encryption
- **gRPC** - Open source RPC framework
- **WebSocket** - Open standard (RFC 6455)
- **Docker** - Open source containerization
- **Docker Compose** - Open source orchestration

## 📜 Tổng Kết Về Open Source

### ✅ Tất Cả Công Nghệ Đều Là Open Source

Dự án IBN Network được xây dựng **100% trên nền tảng Open Source**, đảm bảo:

1. **Không có chi phí bản quyền** - Tất cả software đều miễn phí sử dụng
2. **Tính minh bạch** - Source code có thể được review và audit
3. **Khả năng tùy chỉnh** - Có thể modify và extend theo nhu cầu
4. **Cộng đồng hỗ trợ** - Large community và extensive documentation
5. **Không bị vendor lock-in** - Không phụ thuộc vào proprietary solutions

### 📋 License Summary

| License Type | Số Lượng | Công Nghệ Ví Dụ |
|--------------|----------|-----------------|
| **MIT License** | ~15+ | React, Vite, Tailwind, Zustand, Axios, Chi Router, Zap |
| **Apache 2.0** | ~10+ | TypeScript, Hyperledger Fabric, CouchDB, Prometheus, OpenTelemetry |
| **BSD 3-Clause** | ~5+ | Go, Redis, PostgreSQL (BSD-style) |
| **PostgreSQL License** | 1 | PostgreSQL |
| **BSD 2-Clause** | 1+ | gorilla/websocket |

### 🌐 Open Standards & Protocols

- **HTTP/HTTPS** - Open standard
- **REST API** - Open architectural style
- **JSON** - Open data format
- **JWT** - Open authentication standard
- **TLS/SSL** - Open encryption protocols
- **gRPC** - Open RPC framework
- **WebSocket** - Open real-time communication protocol

## 📦 Services

### Application Layer

| Service | Container | Port(s) | Description | Layer |
|---------|-----------|---------|-------------|-------|
| **Frontend (Prod)** | ibn-frontend | 9999 | React UI (Production Build) | Frontend |
| **Frontend (Dev)** | ibn-frontend-dev | 5173 | React UI (Dev with Hot Reload) | Frontend |
| **Backend API** | ibn-backend | 9900 | RESTful API (85+ endpoints) | Backend |
| **PostgreSQL** | ibn-postgres | 5432 | Primary Database | Backend |
| **Redis** | ibn-redis | 6379 | Cache & Session Store | Backend |
| **OPA** | opa | 8181 | Policy Engine (Authorization) | Backend |

### API Gateway Layer (Load Balanced)

| Service | Container | Port(s) | Description | Layer |
|---------|-----------|---------|-------------|-------|
| **API Gateway LB** | api-gateway-nginx | 9805 | Nginx Load Balancer | Gateway |
| **API Gateway 1** | api-gateway-1 | 9804 | Blockchain Proxy Instance 1 | Gateway |
| **API Gateway 2** | api-gateway-2 | 9802 | Blockchain Proxy Instance 2 | Gateway |
| **API Gateway 3** | api-gateway-3 | 9803 | Blockchain Proxy Instance 3 | Gateway |
| **Admin Service** | admin-service | 9902 (localhost) | Chaincode Lifecycle Management | Gateway |

### Blockchain Network Layer

| Service | Container | Port(s) | Description | Layer |
|---------|-----------|---------|-------------|-------|
| **Orderer 0** | orderer.ibn.vn | 7050, 8443, 9443 | Raft Leader | Network |
| **Orderer 1** | orderer1.ibn.vn | 8050, 9444, 10443 | Raft Follower | Network |
| **Orderer 2** | orderer2.ibn.vn | 9050, 9445, 10444 | Raft Follower | Network |
| **Peer 0** | peer0.org1.ibn.vn | 7051-7052, 9446 | Endorsing Peer + Operations | Network |
| **Peer 1** | peer1.org1.ibn.vn | 8051-8052, 10446 | Endorsing Peer + Operations | Network |
| **Peer 2** | peer2.org1.ibn.vn | 9051-9052, 11446 | Endorsing Peer + Operations | Network |
| **CouchDB 0** | couchdb0 | 5984 | State DB for Peer 0 | Network |
| **CouchDB 1** | couchdb1 | 5985 | State DB for Peer 1 | Network |
| **CouchDB 2** | couchdb2 | 5986 | State DB for Peer 2 | Network |
| **Fabric CA** | ca.org1.ibn.vn | 7054 | Certificate Authority | Network |

### Monitoring & Observability

| Service | Container | Port(s) | Description | Layer |
|---------|-----------|---------|-------------|-------|
| **Prometheus** | prometheus | 9901 | Metrics Collection & Storage | Monitoring |
| **Grafana** | grafana | 9300 | Metrics Visualization & Dashboards | Monitoring |
| **Alertmanager** | alertmanager | 9093 | Alert Management & Routing | Monitoring |
| **cAdvisor** | cadvisor | 8081 | Container Metrics Exporter | Monitoring |
| **Node Exporter** | node-exporter-proxy | 9100 | Host Metrics Exporter | Monitoring |
| **Postgres Exporter** | postgres-exporter | 9187 | PostgreSQL Metrics Exporter | Monitoring |

### Access URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| **Frontend** | http://localhost:9999 | - |
| **Backend API** | http://localhost:9900 | JWT Token |
| **API Gateway** | http://localhost:9805 | API Key |
| **Grafana** | http://localhost:9300 | admin/admin (default) |
| **Prometheus** | http://localhost:9901 | - |
| **Alertmanager** | http://localhost:9093 | - |
| **CouchDB 0** | http://localhost:5984/_utils | admin/adminpw |

## 🔧 Configuration

### Docker Compose

Hệ thống sử dụng `docker-compose.yml` cho production environment. Tất cả services kết nối qua network `ibn-network`.

Xem chi tiết trong [docker-compose.yml](docker-compose.yml)

### Environment Variables

Backend configuration được định nghĩa trong `docker-compose.yml`:

```yaml
# Database
DB_HOST: postgres
DB_USER: gateway
DB_PASSWORD: changeme
DB_NAME: ibn_gateway

# Redis
REDIS_HOST: redis
REDIS_PASSWORD: changeme

# Fabric
FABRIC_PEER_ENDPOINT: peer0.org1.ibn.vn:7051
FABRIC_CHANNEL_NAME: ibnchannel

# JWT
JWT_SECRET: change-this-in-production
```

### Volumes

Persistent data được lưu trong Docker volumes:
- `postgres_data` - PostgreSQL data
- `redis_data` - Redis data
- `peer0.org1.ibn.vn` - Peer ledger data
- `orderer.ibn.vn` - Orderer data
- `couchdb0` - CouchDB state

## 📡 API Endpoints

### Health & Monitoring

```bash
# Health check
GET http://localhost:9900/health

# Readiness check
GET http://localhost:9900/ready

# Cache stats
GET http://localhost:9900/stats
```

### Authentication

```bash
# Register
POST http://localhost:9900/api/v1/auth/register
{
  "email": "admin@ibn.vn",
  "password": "Admin123!",
  "full_name": "Admin User",
  "role": "admin"
}

# Login
POST http://localhost:9900/api/v1/auth/login
{
  "email": "admin@ibn.vn",
  "password": "Admin123!"
}
```

### TeaTrace Chaincode

```bash
# Health check chaincode
GET http://localhost:9900/api/v1/teatrace/health
Authorization: Bearer <token>

# Get all batches
GET http://localhost:9900/api/v1/teatrace/batches
Authorization: Bearer <token>

# Create batch
POST http://localhost:9900/api/v1/teatrace/batches
Authorization: Bearer <token>
{
  "batch_id": "BATCH001",
  "farm_name": "Farm A",
  "harvest_date": "2024-11-13",
  "certification": "Organic",
  "certificate_id": "CERT-001"
}

# Create package
POST http://localhost:9900/api/v1/teatrace/packages
Authorization: Bearer <token>
{
  "package_id": "PKG001",
  "batch_id": "BATCH001",
  "weight": 500.0,
  "production_date": "2024-11-14"
}
```

### QR Code Generation

```bash
# Get QR code PNG for batch
GET http://localhost:9900/api/v1/qrcode/batches/{batchId}

# Get QR code base64 (for frontend)
GET http://localhost:9900/api/v1/qrcode/batches/{batchId}/base64

# Get QR code PNG for package
GET http://localhost:9900/api/v1/qrcode/packages/{packageId}

# Get QR code from transaction ID (auto-detect batch/package)
GET http://localhost:9900/api/v1/qrcode/transactions/{txId}

# Get NFC payload for package
GET http://localhost:9900/api/v1/nfc/packages/{packageId}
```

### Product Verification

```bash
# Verify product by hash (Public endpoint - no auth required)
POST http://localhost:9900/api/v1/teatrace/verify-by-hash
Content-Type: application/json

{
  "hash": "abc123..."
}

# Response:
{
  "success": true,
  "data": {
    "is_valid": true,
    "message": "Sản phẩm thuộc thương hiệu chúng tôi",
    "batch_id": "BATCH001",
    "entity_type": "batch"
  }
}
```

## 🛠️ Development

### Build & Deploy

```bash
# Rebuild backend
cd backend
docker build -t ibn-backend:latest .
cd ..

# Restart backend only
docker-compose restart ibn-backend

# Rebuild and restart
docker-compose up -d --build ibn-backend
```

### Logs

```bash
# All logs
docker-compose logs -f

# Specific service
docker-compose logs -f ibn-backend
docker-compose logs -f peer0.org1.ibn.vn
docker-compose logs -f orderer.ibn.vn

# Last 100 lines
docker-compose logs --tail=100 ibn-backend
```

### Database Access

```bash
# PostgreSQL
docker exec -it ibn-postgres psql -U gateway -d ibn_gateway

# Redis
docker exec -it ibn-redis redis-cli -a changeme

# CouchDB
curl http://admin:adminpw@localhost:5984/_all_dbs
```

## 🔍 Monitoring & Observability

### Monitoring Stack

Hệ thống sử dụng **Prometheus + Grafana + Alertmanager** để giám sát toàn bộ infrastructure:

```bash
# Grafana Dashboard
http://localhost:9300
Username: admin
Password: admin (default)

# Prometheus Metrics
http://localhost:9901

# Alertmanager
http://localhost:9093
```

### Fabric Network Monitoring

```bash
# Orderer health
curl http://localhost:8443/healthz

# Peer 0 health
curl http://localhost:9446/healthz

# Peer 0 metrics (Prometheus format)
curl http://localhost:9446/metrics

# Peer 1 metrics
curl http://localhost:10446/metrics

# Peer 2 metrics
curl http://localhost:11446/metrics
```

### Backend Metrics

```bash
# Get metrics snapshot
curl -H "Authorization: Bearer <token>" \
  http://localhost:9900/api/v1/metrics/snapshot | jq '.'

# Get aggregations
curl -H "Authorization: Bearer <token>" \
  http://localhost:9900/api/v1/metrics/aggregations | jq '.'
```

### Infrastructure Metrics

```bash
# Container metrics (cAdvisor)
curl http://localhost:8081/metrics

# Node/Host metrics
curl http://localhost:9100/metrics

# PostgreSQL metrics
curl http://localhost:9187/metrics

# Prometheus targets status
curl http://localhost:9901/api/v1/targets | jq '.'
```

### Grafana Dashboards

Hệ thống cung cấp các dashboards:
- **IBN Overview** - Tổng quan hệ thống
- **Blockchain Network** - Fabric network metrics
- **Backend Performance** - API performance, cache hit rates
- **Database Monitoring** - PostgreSQL metrics
- **Container Resources** - Docker container resources

## 🛑 Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: Data loss!)
docker-compose down -v

# Stop specific service
docker-compose stop ibn-backend
```

## 🔄 Updating

### Update Backend

```bash
# Stop backend
docker-compose stop ibn-backend

# Rebuild
cd backend
docker build -t ibn-backend:latest .
cd ..

# Start
docker-compose up -d ibn-backend
```

### Update Chaincode

```bash
# Package new chaincode
# Install on peers
# Approve and commit
# See docs/v1.0.1/network.md for details
```

## 🐛 Troubleshooting

### Backend không kết nối được Fabric

```bash
# Kiểm tra peer
docker logs peer0.org1.ibn.vn | tail -50

# Test connection từ backend
docker exec ibn-backend sh -c "nc -zv peer0.org1.ibn.vn 7051"

# Kiểm tra crypto materials
docker exec ibn-backend ls -la /fabric/organizations/
```

### Database connection failed

```bash
# Kiểm tra PostgreSQL
docker exec ibn-postgres pg_isready -U gateway -d ibn_gateway

# Xem logs
docker logs ibn-postgres

# Recreate database
docker-compose down
docker volume rm ibn_postgres_data
docker-compose up -d postgres
```

### Peer không join channel

```bash
# Xem peer logs
docker logs peer0.org1.ibn.vn

# Re-join channel
docker exec -it peer0.org1.ibn.vn peer channel join -b /path/to/channel.block
```

## 📊 Kiến Trúc Chi Tiết

### Frontend → Backend → Gateway → Network Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        FRONTEND LAYER                           │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  React App (Port 3000)                                   │   │
│  │  - Zustand (State)                                       │   │
│  │  - TanStack Query (Data Fetching)                        │   │
│  │  - React Router (Navigation)                             │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST (JSON)
┌─────────────────────────────────────────────────────────────────┐
│                        BACKEND LAYER                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Backend API (Port 9900)                                 │   │
│  │  ├─ Handlers (HTTP endpoints)                            │   │
│  │  ├─ Services (Business logic)                            │   │
│  │  ├─ Repositories (Data access)                           │   │
│  │  └─ Middleware (Auth, Logging, Caching)                  │   │
│  └──────────────────────────────────────────────────────────┘   │
│              ↓                    ↓                    ↓        │
│  ┌──────────────┐      ┌──────────────┐    ┌──────────────┐     │
│  │ PostgreSQL   │      │   Redis      │    │ API Gateway  │     │
│  │  (Port 5432) │      │  (Port 6379) │    │  (Port 9805) │     │
│  └──────────────┘      └──────────────┘    └──────────────┘     │
└─────────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST (JSON)
┌─────────────────────────────────────────────────────────────────┐
│                      API GATEWAY LAYER                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  API Gateway (Port 9805)                                 │   │
│  │  ├─ Fabric Gateway Client                                │   │
│  │  ├─ Transaction Management                               │   │
│  │  ├─ Event System (WebSocket)                             │   │
│  │  └─ Block Explorer                                       │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Admin Service (Port 8090)                               │   │
│  │  ├─ Chaincode Lifecycle Management                       │   │
│  │  └─ Network Operations                                   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                            ↓ gRPC/TLS
┌─────────────────────────────────────────────────────────────────┐
│                        NETWORK LAYER                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Orderer Cluster (Raft Consensus)                        │   │
│  │  ├─ orderer.ibn.vn:7050 (Leader)                         │   │
│  │  ├─ orderer1.ibn.vn:8050 (Follower)                      │   │
│  │  └─ orderer2.ibn.vn:9050 (Follower)                      │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Peer Nodes (Org1MSP)                                    │   │
│  │  ├─ peer0.org1.ibn.vn:7051                               │   │
│  │  ├─ peer1.org1.ibn.vn:8051                               │   │
│  │  └─ peer2.org1.ibn.vn:9051                               │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  CouchDB State Databases                                 │   │
│  │  ├─ couchdb0:5984                                        │   │
│  │  ├─ couchdb1:5985                                        │   │
│  │  └─ couchdb2:5986                                        │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Chaincode: teaTraceCC v1.0 (Sequence 6)                 │   │
│  │  Channel: ibnchannel                                     │   │
│  │  Language: Node.js (TypeScript)                          │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow Example: Tạo Lô Trà Mới

1. **Frontend** → User nhập thông tin lô trà → Gửi POST request đến Backend
2. **Backend** → Validate input → Lưu vào PostgreSQL → Gọi API Gateway
3. **API Gateway** → Tạo Fabric transaction → Gửi đến Peer nodes
4. **Network** → Peer endorse transaction → Orderer consensus → Commit vào blockchain
5. **Network** → Event được emit → API Gateway nhận event
6. **API Gateway** → WebSocket notification → Frontend update real-time

## 📚 Documentation

- [Backend Architecture](docs/v1.0.1/backend.md)
- [Network Architecture](docs/v1.0.1/network.md)
- [API Gateway](docs/v1.0.1/gateway.md)

## 🔐 Security

- JWT authentication với refresh tokens
- API Key authentication cho service-to-service (Backend → Gateway)
- TLS encryption cho Fabric connections
- MSP-based identity management
- Rate limiting
- Audit logging

## 📱 QR Code & Verification

- **QR Code Generation**: Tự động generate QR code từ batch/package/transaction
- **QR Code Content**: Chứa batchId/packageId, verificationHash/blockHash, verifyUrl, txId
- **Frontend Integration**: Component `QRCodeDisplay` để hiển thị và download QR code
- **Verification**: User có thể scan QR code để verify nguồn gốc trên blockchain

## 📝 License

**IBN Network** is licensed under the **Apache License 2.0** (OSI-approved).

### Mục Đích Sử Dụng Giấy Phép

Dự án sử dụng Apache License 2.0 với các mục đích:

1. **Tính Tương Thích Cao** - Tương thích với MIT, BSD, PostgreSQL License
2. **Bảo Vệ Quyền Tác Giả** - Yêu cầu giữ nguyên copyright notice
3. **Khuyến Khích Đóng Góp** - Cho phép tự do sử dụng, sửa đổi, phân phối
4. **Phù Hợp Enterprise** - Được chấp nhận rộng rãi trong môi trường doanh nghiệp
5. **Tuân Thủ OSI** - Được Open Source Initiative (OSI) phê duyệt

**Xem chi tiết:** [LICENSE](LICENSE)

### License Text

```
Copyright 2024 IBN Network (ICTU Blockchain Network)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

**License Type:** Apache 2.0 (OSI-approved Open Source License)  
**Full License Text:** See [LICENSE](LICENSE) file in the root directory  
**License Compatibility:** ✅ Tất cả dependencies đều tương thích (MIT, BSD, Apache 2.0, PostgreSQL License)

---

**Version**: 1.2.0  
**Last Updated**: November 30, 2025  
**Documentation Status**: ✅ Synchronized with production deployment



