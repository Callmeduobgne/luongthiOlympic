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

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Container string `json:"container"`
	Message   string `json:"message"`
	Raw       string `json:"raw"`
}

// LogsService handles querying logs from Loki
type LogsService struct {
	lokiURL  string
	client   *http.Client
	logger   *zap.Logger
}

// NewLogsService creates a new logs service
func NewLogsService(lokiURL string, logger *zap.Logger) *LogsService {
	if lokiURL == "" {
		lokiURL = "http://loki:3100"
	}

	return &LogsService{
		lokiURL: lokiURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// QueryLogs queries logs from Loki with filters
func (s *LogsService) QueryLogs(ctx context.Context, params LogQueryParams) ([]LogEntry, error) {
	// Build Loki query
	query := s.buildLokiQuery(params)
	
	// Build URL
	queryURL := fmt.Sprintf("%s/loki/api/v1/query_range", s.lokiURL)
	reqURL, err := url.Parse(queryURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Loki URL: %w", err)
	}

	// Set query parameters
	q := reqURL.Query()
	q.Set("query", query)
	q.Set("limit", strconv.Itoa(params.Limit))
	
	// Time range: last 1 hour by default
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)
	if params.Since != "" {
		if duration, err := time.ParseDuration(params.Since); err == nil {
			startTime = endTime.Add(-duration)
		}
	}
	
	q.Set("start", strconv.FormatInt(startTime.UnixNano(), 10))
	q.Set("end", strconv.FormatInt(endTime.UnixNano(), 10))
	reqURL.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Warn("Failed to query Loki", 
			zap.Error(err),
			zap.String("loki_url", s.lokiURL),
		)
		// Return empty logs instead of error to avoid breaking UI
		// This allows the UI to continue working even if Loki is unavailable
		return []LogEntry{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Warn("Loki returned error", 
			zap.Int("status", resp.StatusCode),
			zap.String("url", reqURL.String()),
			zap.String("body", string(body)),
		)
		// Return empty logs instead of error to avoid breaking UI
		return []LogEntry{}, nil
	}

	// Parse response
	var lokiResp LokiQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&lokiResp); err != nil {
		s.logger.Warn("Failed to parse Loki response", zap.Error(err))
		return []LogEntry{}, nil
	}

	// Convert to LogEntry
	logs := s.parseLokiResponse(lokiResp, params)
	return logs, nil
}

// LogQueryParams represents query parameters for logs
type LogQueryParams struct {
	Container string // Filter by container name (peer, orderer)
	Level     string // Filter by log level (error, warn, info, debug)
	Limit     int    // Max number of logs (default: 500)
	Since     string // Time range (e.g., "5m", "1h", "30m")
}

// buildLokiQuery builds a LogQL query string
func (s *LogsService) buildLokiQuery(params LogQueryParams) string {
	// Base query: all logs from Docker containers
	// Promtail labels containers with job=peer|orderer|couchdb
	query := `{job=~"peer|orderer|couchdb"}`

	// Add container filter (match container name in labels)
	if params.Container != "" && params.Container != "all" {
		// Match container name in either container label or job label
		query = fmt.Sprintf(`{job=~"peer|orderer|couchdb",container=~".*%s.*"}`, params.Container)
	}

	// Add level filter (match in log line content)
	if params.Level != "" && params.Level != "all" {
		// Use case-insensitive match for log level
		levelPattern := strings.ToUpper(params.Level)
		query = fmt.Sprintf(`%s |= "%s"`, query, levelPattern)
	}

	// Log query for debugging
	s.logger.Debug("Built Loki query", zap.String("query", query))

	return query
}

// parseLokiResponse parses Loki response to LogEntry slice
func (s *LogsService) parseLokiResponse(resp LokiQueryResponse, params LogQueryParams) []LogEntry {
	var logs []LogEntry

	for _, stream := range resp.Data.Result {
		container := "unknown"
		if len(stream.Stream) > 0 {
			// Extract container name from labels
			if c, ok := stream.Stream["container"]; ok {
				container = c
			} else if c, ok := stream.Stream["job"]; ok {
				container = c
			}
		}

		for _, entry := range stream.Values {
			if len(entry) < 2 {
				continue
			}

			timestamp := entry[0]
			message := entry[1]

			// Parse log level from message
			level := s.extractLogLevel(message)

			// Apply level filter
			if params.Level != "" && params.Level != "all" && level != params.Level {
				continue
			}

			// Format timestamp
			ts, err := strconv.ParseInt(timestamp, 10, 64)
			if err == nil {
				t := time.Unix(0, ts)
				timestamp = t.Format("2006-01-02 15:04:05.000")
			}

			logs = append(logs, LogEntry{
				Timestamp: timestamp,
				Level:     level,
				Container: container,
				Message:   message,
				Raw:       message,
			})
		}
	}

	// Reverse to show chronological order (oldest first, newest last)
	// Loki returns newest first, so we reverse for better UX
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs
}

// extractLogLevel extracts log level from message
func (s *LogsService) extractLogLevel(message string) string {
	msg := strings.ToLower(message)
	if strings.Contains(msg, "error") || strings.Contains(msg, "err") {
		return "error"
	}
	if strings.Contains(msg, "warn") || strings.Contains(msg, "warning") {
		return "warn"
	}
	if strings.Contains(msg, "debug") {
		return "debug"
	}
	return "info"
}

// LokiQueryResponse represents Loki API response
type LokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

