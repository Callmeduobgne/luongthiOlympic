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

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Client is a client for Admin Service API
type Client struct {
	baseURL        string
	apiKey         string
	httpClient     *http.Client
	logger         *zap.Logger
	circuitBreaker *CircuitBreaker
	retryConfig    RetryConfig
}

// NewClient creates a new Admin Service client
func NewClient(baseURL, apiKey string, logger *zap.Logger) *Client {
	// Default circuit breaker config (can be overridden via config)
	cb := NewCircuitBreaker(
		3,                    // maxRequests
		10*time.Second,       // interval
		60*time.Second,       // timeout
		0.6,                  // failureRatio
		logger,
	)

	return &Client{
		baseURL:        baseURL,
		apiKey:         apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger:         logger,
		circuitBreaker: cb,
		retryConfig:    DefaultRetryConfig(),
	}
}

// NewClientWithConfig creates a new Admin Service client with custom circuit breaker config
func NewClientWithConfig(baseURL, apiKey string, maxRequests uint32, interval, timeout time.Duration, failureRatio float64, logger *zap.Logger) *Client {
	cb := NewCircuitBreaker(maxRequests, interval, timeout, failureRatio, logger)

	return &Client{
		baseURL:        baseURL,
		apiKey:         apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger:         logger,
		circuitBreaker: cb,
		retryConfig:    DefaultRetryConfig(),
	}
}

// Health checks Admin Service health
func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("admin service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("admin service health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// APIResponse represents Admin Service API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents Admin Service API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

// doRequest performs HTTP request to Admin Service with circuit breaker and retry
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*APIResponse, error) {
	var result *APIResponse
	var resultErr error

	// Execute with circuit breaker and retry
	err := c.circuitBreaker.Execute(ctx, func() error {
		// Retry with exponential backoff
		retryErr := RetryWithBackoff(ctx, c.retryConfig, c.logger, func() error {
			resp, err := c.executeRequest(ctx, method, path, body)
			if err != nil {
				return err
			}
			result = resp
			return nil
		})
		if retryErr != nil {
			resultErr = retryErr
			return retryErr
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if resultErr != nil {
		return nil, resultErr
	}

	return result, nil
}

// executeRequest performs the actual HTTP request (without circuit breaker/retry)
func (c *Client) executeRequest(ctx context.Context, method, path string, body interface{}) (*APIResponse, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	contentType := "application/json"
	
	if body != nil {
		// Check if body is multipart form data (for file uploads)
		if formData, ok := body.(*bytes.Buffer); ok {
			reqBody = formData
			contentType = "" // Let http library set it with boundary
		} else {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("admin service API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(respBody))
	}

	if !apiResp.Success || resp.StatusCode >= 400 {
		return nil, c.formatErrorResponse(resp.StatusCode, apiResp.Error)
	}

	return &apiResp, nil
}

// doRequestWithTimeout performs HTTP request with custom timeout (with circuit breaker and retry)
func (c *Client) doRequestWithTimeout(ctx context.Context, method, path string, body interface{}, timeout time.Duration) (*APIResponse, error) {
	var result *APIResponse
	var resultErr error

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute with circuit breaker and retry
	err := c.circuitBreaker.Execute(timeoutCtx, func() error {
		// Retry with exponential backoff
		retryErr := RetryWithBackoff(timeoutCtx, c.retryConfig, c.logger, func() error {
			// Create temporary HTTP client with custom timeout
			tempClient := &http.Client{
				Timeout: timeout,
			}

			resp, err := c.executeRequestWithClient(timeoutCtx, tempClient, method, path, body)
			if err != nil {
				return err
			}
			result = resp
			return nil
		})
		if retryErr != nil {
			resultErr = retryErr
			return retryErr
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if resultErr != nil {
		return nil, resultErr
	}

	return result, nil
}

// executeRequestWithClient performs the actual HTTP request with a custom HTTP client
func (c *Client) executeRequestWithClient(ctx context.Context, httpClient *http.Client, method, path string, body interface{}) (*APIResponse, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	contentType := "application/json"
	
	if body != nil {
		// Check if body is multipart form data (for file uploads)
		if formData, ok := body.(*bytes.Buffer); ok {
			reqBody = formData
			contentType = "" // Let http library set it with boundary
		} else {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("admin service API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(respBody))
	}

	if !apiResp.Success || resp.StatusCode >= 400 {
		return nil, c.formatErrorResponse(resp.StatusCode, apiResp.Error)
	}

	return &apiResp, nil
}

// formatErrorResponse formats error response into user-friendly message
func (c *Client) formatErrorResponse(statusCode int, apiErr *APIError) error {
	errMsg := "unknown error"
	errDetail := ""
	if apiErr != nil {
		errMsg = apiErr.Message
		if apiErr.Detail != "" {
			errDetail = apiErr.Detail
		}
	}
	
	// Combine message and detail for better error reporting
	if errDetail != "" {
		// Extract meaningful error from detail (e.g., peer command output)
		if strings.Contains(errDetail, "could not parse") || strings.Contains(errDetail, "tar entry") {
			// Chaincode package format error
			errMsg = "Invalid chaincode package format: " + extractPackageError(errDetail)
		} else if strings.Contains(errDetail, "peer command failed") {
			// Extract actual error from peer output
			errMsg = extractPeerError(errDetail)
		} else if strings.Contains(strings.ToLower(errDetail), "peer cli not available") {
			// User-friendly message for peer CLI unavailable
			errMsg = "Peer CLI is not available. Please ensure admin-service container has peer CLI installed and is running."
		} else if strings.Contains(strings.ToLower(errDetail), "connection refused") {
			// User-friendly message for connection errors
			errMsg = "Cannot connect to peer. Please check if peer is running and accessible."
		} else {
			errMsg = errMsg + ": " + errDetail
		}
	}
	
	// Map HTTP status codes to user-friendly messages
	statusMsg := ""
	switch statusCode {
	case http.StatusServiceUnavailable:
		statusMsg = "Admin Service is temporarily unavailable"
	case http.StatusBadRequest:
		statusMsg = "Invalid request"
	case http.StatusInternalServerError:
		statusMsg = "Internal server error"
	default:
		statusMsg = fmt.Sprintf("Request failed with status %d", statusCode)
	}
	
	return fmt.Errorf("%s: %s", statusMsg, errMsg)
}

// extractPackageError extracts meaningful error from chaincode package parsing errors
func extractPackageError(detail string) string {
	// Look for "could not parse" or "tar entry" errors
	if idx := strings.Index(detail, "could not parse"); idx != -1 {
		// Extract the error message after "could not parse"
		start := idx + len("could not parse")
		if end := strings.Index(detail[start:], "\n"); end != -1 {
			return strings.TrimSpace(detail[start : start+end])
		}
		return strings.TrimSpace(detail[start:])
	}
	if idx := strings.Index(detail, "tar entry"); idx != -1 {
		// Extract tar entry error
		if end := strings.Index(detail[idx:], "\n"); end != -1 {
			return strings.TrimSpace(detail[idx : idx+end])
		}
		return strings.TrimSpace(detail[idx:])
	}
	return detail
}

// extractPeerError extracts actual error from peer command output
func extractPeerError(detail string) string {
	// Look for "Error:" in peer output
	if idx := strings.Index(detail, "Error:"); idx != -1 {
		// Extract error message
		start := idx + len("Error:")
		if end := strings.Index(detail[start:], "\n"); end != -1 {
			errorMsg := strings.TrimSpace(detail[start : start+end])
			// Remove redundant prefixes
			errorMsg = strings.TrimPrefix(errorMsg, "chaincode install failed with status: 500 - ")
			errorMsg = strings.TrimPrefix(errorMsg, "failed to invoke backing implementation of 'InstallChaincode': ")
			return errorMsg
		}
		return strings.TrimSpace(detail[start:])
	}
	return detail
}

// UploadPackage uploads a chaincode package file
func (c *Client) UploadPackage(ctx context.Context, fileData []byte, filename string) (string, error) {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	part, err := writer.CreateFormFile("package", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	
	if _, err := part.Write(fileData); err != nil {
		return "", fmt.Errorf("failed to write file data: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request with multipart data
	url := fmt.Sprintf("%s/api/v1/chaincode/upload", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("admin service API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(respBody))
	}

	if !apiResp.Success || resp.StatusCode >= 400 {
		errMsg := "unknown error"
		if apiResp.Error != nil {
			errMsg = apiResp.Error.Message
		}
		return "", fmt.Errorf("admin service returned error (status %d): %s", resp.StatusCode, errMsg)
	}

	// Extract filePath from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	filePath, ok := data["filePath"].(string)
	if !ok {
		return "", fmt.Errorf("filePath not found in response")
	}

	return filePath, nil
}

// InstallChaincode installs a chaincode package
func (c *Client) InstallChaincode(ctx context.Context, packagePath, label string) (string, error) {
	req := map[string]interface{}{
		"packagePath": packagePath,
	}
	if label != "" {
		req["label"] = label
	}

	// Chaincode install can take a long time (peer CLI execution)
	// Create a context with extended timeout (5 minutes)
	installCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	resp, err := c.doRequestWithTimeout(installCtx, "POST", "/api/v1/chaincode/install", req, 5*time.Minute)
	if err != nil {
		return "", err
	}

	// Extract packageId from response
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	packageID, ok := data["packageId"].(string)
	if !ok {
		return "", fmt.Errorf("packageId not found in response")
	}

	return packageID, nil
}

// ApproveChaincode approves a chaincode definition
func (c *Client) ApproveChaincode(ctx context.Context, req *ApproveChaincodeRequest) error {
	_, err := c.doRequest(ctx, "POST", "/api/v1/chaincode/approve", req)
	return err
}

// CommitChaincode commits a chaincode definition
func (c *Client) CommitChaincode(ctx context.Context, req *CommitChaincodeRequest) error {
	_, err := c.doRequest(ctx, "POST", "/api/v1/chaincode/commit", req)
	return err
}

// ListInstalled lists installed chaincodes
func (c *Client) ListInstalled(ctx context.Context, peer string) ([]InstalledChaincode, error) {
	path := "/api/v1/chaincode/installed"
	if peer != "" {
		path += "?peer=" + peer
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		c.logger.Error("Failed to call admin service",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, err
	}

	// Log response for debugging
	c.logger.Info("Admin service response",
		zap.String("path", path),
		zap.Any("data_type", fmt.Sprintf("%T", resp.Data)),
		zap.Any("data", resp.Data),
	)

	// Convert response data to InstalledChaincode slice
	data, ok := resp.Data.([]interface{})
	if !ok {
		c.logger.Warn("Admin service returned unexpected data type",
			zap.String("path", path),
			zap.String("expected", "[]interface{}"),
			zap.String("got", fmt.Sprintf("%T", resp.Data)),
			zap.Any("data", resp.Data),
		)
		return []InstalledChaincode{}, nil
	}

	c.logger.Info("Parsing installed chaincodes",
		zap.Int("count", len(data)),
	)

	chaincodes := make([]InstalledChaincode, 0, len(data))
	for i, item := range data {
		itemJSON, _ := json.Marshal(item)
		var cc InstalledChaincode
		if err := json.Unmarshal(itemJSON, &cc); err != nil {
			c.logger.Warn("Failed to unmarshal chaincode item",
				zap.Int("index", i),
				zap.Error(err),
				zap.String("item", string(itemJSON)),
			)
			continue
		}
		chaincodes = append(chaincodes, cc)
	}

	c.logger.Info("Successfully parsed installed chaincodes",
		zap.Int("count", len(chaincodes)),
	)

	return chaincodes, nil
}

// ListCommitted lists committed chaincodes
func (c *Client) ListCommitted(ctx context.Context, channel string) ([]CommittedChaincode, error) {
	path := "/api/v1/chaincode/committed"
	if channel != "" {
		path += "?channel=" + channel
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	// Convert response data to CommittedChaincode slice
	data, ok := resp.Data.([]interface{})
	if !ok {
		return []CommittedChaincode{}, nil
	}

	chaincodes := make([]CommittedChaincode, 0, len(data))
	for _, item := range data {
		itemJSON, _ := json.Marshal(item)
		var cc CommittedChaincode
		if err := json.Unmarshal(itemJSON, &cc); err == nil {
			chaincodes = append(chaincodes, cc)
		}
	}

	return chaincodes, nil
}

// GetCommittedInfo gets information about a committed chaincode
func (c *Client) GetCommittedInfo(ctx context.Context, channel, name string) (*CommittedChaincode, error) {
	path := fmt.Sprintf("/api/v1/chaincode/committed/info?name=%s", name)
	if channel != "" {
		path += "&channel=" + channel
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	// Convert response data to CommittedChaincode
	itemJSON, _ := json.Marshal(resp.Data)
	var cc CommittedChaincode
	if err := json.Unmarshal(itemJSON, &cc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal committed chaincode: %w", err)
	}

	return &cc, nil
}

// Request types
type ApproveChaincodeRequest struct {
	ChannelName         string   `json:"channelName"`
	Name                string   `json:"name"`
	Version             string   `json:"version"`
	Sequence            int64    `json:"sequence"`
	PackageID           string   `json:"packageId,omitempty"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

type CommitChaincodeRequest struct {
	ChannelName         string   `json:"channelName"`
	Name                string   `json:"name"`
	Version             string   `json:"version"`
	Sequence            int64    `json:"sequence"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

// Response types
type InstalledChaincode struct {
	PackageID string        `json:"packageId"`
	Label     string        `json:"label"`
	Chaincode ChaincodeInfo `json:"chaincode"`
}

type ChaincodeInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

type CommittedChaincode struct {
	Name                 string   `json:"name"`
	Version              string   `json:"version"`
	Sequence             int64    `json:"sequence"`
	EndorsementPlugin    string   `json:"endorsementPlugin"`
	ValidationPlugin     string   `json:"validationPlugin"`
	InitRequired         bool     `json:"initRequired"`
	Collections          []string `json:"collections,omitempty"`
	ApprovedOrganizations []string `json:"approvedOrganizations"`
}

