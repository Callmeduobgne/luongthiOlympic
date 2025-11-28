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

package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ibn-network/backend/internal/services/blockchain/transaction"
	"go.uber.org/zap"
)

// Client wraps API Gateway HTTP client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string // Optional API key for service-to-service auth
	logger     *zap.Logger
}

// Config holds Gateway client configuration
type Config struct {
	BaseURL    string
	APIKey     string // Optional
	Timeout    time.Duration
	Logger     *zap.Logger
}

// NewClient creates a new Gateway client
func NewClient(cfg *Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		apiKey: cfg.APIKey,
		logger: cfg.Logger,
	}
}

// TransactionRequest represents Gateway transaction request (internal)
type TransactionRequest struct {
	ChannelName   string                 `json:"channelName"`
	ChaincodeName string                 `json:"chaincodeName"`
	FunctionName  string                 `json:"functionName"`
	Args          []string               `json:"args,omitempty"`
	TransientData map[string]interface{} `json:"transientData,omitempty"`
	EndorsingOrgs []string               `json:"endorsingOrgs,omitempty"`
}

// TransactionResponse represents Gateway transaction response (internal)
type TransactionResponse struct {
	ID          string    `json:"id"`
	TxID        string    `json:"txId"`
	Status      string    `json:"status"`
	BlockNumber uint64    `json:"blockNumber,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// APIResponse wraps Gateway API response
type APIResponse struct {
	Success bool                   `json:"success"`
	Data    *TransactionResponse   `json:"data,omitempty"`
	Error   *ErrorResponse         `json:"error,omitempty"`
}

// ErrorResponse represents Gateway error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SubmitTransaction submits a transaction via Gateway
// Implements transaction.GatewayClient interface
func (c *Client) SubmitTransaction(ctx context.Context, req *transaction.GatewayTransactionRequest) (*transaction.GatewayTransactionResponse, error) {
	// Convert to internal request format
	gatewayReq := &TransactionRequest{
		ChannelName:   req.ChannelName,
		ChaincodeName: req.ChaincodeName,
		FunctionName:  req.FunctionName,
		Args:          req.Args,
		TransientData: req.TransientData,
		EndorsingOrgs: req.EndorsingOrgs,
		// Note: UserCert, UserKey, MSPID are sent via headers, not in request body
	}

	url := fmt.Sprintf("%s/api/v1/transactions", c.baseURL)

	reqBody, err := json.Marshal(gatewayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	
	// Priority: API Key (from env) > JWT Token (from context)
	// API Key is preferred for service-to-service communication
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
		c.logger.Info("Using API key for Gateway authentication",
			zap.String("key_preview", c.apiKey[:20]+"..."),
			zap.String("url", url),
		)
	} else if token, ok := ctx.Value("jwt_token").(string); ok && token != "" {
		// Fallback to JWT token if API key is not available
		httpReq.Header.Set("Authorization", "Bearer "+token)
		c.logger.Debug("Using JWT token for Gateway authentication",
			zap.String("token_preview", token[:20]+"..."),
		)
	} else {
		c.logger.Warn("No authentication method available for Gateway request",
			zap.String("url", url),
		)
	}

	// Send user certificate to Gateway (required for Fabric authentication)
	// Get from request struct (passed from Backend service)
	if req.UserCert != "" {
		httpReq.Header.Set("X-User-Cert", req.UserCert)
	}
	if req.UserKey != "" {
		httpReq.Header.Set("X-User-Key", req.UserKey)
	}
	if req.MSPID != "" {
		httpReq.Header.Set("X-User-MSPID", req.MSPID)
	}

	c.logger.Debug("Calling Gateway API",
		zap.String("url", url),
		zap.String("channel", gatewayReq.ChannelName),
		zap.String("chaincode", gatewayReq.ChaincodeName),
		zap.String("function", gatewayReq.FunctionName),
	)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("Gateway API call failed", zap.Error(err))
		return nil, fmt.Errorf("gateway API call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var apiErr APIResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != nil {
			return nil, fmt.Errorf("gateway error: %s - %s", apiErr.Error.Code, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success || apiResp.Data == nil {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("gateway error: %s - %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("unexpected response format")
	}

	c.logger.Info("Transaction submitted via Gateway",
		zap.String("tx_id", apiResp.Data.TxID),
		zap.String("status", apiResp.Data.Status),
	)

	// Convert to transaction.GatewayTransactionResponse
	return &transaction.GatewayTransactionResponse{
		ID:          apiResp.Data.ID,
		TxID:        apiResp.Data.TxID,
		Status:      apiResp.Data.Status,
		BlockNumber: apiResp.Data.BlockNumber,
		Timestamp:   apiResp.Data.Timestamp,
	}, nil
}

// GetTransaction retrieves transaction from Gateway
// Implements transaction.GatewayClient interface
func (c *Client) GetTransaction(ctx context.Context, idOrTxID string) (*transaction.GatewayTransactionResponse, error) {
	url := fmt.Sprintf("%s/api/v1/transactions/%s", c.baseURL, idOrTxID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	if token, ok := ctx.Value("jwt_token").(string); ok && token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gateway API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("transaction not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success || apiResp.Data == nil {
		return nil, fmt.Errorf("unexpected response format")
	}

	// Convert to transaction.GatewayTransactionResponse
	return &transaction.GatewayTransactionResponse{
		ID:          apiResp.Data.ID,
		TxID:        apiResp.Data.TxID,
		Status:      apiResp.Data.Status,
		BlockNumber: apiResp.Data.BlockNumber,
		Timestamp:   apiResp.Data.Timestamp,
	}, nil
}

// QueryChaincode performs a read-only query via Gateway
// Implements transaction.GatewayClient interface
func (c *Client) QueryChaincode(ctx context.Context, channelName, chaincodeName, functionName string, args []string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/channels/%s/chaincodes/%s/query", c.baseURL, channelName, chaincodeName)

	reqBody := map[string]interface{}{
		"function": functionName,
		"args":     args,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// Priority: API Key (from env) > JWT Token (from context)
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	} else if token, ok := ctx.Value("jwt_token").(string); ok && token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	// Send user certificate to Gateway (for Fabric authentication with user identity)
	// This allows queries to use user's certificate instead of service account
	if userCert, ok := ctx.Value("user_cert").(string); ok && userCert != "" {
		httpReq.Header.Set("X-User-Cert", userCert)
	}
	if userKey, ok := ctx.Value("user_key").(string); ok && userKey != "" {
		httpReq.Header.Set("X-User-Key", userKey)
	}
	if mspID, ok := ctx.Value("user_msp_id").(string); ok && mspID != "" {
		httpReq.Header.Set("X-User-MSPID", mspID)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gateway API call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		// If not JSON, return raw body
		return body, nil
	}

	// Extract result from data field
	resultJSON, err := json.Marshal(apiResp.Data)
	if err != nil {
		return body, nil
	}

	return resultJSON, nil
}

// Post performs a generic POST request to Gateway
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Priority: API Key (from env) > JWT Token (from context)
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	} else if token, ok := ctx.Value("jwt_token").(string); ok && token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gateway API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gateway returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
