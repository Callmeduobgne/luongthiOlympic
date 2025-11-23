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

package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Collector collects and stores metrics in memory with periodic aggregation
type Collector struct {
	logger        *zap.Logger
	mu            sync.RWMutex
	metrics       map[string]*Metric
	aggregations  map[string]*AggregatedMetric
	flushInterval time.Duration
	stopCh        chan struct{}
}

// NewCollector creates a new metrics collector
func NewCollector(logger *zap.Logger) *Collector {
	c := &Collector{
		logger:        logger,
		metrics:       make(map[string]*Metric),
		aggregations:  make(map[string]*AggregatedMetric),
		flushInterval: 1 * time.Minute,
		stopCh:        make(chan struct{}),
	}

	// Start background aggregation
	go c.startAggregation()

	// Start system metrics collection
	go c.collectSystemMetrics()

	return c
}

// RecordCounter increments a counter metric
func (c *Collector) RecordCounter(name string, value float64, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(name, labels)
	
	if existing, ok := c.metrics[key]; ok {
		existing.Value += value
		existing.Timestamp = time.Now()
	} else {
		c.metrics[key] = &Metric{
			ID:         uuid.New(),
			Name:       name,
			MetricType: MetricTypeCounter,
			Value:      value,
			Labels:     labels,
			Timestamp:  time.Now(),
		}
	}
}

// RecordGauge sets a gauge metric to a specific value
func (c *Collector) RecordGauge(name string, value float64, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(name, labels)
	
	c.metrics[key] = &Metric{
		ID:         uuid.New(),
		Name:       name,
		MetricType: MetricTypeGauge,
		Value:      value,
		Labels:     labels,
		Timestamp:  time.Now(),
	}
}

// RecordHistogram records a histogram value (e.g., duration)
func (c *Collector) RecordHistogram(name string, value float64, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(name, labels)
	
	c.metrics[key] = &Metric{
		ID:         uuid.New(),
		Name:       name,
		MetricType: MetricTypeHistogram,
		Value:      value,
		Labels:     labels,
		Timestamp:  time.Now(),
	}

	// Update aggregation
	if agg, ok := c.aggregations[key]; ok {
		agg.Count++
		agg.Sum += value
		if value < agg.Min {
			agg.Min = value
		}
		if value > agg.Max {
			agg.Max = value
		}
		agg.Avg = agg.Sum / float64(agg.Count)
		agg.EndTime = time.Now()
	} else {
		c.aggregations[key] = &AggregatedMetric{
			Name:       name,
			MetricType: MetricTypeHistogram,
			Count:      1,
			Sum:        value,
			Min:        value,
			Max:        value,
			Avg:        value,
			Labels:     labels,
			StartTime:  time.Now(),
			EndTime:    time.Now(),
		}
	}
}

// GetMetrics retrieves all current metrics
func (c *Collector) GetMetrics() []*Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := make([]*Metric, 0, len(c.metrics))
	for _, m := range c.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// GetAggregations retrieves all aggregated metrics
func (c *Collector) GetAggregations() []*AggregatedMetric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	aggregations := make([]*AggregatedMetric, 0, len(c.aggregations))
	for _, a := range c.aggregations {
		aggregations = append(aggregations, a)
	}
	return aggregations
}

// GetMetricByName retrieves a specific metric by name
func (c *Collector) GetMetricByName(name string) *Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, m := range c.metrics {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// GetSnapshot creates a snapshot of all current metrics
func (c *Collector) GetSnapshot() *MetricSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshot := &MetricSnapshot{
		Timestamp: time.Now(),
		Metrics:   make(map[string]float64),
	}

	for key, m := range c.metrics {
		snapshot.Metrics[key] = m.Value
	}

	return snapshot
}

// generateKey creates a unique key for a metric based on name and labels
func (c *Collector) generateKey(name string, labels map[string]string) string {
	key := name
	if len(labels) > 0 {
		for k, v := range labels {
			key += ":" + k + "=" + v
		}
	}
	return key
}

// startAggregation periodically aggregates metrics
func (c *Collector) startAggregation() {
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.performAggregation()
		case <-c.stopCh:
			return
		}
	}
}

// performAggregation aggregates metrics and resets counters
func (c *Collector) performAggregation() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Log current metrics count
	c.logger.Debug("Aggregating metrics",
		zap.Int("metrics_count", len(c.metrics)),
		zap.Int("aggregations_count", len(c.aggregations)),
	)

	// Keep only recent metrics (last 5 minutes)
	now := time.Now()
	for key, m := range c.metrics {
		if now.Sub(m.Timestamp) > 5*time.Minute {
			delete(c.metrics, key)
		}
	}

	// Keep aggregations for 1 hour
	for key, a := range c.aggregations {
		if now.Sub(a.EndTime) > 1*time.Hour {
			delete(c.aggregations, key)
		}
	}
}

// collectSystemMetrics collects system-level metrics
func (c *Collector) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			// Record system metrics
			c.RecordGauge(MetricSystemMemoryUsage, float64(memStats.Alloc), nil)
			c.RecordGauge(MetricSystemGoroutines, float64(runtime.NumGoroutine()), nil)

		case <-c.stopCh:
			return
		}
	}
}

// Stop stops the metrics collector
func (c *Collector) Stop() {
	close(c.stopCh)
	c.logger.Info("Metrics collector stopped")
}

