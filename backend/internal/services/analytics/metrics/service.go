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
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Service handles metrics business logic
type Service struct {
	collector *Collector
	logger    *zap.Logger
}

// NewService creates a new metrics service
func NewService(logger *zap.Logger) *Service {
	return &Service{
		collector: NewCollector(logger),
		logger:    logger,
	}
}

// RecordAPIRequest records an API request metric
func (s *Service) RecordAPIRequest(method, path string, statusCode int, duration time.Duration) {
	labels := map[string]string{
		"method": method,
		"path":   path,
		"status": string(rune(statusCode)),
	}

	s.collector.RecordCounter(MetricAPIRequestTotal, 1, labels)
	s.collector.RecordHistogram(MetricAPIRequestDuration, float64(duration.Milliseconds()), labels)

	if statusCode >= 400 {
		s.collector.RecordCounter(MetricAPIErrorTotal, 1, labels)
	}
}

// RecordAuthLogin records a login attempt
func (s *Service) RecordAuthLogin(success bool) {
	s.collector.RecordCounter(MetricAuthLoginTotal, 1, map[string]string{
		"success": boolToString(success),
	})

	if !success {
		s.collector.RecordCounter(MetricAuthLoginFailedTotal, 1, nil)
	}
}

// RecordAuthRegister records a registration
func (s *Service) RecordAuthRegister(success bool) {
	s.collector.RecordCounter(MetricAuthRegisterTotal, 1, map[string]string{
		"success": boolToString(success),
	})
}

// SetActiveUsers sets the current number of active users
func (s *Service) SetActiveUsers(count int) {
	s.collector.RecordGauge(MetricAuthActiveUsers, float64(count), nil)
}

// SetActiveAPIKeys sets the current number of active API keys
func (s *Service) SetActiveAPIKeys(count int) {
	s.collector.RecordGauge(MetricAuthActiveAPIKeys, float64(count), nil)
}

// RecordDatabaseQuery records a database query
func (s *Service) RecordDatabaseQuery(queryType string, duration time.Duration, success bool) {
	labels := map[string]string{
		"type":    queryType,
		"success": boolToString(success),
	}

	s.collector.RecordCounter(MetricDBQueryTotal, 1, labels)
	s.collector.RecordHistogram(MetricDBQueryDuration, float64(duration.Milliseconds()), labels)

	if !success {
		s.collector.RecordCounter(MetricDBQueryErrorTotal, 1, labels)
	}
}

// SetDatabaseConnections sets the current database connection metrics
func (s *Service) SetDatabaseConnections(active, idle int) {
	s.collector.RecordGauge(MetricDBConnActive, float64(active), nil)
	s.collector.RecordGauge(MetricDBConnIdle, float64(idle), nil)
}

// RecordCacheOperation records a cache operation
func (s *Service) RecordCacheOperation(operation string, layer string, hit bool) {
	labels := map[string]string{
		"operation": operation,
		"layer":     layer,
	}

	if operation == "get" {
		if hit {
			s.collector.RecordCounter(MetricCacheHitTotal, 1, labels)
		} else {
			s.collector.RecordCounter(MetricCacheMissTotal, 1, labels)
		}
	} else if operation == "set" {
		s.collector.RecordCounter(MetricCacheSetTotal, 1, labels)
	}
}

// SetCacheSize sets the current cache size
func (s *Service) SetCacheSize(layer string, sizeBytes int64) {
	s.collector.RecordGauge(MetricCacheSize, float64(sizeBytes), map[string]string{
		"layer": layer,
	})
}

// RecordBlockchainTransaction records a blockchain transaction
func (s *Service) RecordBlockchainTransaction(txType string, duration time.Duration, success bool) {
	labels := map[string]string{
		"type":    txType,
		"success": boolToString(success),
	}

	s.collector.RecordCounter(MetricBlockchainTxTotal, 1, labels)
	s.collector.RecordHistogram(MetricBlockchainTxDuration, float64(duration.Milliseconds()), labels)

	if success {
		s.collector.RecordCounter(MetricBlockchainTxSuccessTotal, 1, labels)
	} else {
		s.collector.RecordCounter(MetricBlockchainTxFailedTotal, 1, labels)
	}
}

// SetBlockchainMetrics sets blockchain-related metrics
func (s *Service) SetBlockchainMetrics(blockHeight uint64, peerCount int) {
	s.collector.RecordGauge(MetricBlockchainBlockHeight, float64(blockHeight), nil)
	s.collector.RecordGauge(MetricBlockchainPeerCount, float64(peerCount), nil)
}

// RecordEventPublished records an event publication
func (s *Service) RecordEventPublished(eventType string) {
	s.collector.RecordCounter(MetricEventPublishedTotal, 1, map[string]string{
		"type": eventType,
	})
}

// RecordEventDelivered records an event delivery
func (s *Service) RecordEventDelivered(eventType string, success bool) {
	labels := map[string]string{
		"type":    eventType,
		"success": boolToString(success),
	}

	s.collector.RecordCounter(MetricEventDeliveredTotal, 1, labels)

	if !success {
		s.collector.RecordCounter(MetricEventFailedTotal, 1, labels)
	}
}

// RecordWebhookDelivery records a webhook delivery time
func (s *Service) RecordWebhookDelivery(duration time.Duration, success bool) {
	s.collector.RecordHistogram(MetricWebhookDeliveryTime, float64(duration.Milliseconds()), map[string]string{
		"success": boolToString(success),
	})
}

// SetWebsocketConnections sets the current number of websocket connections
func (s *Service) SetWebsocketConnections(count int) {
	s.collector.RecordGauge(MetricWebsocketConnections, float64(count), nil)
}

// GetAllMetrics returns all collected metrics
func (s *Service) GetAllMetrics() []*Metric {
	return s.collector.GetMetrics()
}

// GetAggregations returns all aggregated metrics
func (s *Service) GetAggregations() []*AggregatedMetric {
	return s.collector.GetAggregations()
}

// GetMetricByName returns a specific metric by name
func (s *Service) GetMetricByName(name string) *Metric {
	return s.collector.GetMetricByName(name)
}

// GetSnapshot returns a snapshot of all current metrics
func (s *Service) GetSnapshot() *MetricSnapshot {
	return s.collector.GetSnapshot()
}

// Stop stops the metrics service
func (s *Service) Stop() {
	s.collector.Stop()
}

// Helper functions

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Middleware to track API metrics
type MetricsMiddleware struct {
	service *Service
	logger  *zap.Logger
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(service *Service, logger *zap.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		service: service,
		logger:  logger,
	}
}

// Handler wraps an HTTP handler and records metrics
func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		m.service.RecordAPIRequest(r.Method, r.URL.Path, ww.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

