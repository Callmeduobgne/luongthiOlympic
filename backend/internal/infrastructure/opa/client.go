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

package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client handles communication with OPA server
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient creates a new OPA client
func NewClient(baseURL string, logger *zap.Logger) *Client {
	if baseURL == "" {
		baseURL = "http://opa:8181"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// EvaluateRequest represents an authorization request
type EvaluateRequest struct {
	User      UserInfo      `json:"user"`
	Request   RequestInfo   `json:"request"`
	Resource  ResourceInfo  `json:"resource,omitempty"`
	Environment EnvironmentInfo `json:"environment,omitempty"`
}

// UserInfo contains user information
type UserInfo struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email,omitempty"`
	Roles       []string               `json:"roles"`
	Permissions []PermissionInfo       `json:"permissions"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// PermissionInfo represents a permission
type PermissionInfo struct {
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	Effect     string                 `json:"effect"`
	Scope      string                 `json:"scope,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

// RequestInfo contains request information
type RequestInfo struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Scope    string `json:"scope,omitempty"`
}

// ResourceInfo contains resource attributes for ABAC
type ResourceInfo struct {
	ID         string                 `json:"id,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// EnvironmentInfo contains environment attributes for ABAC
type EnvironmentInfo struct {
	IPAddress string                 `json:"ip_address,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// EvaluateResponse represents OPA evaluation response
type EvaluateResponse struct {
	Result bool `json:"result"`
}

// Evaluate evaluates an authorization request using OPA
func (c *Client) Evaluate(ctx context.Context, req *EvaluateRequest) (bool, error) {
	// Build OPA input
	input := map[string]interface{}{
		"user":        req.User,
		"request":     req.Request,
		"resource":    req.Resource,
		"environment": req.Environment,
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"input": input,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to OPA
	url := fmt.Sprintf("%s/v1/data/authz/allow", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to call OPA: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("OPA returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var opaResp EvaluateResponse
	if err := json.NewDecoder(resp.Body).Decode(&opaResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return opaResp.Result, nil
}

// Health checks OPA server health
func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call OPA: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OPA health check failed with status %d", resp.StatusCode)
	}

	return nil
}

