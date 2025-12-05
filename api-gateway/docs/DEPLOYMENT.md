# Deployment Guide

## Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 16
- Redis 7
- Hyperledger Fabric network running

## Local Development

### 1. Clone and Setup

```bash
cd /home/exp2/ibn/api-gateway

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
nano .env
```

### 2. Install Dependencies

```bash
# Download Go dependencies
make deps

# Install additional tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### 3. Setup Database

```bash
# Start PostgreSQL and Redis
docker-compose -f docker/docker-compose.yml up -d postgres redis

# Run migrations
make migrate-up

# Initialize database with default data
psql -h localhost -U gateway -d ibn_gateway -f scripts/init-db.sql
```

### 4. Generate Code

```bash
# Generate sqlc code
make sqlc

# Generate Swagger docs
make swagger
```

### 5. Setup Fabric Wallet

```bash
# Ensure Fabric network is running
cd ../core
docker-compose up -d

# Setup wallet
cd ../api-gateway
make setup-wallet
```

### 6. Test Connection

```bash
make test-connection
```

### 7. Run Application

```bash
# Development mode
make run

# Production build
make build
./bin/api-gateway
```

## Docker Deployment

### 1. Build and Start

```bash
# Build Docker image
make docker-build

# Start all services (3 gateway instances + Nginx LB)
make docker-up

# View logs
make docker-logs
```

### 2. Verify Deployment

```bash
# Check health
curl http://localhost:8080/health

# Check Swagger UI
open http://localhost:8080/swagger/index.html
```

### 3. Stop Services

```bash
make docker-down
```

## Production Deployment

### Environment Variables

Update `.env` for production:

```bash
# Server
GATEWAY_ENV=production
JWT_SECRET=<strong-random-secret-32-chars-minimum>

# Database
POSTGRES_PASSWORD=<strong-database-password>

# Redis
REDIS_PASSWORD=<strong-redis-password>

# Fabric
FABRIC_PEER_ENDPOINT=peer0.org1.ibn.vn:7051
```

### Security Checklist

- [ ] Change all default passwords
- [ ] Use strong JWT secret (32+ characters)
- [ ] Enable HTTPS/TLS
- [ ] Configure firewall rules
- [ ] Set up log rotation
- [ ] Enable audit logging
- [ ] Review CORS settings
- [ ] Set appropriate rate limits
- [ ] Configure backup strategy

### Scaling

#### Horizontal Scaling

Add more gateway instances in `docker-compose.yml`:

```yaml
gateway4:
  # Same config as gateway1-3
  ports:
    - "8084:8080"
```

Update Nginx upstream in `nginx/nginx.conf`:

```nginx
upstream api_gateway_backend {
    least_conn;
    server gateway1:8080;
    server gateway2:8080;
    server gateway3:8080;
    server gateway4:8080;
}
```

#### Database Scaling

For PostgreSQL:
- Use read replicas for read-heavy workloads
- Implement connection pooling (already configured)
- Use pgBouncer for additional connection pooling

For Redis:
- Use Redis Cluster for high availability
- Configure Redis Sentinel for failover

### Monitoring

Metrics are exported to Prometheus at `/metrics`.

Add to `monitoring/prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'api-gateway'
    static_configs:
      - targets: 
        - 'gateway1:8080'
        - 'gateway2:8080'
        - 'gateway3:8080'
    metrics_path: '/metrics'
```

Import Grafana dashboard from `monitoring/grafana-gateway-dashboard.json`.

### Backup Strategy

#### Database Backup

```bash
# Backup PostgreSQL
docker exec api-gateway-postgres pg_dump -U gateway ibn_gateway > backup_$(date +%Y%m%d).sql

# Restore
docker exec -i api-gateway-postgres psql -U gateway ibn_gateway < backup.sql
```

#### Redis Backup

```bash
# Redis automatically creates snapshots
# Configure in docker-compose.yml:
# command: redis-server --save 60 1 --appendonly yes
```

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to Fabric network

**Solution:**
1. Check if Fabric network is running: `docker ps | grep peer`
2. Verify certificate paths in `.env`
3. Test connection: `make test-connection`

**Problem:** Database connection failed

**Solution:**
1. Check if PostgreSQL is running: `docker ps | grep postgres`
2. Verify credentials in `.env`
3. Check logs: `docker logs api-gateway-postgres`

### Performance Issues

**Problem:** Slow API responses

**Solution:**
1. Check Prometheus metrics
2. Enable Redis caching
3. Review database query performance
4. Check circuit breaker state

### Circuit Breaker

**Problem:** Circuit breaker is open

**Solution:**
1. Check Fabric network health
2. Review circuit breaker logs
3. Adjust circuit breaker thresholds in `.env`
4. Wait for automatic recovery

## Maintenance

### Database Migrations

```bash
# Create new migration
make migrate-create NAME=add_new_table

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down
```

### Log Rotation

Configure in Docker Compose:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### Updating Dependencies

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Rebuild
make build
make docker-build
```

