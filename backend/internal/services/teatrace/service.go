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

package teatrace

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"github.com/ibn-network/backend/internal/infrastructure/gateway"
	"github.com/ibn-network/backend/internal/services/analytics/metrics"
	"go.uber.org/zap"
)

// Service handles Tea Traceability operations
type Service struct {
	gatewayClient *gateway.Client
	metrics       *metrics.Service
	logger        *zap.Logger
}

// NewService creates a new Tea Traceability service
func NewService(gatewayClient *gateway.Client, metrics *metrics.Service, logger *zap.Logger) *Service {
	return &Service{
		gatewayClient: gatewayClient,
		metrics:       metrics,
		logger:        logger,
	}
}

// VerifyByHash verifies an entity by its hash via Gateway
func (s *Service) VerifyByHash(ctx context.Context, hash string) (map[string]interface{}, error) {
	start := time.Now()
	var err error
	
	// Record metric on exit
	defer func() {
		s.metrics.RecordBlockchainTransaction("verify_by_hash", time.Since(start), err == nil)
	}()

	reqBody := map[string]string{
		"hash": hash,
	}

	respBody, err := s.gatewayClient.Post(ctx, "/api/v1/teatrace/verify-by-hash", reqBody)
	if err != nil {
		s.logger.Error("Failed to verify by hash via Gateway", zap.Error(err))
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	var apiResp struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
		Error   interface{}            `json:"error"`
	}

	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		err = fmt.Errorf("verification failed: %v", apiResp.Error)
		return nil, err
	}

	return apiResp.Data, nil
}
