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

package authorization

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles authorization data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new authorization repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Role represents a role
type Role struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	ParentRoleID *uuid.UUID
	Level        int
	IsSystemRole bool
}

// Permission represents a permission
type Permission struct {
	ID           uuid.UUID
	ResourceType string
	ResourceID   *string
	Action       string
	Scope        string
	Conditions   map[string]interface{}
	Effect       string
	Priority     int
}

// GetUserRoles retrieves all active roles for a user
func (r *Repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.parent_role_id, r.level, r.is_system_role
		FROM auth.roles r
		INNER JOIN auth.user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		  AND ur.is_active = TRUE
		  AND (ur.valid_until IS NULL OR ur.valid_until > NOW())
		  AND r.deleted_at IS NULL
		ORDER BY r.level, r.name
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.ParentRoleID,
			&role.Level,
			&role.IsSystemRole,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *Repository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error) {
	query := `
		SELECT p.id, p.resource_type, p.resource_id, p.action, p.scope, 
		       p.conditions, p.effect, p.priority, rp.effect as role_effect
		FROM auth.permissions p
		INNER JOIN auth.role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		  AND p.deleted_at IS NULL
		ORDER BY p.priority DESC, p.resource_type, p.action
	`

	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		var conditionsJSON []byte
		var roleEffect sql.NullString

		err := rows.Scan(
			&perm.ID,
			&perm.ResourceType,
			&perm.ResourceID,
			&perm.Action,
			&perm.Scope,
			&conditionsJSON,
			&perm.Effect,
			&perm.Priority,
			&roleEffect,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		// Parse conditions JSON
		if len(conditionsJSON) > 0 {
			if err := json.Unmarshal(conditionsJSON, &perm.Conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		}

		// Use role effect if provided (override permission effect)
		if roleEffect.Valid && roleEffect.String != "" {
			perm.Effect = roleEffect.String
		}

		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// GetUserDirectPermissions retrieves direct permissions for a user (override roles)
func (r *Repository) GetUserDirectPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error) {
	query := `
		SELECT p.id, p.resource_type, p.resource_id, p.action, p.scope,
		       p.conditions, p.effect, p.priority, up.effect as user_effect
		FROM auth.permissions p
		INNER JOIN auth.user_permissions up ON p.id = up.permission_id
		WHERE up.user_id = $1
		  AND up.is_active = TRUE
		  AND (up.valid_until IS NULL OR up.valid_until > NOW())
		  AND p.deleted_at IS NULL
		ORDER BY p.priority DESC, p.resource_type, p.action
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		var conditionsJSON []byte
		var userEffect sql.NullString

		err := rows.Scan(
			&perm.ID,
			&perm.ResourceType,
			&perm.ResourceID,
			&perm.Action,
			&perm.Scope,
			&conditionsJSON,
			&perm.Effect,
			&perm.Priority,
			&userEffect,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		// Parse conditions JSON
		if len(conditionsJSON) > 0 {
			if err := json.Unmarshal(conditionsJSON, &perm.Conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		}

		// Use user effect if provided (override permission effect)
		if userEffect.Valid && userEffect.String != "" {
			perm.Effect = userEffect.String
		}

		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// GetUserByID retrieves user information
func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	query := `
		SELECT id, email, role, msp_id, created_at
		FROM auth.users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user UserInfo
	var createdAt time.Time
	var role sql.NullString
	var mspID sql.NullString

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&role,
		&mspID,
		&createdAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if role.Valid {
		user.Role = role.String
	}
	if mspID.Valid {
		user.MSPID = &mspID.String
	}

	return &user, nil
}

// UserInfo represents user information
type UserInfo struct {
	ID      uuid.UUID
	Email   string
	Role    string
	MSPID   *string
	Created time.Time
}

