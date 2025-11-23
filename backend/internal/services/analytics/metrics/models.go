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
	"time"

	"github.com/google/uuid"
)

// Metric represents a system or business metric
type Metric struct {
	ID         uuid.UUID              `json:"id"`
	Name       string                 `json:"name"`
	MetricType string                 `json:"metric_type"` // "counter", "gauge", "histogram"
	Value      float64                `json:"value"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// AggregatedMetric represents aggregated metrics over a time period
type AggregatedMetric struct {
	Name       string            `json:"name"`
	MetricType string            `json:"metric_type"`
	Count      int64             `json:"count"`
	Sum        float64           `json:"sum"`
	Min        float64           `json:"min"`
	Max        float64           `json:"max"`
	Avg        float64           `json:"avg"`
	Labels     map[string]string `json:"labels,omitempty"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
}

// QueryMetricsRequest represents metrics query request
type QueryMetricsRequest struct {
	Name       *string           `json:"name,omitempty"`
	MetricType *string           `json:"metric_type,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	StartTime  *time.Time        `json:"start_time,omitempty"`
	EndTime    *time.Time        `json:"end_time,omitempty"`
	Limit      int               `json:"limit"`
}

// Metric types
const (
	MetricTypeCounter   = "counter"
	MetricTypeGauge     = "gauge"
	MetricTypeHistogram = "histogram"
)

// Predefined metric names
const (
	// API Metrics
	MetricAPIRequestTotal    = "api_request_total"
	MetricAPIRequestDuration = "api_request_duration_ms"
	MetricAPIErrorTotal      = "api_error_total"

	// Authentication Metrics
	MetricAuthLoginTotal        = "auth_login_total"
	MetricAuthLoginFailedTotal  = "auth_login_failed_total"
	MetricAuthRegisterTotal     = "auth_register_total"
	MetricAuthActiveUsers       = "auth_active_users"
	MetricAuthActiveAPIKeys     = "auth_active_api_keys"

	// Database Metrics
	MetricDBQueryDuration  = "db_query_duration_ms"
	MetricDBConnActive     = "db_connections_active"
	MetricDBConnIdle       = "db_connections_idle"
	MetricDBQueryTotal     = "db_query_total"
	MetricDBQueryErrorTotal = "db_query_error_total"

	// Cache Metrics
	MetricCacheHitTotal  = "cache_hit_total"
	MetricCacheMissTotal = "cache_miss_total"
	MetricCacheSetTotal  = "cache_set_total"
	MetricCacheSize      = "cache_size_bytes"

	// Blockchain Metrics
	MetricBlockchainTxTotal         = "blockchain_tx_total"
	MetricBlockchainTxDuration      = "blockchain_tx_duration_ms"
	MetricBlockchainTxSuccessTotal  = "blockchain_tx_success_total"
	MetricBlockchainTxFailedTotal   = "blockchain_tx_failed_total"
	MetricBlockchainBlockHeight     = "blockchain_block_height"
	MetricBlockchainPeerCount       = "blockchain_peer_count"

	// Event Metrics
	MetricEventPublishedTotal  = "event_published_total"
	MetricEventDeliveredTotal  = "event_delivered_total"
	MetricEventFailedTotal     = "event_failed_total"
	MetricWebhookDeliveryTime  = "webhook_delivery_time_ms"
	MetricWebsocketConnections = "websocket_connections_active"

	// System Metrics
	MetricSystemCPUUsage    = "system_cpu_usage_percent"
	MetricSystemMemoryUsage = "system_memory_usage_bytes"
	MetricSystemGoroutines  = "system_goroutines_count"
)

// MetricSnapshot represents a snapshot of all metrics at a point in time
type MetricSnapshot struct {
	Timestamp time.Time         `json:"timestamp"`
	Metrics   map[string]float64 `json:"metrics"`
}

