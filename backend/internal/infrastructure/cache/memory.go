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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// MemoryCache represents L1 in-memory cache
type MemoryCache struct {
	cache      *gocache.Cache
	mu         sync.RWMutex
	maxSize    int64 // in bytes
	currentSize int64
	logger     *zap.Logger
	metrics    *CacheMetrics
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	mu          sync.RWMutex
	hits        int64
	misses      int64
	sets        int64
	deletes     int64
	evictions   int64
	size        int64
}

// MemoryCacheConfig holds configuration for L1 cache
type MemoryCacheConfig struct {
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
	MaxSize         int64 // ~100MB
}

// NewMemoryCache creates a new L1 in-memory cache
func NewMemoryCache(cfg *MemoryCacheConfig, logger *zap.Logger) *MemoryCache {
	if cfg == nil {
		cfg = &MemoryCacheConfig{
			DefaultTTL:      5 * time.Minute,
			CleanupInterval: 10 * time.Minute,
			MaxSize:         100 * 1024 * 1024, // 100MB
		}
	}

	c := gocache.New(cfg.DefaultTTL, cfg.CleanupInterval)

	logger.Info("L1 Memory cache initialized",
		zap.Duration("default_ttl", cfg.DefaultTTL),
		zap.Int64("max_size_bytes", cfg.MaxSize),
	)

	return &MemoryCache{
		cache:   c,
		maxSize: cfg.MaxSize,
		logger:  logger,
		metrics: &CacheMetrics{},
	}
}

// Get retrieves a value from cache
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, found := m.cache.Get(key)
	
	if found {
		m.metrics.recordHit()
	} else {
		m.metrics.recordMiss()
	}

	return value, found
}

// GetString retrieves a string value from cache
func (m *MemoryCache) GetString(key string) (string, bool) {
	value, found := m.Get(key)
	if !found {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

// GetJSON retrieves and unmarshals JSON from cache
func (m *MemoryCache) GetJSON(key string, dest interface{}) error {
	value, found := m.Get(key)
	if !found {
		return fmt.Errorf("key not found: %s", key)
	}

	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte")
	}

	return json.Unmarshal(data, dest)
}

// Set stores a value in cache with TTL
func (m *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check size before setting
	size := estimateSize(value)
	if m.currentSize+size > m.maxSize {
		// Simple eviction: clear oldest items
		m.logger.Warn("Cache size limit reached, clearing oldest items",
			zap.Int64("current_size", m.currentSize),
			zap.Int64("max_size", m.maxSize),
		)
		m.cache.Flush()
		m.currentSize = 0
		m.metrics.recordEviction()
	}

	m.cache.Set(key, value, ttl)
	m.currentSize += size
	m.metrics.recordSet()

	return nil
}

// SetJSON marshals and stores JSON in cache
func (m *MemoryCache) SetJSON(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return m.Set(key, data, ttl)
}

// Delete removes a key from cache
func (m *MemoryCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Delete(key)
	m.metrics.recordDelete()
}

// Clear removes all items from cache
func (m *MemoryCache) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Flush()
	m.currentSize = 0
	m.logger.Info("Memory cache cleared")
}

// Exists checks if a key exists in cache
func (m *MemoryCache) Exists(key string) bool {
	_, found := m.Get(key)
	return found
}

// GetMetrics returns cache metrics
func (m *MemoryCache) GetMetrics() CacheMetricsSnapshot {
	return m.metrics.snapshot()
}

// Stats returns cache statistics
func (m *MemoryCache) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := m.metrics.snapshot()
	
	return map[string]interface{}{
		"items":       m.cache.ItemCount(),
		"size_bytes":  m.currentSize,
		"max_size":    m.maxSize,
		"utilization": float64(m.currentSize) / float64(m.maxSize),
		"hits":        metrics.Hits,
		"misses":      metrics.Misses,
		"hit_rate":    metrics.HitRate(),
		"sets":        metrics.Sets,
		"deletes":     metrics.Deletes,
		"evictions":   metrics.Evictions,
	}
}

// estimateSize estimates the size of a value in bytes
func estimateSize(value interface{}) int64 {
	// Simple estimation - can be improved
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int32, int64, float32, float64, bool:
		return 8
	default:
		// For complex types, use JSON marshalling as approximation
		data, err := json.Marshal(v)
		if err != nil {
			return 1024 // Default estimate
		}
		return int64(len(data))
	}
}

// CacheMetrics methods
func (m *CacheMetrics) recordHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hits++
}

func (m *CacheMetrics) recordMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.misses++
}

func (m *CacheMetrics) recordSet() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sets++
}

func (m *CacheMetrics) recordDelete() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deletes++
}

func (m *CacheMetrics) recordEviction() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evictions++
}

func (m *CacheMetrics) snapshot() CacheMetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return CacheMetricsSnapshot{
		Hits:      m.hits,
		Misses:    m.misses,
		Sets:      m.sets,
		Deletes:   m.deletes,
		Evictions: m.evictions,
	}
}

// CacheMetricsSnapshot represents a point-in-time snapshot of metrics
type CacheMetricsSnapshot struct {
	Hits      int64
	Misses    int64
	Sets      int64
	Deletes   int64
	Evictions int64
}

// HitRate calculates cache hit rate
func (s *CacheMetricsSnapshot) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(total)
}

