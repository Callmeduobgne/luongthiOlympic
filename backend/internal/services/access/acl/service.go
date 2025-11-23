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
	"time"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/infrastructure/cache"
	"go.uber.org/zap"
)

// Service handles ACL business logic
type Service struct {
	repo   *Repository
	cache  *cache.MultiLayerCache
	logger *zap.Logger
}

// NewService creates a new ACL service
func NewService(
	repo *Repository,
	cache *cache.MultiLayerCache,
	logger *zap.Logger,
) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// CreatePolicy creates a new ACL policy
func (s *Service) CreatePolicy(ctx context.Context, req *CreatePolicyRequest, createdBy uuid.UUID) (*Policy, error) {
	// Check if policy with same name exists
	existing, _ := s.repo.GetPolicyByName(ctx, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("policy with name %s already exists", req.Name)
	}

	// Validate effect
	if req.Effect != EffectAllow && req.Effect != EffectDeny {
		return nil, fmt.Errorf("invalid effect: %s", req.Effect)
	}

	policy := &Policy{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Actions:      req.Actions,
		Effect:       req.Effect,
		Conditions:   req.Conditions,
		Priority:     req.Priority,
		IsActive:     true,
		CreatedBy:    &createdBy,
	}

	if err := s.repo.CreatePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	s.logger.Info("Policy created",
		zap.String("policy_id", policy.ID.String()),
		zap.String("name", policy.Name),
		zap.String("created_by", createdBy.String()),
	)

	// Invalidate cache
	s.invalidatePolicyCache(ctx)

	return policy, nil
}

// GetPolicy retrieves a policy by ID (with caching)
func (s *Service) GetPolicy(ctx context.Context, id uuid.UUID) (*Policy, error) {
	cacheKey := fmt.Sprintf("policy:%s", id.String())

	var policy Policy
	err := s.cache.Get(ctx, cacheKey, &policy, func(ctx context.Context) (interface{}, error) {
		return s.repo.GetPolicyByID(ctx, id)
	}, &cache.CacheTTLs{
		L1TTL: 10 * time.Minute,
		L2TTL: 1 * time.Hour,
	})

	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// ListPolicies lists all policies (with caching)
func (s *Service) ListPolicies(ctx context.Context, resourceType *string) ([]*Policy, error) {
	cacheKey := "policies:all"
	if resourceType != nil {
		cacheKey = fmt.Sprintf("policies:type:%s", *resourceType)
	}

	var policies []*Policy
	err := s.cache.Get(ctx, cacheKey, &policies, func(ctx context.Context) (interface{}, error) {
		return s.repo.ListPolicies(ctx, resourceType)
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 30 * time.Minute,
	})

	if err != nil {
		return nil, err
	}

	return policies, nil
}

// UpdatePolicy updates a policy
func (s *Service) UpdatePolicy(ctx context.Context, id uuid.UUID, req *UpdatePolicyRequest) (*Policy, error) {
	// Check if policy exists
	policy, err := s.repo.GetPolicyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("Existing policy loaded for update",
		zap.String("policy_id", policy.ID.String()),
		zap.String("policy_name", policy.Name),
	)

	// Build updates map
	updates := make(map[string]interface{})
	
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Actions != nil {
		updates["actions"] = req.Actions
	}
	if req.Effect != nil {
		if *req.Effect != EffectAllow && *req.Effect != EffectDeny {
			return nil, fmt.Errorf("invalid effect: %s", *req.Effect)
		}
		updates["effect"] = *req.Effect
	}
	if req.Conditions != nil {
		updates["conditions"] = req.Conditions
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := s.repo.UpdatePolicy(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	s.logger.Info("Policy updated",
		zap.String("policy_id", id.String()),
		zap.Int("updates", len(updates)),
	)

	// Invalidate cache
	s.invalidatePolicyCache(ctx)
	s.cache.Delete(ctx, fmt.Sprintf("policy:%s", id.String()))

	// Get updated policy
	return s.repo.GetPolicyByID(ctx, id)
}

// DeletePolicy deletes a policy
func (s *Service) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeletePolicy(ctx, id); err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	s.logger.Info("Policy deleted", zap.String("policy_id", id.String()))

	// Invalidate cache
	s.invalidatePolicyCache(ctx)
	s.cache.Delete(ctx, fmt.Sprintf("policy:%s", id.String()))

	return nil
}

// AssignUserPermission assigns a policy to a user
func (s *Service) AssignUserPermission(ctx context.Context, req *AssignUserPermissionRequest, grantedBy uuid.UUID) error {
	// Verify policy exists
	if _, err := s.repo.GetPolicyByID(ctx, req.PolicyID); err != nil {
		return fmt.Errorf("policy not found")
	}

	perm := &UserPermission{
		ID:        uuid.New(),
		UserID:    req.UserID,
		PolicyID:  req.PolicyID,
		GrantedBy: &grantedBy,
		GrantedAt: time.Now(),
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
	}

	if err := s.repo.AssignUserPermission(ctx, perm); err != nil {
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	s.logger.Info("User permission assigned",
		zap.String("user_id", req.UserID.String()),
		zap.String("policy_id", req.PolicyID.String()),
		zap.String("granted_by", grantedBy.String()),
	)

	// Invalidate user permissions cache
	s.cache.Delete(ctx, fmt.Sprintf("user_permissions:%s", req.UserID.String()))

	return nil
}

// AssignRolePermission assigns a policy to a role
func (s *Service) AssignRolePermission(ctx context.Context, req *AssignRolePermissionRequest, grantedBy uuid.UUID) error {
	// Verify policy exists
	if _, err := s.repo.GetPolicyByID(ctx, req.PolicyID); err != nil {
		return fmt.Errorf("policy not found")
	}

	perm := &RolePermission{
		ID:        uuid.New(),
		Role:      req.Role,
		PolicyID:  req.PolicyID,
		GrantedBy: &grantedBy,
		GrantedAt: time.Now(),
		IsActive:  true,
	}

	if err := s.repo.AssignRolePermission(ctx, perm); err != nil {
		return fmt.Errorf("failed to assign role permission: %w", err)
	}

	s.logger.Info("Role permission assigned",
		zap.String("role", req.Role),
		zap.String("policy_id", req.PolicyID.String()),
		zap.String("granted_by", grantedBy.String()),
	)

	// Invalidate role permissions cache
	s.cache.Delete(ctx, fmt.Sprintf("role_permissions:%s", req.Role))

	return nil
}

// CheckPermission checks if a user has permission for a resource-action
func (s *Service) CheckPermission(ctx context.Context, userID uuid.UUID, userRole string, resource string, action string) (*CheckPermissionResponse, error) {
	// Get user permissions from cache
	cacheKey := fmt.Sprintf("user_permissions:%s", userID.String())
	
	var userPerms []*UserPermission
	err := s.cache.Get(ctx, cacheKey, &userPerms, func(ctx context.Context) (interface{}, error) {
		return s.repo.GetUserPermissions(ctx, userID)
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 30 * time.Minute,
	})

	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.Error(err))
		// Continue with role permissions only
	}

	// Get role permissions from cache
	roleCacheKey := fmt.Sprintf("role_permissions:%s", userRole)
	
	var rolePerms []*RolePermission
	err = s.cache.Get(ctx, roleCacheKey, &rolePerms, func(ctx context.Context) (interface{}, error) {
		return s.repo.GetRolePermissions(ctx, userRole)
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 30 * time.Minute,
	})

	if err != nil {
		s.logger.Error("Failed to get role permissions", zap.Error(err))
	}

	// Collect all policy IDs
	policyIDs := make(map[uuid.UUID]bool)
	for _, perm := range userPerms {
		policyIDs[perm.PolicyID] = true
	}
	for _, perm := range rolePerms {
		policyIDs[perm.PolicyID] = true
	}

	// Check each policy
	var matchedPolicies []uuid.UUID
	allowed := false
	explicitDeny := false

	for policyID := range policyIDs {
		policy, err := s.GetPolicy(ctx, policyID)
		if err != nil {
			s.logger.Warn("Failed to get policy", zap.String("policy_id", policyID.String()), zap.Error(err))
			continue
		}

		if !policy.IsActive {
			continue
		}

		// Check if policy matches resource and action
		if s.matchesPolicy(policy, resource, action) {
			matchedPolicies = append(matchedPolicies, policy.ID)

			if policy.Effect == EffectDeny {
				explicitDeny = true
			} else if policy.Effect == EffectAllow {
				allowed = true
			}
		}
	}

	// Deny takes precedence over allow
	if explicitDeny {
		return &CheckPermissionResponse{
			Allowed:         false,
			MatchedPolicies: matchedPolicies,
			Reason:          "Explicitly denied by policy",
		}, nil
	}

	if allowed {
		return &CheckPermissionResponse{
			Allowed:         true,
			MatchedPolicies: matchedPolicies,
			Reason:          "Allowed by policy",
		}, nil
	}

	return &CheckPermissionResponse{
		Allowed: false,
		Reason:  "No matching policy found",
	}, nil
}

// matchesPolicy checks if a policy matches the resource and action
func (s *Service) matchesPolicy(policy *Policy, resource string, action string) bool {
	// Check if action is in policy actions
	actionMatches := false
	for _, policyAction := range policy.Actions {
		if policyAction == action || policyAction == "*" {
			actionMatches = true
			break
		}
	}

	if !actionMatches {
		return false
	}

	// Check if resource matches
	// Simple matching - can be extended with wildcards
	if policy.ResourceID == nil || *policy.ResourceID == "*" || *policy.ResourceID == resource {
		return true
	}

	return false
}

// RevokeUserPermission revokes a user permission
func (s *Service) RevokeUserPermission(ctx context.Context, userID, policyID uuid.UUID) error {
	if err := s.repo.RevokeUserPermission(ctx, userID, policyID); err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	s.logger.Info("User permission revoked",
		zap.String("user_id", userID.String()),
		zap.String("policy_id", policyID.String()),
	)

	// Invalidate cache
	s.cache.Delete(ctx, fmt.Sprintf("user_permissions:%s", userID.String()))

	return nil
}

// RevokeRolePermission revokes a role permission
func (s *Service) RevokeRolePermission(ctx context.Context, role string, policyID uuid.UUID) error {
	if err := s.repo.RevokeRolePermission(ctx, role, policyID); err != nil {
		return fmt.Errorf("failed to revoke role permission: %w", err)
	}

	s.logger.Info("Role permission revoked",
		zap.String("role", role),
		zap.String("policy_id", policyID.String()),
	)

	// Invalidate cache
	s.cache.Delete(ctx, fmt.Sprintf("role_permissions:%s", role))

	return nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *Service) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*UserPermission, error) {
	return s.repo.GetUserPermissions(ctx, userID)
}

// GetRolePermissions retrieves all permissions for a role
func (s *Service) GetRolePermissions(ctx context.Context, role string) ([]*RolePermission, error) {
	return s.repo.GetRolePermissions(ctx, role)
}

// invalidatePolicyCache invalidates all policy-related caches
func (s *Service) invalidatePolicyCache(ctx context.Context) {
	s.cache.Delete(ctx, "policies:all")
	// Also invalidate type-specific caches
	// In production, you might want to track these more systematically
}

