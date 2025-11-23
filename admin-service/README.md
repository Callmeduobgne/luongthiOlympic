# IBN Admin Service

Production-ready Admin Service for Hyperledger Fabric chaincode lifecycle management.

## Overview

Admin Service is a dedicated microservice responsible for:
- Chaincode lifecycle operations (install, approve, commit)
- Chaincode query operations (list installed, list committed)
- Isolated admin operations with proper security

## Architecture

```
Frontend → Backend → Admin Service → Peer CLI → Blockchain
```

**Separation of Concerns:**
- **Admin Service**: Chaincode lifecycle operations
- **API Gateway**: Business transaction routing (invoke/query)
- **Backend**: Business logic and orchestration

## Features

- ✅ Chaincode installation
- ✅ Chaincode approval
- ✅ Chaincode commit
- ✅ List installed chaincodes
- ✅ List committed chaincodes
- ✅ Get committed chaincode info
- ✅ API key authentication
- ✅ Peer CLI integration
- ✅ Production-ready logging

## Configuration

### Environment Variables

```bash
# Server
ADMIN_SERVER_PORT=8090
ADMIN_SERVER_HOST=0.0.0.0
ADMIN_SERVER_ENV=production

# Fabric
ADMIN_FABRIC_CHANNEL=ibnchannel
ADMIN_FABRIC_MSP_ID=Org1MSP
ADMIN_FABRIC_PEER_ENDPOINT=peer0.org1.ibn.vn:7051
ADMIN_FABRIC_USER_CERT_PATH=/app/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/signcerts/Admin@org1.ibn.vn-cert.pem
ADMIN_FABRIC_USER_KEY_PATH=/app/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/keystore
ADMIN_FABRIC_PEER_TLS_CA_PATH=/app/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
ADMIN_FABRIC_PEER_HOST_OVERRIDE=peer0.org1.ibn.vn

# Authentication
ADMIN_AUTH_API_KEY=your-secret-api-key-min-32-chars
ADMIN_AUTH_ENABLED=true

# Logging
ADMIN_LOGGING_LEVEL=info
ADMIN_LOGGING_FORMAT=json
ADMIN_LOGGING_OUTPUT=stdout
```

## API Endpoints

### Health Check
- `GET /health` - Health check (no auth required)
- `GET /ready` - Readiness check (no auth required)

### Chaincode Lifecycle
- `POST /api/v1/chaincode/install` - Install chaincode
- `POST /api/v1/chaincode/approve` - Approve chaincode definition
- `POST /api/v1/chaincode/commit` - Commit chaincode definition

### Chaincode Query
- `GET /api/v1/chaincode/installed` - List installed chaincodes
- `GET /api/v1/chaincode/committed?channel=<channel>` - List committed chaincodes
- `GET /api/v1/chaincode/committed/info?name=<name>&channel=<channel>` - Get committed chaincode info

### Authentication

All API endpoints (except health checks) require API key authentication:

```bash
# Using X-API-Key header
curl -H "X-API-Key: your-secret-api-key" http://localhost:8090/api/v1/chaincode/installed

# Using Authorization header
curl -H "Authorization: Bearer your-secret-api-key" http://localhost:8090/api/v1/chaincode/installed
```

## Development

### Build
```bash
go build ./cmd/server/main.go
```

### Run
```bash
./admin-service
```

### Docker Build
```bash
docker build -f docker/Dockerfile -t ibn-admin-service:latest .
```

## Production Deployment

See `docker-compose.yml` for production deployment configuration.

## Security

- API key authentication for all admin operations
- Internal network only (not exposed to public)
- Isolated from business transaction services
- Audit logging for all operations

## License

Apache 2.0

