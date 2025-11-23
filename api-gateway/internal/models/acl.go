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

// ACLPolicy represents an ACL policy
type ACLPolicy struct {
	ID            string                 `json:"id" db:"id"`
	Name          string                 `json:"name" db:"name"`
	Description   string                 `json:"description,omitempty" db:"description"`
	ResourceType  string                 `json:"resourceType" db:"resource_type"` // channel, chaincode, endpoint, all
	ResourcePattern string               `json:"resourcePattern,omitempty" db:"resource_pattern"`
	Actions       []string               `json:"actions" db:"actions"` // read, write, invoke, query, admin
	Conditions    map[string]interface{} `json:"conditions,omitempty" db:"conditions"`
	Priority      int                    `json:"priority" db:"priority"`
	IsActive      bool                   `json:"isActive" db:"is_active"`
	CreatedAt     time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time              `json:"updatedAt" db:"updated_at"`
}

// ACLPermission represents a predefined permission
type ACLPermission struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description,omitempty" db:"description"`
	ResourceType string    `json:"resourceType" db:"resource_type"`
	Action       string    `json:"action" db:"action"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// UserPermission represents a user's permission assignment
type UserPermission struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"userId" db:"user_id"`
	PolicyID  string     `json:"policyId" db:"policy_id"`
	GrantedBy *string    `json:"grantedBy,omitempty" db:"granted_by"`
	GrantedAt time.Time  `json:"grantedAt" db:"granted_at"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty" db:"expires_at"`
	IsActive  bool       `json:"isActive" db:"is_active"`
	Policy    *ACLPolicy `json:"policy,omitempty"` // Joined policy data
}

// RolePermission represents a role's permission assignment
type RolePermission struct {
	ID        string     `json:"id" db:"id"`
	Role      string     `json:"role" db:"role"`
	PolicyID  string     `json:"policyId" db:"policy_id"`
	GrantedBy *string    `json:"grantedBy,omitempty" db:"granted_by"`
	GrantedAt time.Time  `json:"grantedAt" db:"granted_at"`
	Policy    *ACLPolicy `json:"policy,omitempty"` // Joined policy data
}

// CreatePolicyRequest represents a request to create a new policy
type CreatePolicyRequest struct {
	Name          string                 `json:"name" validate:"required,min=3,max=255"`
	Description   string                 `json:"description,omitempty"`
	ResourceType  string                 `json:"resourceType" validate:"required,oneof=channel chaincode endpoint all"`
	ResourcePattern string               `json:"resourcePattern,omitempty"`
	Actions       []string               `json:"actions" validate:"required,min=1"`
	Conditions    map[string]interface{} `json:"conditions,omitempty"`
	Priority      int                    `json:"priority,omitempty"`
}

// UpdatePolicyRequest represents a request to update a policy
type UpdatePolicyRequest struct {
	Name          *string                `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description   *string                `json:"description,omitempty"`
	ResourceType  *string                `json:"resourceType,omitempty" validate:"omitempty,oneof=channel chaincode endpoint all"`
	ResourcePattern *string              `json:"resourcePattern,omitempty"`
	Actions       []string               `json:"actions,omitempty" validate:"omitempty,min=1"`
	Conditions    map[string]interface{} `json:"conditions,omitempty"`
	Priority      *int                   `json:"priority,omitempty"`
	IsActive      *bool                  `json:"isActive,omitempty"`
}

// CheckPermissionRequest represents a request to check permission
type CheckPermissionRequest struct {
	UserID       string `json:"userId" validate:"required"`
	ResourceType string `json:"resourceType" validate:"required"`
	Resource     string `json:"resource" validate:"required"` // e.g., 'ibnchannel', 'teaTraceCC', '/api/v1/transactions'
	Action       string `json:"action" validate:"required"`   // read, write, invoke, query, admin
}

// CheckPermissionResponse represents the result of a permission check
type CheckPermissionResponse struct {
	Allowed    bool     `json:"allowed"`
	PolicyID   *string  `json:"policyId,omitempty"`
	PolicyName *string  `json:"policyName,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

// ListPoliciesResponse represents a paginated list of policies
type ListPoliciesResponse struct {
	Policies  []ACLPolicy `json:"policies"`
	Total     int          `json:"total"`
	Page      int          `json:"page"`
	PageSize  int          `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// ListPermissionsResponse represents a list of permissions
type ListPermissionsResponse struct {
	Permissions []ACLPermission `json:"permissions"`
	Total       int             `json:"total"`
}

// AssignUserPermissionRequest represents a request to assign a policy to a user
type AssignUserPermissionRequest struct {
	UserID    string     `json:"userId" validate:"required"`
	PolicyID  string     `json:"policyId" validate:"required"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// AssignRolePermissionRequest represents a request to assign a policy to a role
type AssignRolePermissionRequest struct {
	Role     string `json:"role" validate:"required,oneof=user farmer verifier admin"`
	PolicyID string `json:"policyId" validate:"required"`
}

