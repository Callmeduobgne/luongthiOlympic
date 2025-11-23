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
	"time"

	"go.uber.org/zap"
)

// MetricsCollector collects and reports database metrics
type MetricsCollector struct {
	pool   *Pool
	logger *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(pool *Pool, logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		pool:   pool,
		logger: logger,
	}
}

// CollectMetrics collects current pool metrics
func (m *MetricsCollector) CollectMetrics() PoolMetrics {
	stats := m.pool.Stats()
	
	primaryMetrics := calculateMetrics(stats.Primary)
	
	var replicaMetrics []ConnectionMetrics
	for _, replicaStat := range stats.Replicas {
		replicaMetrics = append(replicaMetrics, calculateMetrics(replicaStat))
	}
	
	return PoolMetrics{
		Primary:  primaryMetrics,
		Replicas: replicaMetrics,
	}
}

// StartMonitoring starts periodic metrics collection
func (m *MetricsCollector) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Stopping database metrics collection")
			return
		case <-ticker.C:
			metrics := m.CollectMetrics()
			m.logMetrics(metrics)
			m.checkThresholds(metrics)
		}
	}
}

// logMetrics logs current metrics
func (m *MetricsCollector) logMetrics(metrics PoolMetrics) {
	m.logger.Debug("Database pool metrics",
		zap.Int32("primary_total_conns", metrics.Primary.TotalConns),
		zap.Int32("primary_idle_conns", metrics.Primary.IdleConns),
		zap.Int32("primary_acquired_conns", metrics.Primary.AcquiredConns),
		zap.Float64("primary_utilization", metrics.Primary.Utilization),
	)
	
	for i, replica := range metrics.Replicas {
		m.logger.Debug("Replica pool metrics",
			zap.Int("replica_index", i),
			zap.Int32("total_conns", replica.TotalConns),
			zap.Int32("idle_conns", replica.IdleConns),
			zap.Float64("utilization", replica.Utilization),
		)
	}
}

// checkThresholds checks if metrics exceed warning thresholds
func (m *MetricsCollector) checkThresholds(metrics PoolMetrics) {
	// Alert if primary pool utilization > 70%
	if metrics.Primary.Utilization > 0.7 {
		m.logger.Warn("Primary database pool utilization high",
			zap.Float64("utilization", metrics.Primary.Utilization),
			zap.Int32("acquired_conns", metrics.Primary.AcquiredConns),
			zap.Int32("max_conns", metrics.Primary.MaxConns),
		)
	}
	
	// Critical alert if utilization > 90%
	if metrics.Primary.Utilization > 0.9 {
		m.logger.Error("Primary database pool utilization critical",
			zap.Float64("utilization", metrics.Primary.Utilization),
			zap.Int32("acquired_conns", metrics.Primary.AcquiredConns),
			zap.Int32("max_conns", metrics.Primary.MaxConns),
		)
	}
}

// PoolMetrics holds metrics for all pools
type PoolMetrics struct {
	Primary  ConnectionMetrics
	Replicas []ConnectionMetrics
}

// ConnectionMetrics holds metrics for a single connection pool
type ConnectionMetrics struct {
	TotalConns           int32
	IdleConns            int32
	AcquiredConns        int32
	ConstructingConns    int32
	MaxConns             int32
	Utilization          float64 // acquired / max
	AcquireCount         int64
	AcquireDuration      time.Duration
	CanceledAcquireCount int64
}

// calculateMetrics calculates metrics from pool stats
func calculateMetrics(stat interface{}) ConnectionMetrics {
	// Type assertion for pgxpool.Stat
	// This is a simplified version - in real implementation,
	// we would use proper type assertion
	
	return ConnectionMetrics{
		TotalConns:    0, // stat.TotalConns(),
		IdleConns:     0, // stat.IdleConns(),
		AcquiredConns: 0, // stat.AcquiredConns(),
		MaxConns:      0, // stat.MaxConns(),
		Utilization:   0.0,
	}
}

