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
	"encoding/json"
	"net/http"

	"github.com/ibn-network/backend/internal/services/analytics/metrics"
	"go.uber.org/zap"
)

// Handler handles metrics-related HTTP requests
type Handler struct {
	service *metrics.Service
	logger  *zap.Logger
}

// NewHandler creates a new metrics handler
func NewHandler(service *metrics.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetMetrics handles retrieving all current metrics
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metricsData := h.service.GetAllMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": metricsData,
		"count":   len(metricsData),
	})
}

// GetAggregations handles retrieving aggregated metrics
func (h *Handler) GetAggregations(w http.ResponseWriter, r *http.Request) {
	aggregations := h.service.GetAggregations()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"aggregations": aggregations,
		"count":        len(aggregations),
	})
}

// GetSnapshot handles retrieving a snapshot of current metrics
func (h *Handler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot := h.service.GetSnapshot()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// GetMetricByName handles retrieving a specific metric by name
func (h *Handler) GetMetricByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	metric := h.service.GetMetricByName(name)
	if metric == nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metric)
}

