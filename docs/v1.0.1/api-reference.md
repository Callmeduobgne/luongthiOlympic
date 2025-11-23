# API Reference - IBN Network Backend

**Ngày tạo:** 2025-01-27  
**Version:** 1.0.1  
**Base URL:** `http://localhost:9090/api/v1`  
**Format:** JSON

---

## 📋 Mục Lục

1. [Authentication](#authentication)
2. [Blockchain Operations](#blockchain-operations)
3. [TeaTrace Chaincode](#teatrace-chaincode)
4. [Chaincode Lifecycle](#chaincode-lifecycle)
5. [Events & Subscriptions](#events--subscriptions)
6. [Monitoring & Audit](#monitoring--audit)
7. [Error Codes](#error-codes)

---

## Authentication

### POST /auth/register

Đăng ký tài khoản mới.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "full_name": "Nguyễn Văn A"
}
```

**Response:** `201 Created`
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "Nguyễn Văn A",
    "created_at": "2025-01-27T10:00:00Z"
  },
  "message": "User registered successfully"
}
```

**Errors:**
- `400 Bad Request`: Validation failed
- `409 Conflict`: Email already exists

---

### POST /auth/login

Đăng nhập và nhận JWT tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

**Errors:**
- `401 Unauthorized`: Invalid credentials

---

### POST /auth/refresh

Refresh access token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

---

### POST /auth/logout

Đăng xuất (invalidate refresh token).

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

---

### GET /profile

Lấy thông tin profile của user hiện tại.

**Headers:**
- `Authorization: Bearer <token>` hoặc `X-API-Key: <api_key>`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "full_name": "Nguyễn Văn A",
  "role": "user",
  "created_at": "2025-01-27T10:00:00Z"
}
```

---

### POST /api-keys

Tạo API key mới.

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "name": "My API Key",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response:** `201 Created`
```json
{
  "api_key": "ibn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "name": "My API Key",
  "created_at": "2025-01-27T10:00:00Z",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

⚠️ **Lưu ý:** API key chỉ hiển thị một lần duy nhất khi tạo.

---

## Blockchain Operations

### POST /blockchain/transactions

Gửi transaction lên blockchain.

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "chaincode_name": "teaTraceCC",
  "function": "CreateBatch",
  "args": ["batch001", "Tea Batch 1", "2025-01-27"],
  "channel": "ibnchannel"
}
```

**Response:** `201 Created`
```json
{
  "transaction_id": "uuid",
  "txid": "abc123...",
  "status": "submitted",
  "submitted_at": "2025-01-27T10:00:00Z"
}
```

---

### POST /blockchain/query

Query chaincode (read-only).

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "chaincode_name": "teaTraceCC",
  "function": "GetBatch",
  "args": ["batch001"],
  "channel": "ibnchannel"
}
```

**Response:** `200 OK`
```json
{
  "result": "...",
  "queried_at": "2025-01-27T10:00:00Z"
}
```

---

### GET /blockchain/transactions

Lấy danh sách transactions.

**Headers:**
- `Authorization: Bearer <token>`

**Query Parameters:**
- `limit` (int, default: 50): Số lượng records
- `offset` (int, default: 0): Offset
- `status` (string): Filter by status (submitted, committed, failed)
- `chaincode_name` (string): Filter by chaincode

**Response:** `200 OK`
```json
{
  "transactions": [
    {
      "id": "uuid",
      "txid": "abc123...",
      "status": "committed",
      "chaincode_name": "teaTraceCC",
      "submitted_at": "2025-01-27T10:00:00Z",
      "committed_at": "2025-01-27T10:00:05Z"
    }
  ],
  "total": 100,
  "limit": 50,
  "offset": 0
}
```

---

### GET /blockchain/transactions/{id}

Lấy thông tin transaction theo ID.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "txid": "abc123...",
  "status": "committed",
  "chaincode_name": "teaTraceCC",
  "function": "CreateBatch",
  "args": ["batch001"],
  "channel": "ibnchannel",
  "submitted_at": "2025-01-27T10:00:00Z",
  "committed_at": "2025-01-27T10:00:05Z"
}
```

---

### GET /blockchain/txid/{txid}

Lấy thông tin transaction theo TXID.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "txid": "abc123...",
  "status": "committed",
  "block_number": 100,
  "created_at": "2025-01-27T10:00:00Z"
}
```

---

### GET /blockchain/channel/info

Lấy thông tin channel.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "channel_name": "ibnchannel",
  "block_height": 100,
  "chaincodes": [
    {
      "name": "teaTraceCC",
      "version": "1.0"
    }
  ]
}
```

---

### GET /blockchain/blocks/{number}

Lấy block theo số.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "block_number": 100,
  "block_hash": "abc123...",
  "previous_hash": "def456...",
  "data_hash": "ghi789...",
  "transactions": [
    {
      "txid": "abc123...",
      "type": "ENDORSER_TRANSACTION"
    }
  ],
  "created_at": "2025-01-27T10:00:00Z"
}
```

---

## TeaTrace Chaincode

### GET /teatrace/health

Kiểm tra health của TeaTrace chaincode.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "chaincode": "teaTraceCC",
  "version": "1.0",
  "channel": "ibnchannel"
}
```

---

### POST /teatrace/batches

Tạo tea batch mới.

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "batch_id": "BATCH001",
  "tea_type": "Oolong",
  "origin": "Lâm Đồng",
  "harvest_date": "2025-01-15",
  "quantity_kg": 100.5,
  "farmer_name": "Nguyễn Văn B",
  "certification": "Organic"
}
```

**Response:** `201 Created`
```json
{
  "batch_id": "BATCH001",
  "status": "created",
  "created_at": "2025-01-27T10:00:00Z",
  "txid": "abc123..."
}
```

---

### GET /teatrace/batches

Lấy danh sách tất cả batches.

**Headers:**
- `Authorization: Bearer <token>`

**Query Parameters:**
- `limit` (int, default: 50)
- `offset` (int, default: 0)
- `status` (string): Filter by status

**Response:** `200 OK`
```json
{
  "batches": [
    {
      "batch_id": "BATCH001",
      "tea_type": "Oolong",
      "origin": "Lâm Đồng",
      "status": "created",
      "created_at": "2025-01-27T10:00:00Z"
    }
  ],
  "total": 10,
  "limit": 50,
  "offset": 0
}
```

---

### GET /teatrace/batches/{batchId}

Lấy thông tin batch theo ID.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "batch_id": "BATCH001",
  "tea_type": "Oolong",
  "origin": "Lâm Đồng",
  "harvest_date": "2025-01-15",
  "quantity_kg": 100.5,
  "farmer_name": "Nguyễn Văn B",
  "certification": "Organic",
  "status": "created",
  "hash": "abc123...",
  "created_at": "2025-01-27T10:00:00Z"
}
```

---

### POST /teatrace/batches/{batchId}/verify

Xác thực batch hash.

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "expected_hash": "abc123..."
}
```

**Response:** `200 OK`
```json
{
  "valid": true,
  "message": "Hash matches"
}
```

---

### PUT /teatrace/batches/{batchId}/status

Cập nhật trạng thái batch.

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "status": "processed",
  "notes": "Đã xử lý và đóng gói"
}
```

**Response:** `200 OK`
```json
{
  "batch_id": "BATCH001",
  "status": "processed",
  "updated_at": "2025-01-27T10:30:00Z"
}
```

---

## Chaincode Lifecycle

### POST /chaincode/upload

Upload chaincode package (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`
- `Content-Type: multipart/form-data`

**Form Data:**
- `package` (file): Chaincode package file (.tar.gz)
- `name` (string): Chaincode name
- `version` (string): Chaincode version

**Response:** `201 Created`
```json
{
  "package_id": "myChaincode:1.0:abc123...",
  "name": "myChaincode",
  "version": "1.0",
  "uploaded_at": "2025-01-27T10:00:00Z"
}
```

---

### POST /chaincode/install

Cài đặt chaincode lên peer (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Request:**
```json
{
  "package_id": "myChaincode:1.0:abc123...",
  "peer": "peer0.org1.ibn.vn:7051"
}
```

**Response:** `200 OK`
```json
{
  "message": "Chaincode installed successfully",
  "package_id": "myChaincode:1.0:abc123..."
}
```

---

### POST /chaincode/approve

Phê duyệt chaincode definition (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Request:**
```json
{
  "chaincode_name": "myChaincode",
  "version": "1.0",
  "package_id": "myChaincode:1.0:abc123...",
  "sequence": 1,
  "channel": "ibnchannel"
}
```

**Response:** `200 OK`
```json
{
  "message": "Chaincode approved successfully"
}
```

---

### POST /chaincode/commit

Commit chaincode definition (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Request:**
```json
{
  "chaincode_name": "myChaincode",
  "version": "1.0",
  "sequence": 1,
  "channel": "ibnchannel"
}
```

**Response:** `200 OK`
```json
{
  "message": "Chaincode committed successfully"
}
```

---

### GET /chaincode/installed

Lấy danh sách chaincode đã cài đặt (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Response:** `200 OK`
```json
{
  "chaincodes": [
    {
      "package_id": "myChaincode:1.0:abc123...",
      "name": "myChaincode",
      "version": "1.0",
      "installed_at": "2025-01-27T10:00:00Z"
    }
  ]
}
```

---

### GET /chaincode/committed

Lấy danh sách chaincode đã commit (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Response:** `200 OK`
```json
{
  "chaincodes": [
    {
      "name": "myChaincode",
      "version": "1.0",
      "sequence": 1,
      "channel": "ibnchannel",
      "committed_at": "2025-01-27T10:00:00Z"
    }
  ]
}
```

---

## Events & Subscriptions

### POST /events/subscriptions

Tạo event subscription.

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "event_type": "transaction.committed",
  "webhook_url": "https://example.com/webhook",
  "filters": {
    "chaincode_name": "teaTraceCC",
    "status": "committed"
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "event_type": "transaction.committed",
  "webhook_url": "https://example.com/webhook",
  "active": true,
  "created_at": "2025-01-27T10:00:00Z"
}
```

---

### GET /events/subscriptions

Lấy danh sách subscriptions của user.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "subscriptions": [
    {
      "id": "uuid",
      "event_type": "transaction.committed",
      "webhook_url": "https://example.com/webhook",
      "active": true,
      "created_at": "2025-01-27T10:00:00Z"
    }
  ]
}
```

---

### GET /events/subscriptions/{id}

Lấy thông tin subscription.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "event_type": "transaction.committed",
  "webhook_url": "https://example.com/webhook",
  "filters": {
    "chaincode_name": "teaTraceCC"
  },
  "active": true,
  "created_at": "2025-01-27T10:00:00Z"
}
```

---

### PUT /events/subscriptions/{id}

Cập nhật subscription.

**Headers:**
- `Authorization: Bearer <token>`

**Request:**
```json
{
  "webhook_url": "https://new-url.com/webhook",
  "active": true
}
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "webhook_url": "https://new-url.com/webhook",
  "active": true,
  "updated_at": "2025-01-27T10:30:00Z"
}
```

---

### DELETE /events/subscriptions/{id}

Xóa subscription.

**Headers:**
- `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "message": "Subscription deleted successfully"
}
```

---

## Monitoring & Audit

### GET /metrics

Lấy tất cả metrics (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Response:** `200 OK`
```json
{
  "metrics": [
    {
      "name": "request_count",
      "value": 1000,
      "timestamp": "2025-01-27T10:00:00Z"
    }
  ]
}
```

---

### GET /metrics/summary

Lấy summary metrics (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Response:** `200 OK`
```json
{
  "total_requests": 1000,
  "success_rate": 0.98,
  "average_response_time_ms": 150,
  "error_rate": 0.02
}
```

---

### GET /audit/logs

Query audit logs (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Query Parameters:**
- `limit` (int, default: 50)
- `offset` (int, default: 0)
- `user_id` (uuid): Filter by user
- `action` (string): Filter by action
- `start_date` (datetime): Start date
- `end_date` (datetime): End date

**Response:** `200 OK`
```json
{
  "logs": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "action": "transaction.submitted",
      "resource": "blockchain/transactions",
      "ip_address": "192.168.1.1",
      "created_at": "2025-01-27T10:00:00Z"
    }
  ],
  "total": 100,
  "limit": 50,
  "offset": 0
}
```

---

### GET /audit/search

Tìm kiếm audit logs (Admin only).

**Headers:**
- `Authorization: Bearer <admin_token>`

**Query Parameters:**
- `query` (string): Search query
- `limit` (int, default: 20)

**Response:** `200 OK`
```json
{
  "logs": [...],
  "total": 10
}
```

---

## Error Codes

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

### Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  }
}
```

### Common Error Codes

- `INVALID_REQUEST`: Request validation failed
- `UNAUTHORIZED`: Authentication required
- `FORBIDDEN`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `CONFLICT`: Resource already exists
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Server error

---

**Last Updated:** 2025-01-27  
**Version:** 1.0.1

