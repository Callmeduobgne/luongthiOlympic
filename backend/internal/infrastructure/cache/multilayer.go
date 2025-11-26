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

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// MultiLayerCache implements L1 → L2 → L3 cache lookup
type MultiLayerCache struct {
	l1Cache *MemoryCache
	l2Cache *RedisCache
	db      *pgxpool.Pool
	logger  *zap.Logger
}

// MultiLayerCacheConfig holds configuration for multi-layer cache
type MultiLayerCacheConfig struct {
	L1TTL time.Duration // Default: 5-15 minutes
	L2TTL time.Duration // Default: 30 minutes - 1 hour
}

// NewMultiLayerCache creates a new multi-layer cache service
func NewMultiLayerCache(
	l1Cache *MemoryCache,
	l2Cache *RedisCache,
	db *pgxpool.Pool,
	logger *zap.Logger,
) *MultiLayerCache {
	logger.Info("Multi-layer cache initialized")
	
	return &MultiLayerCache{
		l1Cache: l1Cache,
		l2Cache: l2Cache,
		db:      db,
		logger:  logger,
	}
}

// Get implements cache-aside pattern with L1 → L2 → L3 lookup
// Usage: cache.Get(ctx, "user:123", &user, func(ctx context.Context) (interface{}, error) {
//     return db.QueryUser(ctx, "123")
// })
func (m *MultiLayerCache) Get(
	ctx context.Context,
	key string,
	dest interface{},
	dbFetcher func(context.Context) (interface{}, error),
	ttls *CacheTTLs,
) error {
	if ttls == nil {
		ttls = &CacheTTLs{
			L1TTL: 5 * time.Minute,
			L2TTL: 30 * time.Minute,
		}
	}

	// 1. Check L1 cache (in-memory)
	if data, found := m.l1Cache.Get(key); found {
		m.logger.Debug("L1 cache hit", zap.String("key", key))
		err := m.unmarshalData(data, dest)
		if err != nil {
			// If unmarshal fails, clear corrupted cache and fall through to database
			m.logger.Warn("Failed to unmarshal L1 cache data, clearing cache", zap.String("key", key), zap.Error(err))
			m.l1Cache.Delete(key)
			// Fall through to L2 cache or database query
		} else {
			return nil
		}
	}

	// 2. Check L2 cache (Redis)
	redisData, err := m.l2Cache.Get(ctx, key)
	if err == nil && redisData != "" {
		m.logger.Debug("L2 cache hit", zap.String("key", key))
		
		// Unmarshal Redis data first to get the actual object
		var tempData interface{}
		if err := json.Unmarshal([]byte(redisData), &tempData); err != nil {
			m.logger.Warn("Failed to unmarshal Redis data, clearing cache", zap.String("key", key), zap.Error(err))
			// Clear corrupted cache
			m.l2Cache.Delete(ctx, key)
			// Fall through to database query
		} else {
			// Populate L1 cache with unmarshaled data (not JSON string)
			if err := m.l1Cache.SetJSON(key, tempData, ttls.L1TTL); err != nil {
				m.logger.Warn("Failed to populate L1 cache", zap.Error(err))
			}
			
			// Unmarshal to destination
			return json.Unmarshal([]byte(redisData), dest)
		}
	}

	// 3. Query database (L3)
	m.logger.Debug("Cache miss, querying database", zap.String("key", key))
	
	data, err := dbFetcher(ctx)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	// Populate caches (async to not block)
	go m.populateCaches(context.Background(), key, data, ttls)

	// Marshal data to dest
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	
	return json.Unmarshal(dataBytes, dest)
}

// Set stores data in all cache layers
func (m *MultiLayerCache) Set(ctx context.Context, key string, value interface{}, ttls *CacheTTLs) error {
	if ttls == nil {
		ttls = &CacheTTLs{
			L1TTL: 5 * time.Minute,
			L2TTL: 30 * time.Minute,
		}
	}

	// Set in L1
	if err := m.l1Cache.SetJSON(key, value, ttls.L1TTL); err != nil {
		m.logger.Warn("Failed to set L1 cache", zap.Error(err))
	}

	// Set in L2
	if err := m.l2Cache.SetJSON(ctx, key, value, ttls.L2TTL); err != nil {
		m.logger.Warn("Failed to set L2 cache", zap.Error(err))
		return err
	}

	return nil
}

// Delete removes data from all cache layers
func (m *MultiLayerCache) Delete(ctx context.Context, keys ...string) error {
	// Delete from L1
	for _, key := range keys {
		m.l1Cache.Delete(key)
	}

	// Delete from L2
	if err := m.l2Cache.Delete(ctx, keys...); err != nil {
		m.logger.Warn("Failed to delete from L2 cache", zap.Error(err))
		return err
	}

	return nil
}

// Invalidate is an alias for Delete
func (m *MultiLayerCache) Invalidate(ctx context.Context, pattern string) error {
	// For now, just delete the exact key
	// TODO: Implement pattern-based invalidation
	return m.Delete(ctx, pattern)
}

// WarmCache pre-populates cache with frequently accessed data
func (m *MultiLayerCache) WarmCache(ctx context.Context, keys []CacheWarmupItem) error {
	m.logger.Info("Starting cache warmup", zap.Int("items", len(keys)))
	
	for _, item := range keys {
		if err := m.Set(ctx, item.Key, item.Value, item.TTLs); err != nil {
			m.logger.Warn("Failed to warm cache for key",
				zap.String("key", item.Key),
				zap.Error(err),
			)
			continue
		}
	}
	
	m.logger.Info("Cache warmup completed")
	return nil
}

// GetStats returns statistics for all cache layers
func (m *MultiLayerCache) GetStats(ctx context.Context) map[string]interface{} {
	l1Stats := m.l1Cache.Stats()
	l2Stats, _ := m.l2Cache.Stats(ctx)
	
	return map[string]interface{}{
		"l1_stats": l1Stats,
		"l2_stats": l2Stats,
		"combined_hit_rate": m.calculateCombinedHitRate(),
	}
}

// populateCaches populates L1 and L2 caches asynchronously
func (m *MultiLayerCache) populateCaches(ctx context.Context, key string, data interface{}, ttls *CacheTTLs) {
	// Set in L2 (Redis)
	if err := m.l2Cache.SetJSON(ctx, key, data, ttls.L2TTL); err != nil {
		m.logger.Warn("Failed to populate L2 cache",
			zap.String("key", key),
			zap.Error(err),
		)
	}

	// Set in L1 (Memory)
	if err := m.l1Cache.SetJSON(key, data, ttls.L1TTL); err != nil {
		m.logger.Warn("Failed to populate L1 cache",
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

// unmarshalData unmarshals data to destination
func (m *MultiLayerCache) unmarshalData(data interface{}, dest interface{}) error {
	// Handle []byte (normal case)
	if dataBytes, ok := data.([]byte); ok {
		return json.Unmarshal(dataBytes, dest)
	}
	
	// Handle string (JSON string - corrupted cache case)
	if str, ok := data.(string); ok {
		// Try to unmarshal directly as JSON string
		return json.Unmarshal([]byte(str), dest)
	}
	
	// Handle other types - marshal then unmarshal
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	
	return json.Unmarshal(dataBytes, dest)
}

// calculateCombinedHitRate calculates overall hit rate across all layers
func (m *MultiLayerCache) calculateCombinedHitRate() float64 {
	l1Metrics := m.l1Cache.GetMetrics()
	l2Metrics := m.l2Cache.GetMetrics()
	
	totalHits := l1Metrics.Hits + l2Metrics.Hits
	totalMisses := l1Metrics.Misses + l2Metrics.Misses
	total := totalHits + totalMisses
	
	if total == 0 {
		return 0.0
	}
	
	return float64(totalHits) / float64(total)
}

// CacheTTLs holds TTL values for each cache layer
type CacheTTLs struct {
	L1TTL time.Duration
	L2TTL time.Duration
}

// CacheWarmupItem represents an item to warm up in cache
type CacheWarmupItem struct {
	Key   string
	Value interface{}
	TTLs  *CacheTTLs
}

// CheckRateLimit checks if rate limit is exceeded using L2 cache (Redis)
func (m *MultiLayerCache) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return m.l2Cache.CheckRateLimit(ctx, key, limit, window)
}

