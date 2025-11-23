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

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ibn-network/api-gateway/internal/services/cache"
	"go.uber.org/zap"
)

// WebSocketRateLimitMiddleware provides rate limiting for WebSocket connections
type WebSocketRateLimitMiddleware struct {
	redisService *cache.Service
	enabled      bool
	messages     int
	window       time.Duration
	logger       *zap.Logger
	// In-memory connection tracking (fallback if Redis unavailable)
	connections map[string]*connectionTracker
	mu          sync.RWMutex
}

type connectionTracker struct {
	count            int
	lastReset        time.Time
	messageCount     int
	lastMessageReset time.Time
}

// NewWebSocketRateLimitMiddleware creates a new WebSocket rate limit middleware
func NewWebSocketRateLimitMiddleware(
	redisService *cache.Service,
	enabled bool,
	messages int,
	window time.Duration,
	logger *zap.Logger,
) *WebSocketRateLimitMiddleware {
	return &WebSocketRateLimitMiddleware{
		redisService: redisService,
		enabled:      enabled,
		messages:     messages,
		window:       window,
		logger:       logger,
		connections:  make(map[string]*connectionTracker),
	}
}

// LimitConnection limits WebSocket connections per IP/user
func (m *WebSocketRateLimitMiddleware) LimitConnection(maxPerIP, maxPerUser int, userID string, r *http.Request) error {
	if !m.enabled {
		return nil
	}

	ctx := r.Context()
	// Get client IP (reuse from ratelimit.go)
	ip := getClientIP(r)
	keyIP := "ws:conn:ip:" + ip
	keyUser := "ws:conn:user:" + userID

	// Check IP limit
	if maxPerIP > 0 {
		count, err := m.getConnectionCount(ctx, keyIP)
		if err != nil {
			// Fallback to in-memory tracking
			m.mu.Lock()
			tracker := m.connections[keyIP]
			if tracker == nil {
				tracker = &connectionTracker{count: 0, lastReset: time.Now()}
				m.connections[keyIP] = tracker
			}
			if tracker.count >= maxPerIP {
				m.mu.Unlock()
				m.logger.Warn("WebSocket connection limit exceeded for IP",
					zap.String("ip", ip),
					zap.Int("count", tracker.count),
					zap.Int("max", maxPerIP),
				)
				return http.ErrAbortHandler
			}
			tracker.count++
			m.mu.Unlock()
		} else if count >= maxPerIP {
			m.logger.Warn("WebSocket connection limit exceeded for IP",
				zap.String("ip", ip),
				zap.Int("count", count),
				zap.Int("max", maxPerIP),
			)
			return http.ErrAbortHandler
		}
	}

	// Check user limit
	if maxPerUser > 0 && userID != "" {
		count, err := m.getConnectionCount(ctx, keyUser)
		if err != nil {
			// Fallback to in-memory tracking
			m.mu.Lock()
			tracker := m.connections[keyUser]
			if tracker == nil {
				tracker = &connectionTracker{count: 0, lastReset: time.Now()}
				m.connections[keyUser] = tracker
			}
			if tracker.count >= maxPerUser {
				m.mu.Unlock()
				m.logger.Warn("WebSocket connection limit exceeded for user",
					zap.String("user_id", userID),
					zap.Int("count", tracker.count),
					zap.Int("max", maxPerUser),
				)
				return http.ErrAbortHandler
			}
			tracker.count++
			m.mu.Unlock()
		} else if count >= maxPerUser {
			m.logger.Warn("WebSocket connection limit exceeded for user",
				zap.String("user_id", userID),
				zap.Int("count", count),
				zap.Int("max", maxPerUser),
			)
			return http.ErrAbortHandler
		}
	}

	return nil
}

// IncrementConnection increments connection count
func (m *WebSocketRateLimitMiddleware) IncrementConnection(ip, userID string) {
	if !m.enabled || m.redisService == nil {
		return
	}

	ctx := context.Background()
	keyIP := "ws:conn:ip:" + ip
	keyUser := "ws:conn:user:" + userID

	// Increment in Redis
	_ = m.incrementConnectionCount(ctx, keyIP)
	if userID != "" {
		_ = m.incrementConnectionCount(ctx, keyUser)
	}
}

// DecrementConnection decrements connection count
func (m *WebSocketRateLimitMiddleware) DecrementConnection(ip, userID string) {
	if !m.enabled {
		return
	}

	// Redis-based tracking
	if m.redisService != nil {
		ctx := context.Background()
		keyIP := "ws:conn:ip:" + ip
		keyUser := "ws:conn:user:" + userID

		_ = m.decrementConnectionCount(ctx, keyIP)
		if userID != "" {
			_ = m.decrementConnectionCount(ctx, keyUser)
		}
		return
	}

	// In-memory fallback tracking
	m.mu.Lock()
	defer m.mu.Unlock()

	keyIP := "ws:conn:ip:" + ip
	if tracker, ok := m.connections[keyIP]; ok {
		if tracker.count > 0 {
			tracker.count--
		}
		if tracker.count == 0 {
			delete(m.connections, keyIP)
		}
	}

	if userID != "" {
		keyUser := "ws:conn:user:" + userID
		if tracker, ok := m.connections[keyUser]; ok {
			if tracker.count > 0 {
				tracker.count--
			}
			if tracker.count == 0 {
				delete(m.connections, keyUser)
			}
		}
	}
}

// CheckMessageRate checks if message rate limit is exceeded
func (m *WebSocketRateLimitMiddleware) CheckMessageRate(ip, userID string) error {
	if !m.enabled {
		return nil
	}

	ctx := context.Background()
	key := "ws:msg:" + ip
	if userID != "" {
		key = "ws:msg:user:" + userID
	}

	count, err := m.getConnectionCount(ctx, key)
	if err != nil {
		// Fallback to in-memory tracking
		m.mu.Lock()
		tracker := m.connections[key]
		now := time.Now()
		if tracker == nil || now.Sub(tracker.lastMessageReset) > m.window {
			tracker = &connectionTracker{
				messageCount:     0,
				lastMessageReset: now,
			}
			m.connections[key] = tracker
		}
		if tracker.messageCount >= m.messages {
			m.mu.Unlock()
			return http.ErrAbortHandler
		}
		tracker.messageCount++
		m.mu.Unlock()
		return nil
	}

	if count >= m.messages {
		return http.ErrAbortHandler
	}

	_ = m.incrementConnectionCount(ctx, key)
	return nil
}

// Helper methods
func (m *WebSocketRateLimitMiddleware) getConnectionCount(ctx context.Context, key string) (int, error) {
	if m.redisService == nil {
		return 0, http.ErrAbortHandler
	}
	val, err := m.redisService.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if val == "" {
		return 0, nil
	}
	var count int
	_, err = fmt.Sscanf(val, "%d", &count)
	return count, err
}

func (m *WebSocketRateLimitMiddleware) incrementConnectionCount(ctx context.Context, key string) error {
	if m.redisService == nil {
		return nil
	}
	_, err := m.redisService.Increment(ctx, key)
	if err == nil {
		_ = m.redisService.Expire(ctx, key, m.window)
	}
	return err
}

func (m *WebSocketRateLimitMiddleware) decrementConnectionCount(ctx context.Context, key string) error {
	if m.redisService == nil {
		return nil
	}
	_, err := m.redisService.Decrement(ctx, key)
	return err
}
