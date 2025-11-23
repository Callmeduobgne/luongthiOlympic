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

package admin

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements circuit breaker pattern for Admin Service calls
type CircuitBreaker struct {
	maxRequests    uint32
	interval       time.Duration
	timeout        time.Duration
	failureRatio   float64

	mu                sync.RWMutex
	state             CircuitBreakerState
	failures          uint32
	successes         uint32
	lastFailureTime   time.Time
	nextAttemptTime   time.Time
	requestCount      uint32
	lastResetTime     time.Time

	logger *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxRequests uint32, interval, timeout time.Duration, failureRatio float64, logger *zap.Logger) *CircuitBreaker {
	cb := &CircuitBreaker{
		maxRequests:  maxRequests,
		interval:     interval,
		timeout:      timeout,
		failureRatio: failureRatio,
		state:        StateClosed,
		logger:       logger,
	}
	cb.lastResetTime = time.Now()
	return cb
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should allow the request
	if !cb.shouldAllowRequest() {
		cb.logger.Warn("Circuit breaker is open, rejecting request",
			zap.String("state", cb.stateString()),
			zap.Time("next_attempt", cb.nextAttemptTime),
		)
		return errors.New("circuit breaker is open: admin service is unavailable")
	}

	// Execute the function
	err := fn()

	// Update circuit breaker state based on result
	cb.recordResult(err)

	return err
}

// shouldAllowRequest checks if a request should be allowed
func (cb *CircuitBreaker) shouldAllowRequest() bool {
	now := time.Now()

	switch cb.state {
	case StateClosed:
		// Reset counters if interval has passed
		if now.Sub(cb.lastResetTime) >= cb.interval {
			cb.resetCounters()
		}
		return true

	case StateOpen:
		// Check if timeout has passed, transition to half-open
		if now.After(cb.nextAttemptTime) {
			cb.state = StateHalfOpen
			cb.successes = 0
			cb.failures = 0
			cb.logger.Info("Circuit breaker transitioning to half-open state")
			return true
		}
		return false

	case StateHalfOpen:
		// Allow request to test if service recovered
		return true

	default:
		return false
	}
}

// recordResult records the result of a request
func (cb *CircuitBreaker) recordResult(err error) {
	now := time.Now()

	switch cb.state {
	case StateClosed:
		cb.requestCount++
		if err != nil {
			cb.failures++
			cb.lastFailureTime = now
		} else {
			cb.successes++
		}

		// Check if we should open the circuit
		if cb.requestCount >= cb.maxRequests {
			failureRate := float64(cb.failures) / float64(cb.requestCount)
			if failureRate >= cb.failureRatio {
				cb.state = StateOpen
				cb.nextAttemptTime = now.Add(cb.timeout)
				cb.logger.Warn("Circuit breaker opened",
					zap.Float64("failure_rate", failureRate),
					zap.Uint32("failures", cb.failures),
					zap.Uint32("requests", cb.requestCount),
				)
			}
			cb.resetCounters()
		}

	case StateHalfOpen:
		if err != nil {
			// Still failing, go back to open
			cb.state = StateOpen
			cb.nextAttemptTime = now.Add(cb.timeout)
			cb.logger.Warn("Circuit breaker re-opened after half-open test failed")
		} else {
			// Success, close the circuit
			cb.state = StateClosed
			cb.resetCounters()
			cb.logger.Info("Circuit breaker closed after successful half-open test")
		}
	}
}

// resetCounters resets the counters
func (cb *CircuitBreaker) resetCounters() {
	cb.failures = 0
	cb.successes = 0
	cb.requestCount = 0
	cb.lastResetTime = time.Now()
}

// stateString returns a string representation of the state
func (cb *CircuitBreaker) stateString() string {
	switch cb.state {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// GetState returns the current state (for monitoring)
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	stats := map[string]interface{}{
		"state":        cb.stateString(),
		"failures":     cb.failures,
		"successes":    cb.successes,
		"request_count": cb.requestCount,
	}

	if cb.state == StateOpen {
		stats["next_attempt"] = cb.nextAttemptTime
	}

	return stats
}

