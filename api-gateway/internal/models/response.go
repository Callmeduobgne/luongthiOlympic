package models

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents an error response
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta represents metadata for paginated responses
type Meta struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"pageSize,omitempty"`
	TotalPages int `json:"totalPages,omitempty"`
	TotalCount int `json:"totalCount,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Uptime   int64             `json:"uptime"`
	Services map[string]string `json:"services"`
}

// Error codes
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeInternalServer      = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeBatchExists         = "BATCH_EXISTS"
	ErrCodeBatchNotFound       = "BATCH_NOT_FOUND"
	ErrCodeInvalidStatus       = "INVALID_STATUS"
	ErrCodeVerificationFailed  = "VERIFICATION_FAILED"
	ErrCodeNetworkError        = "NETWORK_ERROR"
	ErrCodeTransactionFailed   = "TRANSACTION_FAILED"
	ErrCodeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"
)

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string, details interface{}) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, page, pageSize, totalCount int) *APIResponse {
	totalPages := (totalCount + pageSize - 1) / pageSize
	return &APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			TotalCount: totalCount,
		},
	}
}

