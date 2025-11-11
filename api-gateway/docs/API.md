# API Documentation

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

### JWT Authentication

Include JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

### API Key Authentication

Include API key in the X-API-Key header:

```
X-API-Key: <api-key>
```

## Endpoints

### Batch Operations

#### Create Batch

**POST** `/batches`

Creates a new tea batch on the blockchain.

**Authorization:** Required (Farmer role)

**Request Body:**
```json
{
  "batchId": "BATCH001",
  "farmLocation": "Moc Chau, Son La",
  "harvestDate": "2024-11-08",
  "processingInfo": "Organic processing, no pesticides",
  "qualityCert": "VN-ORG-2024"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "batchId": "BATCH001",
    "farmLocation": "Moc Chau, Son La",
    "harvestDate": "2024-11-08",
    "processingInfo": "Organic processing, no pesticides",
    "qualityCert": "VN-ORG-2024",
    "hashValue": "a1b2c3d4...",
    "owner": "Org1MSP",
    "timestamp": "2024-11-08T10:00:00.000Z",
    "status": "CREATED"
  }
}
```

#### Get Batch

**GET** `/batches/:id`

Retrieves batch information.

**Authorization:** Optional (Public endpoint)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "batchId": "BATCH001",
    "farmLocation": "Moc Chau, Son La",
    ...
  }
}
```

#### Verify Batch

**POST** `/batches/:id/verify`

Verifies batch hash.

**Authorization:** Required (Farmer, Verifier, or Admin role)

**Request Body:**
```json
{
  "hashInput": "BATCH001|Moc Chau, Son La|2024-11-08|Organic processing|VN-ORG-2024"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "isValid": true,
    "batch": {
      "batchId": "BATCH001",
      "status": "VERIFIED",
      ...
    }
  }
}
```

#### Update Batch Status

**PATCH** `/batches/:id/status`

Updates batch status.

**Authorization:** Required (Farmer or Admin role)

**Request Body:**
```json
{
  "status": "VERIFIED"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "batchId": "BATCH001",
    "status": "VERIFIED",
    ...
  }
}
```

### Health & Monitoring

#### Health Check

**GET** `/health`

Returns health status of all services.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600,
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "fabric": "healthy"
  }
}
```

#### Readiness Check

**GET** `/ready`

Kubernetes readiness probe.

#### Liveness Check

**GET** `/live`

Kubernetes liveness probe.

#### Metrics

**GET** `/metrics`

Prometheus metrics endpoint.

## Error Responses

All errors follow this format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {}
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| BAD_REQUEST | 400 | Invalid request |
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Resource conflict |
| RATE_LIMIT_EXCEEDED | 429 | Too many requests |
| INTERNAL_SERVER_ERROR | 500 | Server error |
| SERVICE_UNAVAILABLE | 503 | Service unavailable |

## Rate Limiting

Default rate limit: 1000 requests per hour per user/IP.

When rate limit is exceeded, you'll receive:

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Max 1000 requests per 1h"
  }
}
```

## Swagger Documentation

Interactive API documentation available at:

```
http://localhost:8080/swagger/index.html
```

