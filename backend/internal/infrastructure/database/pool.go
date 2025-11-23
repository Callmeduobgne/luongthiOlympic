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

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ibn-network/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Pool represents a PostgreSQL connection pool
type Pool struct {
	primary  *pgxpool.Pool
	replicas []*pgxpool.Pool
	logger   *zap.Logger
	config   *config.DatabaseConfig
}

// NewPool creates a new database connection pool
func NewPool(cfg *config.DatabaseConfig, logger *zap.Logger) (*Pool, error) {
	// Create primary pool
	primary, err := createPool(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary pool: %w", err)
	}

	// Create replica pools
	var replicas []*pgxpool.Pool
	for i, replicaCfg := range cfg.ReadReplicas {
		replicaPool, err := createReplicaPool(&replicaCfg, cfg, logger)
		if err != nil {
			logger.Warn("Failed to create replica pool",
				zap.Int("replica_index", i),
				zap.Error(err),
			)
			// Continue even if replica fails
			continue
		}
		replicas = append(replicas, replicaPool)
	}

	logger.Info("Database pools created",
		zap.Int("replicas", len(replicas)),
	)

	return &Pool{
		primary:  primary,
		replicas: replicas,
		logger:   logger,
		config:   cfg,
	}, nil
}

// Primary returns the primary database pool for write operations
func (p *Pool) Primary() *pgxpool.Pool {
	return p.primary
}

// Replica returns a read replica pool (round-robin)
// Falls back to primary if no replicas available
func (p *Pool) Replica() *pgxpool.Pool {
	if len(p.replicas) == 0 {
		return p.primary
	}

	// Simple round-robin selection
	// TODO: Implement more sophisticated load balancing
	now := time.Now().UnixNano()
	index := int(now) % len(p.replicas)
	
	replica := p.replicas[index]
	
	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	if err := replica.Ping(ctx); err != nil {
		p.logger.Warn("Replica unhealthy, falling back to primary",
			zap.Int("replica_index", index),
			zap.Error(err),
		)
		return p.primary
	}
	
	return replica
}

// Close closes all database connections
func (p *Pool) Close() {
	p.primary.Close()
	for _, replica := range p.replicas {
		replica.Close()
	}
	p.logger.Info("Database pools closed")
}

// Stats returns pool statistics
func (p *Pool) Stats() PoolStats {
	primaryStats := p.primary.Stat()
	
	var replicaStats []pgxpool.Stat
	for _, replica := range p.replicas {
		replicaStats = append(replicaStats, *replica.Stat())
	}
	
	return PoolStats{
		Primary:  *primaryStats,
		Replicas: replicaStats,
	}
}

// PoolStats holds statistics for all pools
type PoolStats struct {
	Primary  pgxpool.Stat
	Replicas []pgxpool.Stat
}

// Health checks the health of all database connections
func (p *Pool) Health(ctx context.Context) error {
	// Check primary
	if err := p.primary.Ping(ctx); err != nil {
		return fmt.Errorf("primary database unhealthy: %w", err)
	}
	
	// Check replicas (warning only, not critical)
	for i, replica := range p.replicas {
		if err := replica.Ping(ctx); err != nil {
			p.logger.Warn("Replica unhealthy",
				zap.Int("replica_index", i),
				zap.Error(err),
			)
		}
	}
	
	return nil
}

// createPool creates a connection pool with the given configuration
func createPool(cfg *config.DatabaseConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Connection pool configuration
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MaxConnLifetime = cfg.MaxLifetime
	poolConfig.MaxConnIdleTime = cfg.IdleTimeout
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Create pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection pool created",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
		zap.Int("min_conns", cfg.MinConns),
		zap.Int("max_conns", cfg.MaxConns),
	)

	return pool, nil
}

// createReplicaPool creates a replica connection pool
func createReplicaPool(replicaCfg *config.DatabaseReplicaConfig, primaryCfg *config.DatabaseConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		replicaCfg.Host, replicaCfg.Port, replicaCfg.User, replicaCfg.Password,
		replicaCfg.Database, replicaCfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse replica config: %w", err)
	}

	// Use same pool settings as primary
	poolConfig.MinConns = int32(primaryCfg.MinConns)
	poolConfig.MaxConns = int32(primaryCfg.MaxConns)
	poolConfig.MaxConnLifetime = primaryCfg.MaxLifetime
	poolConfig.MaxConnIdleTime = primaryCfg.IdleTimeout
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Create pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create replica pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping replica: %w", err)
	}

	logger.Info("Replica connection pool created",
		zap.String("host", replicaCfg.Host),
		zap.Int("port", replicaCfg.Port),
	)

	return pool, nil
}

