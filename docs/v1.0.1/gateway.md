# API Gateway Layer - Kiến Trúc & Hướng Dẫn

**Ngày tạo:** 2025-11-12  
**Version:** 2.0.0  
**Last Updated:** 2025-11-27  
**Layer:** API Gateway (REST API cho Blockchain)

---

## 📋 Tổng Quan

API Gateway là lớp trung gian giữa **Frontend/Backend** và **Hyperledger Fabric Network**, cung cấp:
- ✅ RESTful API với 90+ endpoints
- ✅ Authentication & Authorization (JWT + API Key)
- ✅ Rate Limiting & Security
- ✅ Transaction Management
- ✅ Real-time Events (WebSocket)
- ✅ Network Discovery & Monitoring
- ✅ Audit Logging & Metrics

---

## 🏗️ Kiến Trúc Tổng Quan

### System Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                               │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │
│  │   Frontend    │  │    Backend    │  │  External     │           │
│  │   (React)     │  │   (Go API)    │  │   Services    │           │
│  └──────┬────────┘  └──────┬────────┘  └──────┬────────┘           │
│         │                  │                  │                    │
│         └──────────────────┼──────────────────┘                    │
│                            │                                       │
│                            ▼                                       │
└────────────────────────────────────────────────────────────────────┘
                            │
                            │ HTTP/HTTPS
                            │ WebSocket
                            ▼
┌────────────────────────────────────────────────────────────────────┐
│                      API GATEWAY LAYER                             │
│                      (Port 8080)                                   │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    NGINX PROXY                              │   │
│  │  - SSL/TLS Termination                                      │   │
│  │  - Load Balancing                                           │   │
│  │  - CORS Handling                                            │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                         │
│                          ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              HTTP SERVER (Chi Router)                       │   │
│  │  - Request Routing                                          │   │
│  │  - Middleware Stack                                         │   │
│  │  - Swagger Documentation                                    │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                         │
│                          ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              MIDDLEWARE STACK                               │   │
│  │  1. Recovery (Panic Handler)                                │   │
│  │  2. Logger (Request Logging)                                │   │
│  │  3. Tracing (Request ID)                                    │   │
│  │  4. Compression (Gzip)                                      │   │
│  │  5. Audit (Request Audit Log)                               │   │
│  │  6. Authentication (JWT/API Key)                            │   │
│  │  7. Rate Limiting (Redis-based)                             │   │
│  │  8. ACL (Permission Check)                                  │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                         │
│                          ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    HANDLERS LAYER                           │   │
│  │  - AuthHandler      - BatchHandler                          │   │
│  │  - ChaincodeHandler - ChannelHandler                        │   │
│  │  - NetworkHandler   - TransactionHandler                    │   │
│  │  - EventHandler     - ExplorerHandler                       │   │
│  │  - MetricsHandler   - AuditHandler                          │   │
│  │  - ACLHandler       - UserHandler                           │   │
│  │  - DashboardHandler (WebSocket)                             │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                         │
│                          ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    SERVICES LAYER                           │   │
│  │  - AuthService      - BatchService                          │   │
│  │  - ChaincodeService - ChannelService                        │   │
│  │  - NetworkService   - TransactionService                    │   │
│  │  - EventService     - ExplorerService                       │   │
│  │  - MetricsService   - AuditService                          │   │
│  │  - ACLService       - CAService                             │   │
│  │  - IndexerService   - DiscoveryService                      │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                         │
│                          ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              FABRIC GATEWAY SDK                             │   │
│  │  - Gateway Client Connection                                │   │
│  │  - Network & Channel Management                             │   │
│  │  - Contract Invocation                                      │   │
│  │  - Event Streaming                                          │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                         │
└──────────────────────────┼─────────────────────────────────────────┘
                           │
                           │ gRPC/TLS
                           │
                           ▼
┌────────────────────────────────────────────────────────────────────┐
│                    DATA LAYER                                      │
├────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────┐                        │
│  │   PostgreSQL     │  │      Redis       │                        │
│  │   - Users        │  │   - Rate Limit   │                        │
│  │   - Transactions │  │   - Cache        │                        │
│  │   - Audit Logs   │  │   - Sessions     │                        │
│  │   - ACL Policies │  │                  │                        │
│  │   - Events       │  │                  │                        │
│  └──────────────────┘  └──────────────────┘                        │
└────────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────────┐
│              HYPERLEDGER FABRIC NETWORK                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │   Orderers   │  │    Peers     │  │  Fabric CA   │              │
│  │  (3 nodes)   │  │  (3 nodes)   │  │  (1 node)    │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
│         │                │                    │                    │
│         └────────────────┼────────────────────┘                    │
│                          │                                         │
│                    Channel: ibnchannel                             │
│                    Chaincode: teaTraceCC v1.1.0                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## 📁 Cấu Trúc Thư Mục Thực Tế

```
api-gateway/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point, service initialization
├── internal/
│   ├── config/                        # Configuration management
│   │   ├── config.go                 # Main config loader
│   │   ├── fabric.go                 # Fabric network config
│   │   ├── database.go               # PostgreSQL config
│   │   └── redis.go                  # Redis config
│   │
│   ├── handlers/                     # HTTP Request Handlers
│   │   ├── auth/                     # Authentication handlers
│   │   ├── batch.go                  # Tea batch operations
│   │   ├── chaincode/                # Chaincode invoke/query
│   │   ├── channel/                 # Channel management
│   │   ├── network/                  # Network discovery
│   │   ├── transaction/              # Transaction management
│   │   ├── event/                    # Event subscriptions
│   │   ├── explorer/                # Block explorer
│   │   ├── metrics/                  # Metrics endpoints
│   │   ├── audit/                    # Audit logs
│   │   ├── acl/                      # Access control
│   │   ├── users/                    # User management
│   │   ├── dashboard/                # WebSocket dashboard
│   │   └── health.go                 # Health checks
│   │
│   ├── services/                     # Business Logic Layer
│   │   ├── auth/                     # Authentication service
│   │   ├── chaincode/                # Chaincode operations
│   │   ├── channel/                 # Channel operations
│   │   ├── network/                  # Network discovery
│   │   │   ├── service.go           # Network info service
│   │   │   └── discovery_service.go # Discovery service
│   │   ├── transaction/              # Transaction service
│   │   ├── event/                    # Event service
│   │   ├── explorer/                # Block explorer service
│   │   ├── metrics/                  # Metrics service
│   │   ├── audit/                    # Audit service
│   │   ├── acl/                      # ACL service
│   │   ├── ca/                       # Fabric CA service
│   │   ├── indexer/                  # Block indexer
│   │   ├── fabric/                   # Fabric Gateway SDK wrapper
│   │   │   ├── gateway.go           # Gateway connection
│   │   │   ├── chaincode.go         # Chaincode service
│   │   │   └── contract.go          # Contract service
│   │   └── cache/                    # Redis cache service
│   │
│   ├── middleware/                   # HTTP Middleware
│   │   ├── auth.go                   # JWT/API Key authentication
│   │   ├── rate_limit.go            # Rate limiting
│   │   ├── logger.go                 # Request logging
│   │   ├── audit.go                  # Audit logging
│   │   ├── cors.go                   # CORS handling
│   │   ├── recovery.go               # Panic recovery
│   │   ├── tracing.go                # Request tracing
│   │   └── websocket_rate_limit.go   # WebSocket rate limiting
│   │
│   ├── routes/                       # Route Configuration
│   │   └── routes.go                 # Chi router setup
│   │
│   ├── models/                       # Data Models
│   │   ├── user.go
│   │   ├── transaction.go
│   │   ├── network.go
│   │   └── ...
│   │
│   ├── repository/                   # Database Repositories
│   │   └── ...
│   │
│   └── utils/                        # Utilities
│       ├── logger.go                 # Zap logger setup
│       └── ...
│
├── migrations/                       # Database Migrations
│   └── *.sql
│
├── docker/                          # Docker Configuration
│   ├── docker-compose.yml
│   └── nginx/
│       └── nginx.conf               # Nginx reverse proxy
│
└── go.mod                           # Go Dependencies
```

---

## 🔄 Request Flow Chi Tiết

### 1. Request Flow: Chaincode Invoke

```
Client Request
    │
    │ POST /api/v1/channels/ibnchannel/chaincodes/teaTraceCC/invoke
    │ Headers: Authorization: Bearer <JWT>
    │ Body: { "function": "createBatch", "args": [...] }
    │
    ▼
┌─────────────────────────────────────┐
│  NGINX Proxy (Port 80/443)          │
│  - SSL Termination                  │
│  - CORS Headers                     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  HTTP Server (Port 8080)            │
│  Chi Router                         │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  MIDDLEWARE STACK                   │
│  1. Recovery → Catch panics         │
│  2. Logger → Log request            │
│  3. Tracing → Add Request ID        │
│  4. Compression → Gzip (skip WS)    │
│  5. Audit → Log to DB               │
│  6. Auth → Validate JWT/API Key     │
│  7. Rate Limit → Check Redis        │
│  8. ACL → Check permissions         │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  ChaincodeHandler.Invoke()          │
│  - Parse request body               │
│  - Validate parameters              │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  ChaincodeService.Invoke()          │
│  - Prepare transaction proposal     │
│  - Endorse with peers               │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Fabric Gateway SDK                 │
│  - Get Network (ibnchannel)         │
│  - Get Contract (teaTraceCC)        │
│  - Submit Transaction               │
└──────────────┬──────────────────────┘
               │
               │ gRPC/TLS
               ▼
┌─────────────────────────────────────┐
│  Hyperledger Fabric Network         │
│  - Peer0: Endorse                   │
│  - Peer1: Endorse                   │
│  - Peer2: Endorse                   │
│  - Orderer: Order & Commit          │
└──────────────┬──────────────────────┘
               │
               │ Transaction ID
               ▼
┌─────────────────────────────────────┐
│  Response                           │
│  {                                  │
│    "transactionId": "tx123...",     │
│    "status": "committed"            │
│  }                                  │
└─────────────────────────────────────┘
```

### 2. Request Flow: WebSocket Dashboard

```
Client WebSocket Connection
    │
    │ GET /api/v1/dashboard/ws/ibnchannel?token=<JWT>
    │ Upgrade: websocket
    │
    ▼
┌─────────────────────────────────────┐
│  NGINX Proxy                        │
│  - WebSocket Upgrade                │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  HTTP Server                        │
│  - Skip compression (WebSocket)     │
│  - Skip audit (WebSocket)           │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  DashboardHandler.HandleWebSocket() │
│  - Validate origin                  │
│  - Check rate limit                 │
│  - Upgrade connection               │
└──────────────┬──────────────────────┘
               │
               │ WebSocket Connection
               ▼
┌─────────────────────────────────────┐
│  WebSocket Service                  │
│  - Subscribe to metrics             │
│  - Subscribe to blocks              │
│  - Subscribe to network info        │
│  - Send updates every 5s            │
└─────────────────────────────────────┘
```

---

## 🛠️ Middleware Stack (Thứ Tự Thực Thi)

### Global Middleware (Áp dụng cho tất cả requests)

```go
1. Recovery Middleware
   └─> Catch panics, return 500 error

2. Logger Middleware
   └─> Log request: method, path, IP, user agent

3. Tracing Middleware
   └─> Generate Request ID, add to context

4. RequestID Middleware (Chi)
   └─> Ensure Request ID exists

5. RealIP Middleware (Chi)
   └─> Extract real client IP from headers

6. Compression Middleware
   └─> Gzip response (skip WebSocket)

7. Audit Middleware
   └─> Log request to audit_logs table (skip WebSocket)
```

### Route-Specific Middleware (Áp dụng cho protected routes)

```go
8. Authentication Middleware
   └─> Validate JWT token or API Key
       ├─> Check Authorization header (Bearer token)
       ├─> Check X-API-Key header
       ├─> Check query parameter (for WebSocket)
       └─> Add user info to context

9. Rate Limiting Middleware
   └─> Check Redis for rate limit
       ├─> Default: 1000 req/min
       ├─> Login: 5 req/15min (anti-brute force)
       └─> WebSocket: 100 messages/min

10. ACL Middleware (Optional)
    └─> Check user permissions
        ├─> Resource-based permissions
        ├─> Role-based permissions
        └─> Pattern matching
```

---

## 🔌 API Endpoints Overview

### Authentication & Authorization
```
POST   /api/v1/auth/login              # Login (JWT)
POST   /api/v1/auth/refresh            # Refresh token
POST   /api/v1/auth/register            # Register user
POST   /api/v1/auth/api-keys            # Generate API key (Auth)
GET    /api/v1/auth/api-keys            # List API keys (Auth)
DELETE /api/v1/auth/api-keys/{id}       # Revoke API key (Auth)
```

### Chaincode Operations (Business Transactions)
```
POST   /api/v1/channels/{channel}/chaincodes/{name}/invoke  # Invoke (Auth)
POST   /api/v1/channels/{channel}/chaincodes/{name}/query   # Query (Auth)
```

### Batch Operations (Tea Traceability)
```
GET    /api/v1/batches/{id}            # Get batch (Public)
POST   /api/v1/batches                 # Create batch (Auth)
POST   /api/v1/batches/{id}/verify     # Verify batch (Auth)
PATCH  /api/v1/batches/{id}/status     # Update status (Auth)
```

### Network Discovery
```
GET    /api/v1/network/info             # Network overview (Auth)
GET    /api/v1/network/peers           # List peers (Auth)
GET    /api/v1/network/peers/{id}      # Peer details (Auth)
GET    /api/v1/network/orderers        # List orderers (Auth)
GET    /api/v1/network/orderers/{id}   # Orderer details (Auth)
GET    /api/v1/network/cas             # List CAs (Auth)
GET    /api/v1/network/topology        # Network topology (Auth)
GET    /api/v1/network/channels        # List channels (Auth)
GET    /api/v1/network/channels/{name} # Channel info (Auth)
GET    /api/v1/network/health/peers    # Peer health (Auth)
GET    /api/v1/network/health/orderers # Orderer health (Auth)
```

### Channel Management
```
POST   /api/v1/channels                # Create channel (Admin)
GET    /api/v1/channels/{name}/config  # Get config (Auth)
PATCH  /api/v1/channels/{name}/config  # Update config (Admin)
POST   /api/v1/channels/{name}/join    # Join peer (Admin)
GET    /api/v1/channels/{name}/members # List members (Auth)
GET    /api/v1/channels/{name}/peers   # List peers (Auth)
```

### Transactions
```
POST   /api/v1/transactions            # Submit transaction (Auth)
GET    /api/v1/transactions            # List transactions (Auth)
GET    /api/v1/transactions/{id}       # Get transaction (Auth)
```

### Blocks & Explorer
```
GET    /api/v1/blocks/{channel}        # List blocks (Auth)
GET    /api/v1/blocks/{channel}/latest # Latest block (Auth)
GET    /api/v1/blocks/{channel}/{number} # Get block (Auth)
```

### Metrics
```
GET    /api/v1/metrics/summary         # Summary metrics (Auth)
GET    /api/v1/metrics/transactions    # Transaction metrics (Auth)
GET    /api/v1/metrics/blocks          # Block metrics (Auth)
GET    /api/v1/metrics/performance     # Performance metrics (Auth)
GET    /api/v1/metrics/peers          # Peer metrics (Auth)
```

### Events
```
POST   /api/v1/events/subscriptions   # Subscribe to events (Auth)
```

### Users & Identity (Fabric CA)
```
GET    /api/v1/users                   # List users (Auth)
GET    /api/v1/users/{id}              # Get user (Auth)
POST   /api/v1/users/enroll            # Enroll user (Admin)
POST   /api/v1/users/register          # Register user (Admin)
POST   /api/v1/users/{id}/reenroll     # Reenroll user (Admin)
DELETE /api/v1/users/{id}/revoke       # Revoke certificate (Admin)
GET    /api/v1/users/{id}/certificate  # Get certificate (Admin)
```

### ACL (Access Control)
```
GET    /api/v1/acl/policies           # List policies (Auth)
POST   /api/v1/acl/policies           # Create policy (Admin)
GET    /api/v1/acl/policies/{id}       # Get policy (Auth)
PATCH  /api/v1/acl/policies/{id}       # Update policy (Admin)
DELETE /api/v1/acl/policies/{id}      # Delete policy (Admin)
GET    /api/v1/acl/permissions         # List permissions (Auth)
POST   /api/v1/acl/check               # Check permission (Auth)
```

### Dashboard WebSocket
```
GET    /api/v1/dashboard/ws/{channel}  # WebSocket connection
       Query: ?token=<JWT>
```

### Health & Monitoring
```
GET    /health                         # Health check (Public)
GET    /ready                          # Readiness check (Public)
GET    /live                           # Liveness check (Public)
GET    /metrics                        # Prometheus metrics (Public)
GET    /swagger/*                      # Swagger docs (Public)
```

---

## 🔐 Authentication Flow

### JWT Authentication

```
1. Client Login
   POST /api/v1/auth/login
   Body: { "email": "user@example.com", "password": "..." }
   
   ↓
   
2. AuthService.ValidateCredentials()
   - Check email/password in database
   - Generate JWT token (access + refresh)
   
   ↓
   
3. Response
   {
     "accessToken": "eyJhbGc...",
     "refreshToken": "eyJhbGc...",
     "expiresIn": 3600
   }
   
   ↓
   
4. Client uses token
   GET /api/v1/network/peers
   Header: Authorization: Bearer <accessToken>
   
   ↓
   
5. AuthMiddleware.Authenticate()
   - Extract token from header
   - Validate JWT signature
   - Check expiration
   - Add user info to context
   
   ↓
   
6. Handler processes request
```

### API Key Authentication

```
1. Generate API Key (Admin)
   POST /api/v1/auth/api-keys
   Header: Authorization: Bearer <JWT>
   
   ↓
   
2. Response
   {
     "apiKey": "ibn_abc123...",
     "expiresAt": "2025-12-31T23:59:59Z"
   }
   
   ↓
   
3. Client uses API Key
   GET /api/v1/network/peers
   Header: X-API-Key: ibn_abc123...
   
   ↓
   
4. AuthMiddleware.Authenticate()
   - Extract API key from header
   - Validate in database
   - Check expiration
   - Add user info to context
   
   ↓
   
5. Handler processes request
```

---

## 💾 Data Flow: Batch Operations

### Create Batch Flow

```
1. Client Request
   POST /api/v1/batches
   Header: Authorization: Bearer <JWT>
   Body: {
     "batchId": "BATCH001",
     "farmLocation": "Moc Chau",
     "harvestDate": "2024-11-12",
     "processingInfo": "Organic",
     "qualityCert": "VN-ORG-2024"
   }
   
   ↓
   
2. BatchHandler.CreateBatch()
   - Validate request
   - Check authentication
   
   ↓
   
3. BatchService.CreateBatch()
   - Check Redis cache (5min TTL)
   - Prepare chaincode invoke
   
   ↓
   
4. ChaincodeService.Invoke()
   - Function: "createBatch"
   - Args: [batchId, farmLocation, harvestDate, processingInfo, qualityCert]
   
   ↓
   
5. Fabric Gateway SDK
   - Submit transaction to Fabric
   - Wait for commit
   
   ↓
   
6. Response
   {
     "batchId": "BATCH001",
     "status": "CREATED",
     "transactionId": "tx123...",
     "hashValue": "abc123..."
   }
   
   ↓
   
7. Cache Result
   - Store in Redis (5min TTL)
   - Key: batch:BATCH001
```

---

## 🔧 Service Initialization Flow

### Startup Sequence (main.go)

```
1. Load Configuration
   config.Load()
   ├─> Server config (host, port)
   ├─> Database config (PostgreSQL)
   ├─> Redis config
   ├─> Fabric config (channel, chaincode, certificates)
   └─> JWT config

2. Initialize Logger
   utils.NewLogger()
   └─> Zap logger (JSON format)

3. Connect to PostgreSQL
   config.NewPostgresPool()
   └─> Connection pool (min: 5, max: 25)

4. Connect to Redis
   cache.NewService()
   └─> Redis client for rate limiting & cache

5. Initialize Fabric Gateway
   fabric.NewGatewayService()
   ├─> Load certificates
   ├─> Create gRPC connection
   ├─> Create Gateway client
   └─> Connect to peer0.org1.ibn.vn:7051

6. Initialize Services
   ├─> AuthService
   ├─> ChaincodeService
   ├─> TransactionService
   ├─> NetworkService
   ├─> DiscoveryService
   ├─> ChannelService
   ├─> EventService
   ├─> ExplorerService
   ├─> MetricsService
   ├─> AuditService
   ├─> ACLService
   └─> BatchService

7. Initialize Handlers
   ├─> AuthHandler
   ├─> ChaincodeHandler
   ├─> TransactionHandler
   ├─> NetworkHandler
   ├─> ChannelHandler
   ├─> EventHandler
   ├─> ExplorerHandler
   ├─> MetricsHandler
   ├─> AuditHandler
   ├─> ACLHandler
   ├─> UserHandler
   └─> BatchHandler

8. Initialize Middleware
   ├─> AuthMiddleware
   ├─> RateLimitMiddleware
   ├─> LoggerMiddleware
   ├─> AuditMiddleware
   ├─> CORSMiddleware
   ├─> RecoveryMiddleware
   └─> TracingMiddleware

9. Setup Routes
   routes.SetupRoutes()
   └─> Chi router with all endpoints

10. Start HTTP Server
    http.ListenAndServe(":8080", router)
    └─> Server running on port 8080
```

---

## 📊 Component Interactions

### Service Dependencies

```
FabricGatewayService (Core)
    │
    ├─> ChaincodeService
    │   └─> ChaincodeHandler
    │
    ├─> ContractService
    │   └─> BatchHandler
    │
    ├─> NetworkService
    │   └─> NetworkHandler
    │
    └─> DiscoveryService
        └─> DiscoveryHandler

PostgreSQL
    │
    ├─> AuthService
    │   └─> AuthHandler
    │
    ├─> TransactionService
    │   └─> TransactionHandler
    │
    ├─> AuditService
    │   └─> AuditHandler
    │
    ├─> MetricsService
    │   └─> MetricsHandler
    │
    ├─> ACLService
    │   └─> ACLHandler
    │
    └─> IndexerService
        └─> ExplorerService
            └─> ExplorerHandler

Redis
    │
    ├─> RateLimitMiddleware
    ├─> CacheService
    │   └─> BatchService
    └─> WebSocketRateLimitMiddleware
```

---

## 🎯 Key Features

### 1. Authentication & Authorization
- ✅ JWT tokens (access + refresh)
- ✅ API Keys (service-to-service)
- ✅ Role-based access control (RBAC)
- ✅ Resource-based permissions (ACL)
- ✅ Password hashing (bcrypt)

### 2. Security
- ✅ Rate limiting (Redis-based)
- ✅ CORS protection
- ✅ TLS/SSL (via Nginx)
- ✅ Request validation
- ✅ Audit logging

### 3. Performance
- ✅ Connection pooling (PostgreSQL)
- ✅ Redis caching (batch operations)
- ✅ Compression (Gzip)
- ✅ Parallel requests (network discovery)

### 4. Monitoring
- ✅ Request logging (structured)
- ✅ Audit logs (database)
- ✅ Metrics (Prometheus)
- ✅ Health checks

### 5. Real-time
- ✅ WebSocket support (dashboard)
- ✅ Event subscriptions
- ✅ Block indexer (background)

---

## 📈 Thống Kê

### Endpoints
- **Tổng số:** ~90+ endpoints
- **Public:** 5 endpoints (health, metrics, swagger)
- **Protected:** 85+ endpoints (require auth)

### Services
- **Tổng số:** 14 services
- **Core:** Fabric Gateway, Auth, Chaincode
- **Business:** Batch, Transaction, Network
- **Support:** Metrics, Audit, ACL, Explorer

### Database
- **Migrations:** 5 migrations
- **Tables:** 15+ tables
- **Connection Pool:** 5-25 connections

### Middleware
- **Global:** 7 middleware
- **Route-specific:** 3 middleware
- **Custom:** 9 middleware

---

## 🚀 Quick Start

### 1. Configuration

```bash
# Environment variables
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_NAME=ibn_gateway
REDIS_HOST=redis
REDIS_PORT=6379
FABRIC_CHANNEL=ibnchannel
FABRIC_CHAINCODE=teaTraceCC
FABRIC_PEER_ENDPOINT=peer0.org1.ibn.vn:7051
```

### 2. Run

```bash
# Development
go run cmd/server/main.go

# Production (Docker)
docker-compose up -d
```

### 3. Test

```bash
# Health check
curl http://localhost:8080/health

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ibn.vn","password":"..."}'

# Get network info
curl http://localhost:8080/api/v1/network/info \
  -H "Authorization: Bearer <token>"
```

---

## 📝 Notes

- ✅ **Lifecycle operations** (install/approve/commit) đã chuyển sang Admin Service
- ✅ **API Gateway** chỉ xử lý business transactions (invoke/query)
- ✅ **WebSocket** có rate limiting riêng (100 messages/min)
- ✅ **Batch operations** có Redis cache (5min TTL)
- ✅ **Network discovery** sử dụng config-based discovery
- ✅ **Channel operations** (write) trả về "pending" - cần peer CLI để thực thi

---

**Last Updated:** 2025-01-27
