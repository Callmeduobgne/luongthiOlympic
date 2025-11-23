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

	"github.com/ibn-network/backend/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCache represents L2 Redis cache
type RedisCache struct {
	client  *redis.Client
	logger  *zap.Logger
	metrics *CacheMetrics
}

// NewRedisCache creates a new L2 Redis cache
func NewRedisCache(cfg *config.RedisConfig, logger *zap.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("L2 Redis cache connected", zap.String("address", cfg.Address()))

	return &RedisCache{
		client:  client,
		logger:  logger,
		metrics: &CacheMetrics{},
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		r.metrics.recordMiss()
		return "", nil // Key not found, not an error
	}
	if err != nil {
		r.logger.Error("Failed to get from Redis", zap.String("key", key), zap.Error(err))
		return "", err
	}
	
	r.metrics.recordHit()
	return val, nil
}

// Set stores a value in Redis with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		r.logger.Error("Failed to set in Redis", zap.String("key", key), zap.Error(err))
		return err
	}
	
	r.metrics.recordSet()
	return nil
}

// GetJSON retrieves and unmarshals JSON from Redis
func (r *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return fmt.Errorf("key not found: %s", key)
	}
	
	return json.Unmarshal([]byte(val), dest)
}

// SetJSON marshals and stores JSON in Redis
func (r *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return r.Set(ctx, key, data, expiration)
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		r.logger.Error("Failed to delete from Redis", zap.Strings("keys", keys), zap.Error(err))
		return err
	}
	
	r.metrics.recordDelete()
	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to check existence in Redis", zap.String("key", key), zap.Error(err))
		return false, err
	}
	
	return count > 0, nil
}

// Increment increments a counter in Redis
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to increment in Redis", zap.String("key", key), zap.Error(err))
		return 0, err
	}
	
	return val, nil
}

// Expire sets expiration on a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		r.logger.Error("Failed to set expiration in Redis", zap.String("key", key), zap.Error(err))
		return err
	}
	
	return nil
}

// TTL gets the time to live for a key
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to get TTL from Redis", zap.String("key", key), zap.Error(err))
		return 0, err
	}
	
	return ttl, nil
}

// CheckRateLimit checks if rate limit is exceeded
func (r *RedisCache) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := r.Increment(ctx, key)
	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := r.Expire(ctx, key, window); err != nil {
			return false, err
		}
	}

	return count <= int64(limit), nil
}

// GetMetrics returns cache metrics
func (r *RedisCache) GetMetrics() CacheMetricsSnapshot {
	return r.metrics.snapshot()
}

// Stats returns Redis statistics
func (r *RedisCache) Stats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}
	
	metrics := r.metrics.snapshot()
	
	return map[string]interface{}{
		"redis_info": info,
		"hits":       metrics.Hits,
		"misses":     metrics.Misses,
		"hit_rate":   metrics.HitRate(),
		"sets":       metrics.Sets,
		"deletes":    metrics.Deletes,
	}, nil
}

// Health checks Redis health
func (r *RedisCache) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	if r.client == nil {
		return nil
	}
	
	if err := r.client.Close(); err != nil {
		r.logger.Error("Failed to close Redis connection", zap.Error(err))
		return err
	}
	
	r.logger.Info("Redis connection closed")
	return nil
}

