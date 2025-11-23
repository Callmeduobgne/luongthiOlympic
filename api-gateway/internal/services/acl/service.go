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

package acl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// Service handles ACL operations
type Service struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewService creates a new ACL service
func NewService(dbPool *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{
		db:     dbPool,
		logger: logger,
	}
}

// CreatePolicy creates a new ACL policy
func (s *Service) CreatePolicy(ctx context.Context, req *models.CreatePolicyRequest) (*models.ACLPolicy, error) {
	s.logger.Info("Creating ACL policy", zap.String("name", req.Name))

	// Convert actions to PostgreSQL array
	actionsJSON, err := json.Marshal(req.Actions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal actions: %w", err)
	}

	// Convert conditions to JSONB
	conditionsJSON, err := json.Marshal(req.Conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	// Set default priority if not provided
	priority := req.Priority
	if priority == 0 {
		priority = 0
	}

	var id uuid.UUID
	var createdAt, updatedAt time.Time
	err = s.db.QueryRow(ctx, `
		INSERT INTO acl_policies (name, description, resource_type, resource_pattern, actions, conditions, priority, is_active)
		VALUES ($1, $2, $3, $4, $5::text[], $6::jsonb, $7, true)
		RETURNING id, created_at, updated_at
	`, req.Name, req.Description, req.ResourceType, req.ResourcePattern, string(actionsJSON), string(conditionsJSON), priority).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("policy with name '%s' already exists", req.Name)
		}
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	policy := &models.ACLPolicy{
		ID:            id.String(),
		Name:          req.Name,
		Description:   req.Description,
		ResourceType:  req.ResourceType,
		ResourcePattern: req.ResourcePattern,
		Actions:       req.Actions,
		Conditions:    req.Conditions,
		Priority:      priority,
		IsActive:      true,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return policy, nil
}

// GetPolicy gets a policy by ID
func (s *Service) GetPolicy(ctx context.Context, policyID string) (*models.ACLPolicy, error) {
	var id uuid.UUID
	var name, description, resourceType, resourcePattern string
	var actionsJSON string
	var conditionsJSON []byte
	var priority int
	var isActive bool
	var createdAt, updatedAt time.Time

	err := s.db.QueryRow(ctx, `
		SELECT id, name, description, resource_type, resource_pattern, actions::text, conditions, priority, is_active, created_at, updated_at
		FROM acl_policies
		WHERE id = $1
	`, policyID).Scan(&id, &name, &description, &resourceType, &resourcePattern, &actionsJSON, &conditionsJSON, &priority, &isActive, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("policy not found: %w", err)
	}

	// Parse actions
	var actions []string
	if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
		// Fallback: try to parse as array directly
		actions = []string{actionsJSON}
	}

	// Parse conditions
	var conditions map[string]interface{}
	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
			conditions = make(map[string]interface{})
		}
	} else {
		conditions = make(map[string]interface{})
	}

	policy := &models.ACLPolicy{
		ID:            id.String(),
		Name:          name,
		Description:   description,
		ResourceType:  resourceType,
		ResourcePattern: resourcePattern,
		Actions:       actions,
		Conditions:    conditions,
		Priority:      priority,
		IsActive:      isActive,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return policy, nil
}

// ListPolicies lists all policies with pagination
func (s *Service) ListPolicies(ctx context.Context, page, pageSize int, resourceType string) (*models.ListPoliciesResponse, error) {
	offset := (page - 1) * pageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	query := `
		SELECT id, name, description, resource_type, resource_pattern, actions::text, conditions, priority, is_active, created_at, updated_at
		FROM acl_policies
		WHERE ($1::text IS NULL OR resource_type = $1)
		ORDER BY priority DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	var resourceTypeFilter interface{}
	if resourceType != "" {
		resourceTypeFilter = resourceType
	}

	rows, err := s.db.Query(ctx, query, resourceTypeFilter, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	defer rows.Close()

	policies := []models.ACLPolicy{}
	for rows.Next() {
		var id uuid.UUID
		var name, description, resourceType, resourcePattern string
		var actionsJSON string
		var conditionsJSON []byte
		var priority int
		var isActive bool
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &description, &resourceType, &resourcePattern, &actionsJSON, &conditionsJSON, &priority, &isActive, &createdAt, &updatedAt); err != nil {
			continue
		}

		// Parse actions
		var actions []string
		if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
			actions = []string{actionsJSON}
		}

		// Parse conditions
		var conditions map[string]interface{}
		if len(conditionsJSON) > 0 {
			if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
				conditions = make(map[string]interface{})
			}
		} else {
			conditions = make(map[string]interface{})
		}

		policies = append(policies, models.ACLPolicy{
			ID:            id.String(),
			Name:          name,
			Description:   description,
			ResourceType:  resourceType,
			ResourcePattern: resourcePattern,
			Actions:       actions,
			Conditions:    conditions,
			Priority:      priority,
			IsActive:      isActive,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		})
	}

	// Count total
	var total int
	countQuery := `
		SELECT COUNT(*) FROM acl_policies
		WHERE ($1::text IS NULL OR resource_type = $1)
	`
	err = s.db.QueryRow(ctx, countQuery, resourceTypeFilter).Scan(&total)
	if err != nil {
		s.logger.Warn("Failed to count policies", zap.Error(err))
		total = len(policies)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &models.ListPoliciesResponse{
		Policies:  policies,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdatePolicy updates a policy
func (s *Service) UpdatePolicy(ctx context.Context, policyID string, req *models.UpdatePolicyRequest) (*models.ACLPolicy, error) {
	s.logger.Info("Updating ACL policy", zap.String("policyId", policyID))

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.ResourceType != nil {
		updates = append(updates, fmt.Sprintf("resource_type = $%d", argIndex))
		args = append(args, *req.ResourceType)
		argIndex++
	}
	if req.ResourcePattern != nil {
		updates = append(updates, fmt.Sprintf("resource_pattern = $%d", argIndex))
		args = append(args, *req.ResourcePattern)
		argIndex++
	}
	if req.Actions != nil {
		actionsJSON, err := json.Marshal(req.Actions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal actions: %w", err)
		}
		updates = append(updates, fmt.Sprintf("actions = $%d::text[]", argIndex))
		args = append(args, string(actionsJSON))
		argIndex++
	}
	if req.Conditions != nil {
		conditionsJSON, err := json.Marshal(req.Conditions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal conditions: %w", err)
		}
		updates = append(updates, fmt.Sprintf("conditions = $%d::jsonb", argIndex))
		args = append(args, string(conditionsJSON))
		argIndex++
	}
	if req.Priority != nil {
		updates = append(updates, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *req.Priority)
		argIndex++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updates) == 0 {
		return s.GetPolicy(ctx, policyID)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, policyID)

	query := fmt.Sprintf(`
		UPDATE acl_policies
		SET %s
		WHERE id = $%d
		RETURNING id, name, description, resource_type, resource_pattern, actions::text, conditions, priority, is_active, created_at, updated_at
	`, strings.Join(updates, ", "), argIndex)

	var id uuid.UUID
	var name, description, resourceType, resourcePattern string
	var actionsJSON string
	var conditionsJSON []byte
	var priority int
	var isActive bool
	var createdAt, updatedAt time.Time

	err := s.db.QueryRow(ctx, query, args...).Scan(&id, &name, &description, &resourceType, &resourcePattern, &actionsJSON, &conditionsJSON, &priority, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	// Parse actions
	var actions []string
	if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
		actions = []string{actionsJSON}
	}

	// Parse conditions
	var conditions map[string]interface{}
	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
			conditions = make(map[string]interface{})
		}
	} else {
		conditions = make(map[string]interface{})
	}

	policy := &models.ACLPolicy{
		ID:            id.String(),
		Name:          name,
		Description:   description,
		ResourceType:  resourceType,
		ResourcePattern: resourcePattern,
		Actions:       actions,
		Conditions:    conditions,
		Priority:      priority,
		IsActive:      isActive,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return policy, nil
}

// DeletePolicy deletes a policy
func (s *Service) DeletePolicy(ctx context.Context, policyID string) error {
	s.logger.Info("Deleting ACL policy", zap.String("policyId", policyID))

	result, err := s.db.Exec(ctx, "DELETE FROM acl_policies WHERE id = $1", policyID)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("policy not found")
	}

	return nil
}

// ListPermissions lists all predefined permissions
func (s *Service) ListPermissions(ctx context.Context, resourceType string) (*models.ListPermissionsResponse, error) {
	query := `
		SELECT id, name, description, resource_type, action, created_at
		FROM acl_permissions
		WHERE ($1::text IS NULL OR resource_type = $1)
		ORDER BY resource_type, action
	`

	var resourceTypeFilter interface{}
	if resourceType != "" {
		resourceTypeFilter = resourceType
	}

	rows, err := s.db.Query(ctx, query, resourceTypeFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	permissions := []models.ACLPermission{}
	for rows.Next() {
		var id uuid.UUID
		var name, description, resourceType, action string
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &description, &resourceType, &action, &createdAt); err != nil {
			continue
		}

		permissions = append(permissions, models.ACLPermission{
			ID:           id.String(),
			Name:         name,
			Description:  description,
			ResourceType: resourceType,
			Action:       action,
			CreatedAt:    createdAt,
		})
	}

	return &models.ListPermissionsResponse{
		Permissions: permissions,
		Total:       len(permissions),
	}, nil
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *Service) CheckPermission(ctx context.Context, req *models.CheckPermissionRequest) (*models.CheckPermissionResponse, error) {
	s.logger.Info("Checking permission",
		zap.String("userId", req.UserID),
		zap.String("resourceType", req.ResourceType),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action),
	)

	// Get user's role
	var userRole string
	err := s.db.QueryRow(ctx, "SELECT role FROM users WHERE id = $1", req.UserID).Scan(&userRole)
	if err != nil {
		return &models.CheckPermissionResponse{
			Allowed: false,
			Reason:  "User not found",
		}, nil
	}

	// Check role-based permissions first (higher priority)
	roleAllowed, rolePolicy := s.checkRolePermission(ctx, userRole, req.ResourceType, req.Resource, req.Action)
	if roleAllowed {
		return &models.CheckPermissionResponse{
			Allowed:    true,
			PolicyID:   &rolePolicy.ID,
			PolicyName: &rolePolicy.Name,
			Reason:     fmt.Sprintf("Allowed by role policy: %s", rolePolicy.Name),
		}, nil
	}

	// Check user-specific permissions
	userAllowed, userPolicy := s.checkUserPermission(ctx, req.UserID, req.ResourceType, req.Resource, req.Action)
	if userAllowed {
		return &models.CheckPermissionResponse{
			Allowed:    true,
			PolicyID:   &userPolicy.ID,
			PolicyName: &userPolicy.Name,
			Reason:     fmt.Sprintf("Allowed by user policy: %s", userPolicy.Name),
		}, nil
	}

	return &models.CheckPermissionResponse{
		Allowed: false,
		Reason:  "No matching policy found",
	}, nil
}

// checkRolePermission checks if a role has permission
func (s *Service) checkRolePermission(ctx context.Context, role, resourceType, resource, action string) (bool, *models.ACLPolicy) {
	query := `
		SELECT p.id, p.name, p.resource_type, p.resource_pattern, p.actions::text, p.priority
		FROM role_permissions rp
		JOIN acl_policies p ON rp.policy_id = p.id
		WHERE rp.role = $1 AND p.is_active = true
		ORDER BY p.priority DESC
	`

	rows, err := s.db.Query(ctx, query, role)
	if err != nil {
		return false, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, policyResourceType, resourcePattern, actionsJSON string
		var priority int

		if err := rows.Scan(&id, &name, &policyResourceType, &resourcePattern, &actionsJSON, &priority); err != nil {
			continue
		}

		// Parse actions
		var actions []string
		if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
			continue
		}

		// Check if policy matches
		if s.matchesPolicy(policyResourceType, resourcePattern, resourceType, resource, action, actions) {
			return true, &models.ACLPolicy{
				ID:            id.String(),
				Name:          name,
				ResourceType:  policyResourceType,
				ResourcePattern: resourcePattern,
				Actions:       actions,
				Priority:      priority,
			}
		}
	}

	return false, nil
}

// checkUserPermission checks if a user has specific permission
func (s *Service) checkUserPermission(ctx context.Context, userID, resourceType, resource, action string) (bool, *models.ACLPolicy) {
	query := `
		SELECT p.id, p.name, p.resource_type, p.resource_pattern, p.actions::text, p.priority
		FROM user_permissions up
		JOIN acl_policies p ON up.policy_id = p.id
		WHERE up.user_id = $1 AND up.is_active = true AND p.is_active = true
		AND (up.expires_at IS NULL OR up.expires_at > NOW())
		ORDER BY p.priority DESC
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return false, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, policyResourceType, resourcePattern, actionsJSON string
		var priority int

		if err := rows.Scan(&id, &name, &policyResourceType, &resourcePattern, &actionsJSON, &priority); err != nil {
			continue
		}

		// Parse actions
		var actions []string
		if err := json.Unmarshal([]byte(actionsJSON), &actions); err != nil {
			continue
		}

		// Check if policy matches
		if s.matchesPolicy(policyResourceType, resourcePattern, resourceType, resource, action, actions) {
			return true, &models.ACLPolicy{
				ID:            id.String(),
				Name:          name,
				ResourceType:  policyResourceType,
				ResourcePattern: resourcePattern,
				Actions:       actions,
				Priority:      priority,
			}
		}
	}

	return false, nil
}

// matchesPolicy checks if a policy matches the resource and action
func (s *Service) matchesPolicy(policyResourceType, resourcePattern, resourceType, resource, action string, allowedActions []string) bool {
	// Check resource type
	if policyResourceType != "all" && policyResourceType != resourceType {
		return false
	}

	// Check action
	actionAllowed := false
	for _, allowedAction := range allowedActions {
		if allowedAction == action || allowedAction == "admin" {
			actionAllowed = true
			break
		}
	}
	if !actionAllowed {
		return false
	}

	// Check resource pattern
	if resourcePattern == "" || resourcePattern == "*" {
		return true
	}

	// Simple pattern matching (supports * wildcard)
	return s.matchPattern(resourcePattern, resource)
}

// matchPattern performs simple wildcard pattern matching
func (s *Service) matchPattern(pattern, text string) bool {
	// Exact match
	if pattern == text {
		return true
	}

	// Wildcard * matches everything
	if pattern == "*" {
		return true
	}

	// Check if pattern contains wildcard
	if strings.Contains(pattern, "*") {
		// Convert pattern to regex-like matching
		parts := strings.Split(pattern, "*")
		if len(parts) == 0 {
			return true
		}

		// Check if text starts with first part
		if parts[0] != "" && !strings.HasPrefix(text, parts[0]) {
			return false
		}

		// Check if text ends with last part
		if len(parts) > 1 && parts[len(parts)-1] != "" {
			if !strings.HasSuffix(text, parts[len(parts)-1]) {
				return false
			}
		}

		// For simple cases, check if all parts are in order
		remaining := text
		for i, part := range parts {
			if part == "" {
				continue
			}
			idx := strings.Index(remaining, part)
			if idx == -1 {
				return false
			}
			if i == 0 && idx != 0 {
				return false
			}
			remaining = remaining[idx+len(part):]
		}

		return true
	}

	// No wildcard, exact match required
	return pattern == text
}

