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

package audit

import (
	"time"

	"github.com/google/uuid"
)

// Log represents an audit log entry
type Log struct {
	ID           uuid.UUID              `json:"id"`
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Status       string                 `json:"status"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	UserAgent    *string                `json:"user_agent,omitempty"`
	RequestID    *string                `json:"request_id,omitempty"`
	Method       *string                `json:"method,omitempty"`
	Path         *string                `json:"path,omitempty"`
	DurationMs   *int                   `json:"duration_ms,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// CreateLogRequest represents audit log creation request
type CreateLogRequest struct {
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Status       string                 `json:"status"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	UserAgent    *string                `json:"user_agent,omitempty"`
	RequestID    *string                `json:"request_id,omitempty"`
	Method       *string                `json:"method,omitempty"`
	Path         *string                `json:"path,omitempty"`
	DurationMs   *int                   `json:"duration_ms,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// QueryLogsRequest represents audit log query request
type QueryLogsRequest struct {
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	Action       *string    `json:"action,omitempty"`
	ResourceType *string    `json:"resource_type,omitempty"`
	Status       *string    `json:"status,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

// Status constants
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

// Action constants
const (
	ActionLogin           = "auth.login"
	ActionLogout          = "auth.logout"
	ActionRegister        = "auth.register"
	ActionCreateAPIKey    = "auth.create_api_key"
	ActionRevokeAPIKey    = "auth.revoke_api_key"
	ActionCreatePolicy    = "acl.create_policy"
	ActionUpdatePolicy    = "acl.update_policy"
	ActionDeletePolicy    = "acl.delete_policy"
	ActionAssignPermission = "acl.assign_permission"
	ActionRevokePermission = "acl.revoke_permission"
	ActionSubmitTransaction = "blockchain.submit_transaction"
	ActionQueryChaincode  = "blockchain.query_chaincode"
	ActionInvokeChaincode = "blockchain.invoke_chaincode"
)

