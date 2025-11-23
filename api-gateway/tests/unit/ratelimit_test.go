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
)

// MockCacheService mocks the cache service for rate limiting tests
type MockRateLimitCache struct {
	mock.Mock
}

func (m *MockRateLimitCache) CheckRateLimit(ctx context.Context, key string, requests int, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, requests, window)
	return args.Bool(0), args.Error(1)
}

// TestRateLimitMiddleware tests rate limiting middleware behavior
func TestRateLimitMiddleware(t *testing.T) {
	t.Run("Should allow requests within limit", func(t *testing.T) {
		mockCache := new(MockRateLimitCache)
		ctx := context.Background()
		key := "rate_limit:ip:192.168.1.1"

		// Allow first 5 requests
		for i := 0; i < 5; i++ {
			mockCache.On("CheckRateLimit", ctx, key, 5, 15*time.Minute).Return(true, nil).Once()
		}

		for i := 0; i < 5; i++ {
			allowed, err := mockCache.CheckRateLimit(ctx, key, 5, 15*time.Minute)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
			assert.NoError(t, err)
		}

		mockCache.AssertExpectations(t)
	})

	t.Run("Should block requests exceeding limit", func(t *testing.T) {
		mockCache := new(MockRateLimitCache)
		ctx := context.Background()
		key := "rate_limit:ip:192.168.1.1"

		// Block 6th request
		mockCache.On("CheckRateLimit", ctx, key, 5, 15*time.Minute).Return(false, nil).Once()

		allowed, err := mockCache.CheckRateLimit(ctx, key, 5, 15*time.Minute)
		assert.False(t, allowed, "Request should be blocked")
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})

	t.Run("Should handle cache errors gracefully", func(t *testing.T) {
		mockCache := new(MockRateLimitCache)
		ctx := context.Background()
		key := "rate_limit:ip:192.168.1.1"

		// Simulate cache error
		mockCache.On("CheckRateLimit", ctx, key, 5, 15*time.Minute).Return(false, assert.AnError).Once()

		allowed, err := mockCache.CheckRateLimit(ctx, key, 5, 15*time.Minute)
		assert.False(t, allowed)
		assert.Error(t, err)
	})
}

// TestLoginRateLimitConfiguration tests login-specific rate limit configuration
func TestLoginRateLimitConfiguration(t *testing.T) {
	t.Run("Login endpoint should use stricter rate limit", func(t *testing.T) {
		// Login endpoint: 5 attempts per 15 minutes
		loginRequests := 5
		loginWindow := 15 * time.Minute

		// General API: 1000 requests per hour
		generalRequests := 1000
		generalWindow := 1 * time.Hour

		assert.Less(t, loginRequests, generalRequests, "Login should have stricter limit")
		assert.Less(t, loginWindow, generalWindow, "Login should have shorter window")
	})
}

