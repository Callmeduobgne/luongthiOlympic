// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

