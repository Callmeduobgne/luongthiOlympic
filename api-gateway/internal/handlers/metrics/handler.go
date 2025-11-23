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
	"time"

	"github.com/ibn-network/api-gateway/internal/models"
	metricsservice "github.com/ibn-network/api-gateway/internal/services/metrics"
	"go.uber.org/zap"
)

// Handler handles metrics operations
type Handler struct {
	metricsService *metricsservice.Service
	logger         *zap.Logger
}

// NewHandler creates a new metrics handler
func NewHandler(metricsService *metricsservice.Service, logger *zap.Logger) *Handler {
	return &Handler{
		metricsService: metricsService,
		logger:         logger,
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// GetTransactionMetrics godoc
// @Summary Get transaction metrics
// @Description Get transaction-related metrics
// @Tags metrics
// @Produce json
// @Param channel query string false "Channel name"
// @Param startTime query string false "Start time (RFC3339)"
// @Param endTime query string false "End time (RFC3339)"
// @Success 200 {object} models.APIResponse{data=models.TransactionMetricsResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /metrics/transactions [get]
func (h *Handler) GetTransactionMetrics(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channel")
	
	var startTime, endTime *time.Time
	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	metrics, err := h.metricsService.GetTransactionMetrics(r.Context(), channelName, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get transaction metrics", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get transaction metrics",
			err.Error(),
		))
		return
	}

	// Convert to response model
	response := &models.TransactionMetricsResponse{
		Total:           metrics.Total,
		Valid:           metrics.Valid,
		Invalid:         metrics.Invalid,
		Submitted:      metrics.Submitted,
		SuccessRate:     metrics.SuccessRate,
		AverageDuration: metrics.AverageDuration,
		ByChannel:       metrics.ByChannel,
		ByChaincode:     metrics.ByChaincode,
		ByStatus:        metrics.ByStatus,
		Last24Hours:     metrics.Last24Hours,
		Last7Days:       metrics.Last7Days,
		Last30Days:      metrics.Last30Days,
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetBlockMetrics godoc
// @Summary Get block metrics
// @Description Get block-related metrics
// @Tags metrics
// @Produce json
// @Param channel query string false "Channel name"
// @Success 200 {object} models.APIResponse{data=models.BlockMetricsResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /metrics/blocks [get]
func (h *Handler) GetBlockMetrics(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channel")

	metrics, err := h.metricsService.GetBlockMetrics(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to get block metrics", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get block metrics",
			err.Error(),
		))
		return
	}

	// Convert to response model
	response := &models.BlockMetricsResponse{
		Total:            metrics.Total,
		Last24Hours:      metrics.Last24Hours,
		Last7Days:        metrics.Last7Days,
		AverageBlockTime: metrics.AverageBlockTime,
		LargestBlock:     metrics.LargestBlock,
		ByChannel:        metrics.ByChannel,
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetPerformanceMetrics godoc
// @Summary Get performance metrics
// @Description Get performance-related metrics from audit logs
// @Tags metrics
// @Produce json
// @Param startTime query string false "Start time (RFC3339)"
// @Param endTime query string false "End time (RFC3339)"
// @Success 200 {object} models.APIResponse{data=models.PerformanceMetricsResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /metrics/performance [get]
func (h *Handler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	var startTime, endTime *time.Time
	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	metrics, err := h.metricsService.GetPerformanceMetrics(r.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get performance metrics", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get performance metrics",
			err.Error(),
		))
		return
	}

	// Convert to response model
	response := &models.PerformanceMetricsResponse{
		AverageResponseTime: metrics.AverageResponseTime,
		P95ResponseTime:    metrics.P95ResponseTime,
		P99ResponseTime:    metrics.P99ResponseTime,
		RequestsPerSecond:  metrics.RequestsPerSecond,
		ErrorRate:          metrics.ErrorRate,
		TotalRequests:      metrics.TotalRequests,
		SuccessfulRequests: metrics.SuccessfulRequests,
		FailedRequests:     metrics.FailedRequests,
		ByEndpoint:         metrics.ByEndpoint,
		ByStatus:           metrics.ByStatus,
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetPeerMetrics godoc
// @Summary Get peer metrics
// @Description Get peer-related metrics
// @Tags metrics
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.PeerMetricsResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /metrics/peers [get]
func (h *Handler) GetPeerMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricsService.GetPeerMetrics(r.Context())
	if err != nil {
		h.logger.Error("Failed to get peer metrics", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get peer metrics",
			err.Error(),
		))
		return
	}

	// Convert to response model
	response := &models.PeerMetricsResponse{
		TotalPeers:    metrics.TotalPeers,
		ActivePeers:   metrics.ActivePeers,
		InactivePeers: metrics.InactivePeers,
		ByChannel:     metrics.ByChannel,
		ByMSP:         metrics.ByMSP,
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetMetricsSummary godoc
// @Summary Get metrics summary
// @Description Get overall metrics summary
// @Tags metrics
// @Produce json
// @Param channel query string false "Channel name"
// @Success 200 {object} models.APIResponse{data=models.MetricsSummaryResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /metrics/summary [get]
func (h *Handler) GetMetricsSummary(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channel")

	summary, err := h.metricsService.GetMetricsSummary(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to get metrics summary", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get metrics summary",
			err.Error(),
		))
		return
	}

	// Convert to response model
	response := &models.MetricsSummaryResponse{
		Transactions: models.TransactionMetricsResponse{
			Total:           summary.Transactions.Total,
			Valid:           summary.Transactions.Valid,
			Invalid:         summary.Transactions.Invalid,
			Submitted:      summary.Transactions.Submitted,
			SuccessRate:     summary.Transactions.SuccessRate,
			AverageDuration: summary.Transactions.AverageDuration,
			ByChannel:       summary.Transactions.ByChannel,
			ByChaincode:     summary.Transactions.ByChaincode,
			ByStatus:        summary.Transactions.ByStatus,
			Last24Hours:     summary.Transactions.Last24Hours,
			Last7Days:       summary.Transactions.Last7Days,
			Last30Days:      summary.Transactions.Last30Days,
		},
		Blocks: models.BlockMetricsResponse{
			Total:            summary.Blocks.Total,
			Last24Hours:      summary.Blocks.Last24Hours,
			Last7Days:        summary.Blocks.Last7Days,
			AverageBlockTime: summary.Blocks.AverageBlockTime,
			LargestBlock:     summary.Blocks.LargestBlock,
			ByChannel:        summary.Blocks.ByChannel,
		},
		Performance: models.PerformanceMetricsResponse{
			AverageResponseTime: summary.Performance.AverageResponseTime,
			P95ResponseTime:    summary.Performance.P95ResponseTime,
			P99ResponseTime:    summary.Performance.P99ResponseTime,
			RequestsPerSecond:  summary.Performance.RequestsPerSecond,
			ErrorRate:          summary.Performance.ErrorRate,
			TotalRequests:      summary.Performance.TotalRequests,
			SuccessfulRequests: summary.Performance.SuccessfulRequests,
			FailedRequests:     summary.Performance.FailedRequests,
			ByEndpoint:         summary.Performance.ByEndpoint,
			ByStatus:           summary.Performance.ByStatus,
		},
		Peers: models.PeerMetricsResponse{
			TotalPeers:    summary.Peers.TotalPeers,
			ActivePeers:   summary.Peers.ActivePeers,
			InactivePeers: summary.Peers.InactivePeers,
			ByChannel:     summary.Peers.ByChannel,
			ByMSP:         summary.Peers.ByMSP,
		},
		Timestamp: summary.Timestamp,
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

