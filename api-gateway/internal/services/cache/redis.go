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

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Service represents a Redis cache service
type Service struct {
	client *redis.Client
	logger *zap.Logger
}

// NewService creates a new Redis cache service
func NewService(cfg *config.RedisConfig, logger *zap.Logger) (*Service, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis", zap.String("address", cfg.Address()))

	return &Service{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from cache
func (s *Service) Get(ctx context.Context, key string) (string, error) {
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		s.logger.Error("Failed to get from cache", zap.String("key", key), zap.Error(err))
		return "", err
	}
	return val, nil
}

// Set stores a value in cache with expiration
func (s *Service) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := s.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		s.logger.Error("Failed to set cache", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// GetJSON retrieves and unmarshals JSON from cache
func (s *Service) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}

// SetJSON marshals and stores JSON in cache
func (s *Service) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return s.Set(ctx, key, data, expiration)
}

// Delete removes a key from cache
func (s *Service) Delete(ctx context.Context, keys ...string) error {
	err := s.client.Del(ctx, keys...).Err()
	if err != nil {
		s.logger.Error("Failed to delete from cache", zap.Strings("keys", keys), zap.Error(err))
		return err
	}
	return nil
}

// Exists checks if a key exists in cache
func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		s.logger.Error("Failed to check existence", zap.String("key", key), zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// Increment increments a counter
func (s *Service) Increment(ctx context.Context, key string) (int64, error) {
	val, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		s.logger.Error("Failed to increment", zap.String("key", key), zap.Error(err))
		return 0, err
	}
	return val, nil
}

// Decrement decrements a counter (never returns negative)
func (s *Service) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := s.client.Decr(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		s.logger.Error("Failed to decrement", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	// If the counter drops below zero, clean up the key
	if val <= 0 {
		if delErr := s.client.Del(ctx, key).Err(); delErr != nil {
			s.logger.Warn("Failed to delete key after decrement", zap.String("key", key), zap.Error(delErr))
		}
		return 0, nil
	}

	return val, nil
}

// Expire sets expiration on a key
func (s *Service) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := s.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		s.logger.Error("Failed to set expiration", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// TTL gets the time to live for a key
func (s *Service) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		s.logger.Error("Failed to get TTL", zap.String("key", key), zap.Error(err))
		return 0, err
	}
	return ttl, nil
}

// CheckRateLimit checks if rate limit is exceeded
func (s *Service) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := s.Increment(ctx, key)
	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := s.Expire(ctx, key, window); err != nil {
			return false, err
		}
	}

	return count <= int64(limit), nil
}

// Close closes the Redis connection
func (s *Service) Close() error {
	if s.client == nil {
		return nil
	}
	if err := s.client.Close(); err != nil {
		// Ignore "client is closed" error as it means already closed
		if err.Error() != "redis: client is closed" {
			s.logger.Error("Failed to close Redis connection", zap.Error(err))
			return err
		}
		return nil
	}
	s.logger.Info("Redis connection closed")
	s.client = nil
	return nil
}

// Health checks Redis health
func (s *Service) Health(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
