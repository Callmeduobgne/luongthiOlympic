package acl

import (
	"time"

	"github.com/google/uuid"
)

// Policy represents an ACL policy
type Policy struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Actions      []string               `json:"actions"`
	Effect       string                 `json:"effect"`
	Conditions   map[string]interface{} `json:"conditions,omitempty"`
	Priority     int                    `json:"priority"`
	IsActive     bool                   `json:"is_active"`
	CreatedBy    *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeletedAt    *time.Time             `json:"deleted_at,omitempty"`
}

// Permission represents a resource-action permission
type Permission struct {
	ID        uuid.UUID              `json:"id"`
	PolicyID  uuid.UUID              `json:"policy_id"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Granted   bool                   `json:"granted"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// UserPermission represents a user-policy assignment
type UserPermission struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	PolicyID  uuid.UUID              `json:"policy_id"`
	GrantedBy *uuid.UUID             `json:"granted_by,omitempty"`
	GrantedAt time.Time              `json:"granted_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	IsActive  bool                   `json:"is_active"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// RolePermission represents a role-policy assignment
type RolePermission struct {
	ID        uuid.UUID  `json:"id"`
	Role      string     `json:"role"`
	PolicyID  uuid.UUID  `json:"policy_id"`
	GrantedBy *uuid.UUID `json:"granted_by,omitempty"`
	GrantedAt time.Time  `json:"granted_at"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreatePolicyRequest represents policy creation request
type CreatePolicyRequest struct {
	Name         string                 `json:"name" validate:"required"`
	Description  *string                `json:"description,omitempty"`
	ResourceType string                 `json:"resource_type" validate:"required"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Actions      []string               `json:"actions" validate:"required,min=1"`
	Effect       string                 `json:"effect" validate:"required,oneof=allow deny"`
	Conditions   map[string]interface{} `json:"conditions,omitempty"`
	Priority     int                    `json:"priority"`
}

// UpdatePolicyRequest represents policy update request
type UpdatePolicyRequest struct {
	Name         *string                `json:"name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Actions      []string               `json:"actions,omitempty"`
	Effect       *string                `json:"effect,omitempty"`
	Conditions   map[string]interface{} `json:"conditions,omitempty"`
	Priority     *int                   `json:"priority,omitempty"`
	IsActive     *bool                  `json:"is_active,omitempty"`
}

// AssignUserPermissionRequest represents user permission assignment
type AssignUserPermissionRequest struct {
	UserID    uuid.UUID  `json:"user_id" validate:"required"`
	PolicyID  uuid.UUID  `json:"policy_id" validate:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// AssignRolePermissionRequest represents role permission assignment
type AssignRolePermissionRequest struct {
	Role     string    `json:"role" validate:"required"`
	PolicyID uuid.UUID `json:"policy_id" validate:"required"`
}

// CheckPermissionRequest represents permission check request
type CheckPermissionRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Resource string    `json:"resource" validate:"required"`
	Action   string    `json:"action" validate:"required"`
}

// CheckPermissionResponse represents permission check response
type CheckPermissionResponse struct {
	Allowed    bool     `json:"allowed"`
	MatchedPolicies []uuid.UUID `json:"matched_policies,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

// Effect constants
const (
	EffectAllow = "allow"
	EffectDeny  = "deny"
)

// Common resource types
const (
	ResourceTypeChannel    = "channel"
	ResourceTypeChaincode  = "chaincode"
	ResourceTypeTransaction = "transaction"
	ResourceTypeUser       = "user"
	ResourceTypeAPIKey     = "api_key"
	ResourceTypePolicy     = "policy"
)

// Common actions
const (
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionDelete = "delete"
	ActionExecute = "execute"
	ActionManage = "manage"
)

