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
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/infrastructure/opa"
	"go.uber.org/zap"
)

// Service handles authorization operations
type Service struct {
	repo         *Repository
	opaClient    *opa.Client
	cache        PermissionCache
	logger       *zap.Logger
	opaEnabled   bool
}

// PermissionCache interface for caching permissions
type PermissionCache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

// NewService creates a new authorization service
func NewService(repo *Repository, opaClient *opa.Client, cache PermissionCache, logger *zap.Logger, opaEnabled bool) *Service {
	return &Service{
		repo:       repo,
		opaClient:  opaClient,
		cache:      cache,
		logger:     logger,
		opaEnabled: opaEnabled,
	}
}

// AuthorizeRequest represents an authorization request
type AuthorizeRequest struct {
	UserID      uuid.UUID
	Resource    string
	Action      string
	Scope       string
	ResourceID  *string
	IPAddress   string
	ResourceAttrs map[string]interface{}
}

// Authorize checks if a user is authorized to perform an action
func (s *Service) Authorize(ctx context.Context, req *AuthorizeRequest) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("authz:%s:%s:%s:%s", req.UserID.String(), req.Resource, req.Action, req.Scope)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		s.logger.Debug("Authorization cache hit", zap.String("key", cacheKey))
		return cached.(bool), nil
	}

	// Get user information
	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user roles
	roles, err := s.repo.GetUserRoles(ctx, req.UserID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Collect all permissions from roles
	var allPermissions []Permission
	for _, role := range roles {
		perms, err := s.repo.GetRolePermissions(ctx, role.ID)
		if err != nil {
			s.logger.Warn("Failed to get role permissions", zap.String("role", role.Name), zap.Error(err))
			continue
		}
		allPermissions = append(allPermissions, perms...)
	}

	// Get direct user permissions (override roles)
	directPerms, err := s.repo.GetUserDirectPermissions(ctx, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get direct permissions", zap.Error(err))
	} else {
		// Direct permissions override role permissions
		allPermissions = append(allPermissions, directPerms...)
	}

	// Convert to OPA format
	opaPermissions := make([]opa.PermissionInfo, 0, len(allPermissions))
	for _, perm := range allPermissions {
		opaPermissions = append(opaPermissions, opa.PermissionInfo{
			Resource:   perm.ResourceType,
			Action:     perm.Action,
			Effect:     perm.Effect,
			Scope:      perm.Scope,
			Conditions: perm.Conditions,
		})
	}

	// Extract role names
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Build OPA request
	opaReq := &opa.EvaluateRequest{
		User: opa.UserInfo{
			ID:          req.UserID.String(),
			Email:       user.Email,
			Roles:       roleNames,
			Permissions: opaPermissions,
		},
		Request: opa.RequestInfo{
			Resource: req.Resource,
			Action:   req.Action,
			Scope:    req.Scope,
		},
		Resource: opa.ResourceInfo{
			ID:         getStringPtr(req.ResourceID),
			Type:       req.Resource,
			Attributes: req.ResourceAttrs,
		},
		Environment: opa.EnvironmentInfo{
			IPAddress:  req.IPAddress,
			Timestamp:  time.Now().Format(time.RFC3339),
			Attributes: make(map[string]interface{}),
		},
	}

	// Evaluate with OPA if enabled
	if s.opaEnabled && s.opaClient != nil {
		allowed, err := s.opaClient.Evaluate(ctx, opaReq)
		if err != nil {
			s.logger.Warn("OPA evaluation failed, falling back to basic check", zap.Error(err))
			// Fallback to basic permission check
			return s.basicPermissionCheck(allPermissions, req), nil
		}

		// Cache result (short TTL for security)
		s.cache.Set(ctx, cacheKey, allowed, 30*time.Second)
		return allowed, nil
	}

	// Fallback to basic permission check if OPA is disabled
	allowed := s.basicPermissionCheck(allPermissions, req)
	s.cache.Set(ctx, cacheKey, allowed, 30*time.Second)
	return allowed, nil
}

// basicPermissionCheck performs a basic permission check without OPA
func (s *Service) basicPermissionCheck(permissions []Permission, req *AuthorizeRequest) bool {
	for _, perm := range permissions {
		// Check resource match
		if !resourceMatches(perm.ResourceType, req.Resource) {
			continue
		}

		// Check action match
		if perm.Action != "*" && perm.Action != req.Action {
			continue
		}

		// Check scope match
		if perm.Scope != "" && perm.Scope != req.Scope {
			continue
		}

		// Check effect
		if perm.Effect == "deny" {
			return false
		}
		if perm.Effect == "allow" {
			return true
		}
	}

	return false
}

// resourceMatches checks if a resource matches a pattern
func resourceMatches(pattern, resource string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == resource {
		return true
	}
	// Simple prefix matching (can be enhanced)
	return len(pattern) > 0 && len(resource) >= len(pattern) && resource[:len(pattern)] == pattern
}

// getStringPtr returns a pointer to a string
func getStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

