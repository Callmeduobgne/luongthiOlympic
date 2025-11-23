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
	"strings"
	"time"

	"go.uber.org/zap"
)

// LogEntry represents a log entry from Loki
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Container string `json:"container"`
	Message   string `json:"message"`
	Raw       string `json:"raw"`
}

// QueryLogsRequest represents a request to query logs
type QueryLogsRequest struct {
	Containers []string
	Since      string // e.g., "1h", "30m"
	Limit      int
	Search     string
}

// Service handles network log queries
type Service struct {
	lokiURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewService creates a new network logs service
func NewService(lokiURL string, logger *zap.Logger) *Service {
	return &Service{
		lokiURL: lokiURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// QueryLogs queries logs from Loki
func (s *Service) QueryLogs(ctx context.Context, req *QueryLogsRequest) ([]*LogEntry, error) {
	// Build Loki query
	query := buildLokiQuery(req)
	
	// Build URL
	u, err := url.Parse(s.lokiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Loki URL: %w", err)
	}
	u.Path = "/loki/api/v1/query_range"
	
	// Add query parameters
	params := url.Values{}
	params.Set("query", query)
	
	// Parse since duration
	sinceDuration := parseSince(req.Since)
	now := time.Now()
	start := now.Add(-sinceDuration)
	end := now
	
	params.Set("start", fmt.Sprintf("%d", start.UnixNano()))
	params.Set("end", fmt.Sprintf("%d", end.UnixNano()))
	params.Set("limit", fmt.Sprintf("%d", req.Limit))
	
	u.RawQuery = params.Encode()
	
	// Make request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	s.logger.Debug("Querying Loki",
		zap.String("url", u.String()),
		zap.String("query", query),
	)
	
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query Loki: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Loki returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var lokiResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Stream map[string]string `json:"stream"`
				Values [][]string        `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&lokiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Loki response: %w", err)
	}
	
	// Convert to LogEntry
	var logs []*LogEntry
	for _, result := range lokiResp.Data.Result {
		container := result.Stream["container"]
		if container == "" {
			container = result.Stream["job"] // Fallback to job label
		}
		
		for _, value := range result.Values {
			if len(value) < 2 {
				continue
			}
			
			timestamp := value[0]
			logLine := value[1]
			
			// Parse log line to extract level and message
			level, message := parseLogLine(logLine)
			
			logs = append(logs, &LogEntry{
				Timestamp: timestamp,
				Level:     level,
				Container: container,
				Message:   message,
				Raw:       logLine,
			})
		}
	}
	
	s.logger.Info("Queried logs from Loki",
		zap.Int("count", len(logs)),
		zap.String("query", query),
	)
	
	return logs, nil
}

// buildLokiQuery builds a LogQL query from the request
func buildLokiQuery(req *QueryLogsRequest) string {
	var parts []string
	
	// Container filter
	if len(req.Containers) > 0 {
		containerFilter := strings.Join(req.Containers, "|")
		parts = append(parts, fmt.Sprintf(`{container=~"%s"}`, containerFilter))
	} else {
		parts = append(parts, `{container=~".+"}`) // Match all containers
	}
	
	// Search filter
	if req.Search != "" {
		parts = append(parts, fmt.Sprintf(`|= "%s"`, req.Search))
	}
	
	return strings.Join(parts, " ")
}

// parseSince parses a duration string like "1h", "30m" into time.Duration
func parseSince(since string) time.Duration {
	if since == "" {
		return 1 * time.Hour // Default: 1 hour
	}
	
	duration, err := time.ParseDuration(since)
	if err != nil {
		return 1 * time.Hour // Default on error
	}
	
	return duration
}

// parseLogLine parses a log line to extract level and message
func parseLogLine(logLine string) (level, message string) {
	// Try to parse JSON log format
	var jsonLog struct {
		Level   string `json:"level"`
		Message string `json:"msg"`
		Log     string `json:"log"`
	}
	
	if err := json.Unmarshal([]byte(logLine), &jsonLog); err == nil {
		if jsonLog.Level != "" {
			level = strings.ToLower(jsonLog.Level)
		} else {
			level = "info"
		}
		
		if jsonLog.Message != "" {
			message = jsonLog.Message
		} else if jsonLog.Log != "" {
			message = jsonLog.Log
		} else {
			message = logLine
		}
		return
	}
	
	// Try to parse common log formats
	upperLog := strings.ToUpper(logLine)
	if strings.Contains(upperLog, "ERROR") || strings.Contains(upperLog, "[ERROR]") {
		level = "error"
	} else if strings.Contains(upperLog, "WARN") || strings.Contains(upperLog, "[WARN]") {
		level = "warn"
	} else if strings.Contains(upperLog, "DEBUG") || strings.Contains(upperLog, "[DEBUG]") {
		level = "debug"
	} else {
		level = "info"
	}
	
	message = logLine
	return
}

