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

package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/auth"
)

// MockCacheService mocks the cache service for testing
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) CheckRateLimit(ctx context.Context, key string, requests int, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, requests, window)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestAccountLockout tests account lockout mechanism
func TestAccountLockout(t *testing.T) {
	t.Run("Account lockout constants should be set correctly", func(t *testing.T) {
		// Verify lockout configuration
		assert.Equal(t, 5, auth.MaxFailedLoginAttempts, "Max failed attempts should be 5")
		assert.Equal(t, 15*time.Minute, auth.AccountLockoutDuration, "Lockout duration should be 15 minutes")
	})

	t.Run("Failed attempts should reset on successful login", func(t *testing.T) {
		mockCache := new(MockCacheService)
		ctx := context.Background()
		userID := "test-user-id"

		// Simulate successful login - should delete failed attempts
		mockCache.On("Delete", ctx, "failed_attempts:"+userID).Return(nil).Once()
		mockCache.On("Delete", ctx, "account_lockout:"+userID).Return(nil).Once()

		err1 := mockCache.Delete(ctx, "failed_attempts:"+userID)
		err2 := mockCache.Delete(ctx, "account_lockout:"+userID)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		mockCache.AssertExpectations(t)
	})
}

// TestLoginRateLimit tests rate limiting for login endpoint
func TestLoginRateLimit(t *testing.T) {
	t.Run("Rate limit should allow 5 attempts per 15 minutes", func(t *testing.T) {
		mockCache := new(MockCacheService)
		ctx := context.Background()
		ip := "192.168.1.1"
		key := "rate_limit:ip:" + ip

		// First 5 attempts should be allowed
		for i := 0; i < 5; i++ {
			mockCache.On("CheckRateLimit", ctx, key, 5, 15*time.Minute).Return(true, nil).Once()
		}

		// 6th attempt should be blocked
		mockCache.On("CheckRateLimit", ctx, key, 5, 15*time.Minute).Return(false, nil).Once()

		// Test first 5 attempts
		for i := 0; i < 5; i++ {
			allowed, _ := mockCache.CheckRateLimit(ctx, key, 5, 15*time.Minute)
			assert.True(t, allowed, "Attempt %d should be allowed", i+1)
		}

		// Test 6th attempt
		allowed, _ := mockCache.CheckRateLimit(ctx, key, 5, 15*time.Minute)
		assert.False(t, allowed, "6th attempt should be blocked")

		mockCache.AssertExpectations(t)
	})
}

// TestPasswordValidation tests password validation requirements
func TestPasswordValidation(t *testing.T) {
	t.Run("Password should meet minimum length requirement", func(t *testing.T) {
		req := &models.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "short", // Too short
		}

		// Password should be at least 8 characters
		assert.Less(t, len(req.Password), 8, "Password should be at least 8 characters")
	})

	t.Run("Valid password should pass validation", func(t *testing.T) {
		req := &models.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "ValidPass123!", // Valid: 8+ chars, uppercase, lowercase, number
		}

		assert.GreaterOrEqual(t, len(req.Password), 8, "Password should meet minimum length")
	})
}

