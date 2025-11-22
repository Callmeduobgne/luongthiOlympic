# IBN Network - ICTU Blockchain Network

> Hệ thống truy xuất nguồn gốc sản phẩm trà dựa trên blockchain, đảm bảo tính minh bạch, bất biến và có thể kiểm chứng trong toàn bộ chuỗi cung ứng.

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

4. **Enterprise-Grade Security**
   - JWT authentication + API Keys
   - TLS encryption cho tất cả blockchain connections
   - Role-based access control (RBAC)
   - Audit logging đầy đủ

## 🏗️ Kiến Trúc Hệ Thống

Hệ thống IBN Network được xây dựng theo kiến trúc **4 tầng (layers)** với sự tách biệt rõ ràng về trách nhiệm:

```
┌─────────────────────────────────────────────────────────────┐
│                    FRONTEND LAYER                           │
│  React + TypeScript + Vite + Tailwind CSS                   │
│  - User Interface                                           │
│  - State Management (Zustand)                               │
│  - Data Fetching (TanStack Query)                           │
└─────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST
┌─────────────────────────────────────────────────────────────┐
│                    BACKEND LAYER                            │
│  Go + Chi Router + PostgreSQL + Redis                       │
│  - Business Logic                                           │
│  - Authentication & Authorization                           │
│  - Caching & Metrics                                        │
└─────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST
┌─────────────────────────────────────────────────────────────┐
│                  API GATEWAY LAYER                          │
│  Go + Fabric Gateway SDK                                    │
│  - Blockchain Operations Proxy                              │
│  - Transaction Management                                   │
│  - Event System                                             │
└─────────────────────────────────────────────────────────────┘
                            ↓ gRPC/TLS
┌─────────────────────────────────────────────────────────────┐
│                  NETWORK LAYER                              │
│  Hyperledger Fabric 2.5.9                                   │
│  - Orderer Cluster (Raft Consensus)                         │
│  - Peer Nodes + CouchDB                                     │
│  - Chaincode (teaTraceCC)                                   │
└─────────────────────────────────────────────────────────────┘
```

### 📊 Tổng Quan Các Tầng

| Tầng | Công Nghệ | Vai Trò | Port |
|------|-----------|---------|------|
| **Frontend** | React, TypeScript, Vite | User Interface | 3000 |
| **Backend** | Go, Chi, PostgreSQL, Redis | Business Logic & API | 9090 |
| **API Gateway** | Go, Fabric Gateway SDK | Blockchain Proxy | 8080 |
| **Network** | Hyperledger Fabric 2.5.9 | Blockchain Network | 7050-9051 |

## 🚀 Quick Start

### Yêu Cầu

- Docker 20.10+
- Docker Compose 1.29+
- 8GB RAM minimum
- 20GB disk space

### Khởi Động Toàn Bộ Hệ Thống

```bash
# Clone repository
cd /home/exp2/ibn

# Khởi động tất cả services (Production)
docker-compose up -d

# Hoặc sử dụng script nếu có
./start-ibn.sh
```

### Kiểm Tra Status

```bash
# Xem tất cả containers
docker-compose ps

# Xem logs
docker-compose logs -f

# Health check
curl http://localhost:9090/health | jq '.'
```

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

| Service | Container | Port | Description | Layer |
|---------|-----------|------|-------------|-------|
| **Frontend** | ibn-frontend | 3000 | React UI | Frontend |
| **Backend API** | ibn-backend | 9090 | RESTful API | Backend |
| **API Gateway** | api-gateway | 8080 | Blockchain Proxy | Gateway |
| **Admin Service** | admin-service | 8090 | Chaincode Management | Gateway |
| **PostgreSQL** | ibn-postgres | 5432 | Database | Backend |
| **Redis** | ibn-redis | 6379 | Cache | Backend |
| **Orderer 0** | orderer.ibn.vn | 7050 | Raft leader | Network |
| **Orderer 1** | orderer1.ibn.vn | 8050 | Raft follower | Network |
| **Orderer 2** | orderer2.ibn.vn | 9050 | Raft follower | Network |
| **Peer 0** | peer0.org1.ibn.vn | 7051 | Endorsing peer | Network |
| **Peer 1** | peer1.org1.ibn.vn | 8051 | Endorsing peer | Network |
| **Peer 2** | peer2.org1.ibn.vn | 9051 | Endorsing peer | Network |
| **CouchDB 0** | couchdb0 | 5984 | State DB | Network |
| **CouchDB 1** | couchdb1 | 5985 | State DB | Network |
| **CouchDB 2** | couchdb2 | 5986 | State DB | Network |

## 🔧 Configuration

### Docker Compose

Hệ thống sử dụng `docker-compose.yml` cho production environment. Tất cả services kết nối qua network `ibn-network`.

Xem chi tiết trong [DOCKER_COMPOSE_GUIDE.md](DOCKER_COMPOSE_GUIDE.md)

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
GET http://localhost:9090/health

# Readiness check
GET http://localhost:9090/ready

# Cache stats
GET http://localhost:9090/stats
```

### Authentication

```bash
# Register
POST http://localhost:9090/api/v1/auth/register
{
  "email": "admin@ibn.vn",
  "password": "Admin123!",
  "full_name": "Admin User",
  "role": "admin"
}

# Login
POST http://localhost:9090/api/v1/auth/login
{
  "email": "admin@ibn.vn",
  "password": "Admin123!"
}
```

### TeaTrace Chaincode

```bash
# Health check chaincode
GET http://localhost:9090/api/v1/teatrace/health
Authorization: Bearer <token>

# Get all batches
GET http://localhost:9090/api/v1/teatrace/batches
Authorization: Bearer <token>

# Create batch
POST http://localhost:9090/api/v1/teatrace/batches
Authorization: Bearer <token>
{
  "batch_id": "BATCH001",
  "farm_name": "Farm A",
  "harvest_date": "2024-11-13",
  "certification": "Organic",
  "certificate_id": "CERT-001"
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

## 🔍 Monitoring

### Fabric Network

```bash
# Orderer health
curl http://localhost:8443/healthz

# Peer health
curl http://localhost:9446/healthz

# Peer metrics (Prometheus format)
curl http://localhost:9446/metrics
```

### Backend Metrics

```bash
# Get metrics snapshot
curl -H "Authorization: Bearer <token>" \
  http://localhost:9090/api/v1/metrics/snapshot | jq '.'

# Get aggregations
curl -H "Authorization: Bearer <token>" \
  http://localhost:9090/api/v1/metrics/aggregations | jq '.'
```

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
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  React App (Port 3000)                                    │  │
│  │  - Zustand (State)                                        │  │
│  │  - TanStack Query (Data Fetching)                         │  │
│  │  - React Router (Navigation)                              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST (JSON)
┌─────────────────────────────────────────────────────────────────┐
│                        BACKEND LAYER                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Backend API (Port 9090)                                   │  │
│  │  ├─ Handlers (HTTP endpoints)                             │  │
│  │  ├─ Services (Business logic)                              │  │
│  │  ├─ Repositories (Data access)                            │  │
│  │  └─ Middleware (Auth, Logging, Caching)                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│              ↓                    ↓                    ↓         │
│  ┌──────────────┐      ┌──────────────┐    ┌──────────────┐  │
│  │ PostgreSQL   │      │   Redis      │    │ API Gateway  │  │
│  │  (Port 5432) │      │  (Port 6379) │    │  (Port 8080)  │  │
│  └──────────────┘      └──────────────┘    └──────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            ↓ HTTP/REST (JSON)
┌─────────────────────────────────────────────────────────────────┐
│                      API GATEWAY LAYER                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  API Gateway (Port 8080)                                  │  │
│  │  ├─ Fabric Gateway Client                                │  │
│  │  ├─ Transaction Management                                │  │
│  │  ├─ Event System (WebSocket)                              │  │
│  │  └─ Block Explorer                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Admin Service (Port 8090)                                 │  │
│  │  ├─ Chaincode Lifecycle Management                        │  │
│  │  └─ Network Operations                                    │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            ↓ gRPC/TLS
┌─────────────────────────────────────────────────────────────────┐
│                        NETWORK LAYER                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Orderer Cluster (Raft Consensus)                        │  │
│  │  ├─ orderer.ibn.vn:7050 (Leader)                         │  │
│  │  ├─ orderer1.ibn.vn:8050 (Follower)                      │  │
│  │  └─ orderer2.ibn.vn:9050 (Follower)                      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Peer Nodes (Org1MSP)                                     │  │
│  │  ├─ peer0.org1.ibn.vn:7051                                │  │
│  │  ├─ peer1.org1.ibn.vn:8051                                │  │
│  │  └─ peer2.org1.ibn.vn:9051                                │  │
│  └──────────────────────────────────────────────────────────┘  │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  CouchDB State Databases                                  │  │
│  │  ├─ couchdb0:5984                                         │  │
│  │  ├─ couchdb1:5985                                         │  │
│  │  └─ couchdb2:5986                                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Chaincode: teaTraceCC v1.0 (Sequence 6)                │  │
│  │  Channel: ibnchannel                                     │  │
│  │  Language: Node.js (TypeScript)                          │  │
│  └──────────────────────────────────────────────────────────┘  │
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
- [Chaincode Documentation](docs/v1.0.1/tea_1.0.md)

## 🔐 Security

- JWT authentication với refresh tokens
- API key support
- TLS encryption cho Fabric connections
- MSP-based identity management
- Rate limiting
- Audit logging

## 📝 License

Copyright © 2024 IBN Network

---

**Version**: 1.0.0  
**Last Updated**: November 2024

