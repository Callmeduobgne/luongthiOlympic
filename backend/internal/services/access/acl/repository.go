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
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles ACL data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new ACL repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreatePolicy creates a new ACL policy
func (r *Repository) CreatePolicy(ctx context.Context, policy *Policy) error {
	query := `
		INSERT INTO access.acl_policies 
		(id, name, description, resource_type, resource_id, actions, effect, conditions, priority, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.ResourceType,
		policy.ResourceID, policy.Actions, policy.Effect, policy.Conditions,
		policy.Priority, policy.IsActive, policy.CreatedBy,
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// GetPolicyByID retrieves a policy by ID
func (r *Repository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*Policy, error) {
	query := `
		SELECT id, name, description, resource_type, resource_id, actions, effect,
		       conditions, priority, is_active, created_by, created_at, updated_at, deleted_at
		FROM access.acl_policies
		WHERE id = $1 AND deleted_at IS NULL
	`

	policy := &Policy{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.ResourceType,
		&policy.ResourceID, &policy.Actions, &policy.Effect, &policy.Conditions,
		&policy.Priority, &policy.IsActive, &policy.CreatedBy,
		&policy.CreatedAt, &policy.UpdatedAt, &policy.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("policy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	return policy, nil
}

// GetPolicyByName retrieves a policy by name
func (r *Repository) GetPolicyByName(ctx context.Context, name string) (*Policy, error) {
	query := `
		SELECT id, name, description, resource_type, resource_id, actions, effect,
		       conditions, priority, is_active, created_by, created_at, updated_at, deleted_at
		FROM access.acl_policies
		WHERE name = $1 AND deleted_at IS NULL
	`

	policy := &Policy{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.ResourceType,
		&policy.ResourceID, &policy.Actions, &policy.Effect, &policy.Conditions,
		&policy.Priority, &policy.IsActive, &policy.CreatedBy,
		&policy.CreatedAt, &policy.UpdatedAt, &policy.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("policy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	return policy, nil
}

// ListPolicies lists all active policies
func (r *Repository) ListPolicies(ctx context.Context, resourceType *string) ([]*Policy, error) {
	query := `
		SELECT id, name, description, resource_type, resource_id, actions, effect,
		       conditions, priority, is_active, created_by, created_at, updated_at, deleted_at
		FROM access.acl_policies
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argPos := 1

	if resourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argPos)
		args = append(args, *resourceType)
		argPos++
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	defer rows.Close()

	var policies []*Policy
	for rows.Next() {
		policy := &Policy{}
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.ResourceType,
			&policy.ResourceID, &policy.Actions, &policy.Effect, &policy.Conditions,
			&policy.Priority, &policy.IsActive, &policy.CreatedBy,
			&policy.CreatedAt, &policy.UpdatedAt, &policy.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// UpdatePolicy updates a policy
func (r *Repository) UpdatePolicy(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	query := "UPDATE access.acl_policies SET "
	args := []interface{}{}
	argPos := 1

	for field, value := range updates {
		if argPos > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, argPos)
		args = append(args, value)
		argPos++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

// DeletePolicy soft deletes a policy
func (r *Repository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE access.acl_policies
		SET deleted_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}

// AssignUserPermission assigns a policy to a user
func (r *Repository) AssignUserPermission(ctx context.Context, perm *UserPermission) error {
	query := `
		INSERT INTO access.user_permissions
		(id, user_id, policy_id, granted_by, granted_at, expires_at, is_active, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		perm.ID, perm.UserID, perm.PolicyID, perm.GrantedBy,
		perm.GrantedAt, perm.ExpiresAt, perm.IsActive, perm.Metadata,
	).Scan(&perm.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to assign user permission: %w", err)
	}

	return nil
}

// GetUserPermissions retrieves all permissions for a user
func (r *Repository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*UserPermission, error) {
	query := `
		SELECT id, user_id, policy_id, granted_by, granted_at, expires_at, is_active, metadata, created_at
		FROM access.user_permissions
		WHERE user_id = $1 AND is_active = TRUE
		  AND (expires_at IS NULL OR expires_at > NOW())
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*UserPermission
	for rows.Next() {
		perm := &UserPermission{}
		err := rows.Scan(
			&perm.ID, &perm.UserID, &perm.PolicyID, &perm.GrantedBy,
			&perm.GrantedAt, &perm.ExpiresAt, &perm.IsActive,
			&perm.Metadata, &perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// AssignRolePermission assigns a policy to a role
func (r *Repository) AssignRolePermission(ctx context.Context, perm *RolePermission) error {
	query := `
		INSERT INTO access.role_permissions
		(id, role, policy_id, granted_by, granted_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		perm.ID, perm.Role, perm.PolicyID, perm.GrantedBy,
		perm.GrantedAt, perm.IsActive,
	).Scan(&perm.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to assign role permission: %w", err)
	}

	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *Repository) GetRolePermissions(ctx context.Context, role string) ([]*RolePermission, error) {
	query := `
		SELECT id, role, policy_id, granted_by, granted_at, is_active, created_at
		FROM access.role_permissions
		WHERE role = $1 AND is_active = TRUE
	`

	rows, err := r.db.Query(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*RolePermission
	for rows.Next() {
		perm := &RolePermission{}
		err := rows.Scan(
			&perm.ID, &perm.Role, &perm.PolicyID, &perm.GrantedBy,
			&perm.GrantedAt, &perm.IsActive, &perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// RevokeUserPermission revokes a user permission
func (r *Repository) RevokeUserPermission(ctx context.Context, userID, policyID uuid.UUID) error {
	query := `
		UPDATE access.user_permissions
		SET is_active = FALSE
		WHERE user_id = $1 AND policy_id = $2
	`

	_, err := r.db.Exec(ctx, query, userID, policyID)
	if err != nil {
		return fmt.Errorf("failed to revoke user permission: %w", err)
	}

	return nil
}

// RevokeRolePermission revokes a role permission
func (r *Repository) RevokeRolePermission(ctx context.Context, role string, policyID uuid.UUID) error {
	query := `
		UPDATE access.role_permissions
		SET is_active = FALSE
		WHERE role = $1 AND policy_id = $2
	`

	_, err := r.db.Exec(ctx, query, role, policyID)
	if err != nil {
		return fmt.Errorf("failed to revoke role permission: %w", err)
	}

	return nil
}

