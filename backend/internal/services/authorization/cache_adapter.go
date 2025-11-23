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
	"errors"
	"time"

	"github.com/ibn-network/backend/internal/infrastructure/cache"
)

var ErrCacheMiss = errors.New("cache miss")

// CacheAdapter adapts MultiLayerCache to PermissionCache interface
type CacheAdapter struct {
	cache *cache.MultiLayerCache
}

// NewCacheAdapter creates a new cache adapter
func NewCacheAdapter(multiCache *cache.MultiLayerCache) *CacheAdapter {
	return &CacheAdapter{
		cache: multiCache,
	}
}

// Get retrieves a value from cache
func (a *CacheAdapter) Get(ctx context.Context, key string) (interface{}, error) {
	var value interface{}
	err := a.cache.Get(ctx, key, &value, func(ctx context.Context) (interface{}, error) {
		return nil, ErrCacheMiss // Return cache miss if not found
	}, nil)
	if err == ErrCacheMiss {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Set stores a value in cache
func (a *CacheAdapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Use L1 cache for authorization (fast, short TTL)
	// Don't cache in L2 for authorization (security - permissions can change)
	ttls := &cache.CacheTTLs{
		L1TTL: ttl,
		L2TTL: 0, // Don't cache in L2 for authorization
	}
	return a.cache.Set(ctx, key, value, ttls)
}

// Delete removes keys from cache
func (a *CacheAdapter) Delete(ctx context.Context, keys ...string) error {
	return a.cache.Delete(ctx, keys...)
}

