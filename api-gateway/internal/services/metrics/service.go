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
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/audit"
	"github.com/ibn-network/api-gateway/internal/services/explorer"
	"github.com/ibn-network/api-gateway/internal/services/transaction"
	"go.uber.org/zap"
)

// Service provides metrics aggregation
type Service struct {
	db              *pgxpool.Pool
	transactionService *transaction.Service
	explorerService    *explorer.Service
	auditService       *audit.Service
	logger          *zap.Logger
}

// NewService creates a new metrics service
func NewService(
	db *pgxpool.Pool,
	txService *transaction.Service,
	explorerService *explorer.Service,
	auditService *audit.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		db:              db,
		transactionService: txService,
		explorerService:    explorerService,
		auditService:       auditService,
		logger:          logger,
	}
}

// TransactionMetrics represents transaction-related metrics
type TransactionMetrics struct {
	Total           int64   `json:"total"`
	Valid           int64   `json:"valid"`
	Invalid         int64   `json:"invalid"`
	Submitted       int64   `json:"submitted"`
	SuccessRate     float64 `json:"successRate"`
	AverageDuration float64 `json:"averageDuration"` // milliseconds
	ByChannel       map[string]int64 `json:"byChannel"`
	ByChaincode     map[string]int64 `json:"byChaincode"`
	ByStatus        map[string]int64 `json:"byStatus"`
	Last24Hours     int64   `json:"last24Hours"`
	Last7Days       int64   `json:"last7Days"`
	Last30Days      int64   `json:"last30Days"`
}

// BlockMetrics represents block-related metrics
type BlockMetrics struct {
	Total           int64   `json:"total"`
	Last24Hours     int64   `json:"last24Hours"`
	Last7Days       int64   `json:"last7Days"`
	AverageBlockTime float64 `json:"averageBlockTime"` // seconds
	LargestBlock    uint64  `json:"largestBlock"` // transaction count
	ByChannel       map[string]int64 `json:"byChannel"`
}

// PerformanceMetrics represents performance-related metrics
type PerformanceMetrics struct {
	AverageResponseTime float64 `json:"averageResponseTime"` // milliseconds
	P95ResponseTime    float64 `json:"p95ResponseTime"` // milliseconds
	P99ResponseTime    float64 `json:"p99ResponseTime"` // milliseconds
	RequestsPerSecond  float64 `json:"requestsPerSecond"`
	ErrorRate          float64 `json:"errorRate"` // percentage
	TotalRequests      int64   `json:"totalRequests"`
	SuccessfulRequests int64   `json:"successfulRequests"`
	FailedRequests     int64   `json:"failedRequests"`
	ByEndpoint         map[string]int64 `json:"byEndpoint"`
	ByStatus           map[string]int64 `json:"byStatus"`
}

// PeerMetrics represents peer-related metrics
type PeerMetrics struct {
	TotalPeers    int64            `json:"totalPeers"`
	ActivePeers   int64            `json:"activePeers"`
	InactivePeers int64            `json:"inactivePeers"`
	ByChannel     map[string]int64 `json:"byChannel"`
	ByMSP         map[string]int64 `json:"byMSP"`
}

// MetricsSummary represents overall metrics summary
type MetricsSummary struct {
	Transactions TransactionMetrics `json:"transactions"`
	Blocks       BlockMetrics        `json:"blocks"`
	Performance  PerformanceMetrics  `json:"performance"`
	Peers        PeerMetrics         `json:"peers"`
	Timestamp    time.Time           `json:"timestamp"`
}

// GetTransactionMetrics retrieves transaction metrics
func (s *Service) GetTransactionMetrics(ctx context.Context, channelName string, startTime, endTime *time.Time) (*TransactionMetrics, error) {
	s.logger.Info("Getting transaction metrics",
		zap.String("channel", channelName),
	)

	// Build query
	query := &models.TransactionListQuery{
		ChannelName: channelName,
		Limit:       10000, // Get all for metrics
		Offset:      0,
	}
	if startTime != nil {
		query.StartTime = startTime
	}
	if endTime != nil {
		query.EndTime = endTime
	}

	// Get all transactions
	transactions, total, err := s.transactionService.ListTransactions(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	metrics := &TransactionMetrics{
		Total:       total,
		ByChannel:   make(map[string]int64),
		ByChaincode: make(map[string]int64),
		ByStatus:    make(map[string]int64),
	}

	now := time.Now()
	last24Hours := now.Add(-24 * time.Hour)
	last7Days := now.Add(-7 * 24 * time.Hour)
	last30Days := now.Add(-30 * 24 * time.Hour)

	var totalDuration int64
	var durationCount int64

	for _, tx := range transactions {
		// Count by status
		metrics.ByStatus[string(tx.Status)]++
		switch tx.Status {
		case models.TransactionStatusValid:
			metrics.Valid++
		case models.TransactionStatusInvalid:
			metrics.Invalid++
		case models.TransactionStatusSubmitted:
			metrics.Submitted++
		}

		// Count by channel
		if tx.ChannelName != "" {
			metrics.ByChannel[tx.ChannelName]++
		}

		// Count by chaincode
		if tx.ChaincodeName != "" {
			metrics.ByChaincode[tx.ChaincodeName]++
		}

		// Count by time range
		if tx.Timestamp.After(last24Hours) {
			metrics.Last24Hours++
		}
		if tx.Timestamp.After(last7Days) {
			metrics.Last7Days++
		}
		if tx.Timestamp.After(last30Days) {
			metrics.Last30Days++
		}

		// Calculate duration (if available in details)
		// Note: Duration would need to be stored in transaction record
	}

	// Calculate success rate
	if metrics.Total > 0 {
		metrics.SuccessRate = float64(metrics.Valid) / float64(metrics.Total) * 100
	}

	// Calculate average duration
	if durationCount > 0 {
		metrics.AverageDuration = float64(totalDuration) / float64(durationCount)
	}

	return metrics, nil
}

// GetBlockMetrics retrieves block metrics
func (s *Service) GetBlockMetrics(ctx context.Context, channelName string) (*BlockMetrics, error) {
	s.logger.Info("Getting block metrics",
		zap.String("channel", channelName),
	)

	// Get all blocks for the channel
	blocks, total, err := s.explorerService.ListBlocks(ctx, channelName, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocks: %w", err)
	}

	metrics := &BlockMetrics{
		Total:     total,
		ByChannel: make(map[string]int64),
	}

	now := time.Now()
	last24Hours := now.Add(-24 * time.Hour)
	last7Days := now.Add(-7 * 24 * time.Hour)

	var totalBlockTime int64
	var blockTimeCount int64
	var largestBlockTxCount uint64

	for _, block := range blocks {
		// Count by channel
		metrics.ByChannel[channelName]++

		// Count by time range
		blockTime, err := time.Parse(time.RFC3339, block.Timestamp)
		if err == nil {
			if blockTime.After(last24Hours) {
				metrics.Last24Hours++
			}
			if blockTime.After(last7Days) {
				metrics.Last7Days++
			}

			// Calculate block time (time between blocks)
			if blockTimeCount > 0 {
				// This would need previous block time
				// For now, we'll skip this calculation
			}
			blockTimeCount++
		}

		// Track largest block
		if uint64(block.TransactionCount) > largestBlockTxCount {
			largestBlockTxCount = uint64(block.TransactionCount)
		}
	}

	metrics.LargestBlock = largestBlockTxCount

	// Calculate average block time
	if blockTimeCount > 1 && totalBlockTime > 0 {
		metrics.AverageBlockTime = float64(totalBlockTime) / float64(blockTimeCount-1)
	}

	return metrics, nil
}

// GetPerformanceMetrics retrieves performance metrics from audit logs
func (s *Service) GetPerformanceMetrics(ctx context.Context, startTime, endTime *time.Time) (*PerformanceMetrics, error) {
	s.logger.Info("Getting performance metrics")

	// Default to last 24 hours if not specified
	if startTime == nil {
		last24Hours := time.Now().Add(-24 * time.Hour)
		startTime = &last24Hours
	}
	if endTime == nil {
		now := time.Now()
		endTime = &now
	}

	// Get audit logs for the time range
	logs, _, err := s.auditService.ListLogsByDateRange(ctx, *startTime, *endTime, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	metrics := &PerformanceMetrics{
		ByEndpoint: make(map[string]int64),
		ByStatus:   make(map[string]int64),
	}

	var totalDuration int64
	var durationCount int64
	var durations []int64

	for _, log := range logs {
		metrics.TotalRequests++

		// Count by status
		metrics.ByStatus[log.Status]++
		if log.Status == "OK" || log.Status == "200" {
			metrics.SuccessfulRequests++
		} else {
			metrics.FailedRequests++
		}

		// Extract duration from details
		if len(log.Details) > 0 {
			// Parse details JSON to get duration
			// This would require unmarshaling the details JSON
			// For now, we'll skip this
		}

		// Extract endpoint from action or path in details
		// This would require parsing the details JSON
	}

	// Calculate error rate
	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
	}

	// Calculate average response time
	if durationCount > 0 {
		metrics.AverageResponseTime = float64(totalDuration) / float64(durationCount)
	}

	// Calculate P95 and P99 (would need sorted durations)
	if len(durations) > 0 {
		// Sort durations and calculate percentiles
		// For now, we'll skip this
	}

	// Calculate requests per second
	duration := endTime.Sub(*startTime).Seconds()
	if duration > 0 {
		metrics.RequestsPerSecond = float64(metrics.TotalRequests) / duration
	}

	return metrics, nil
}

// GetPeerMetrics retrieves peer metrics
// Note: This is a placeholder as we don't have peer discovery yet
func (s *Service) GetPeerMetrics(ctx context.Context) (*PeerMetrics, error) {
	s.logger.Info("Getting peer metrics")

	// Placeholder implementation
	// In the future, this would query peer status from Fabric network
	metrics := &PeerMetrics{
		TotalPeers:    0,
		ActivePeers:   0,
		InactivePeers: 0,
		ByChannel:     make(map[string]int64),
		ByMSP:         make(map[string]int64),
	}

	return metrics, nil
}

// GetMetricsSummary retrieves overall metrics summary
func (s *Service) GetMetricsSummary(ctx context.Context, channelName string) (*MetricsSummary, error) {
	s.logger.Info("Getting metrics summary",
		zap.String("channel", channelName),
	)

	// Get all metrics
	txMetrics, err := s.GetTransactionMetrics(ctx, channelName, nil, nil)
	if err != nil {
		s.logger.Warn("Failed to get transaction metrics", zap.Error(err))
		txMetrics = &TransactionMetrics{
			ByChannel:   make(map[string]int64),
			ByChaincode: make(map[string]int64),
			ByStatus:    make(map[string]int64),
		}
	}

	blockMetrics, err := s.GetBlockMetrics(ctx, channelName)
	if err != nil {
		s.logger.Warn("Failed to get block metrics", zap.Error(err))
		blockMetrics = &BlockMetrics{
			ByChannel: make(map[string]int64),
		}
	}

	perfMetrics, err := s.GetPerformanceMetrics(ctx, nil, nil)
	if err != nil {
		s.logger.Warn("Failed to get performance metrics", zap.Error(err))
		perfMetrics = &PerformanceMetrics{
			ByEndpoint: make(map[string]int64),
			ByStatus:   make(map[string]int64),
		}
	}

	peerMetrics, err := s.GetPeerMetrics(ctx)
	if err != nil {
		s.logger.Warn("Failed to get peer metrics", zap.Error(err))
		peerMetrics = &PeerMetrics{
			ByChannel: make(map[string]int64),
			ByMSP:     make(map[string]int64),
		}
	}

	return &MetricsSummary{
		Transactions: *txMetrics,
		Blocks:       *blockMetrics,
		Performance:  *perfMetrics,
		Peers:        *peerMetrics,
		Timestamp:    time.Now(),
	}, nil
}

