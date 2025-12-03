# GitHub Copilot Instructions - IBN Network Backend

**Version:** 1.0.1  
**Last Updated:** 2025-01-27

---

## Project Context

IBN Network là hệ thống blockchain dựa trên Hyperledger Fabric với Backend API được viết bằng Go. Hệ thống quản lý tea traceability thông qua chaincode TeaTraceCC.

### Technology Stack
- **Language:** Go 1.24.0
- **HTTP Router:** go-chi/chi v5
- **Database:** PostgreSQL 15 với connection pooling
- **Cache:** Redis 7 (multi-layer: L1 Memory + L2 Redis)
- **Blockchain:** Hyperledger Fabric 2.5.9 (qua API Gateway)
- **Logging:** go.uber.org/zap
- **Deployment:** Docker Compose

### Architecture
- **Pattern:** Layered Architecture (Presentation → Business Logic → Data Access → Infrastructure)
- **Design:** Domain-Driven Design (DDD)
- **Style:** Monolithic

---

## Code Style Guidelines

### Go Conventions
- Sử dụng `gofmt` standards
- Package names: lowercase, single word
- Exported: PascalCase
- Private: camelCase
- Error handling: Luôn kiểm tra và return errors
- Context: Truyền `context.Context` cho async operations

### File Structure
```
backend/internal/
├── handlers/        # HTTP handlers (Presentation Layer)
├── services/        # Business logic (Service Layer)
├── infrastructure/ # Database, Cache, Gateway
└── middleware/      # HTTP middleware
```

### Naming
- Handlers: `{Resource}Handler`
- Services: `{Domain}Service`
- Repositories: `{Resource}Repository`
- Models: PascalCase

---

## API Development Rules

### Endpoints
- Base: `/api/v1`
- RESTful: GET (read), POST (create), PUT (update), DELETE (delete)
- Auth: JWT Bearer Token hoặc API Key header
- Format: JSON

### Request Handling
```go
// 1. Parse request body
var req RequestModel
json.NewDecoder(r.Body).Decode(&req)

// 2. Validate input
if err := utils.ValidateStruct(&req); err != nil {
    respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
    return
}

// 3. Call service
result, err := h.service.Method(r.Context(), &req)
if err != nil {
    respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
    return
}

// 4. Return response
respondSuccess(w, http.StatusOK, result)
```

### Response Format
- Success: `{"success": true, "data": {...}}`
- Error: `{"error": {"code": "...", "message": "...", "details": {...}}}`

---

## Database Rules

### Schema Organization
- Schemas: `auth`, `blockchain`, `events`, `access`, `audit`
- Tables: Snake_case
- Migrations: Numbered prefix (001_, 002_, ...)
- Queries: Use `sqlc` for type-safe queries

### Database Access Pattern
```go
// Use connection pool
db := infrastructure.GetDB()

// Use transactions for complex operations
tx, err := db.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

// Commit on success
return tx.Commit(ctx)
```

---

## Caching Strategy

### Multi-Layer Cache
- **L1:** In-memory (5-15 min TTL) - Hot data
- **L2:** Redis (30 min - 1h TTL) - User permissions, API keys
- **L3:** Database - Persistent data

### Cache Pattern Example
```go
// Check L1 cache
if data := l1Cache.Get(key); data != nil {
    return data
}

// Check L2 cache (Redis)
if data := redis.Get(key); data != nil {
    l1Cache.Set(key, data, 5*time.Minute)
    return data
}

// Query database
data := db.Query(key)
redis.Set(key, data, 30*time.Minute)
l1Cache.Set(key, data, 5*time.Minute)
return data
```

---

## Blockchain Integration

### ⚠️ CRITICAL: Gateway Architecture
**Backend KHÔNG kết nối trực tiếp với Fabric!**

```
Backend → API Gateway (HTTP) → Fabric Network
```

### Gateway Client Usage
```go
// Location: internal/infrastructure/gateway/client.go
gatewayClient := infrastructure.NewGatewayClient(cfg)

// Submit transaction
txID, err := gatewayClient.SubmitTransaction(ctx, &TransactionRequest{
    ChaincodeName: "teaTraceCC",
    Function: "CreateBatch",
    Args: []string{"batch001"},
    Channel: "ibnchannel",
})

// Query chaincode
result, err := gatewayClient.QueryChaincode(ctx, &QueryRequest{
    ChaincodeName: "teaTraceCC",
    Function: "GetBatch",
    Args: []string{"batch001"},
    Channel: "ibnchannel",
})
```

### Configuration
```go
GATEWAY_ENABLED=true (REQUIRED)
GATEWAY_BASE_URL=http://api-gateway:8080
GATEWAY_API_KEY=optional-service-key
GATEWAY_TIMEOUT=30s
```

---

## Error Handling

### Error Response Format
```go
type ErrorResponse struct {
    Error struct {
        Code    string      `json:"code"`
        Message string      `json:"message"`
        Details interface{} `json:"details,omitempty"`
    } `json:"error"`
}
```

### Error Handling Pattern
```go
func (h *Handler) Method(w http.ResponseWriter, r *http.Request) {
    // ... processing ...
    
    if err != nil {
        h.logger.Error("Operation failed", 
            zap.Error(err),
            zap.String("user_id", userID),
            zap.String("request_id", requestID),
        )
        
        // Don't expose internal errors
        respondError(w, http.StatusInternalServerError, 
            "INTERNAL_ERROR", 
            "An error occurred. Please try again later.")
        return
    }
}
```

### Common Error Codes
- `INVALID_REQUEST`: Validation failed
- `UNAUTHORIZED`: Authentication required
- `FORBIDDEN`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `CONFLICT`: Resource already exists
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Server error

---

## Logging

### Structured Logging
```go
import "go.uber.org/zap"

// Logger initialization
logger, _ := zap.NewProduction()
defer logger.Sync()

// Log with context
logger.Info("User logged in",
    zap.String("user_id", userID),
    zap.String("email", email),
    zap.Time("timestamp", time.Now()),
)

// Error logging
logger.Error("Operation failed",
    zap.Error(err),
    zap.String("operation", "create_batch"),
    zap.Any("request", req),
)
```

### Log Levels
- **DEBUG:** Detailed information for debugging
- **INFO:** General informational messages
- **WARN:** Warning messages (non-critical issues)
- **ERROR:** Error messages (needs attention)

---

## Authentication & Authorization

### JWT Authentication
```go
// Generate tokens
accessToken, refreshToken, err := auth.GenerateTokens(userID, email, role)

// Validate token
claims, err := auth.ValidateToken(tokenString)

// Extract user from context (after middleware)
userID := r.Context().Value("user_id").(string)
```

### API Key Authentication
```go
// Create API key
apiKey := fmt.Sprintf("ibn_%s", generateRandomString(32))

// Validate API key
userID, err := auth.ValidateAPIKey(apiKey)
```

### Middleware Usage
```go
// Apply auth middleware
router.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware)
    r.Get("/profile", handler.GetProfile)
})
```

---

## Service Layer Pattern

### Service Structure
```go
type Service struct {
    repository *Repository
    cache      *Cache
    logger     *zap.Logger
}

func NewService(repo *Repository, cache *Cache, logger *zap.Logger) *Service {
    return &Service{
        repository: repo,
        cache:      cache,
        logger:     logger,
    }
}

func (s *Service) Method(ctx context.Context, req *Request) (*Response, error) {
    // 1. Validate business rules
    // 2. Check cache
    // 3. Call repository
    // 4. Update cache
    // 5. Return result
}
```

---

## Testing Guidelines

### Unit Test Pattern
```go
func TestService_Method(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Response
        wantErr bool
    }{
        {
            name:  "success case",
            input: "valid_input",
            want:  &Response{...},
        },
        {
            name:    "error case",
            input:   "invalid_input",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## Common Tasks

### Adding New Endpoint
1. Create request/response models
2. Add service method
3. Add handler method
4. Add route in router
5. Add Swagger annotations
6. Write tests

### Adding Database Table
1. Create migration file: `migrations/XXX_table_name.up.sql`
2. Create down migration: `XXX_table_name.down.sql`
3. Generate sqlc queries
4. Create repository
5. Update service layer

### Adding Cache
1. Check if data is cacheable
2. Add cache key generation
3. Implement cache-aside pattern
4. Set appropriate TTL
5. Handle cache invalidation

---

## Important Reminders

1. ⚠️ **Backend không kết nối trực tiếp với Fabric** - phải qua API Gateway
2. ✅ **Luôn validate input** trước khi process
3. ✅ **Log errors với context** (user_id, request_id, etc.)
4. ✅ **Sử dụng transactions** cho operations phức tạp
5. ✅ **Cache frequently accessed data**
6. ✅ **Return appropriate HTTP status codes**
7. ✅ **Don't expose internal errors** to clients

---

## References

- **Backend Architecture:** `docs/v1.0.1/backend.md`
- **API Reference:** `docs/v1.0.1/api-reference.md`
- **Cursor Rules:** `.cursorrules`

---

**When in doubt, refer to existing code patterns in the codebase!**

