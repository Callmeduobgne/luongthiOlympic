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

package ca

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ibn-network/api-gateway/internal/config"
	"go.uber.org/zap"
)

// Client provides Fabric CA client functionality
type Client struct {
	url        string
	caName     string
	adminUser  string
	adminPass  string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient creates a new Fabric CA client
func NewClient(cfg *config.CAConfig, logger *zap.Logger) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("CA URL is required")
	}

	// Setup HTTP client with TLS
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	var tlsConfig tls.Config
	if cfg.TLSCertPath != "" {
		caCert, err := os.ReadFile(cfg.TLSCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA TLS certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA TLS certificate")
		}
		tlsConfig.RootCAs = caCertPool
	} else {
		// Development fallback (not recommended)
		tlsConfig.InsecureSkipVerify = true
	}

	if cfg.MSPDir != "" {
		adminUser := cfg.AdminUser
		if adminUser == "" {
			adminUser = "admin"
		}
		if certPEM, keyPEM, err := loadIdentityCredentials(cfg.MSPDir, adminCandidatePaths(cfg.MSPDir, adminUser, cfg.MSPID)); err == nil {
			if tlsCert, err := tls.X509KeyPair(certPEM, keyPEM); err == nil {
				tlsConfig.Certificates = []tls.Certificate{tlsCert}
			} else {
				logger.Warn("Failed to parse admin TLS certificate", zap.Error(err))
			}
		} else {
			logger.Warn("Failed to load admin TLS credentials", zap.Error(err))
		}
	}

	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tlsConfig,
	}

	return &Client{
		url:        cfg.URL,
		caName:     cfg.CAName,
		adminUser:  cfg.AdminUser,
		adminPass:  cfg.AdminPass,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// EnrollRequest represents an enrollment request
type EnrollRequest struct {
	Name     string   `json:"name"`
	Secret   string   `json:"secret"`
	Profile  string   `json:"profile,omitempty"`
	Label    string   `json:"label,omitempty"`
	AttrReqs []string `json:"attr_reqs,omitempty"`
}

// EnrollResponse represents an enrollment response
type EnrollResponse struct {
	Result struct {
		Cert       string `json:"Cert"`
		ServerInfo struct {
			CAName  string `json:"CAName"`
			CAChain string `json:"CAChain"`
		} `json:"ServerInfo"`
		Version string `json:"Version"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// Enroll enrolls a user with Fabric CA
func (c *Client) Enroll(req *EnrollRequest) (*EnrollResponse, error) {
	url := fmt.Sprintf("%s/api/v1/enroll", c.url)
	if c.caName != "" {
		url = fmt.Sprintf("%s/api/v1/enroll?ca=%s", c.url, c.caName)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enroll request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.adminUser != "" && c.adminPass != "" {
		httpReq.SetBasicAuth(c.adminUser, c.adminPass)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send enroll request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Success bool `json:"success"`
			Errors  []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && len(errorResp.Errors) > 0 {
			return nil, fmt.Errorf("enroll failed: %s", errorResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("enroll failed with status %d: %s", resp.StatusCode, string(body))
	}

	var enrollResp EnrollResponse
	if err := json.Unmarshal(body, &enrollResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal enroll response: %w", err)
	}

	if !enrollResp.Success {
		if len(enrollResp.Errors) > 0 {
			return nil, fmt.Errorf("enroll failed: %s", enrollResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("enroll failed: unknown error")
	}

	return &enrollResp, nil
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Name           string            `json:"name"`
	Secret         string            `json:"secret,omitempty"`
	Type           string            `json:"type,omitempty"`
	Affiliation    string            `json:"affiliation,omitempty"`
	Attributes     map[string]string `json:"attrs,omitempty"`
	MaxEnrollments int               `json:"max_enrollments,omitempty"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	Result struct {
		Secret string `json:"secret"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// Register registers a new identity with Fabric CA using mutual TLS
func (c *Client) Register(req *RegisterRequest) (*RegisterResponse, error) {
	url := fmt.Sprintf("%s/api/v1/register", c.url)
	if c.caName != "" {
		url = fmt.Sprintf("%s/api/v1/register?ca=%s", c.url, c.caName)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal register request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.adminUser != "" && c.adminPass != "" {
		httpReq.SetBasicAuth(c.adminUser, c.adminPass)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send register request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Success bool `json:"success"`
			Errors  []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && len(errorResp.Errors) > 0 {
			return nil, fmt.Errorf("register failed: %s", errorResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("register failed with status %d: %s", resp.StatusCode, string(body))
	}

	var registerResp RegisterResponse
	if err := json.Unmarshal(body, &registerResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal register response: %w", err)
	}

	if !registerResp.Success {
		if len(registerResp.Errors) > 0 {
			return nil, fmt.Errorf("register failed: %s", registerResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("register failed: unknown error")
	}

	return &registerResp, nil
}

// ReenrollRequest represents a re-enrollment request
type ReenrollRequest struct {
	Label string `json:"label,omitempty"`
}

// ReenrollResponse represents a re-enrollment response
type ReenrollResponse struct {
	Result struct {
		Cert       string `json:"Cert"`
		ServerInfo struct {
			CAName  string `json:"CAName"`
			CAChain string `json:"CAChain"`
		} `json:"ServerInfo"`
		Version string `json:"Version"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// Reenroll re-enrolls a user (renews certificate)
func (c *Client) Reenroll(req *ReenrollRequest, cert, key []byte) (*ReenrollResponse, error) {
	url := fmt.Sprintf("%s/api/v1/reenroll", c.url)
	if c.caName != "" {
		url = fmt.Sprintf("%s/api/v1/reenroll?ca=%s", c.url, c.caName)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reenroll request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send reenroll request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Success bool `json:"success"`
			Errors  []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && len(errorResp.Errors) > 0 {
			return nil, fmt.Errorf("reenroll failed: %s", errorResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("reenroll failed with status %d: %s", resp.StatusCode, string(body))
	}

	var reenrollResp ReenrollResponse
	if err := json.Unmarshal(body, &reenrollResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reenroll response: %w", err)
	}

	if !reenrollResp.Success {
		if len(reenrollResp.Errors) > 0 {
			return nil, fmt.Errorf("reenroll failed: %s", reenrollResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("reenroll failed: unknown error")
	}

	return &reenrollResp, nil
}

// RevokeRequest represents a revocation request
type RevokeRequest struct {
	Name   string `json:"name,omitempty"`
	Serial string `json:"serial,omitempty"`
	AKI    string `json:"aki,omitempty"`
	Reason string `json:"reason,omitempty"`
	GenCRL bool   `json:"gencrl,omitempty"`
}

// RevokeResponse represents a revocation response
type RevokeResponse struct {
	Result struct {
		RevokedCerts []struct {
			Serial string `json:"Serial"`
			AKI    string `json:"AKI"`
		} `json:"RevokedCerts"`
		CRL string `json:"CRL,omitempty"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// Revoke revokes a certificate using admin mutual TLS identity
func (c *Client) Revoke(req *RevokeRequest) (*RevokeResponse, error) {
	url := fmt.Sprintf("%s/api/v1/revoke", c.url)
	if c.caName != "" {
		url = fmt.Sprintf("%s/api/v1/revoke?ca=%s", c.url, c.caName)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal revoke request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send revoke request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Success bool `json:"success"`
			Errors  []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && len(errorResp.Errors) > 0 {
			return nil, fmt.Errorf("revoke failed: %s", errorResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("revoke failed with status %d: %s", resp.StatusCode, string(body))
	}

	var revokeResp RevokeResponse
	if err := json.Unmarshal(body, &revokeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal revoke response: %w", err)
	}

	if !revokeResp.Success {
		if len(revokeResp.Errors) > 0 {
			return nil, fmt.Errorf("revoke failed: %s", revokeResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("revoke failed: unknown error")
	}

	return &revokeResp, nil
}
