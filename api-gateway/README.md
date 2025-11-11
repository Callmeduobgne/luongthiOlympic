# IBN API Gateway

Production-ready API Gateway for Hyperledger Fabric IBN Network - Tea Traceability System

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Features

- ✅ RESTful API for tea batch operations
- ✅ JWT & API Key authentication
- ✅ Redis-based rate limiting with sliding window
- ✅ PostgreSQL with type-safe sqlc queries (NO GORM)
- ✅ Fabric Gateway integration with circuit breaker
- ✅ OpenTelemetry distributed tracing
- ✅ Prometheus metrics
- ✅ OpenAPI/Swagger documentation
- ✅ Graceful shutdown with 30s timeout
- ✅ Nginx load balancer (3 instances)
- ✅ Docker & Kubernetes ready
- ✅ Structured logging with zap
- ✅ Configuration validation with Viper

## Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| **Language** | Go | 1.21+ |
| **Web Framework** | Chi Router | v5 |
| **Database** | PostgreSQL + pgx + sqlc | 16 |
| **Migrations** | golang-migrate | v4 |
| **Cache** | Redis | 7 |
| **Blockchain** | Hyperledger Fabric | 2.5 |
| **Load Balancer** | Nginx | latest |
| **Config** | Viper + Validator | latest |
| **Logging** | zap | latest |
| **Tracing** | OpenTelemetry | latest |
| **Resilience** | gobreaker | latest |
| **Monitoring** | Prometheus + Grafana | latest |

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 16
- Redis 7
- Hyperledger Fabric network running

### Installation

```bash
# Clone repository
cd /home/exp2/ibn/api-gateway

# Install dependencies
make deps

# Setup environment
cp .env.example .env
# Edit .env with your configuration

# Start database services
docker-compose -f docker/docker-compose.yml up -d postgres redis

# Run database migrations
make migrate-up

# Generate sqlc code
make sqlc

# Setup Fabric wallet
make setup-wallet

# Test connections
make test-connection

# Build application
make build

# Run application
make run
```

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Start all services (Postgres, Redis, 3 Gateway instances, Nginx)
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

Access the API:
- **API Endpoint**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics

## API Endpoints

### Batch Operations

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/batches` | Create new tea batch | ✅ Farmer |
| GET | `/api/v1/batches/:id` | Get batch info | ❌ Public |
| POST | `/api/v1/batches/:id/verify` | Verify batch hash | ✅ Farmer/Verifier/Admin |
| PATCH | `/api/v1/batches/:id/status` | Update batch status | ✅ Farmer/Admin |

### Health & Monitoring

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check (all services) |
| GET | `/ready` | Readiness check (K8s) |
| GET | `/live` | Liveness check (K8s) |
| GET | `/metrics` | Prometheus metrics |
| GET | `/swagger/*` | Swagger documentation |

## Configuration

See `.env.example` for all configuration options.

### Critical Configuration

```bash
# Server
GATEWAY_PORT=8080
GATEWAY_ENV=production  # development | staging | production

# Database (with connection pooling)
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=ibn_gateway
POSTGRES_MAX_OPEN_CONNS=25
POSTGRES_MAX_IDLE_CONNS=10
POSTGRES_CONN_MAX_LIFETIME=5m

# Fabric Network
FABRIC_CHANNEL=ibnchannel
FABRIC_CHAINCODE=teaTraceCC
FABRIC_MSP_ID=Org1MSP
FABRIC_PEER_ENDPOINT=localhost:7051

# Circuit Breaker
CIRCUIT_BREAKER_MAX_REQUESTS=3
CIRCUIT_BREAKER_INTERVAL=10s
CIRCUIT_BREAKER_TIMEOUT=60s
CIRCUIT_BREAKER_FAILURE_RATIO=0.6

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1h
```

## Development

### Available Commands

```bash
make help              # Show all commands
make build             # Build application
make run               # Run application
make test              # Run unit tests
make test-integration  # Run integration tests
make lint              # Run linter
make fmt               # Format code
make sqlc              # Generate sqlc code
make swagger           # Generate Swagger docs
make migrate-up        # Run migrations
make migrate-down      # Rollback migrations
make docker-build      # Build Docker image
make docker-up         # Start all services
make docker-down       # Stop all services
```

### Code Generation

#### Generate sqlc code

```bash
# After modifying SQL queries in queries/
make sqlc
```

#### Generate Swagger docs

```bash
# After updating API handlers with godoc comments
make swagger
```

### Database Migrations

```bash
# Create new migration
make migrate-create NAME=add_new_table

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down
```

## Architecture

### System Architecture

```
Client → Nginx LB → API Gateway (3 instances) → Fabric Network
                         ↓
                    PostgreSQL + Redis
                         ↓
                    Prometheus + Grafana
```

### Components

- **API Gateway**: Chi router with middleware stack
- **Fabric Service**: Gateway SDK with circuit breaker & retry logic
- **Database**: PostgreSQL with sqlc (type-safe SQL, NO GORM)
- **Cache**: Redis for rate limiting & query caching
- **Load Balancer**: Nginx with health checks
- **Monitoring**: Prometheus + Grafana + OpenTelemetry

### Middleware Stack

Request flow through middleware:

```
Request
  → Recovery (panic recovery)
  → CORS
  → Logger (with correlation ID)
  → Tracing (OpenTelemetry)
  → Authentication (JWT/API Key)
  → Rate Limiting (Redis-based)
  → Validation
  → Handler
```

## Monitoring

### Prometheus Metrics

Available at `http://localhost:8080/metrics`

**Custom Metrics:**
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration histogram
- `blockchain_transactions_total` - Blockchain transactions
- `blockchain_transaction_duration_seconds` - Transaction duration
- `cache_hits_total` - Cache hits
- `cache_misses_total` - Cache misses
- `circuit_breaker_state` - Circuit breaker state

### Grafana Dashboard

Import dashboard from `monitoring/grafana-gateway-dashboard.json`

Access Grafana at: http://localhost:3000

### OpenTelemetry Tracing

Enable in `.env`:

```bash
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

## Security

### Authentication

Two authentication methods supported:

1. **JWT Token** (for user sessions)
   ```
   Authorization: Bearer <jwt-token>
   ```

2. **API Key** (for service-to-service)
   ```
   X-API-Key: <api-key>
   ```

### Rate Limiting

Redis-based sliding window rate limiting:
- Default: 1000 requests per hour
- Per user/API key/IP
- Configurable in `.env`

### MSP Authorization

Blockchain operations require specific MSP roles:

| Operation | Allowed MSPs |
|-----------|--------------|
| Create Batch | Org1MSP (Farmer) |
| Verify Batch | Org1MSP, Org2MSP, Org3MSP |
| Get Batch Info | Public |
| Update Status | Org1MSP (Farmer), Org3MSP (Admin) |

### TLS Configuration

All Fabric communications use TLS:
- Peer connections: grpcs://
- Certificate validation enabled
- Hostname override for localhost development

## Performance

### Connection Pooling

**PostgreSQL:**
- Max Open Connections: 25
- Max Idle Connections: 10
- Connection Max Lifetime: 5m
- Health check period: 30s

**Circuit Breaker:**
- Max Requests (half-open): 3
- Interval: 10s
- Timeout: 60s
- Failure Ratio: 0.6

### Caching Strategy

**Batch queries cached for 5 minutes:**
- Key format: `batch:{batchId}`
- Automatic invalidation on updates
- Cache hit metrics tracked

### Load Balancing

Nginx with `least_conn` algorithm:
- 3 gateway instances
- Health checks every 30s
- Automatic failover
- Connection keepalive

## Testing

```bash
# Run all tests
make test

# Run integration tests
make test-integration

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Troubleshooting

### Common Issues

**Problem:** Go not found

**Solution:**
```bash
sudo snap install go
# or
sudo apt install golang-go
```

**Problem:** Cannot connect to Fabric

**Solution:**
```bash
# Check Fabric network
cd ../core
docker ps | grep peer

# Verify certificates
ls -la ../core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/

# Test connection
cd ../api-gateway
make test-connection
```

**Problem:** Database migration failed

**Solution:**
```bash
# Check PostgreSQL logs
docker logs api-gateway-postgres

# Retry migration
make migrate-down
make migrate-up
```

## License

Apache 2.0

## Support

For issues and questions:
- Email: support@ibn.vn
- Documentation: See `docs/` directory

