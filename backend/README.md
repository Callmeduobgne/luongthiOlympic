# IBN Backend API

Backend API cho IBN Network - Blockchain traceability system sử dụng Hyperledger Fabric.

## 🎯 Tính Năng

### ✅ Đã Triển Khai

- **Authentication & Authorization**
  - JWT-based authentication
  - API Key management
  - Role-based access control (RBAC)
  - Refresh token mechanism

- **Blockchain Integration**
  - Kết nối với Hyperledger Fabric network
  - TeaTrace chaincode integration
  - Transaction management
  - Query/Invoke operations

- **Caching Strategy**
  - Multi-layer cache (L1: Memory, L2: Redis, L3: PostgreSQL)
  - Cache-aside pattern
  - Write-through/Write-behind

- **Analytics & Monitoring**
  - Audit logging với batch writes
  - Real-time metrics collection
  - System health monitoring

- **Event System**
  - Event subscriptions
  - Webhook delivery với retry
  - WebSocket support (planned)

## 🚀 Quick Start

### Yêu Cầu

- Docker
- Hyperledger Fabric network đang chạy
- PostgreSQL (port 5432)
- Redis (port 6379)

### Khởi Động

```bash
# Di chuyển vào thư mục backend
cd /home/exp2/ibn/backend

# Quick start (tự động build, migration, start)
./start.sh

# Hoặc sử dụng Make
make start

# Hoặc từng bước
make docker-build
make migrate-up
make docker-up
```

Backend sẽ khởi động tại: **http://localhost:9090**

### Test API

```bash
# Health check
curl http://localhost:9090/health

# Register user
curl -X POST http://localhost:9090/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ibn.vn","password":"Test123456!","full_name":"Admin","role":"admin"}'

# Login
curl -X POST http://localhost:9090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ibn.vn","password":"Test123456!"}'

# Test chaincode (với token từ login)
curl http://localhost:9090/api/v1/teatrace/health \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📚 Documentation

- [Docker Deployment Guide](DOCKER_DEPLOYMENT.md) - Chi tiết về Docker setup
- [Architecture Design](../docs/v1.0.1/backend.md) - Kiến trúc hệ thống
- [API Documentation](#api-endpoints) - Danh sách API endpoints

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│           IBN Backend (Port 9090)           │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────────┐  ┌──────────────┐        │
│  │   Handlers   │  │  Middleware  │        │
│  │  (HTTP API)  │  │  (Auth/Log)  │        │
│  └──────────────┘  └──────────────┘        │
│          ↓                 ↓                │
│  ┌─────────────────────────────────────┐   │
│  │          Services Layer             │   │
│  │  Auth │ ACL │ Audit │ Metrics │... │   │
│  └─────────────────────────────────────┘   │
│          ↓                ↓        ↓        │
│  ┌──────────┐  ┌────────┐  ┌────────────┐ │
│  │PostgreSQL│  │ Redis  │  │   Fabric   │ │
│  └──────────┘  └────────┘  └────────────┘ │
└─────────────────────────────────────────────┘
```

## 📡 API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| POST | `/api/v1/auth/register` | Register user |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh token |

### Protected Endpoints (Require Authentication)

#### User Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/profile` | Get user profile |
| POST | `/api/v1/api-keys` | Create API key |

#### TeaTrace Chaincode
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/teatrace/health` | Chaincode health |
| POST | `/api/v1/teatrace/batches` | Create tea batch |
| GET | `/api/v1/teatrace/batches` | Get all batches |
| GET | `/api/v1/teatrace/batches/{id}` | Get batch by ID |
| POST | `/api/v1/teatrace/batches/{id}/verify` | Verify batch |
| PUT | `/api/v1/teatrace/batches/{id}/status` | Update status |

#### Blockchain Operations
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/blockchain/transactions` | Submit transaction |
| POST | `/api/v1/blockchain/query` | Query chaincode |
| GET | `/api/v1/blockchain/transactions/{id}` | Get transaction |

#### Analytics
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/metrics/snapshot` | Get metrics |
| GET | `/api/v1/audit/logs` | Query audit logs |

## 🔧 Configuration

### Environment Variables

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=gateway
DB_PASSWORD=changeme
DB_NAME=ibn_gateway

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=changeme

# Fabric
FABRIC_MSP_ID=Org1MSP
FABRIC_CRYPTO_PATH=/fabric/organizations
FABRIC_PEER_ENDPOINT=peer0.org1.ibn.vn:7051
FABRIC_CHANNEL_NAME=ibnchannel

# JWT
JWT_SECRET=your-secret-key
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
```

## 🛠️ Development

### Build

```bash
# Local build
make build

# Docker build
make docker-build
```

### Run Tests

```bash
make test
```

### View Logs

```bash
# Docker logs
docker logs -f ibn-backend-dev

# Or with Make
make docker-dev-logs
```

## 🐳 Docker Commands

```bash
# Quick start (all-in-one)
./start.sh
# Or
make start

# Manual control
make docker-build    # Build image
make docker-up       # Start container
make docker-down     # Stop container
make docker-restart  # Restart container
make docker-logs     # View logs

# Check status
make status
```

## 📊 Monitoring

### Health Endpoints

- **`/health`** - Basic health check
- **`/ready`** - Readiness probe (DB, Redis, Fabric)
- **`/stats`** - Cache statistics

### Metrics

Access metrics at `/api/v1/metrics/snapshot` (requires authentication)

```json
{
  "timestamp": "2025-11-13T10:00:00Z",
  "metrics": {
    "api_request_total": 1234,
    "blockchain_tx_total": 56,
    "cache_hit_total": 890,
    "db_connections_active": 5
  }
}
```

## 🔐 Security

- JWT-based authentication
- API key support
- TLS connections to Fabric
- MSP-based identity
- Audit logging for all operations
- Rate limiting (configurable)
- Input validation

## 🚀 Deployment

### Production

```bash
# Build production image
docker build -t ibn-backend:1.0.0 .

# Run with production config
docker-compose -f docker-compose.yml up -d
```

### Environment-specific

- **Development**: `docker-compose.dev.yml`
- **Production**: `docker-compose.yml`

## 📝 Migration

Run database migrations:

```bash
# Using Docker exec
docker exec ibn-backend-dev /app/server migrate up

# Or manually with psql
for f in migrations/*.up.sql; do
    psql -h localhost -U gateway -d ibn_gateway -f "$f"
done
```

## 🐛 Troubleshooting

### Container không khởi động

```bash
# Xem logs
docker logs ibn-backend-dev

# Kiểm tra health
docker inspect ibn-backend-dev | jq '.[0].State.Health'
```

### Không kết nối được Fabric

```bash
# Kiểm tra Fabric network
docker ps | grep peer0.org1.ibn.vn

# Test connection
docker exec ibn-backend-dev nc -zv peer0.org1.ibn.vn 7051
```

## 📚 Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: Chi Router
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Blockchain**: Hyperledger Fabric 2.5
- **Logging**: Zap
- **Authentication**: JWT (golang-jwt/jwt/v5)

## 📄 License

Copyright © 2024 IBN Network

## 👥 Contributors

IBN Development Team

---

**Version**: 1.0.0  
**Last Updated**: November 2025
