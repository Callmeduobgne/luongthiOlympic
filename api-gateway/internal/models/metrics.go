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

package models

import "time"

// TransactionMetricsResponse represents transaction metrics in API response
type TransactionMetricsResponse struct {
	Total           int64            `json:"total"`
	Valid           int64            `json:"valid"`
	Invalid         int64            `json:"invalid"`
	Submitted       int64            `json:"submitted"`
	SuccessRate     float64          `json:"successRate"`
	AverageDuration float64          `json:"averageDuration"`
	ByChannel       map[string]int64 `json:"byChannel"`
	ByChaincode     map[string]int64 `json:"byChaincode"`
	ByStatus        map[string]int64 `json:"byStatus"`
	Last24Hours     int64            `json:"last24Hours"`
	Last7Days       int64            `json:"last7Days"`
	Last30Days      int64            `json:"last30Days"`
}

// BlockMetricsResponse represents block metrics in API response
type BlockMetricsResponse struct {
	Total           int64            `json:"total"`
	Last24Hours     int64            `json:"last24Hours"`
	Last7Days       int64            `json:"last7Days"`
	AverageBlockTime float64          `json:"averageBlockTime"`
	LargestBlock    uint64           `json:"largestBlock"`
	ByChannel       map[string]int64 `json:"byChannel"`
}

// PerformanceMetricsResponse represents performance metrics in API response
type PerformanceMetricsResponse struct {
	AverageResponseTime float64          `json:"averageResponseTime"`
	P95ResponseTime    float64          `json:"p95ResponseTime"`
	P99ResponseTime    float64          `json:"p99ResponseTime"`
	RequestsPerSecond  float64          `json:"requestsPerSecond"`
	ErrorRate          float64          `json:"errorRate"`
	TotalRequests      int64            `json:"totalRequests"`
	SuccessfulRequests int64            `json:"successfulRequests"`
	FailedRequests     int64            `json:"failedRequests"`
	ByEndpoint         map[string]int64 `json:"byEndpoint"`
	ByStatus           map[string]int64 `json:"byStatus"`
}

// PeerMetricsResponse represents peer metrics in API response
type PeerMetricsResponse struct {
	TotalPeers    int64            `json:"totalPeers"`
	ActivePeers   int64            `json:"activePeers"`
	InactivePeers int64            `json:"inactivePeers"`
	ByChannel     map[string]int64 `json:"byChannel"`
	ByMSP         map[string]int64 `json:"byMSP"`
}

// MetricsSummaryResponse represents overall metrics summary in API response
type MetricsSummaryResponse struct {
	Transactions TransactionMetricsResponse `json:"transactions"`
	Blocks       BlockMetricsResponse        `json:"blocks"`
	Performance  PerformanceMetricsResponse `json:"performance"`
	Peers        PeerMetricsResponse         `json:"peers"`
	Timestamp    time.Time                   `json:"timestamp"`
}

