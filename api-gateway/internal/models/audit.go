package models

import "time"

// AuditLogResponse represents an audit log in API response
type AuditLogResponse struct {
	ID           int64                  `json:"id"`
	UserID       string                 `json:"userId,omitempty"`
	ApiKeyID     string                 `json:"apiKeyId,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resourceType,omitempty"`
	ResourceID   string                 `json:"resourceId,omitempty"`
	TxID         string                 `json:"txId,omitempty"`
	Status       string                 `json:"status"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IpAddress    string                 `json:"ipAddress,omitempty"`
	UserAgent    string                 `json:"userAgent,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
}

// AuditLogListQuery represents query parameters for listing audit logs
type AuditLogListQuery struct {
	UserID    string     `json:"userId,omitempty"`
	Action    string     `json:"action,omitempty"`
	TxID      string     `json:"txId,omitempty"`
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

