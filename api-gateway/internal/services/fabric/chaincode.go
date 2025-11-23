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

package fabric

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/ibn-network/api-gateway/internal/models"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ChaincodeService provides generic chaincode operations
type ChaincodeService struct {
	gateway *GatewayService
}

// NewChaincodeService creates a new chaincode service
func NewChaincodeService(gateway *GatewayService) *ChaincodeService {
	return &ChaincodeService{
		gateway: gateway,
	}
}

// Invoke invokes a chaincode function (submit transaction)
func (s *ChaincodeService) Invoke(
	ctx context.Context,
	channelName string,
	chaincodeName string,
	req *models.InvokeRequest,
) (*models.InvokeResponse, error) {
	ctx, span := s.gateway.tracer.Start(ctx, "ChaincodeService.Invoke")
	defer span.End()

	span.SetAttributes(
		attribute.String("channel", channelName),
		attribute.String("chaincode", chaincodeName),
		attribute.String("function", req.Function),
		attribute.Int("args.count", len(req.Args)),
	)

	// Check if user cert is provided in context (from Backend)
	userCert := ""
	userKey := ""
	userMSPID := ""
	if cert, ok := ctx.Value("user_cert").(string); ok && cert != "" {
		userCert = cert
	}
	if key, ok := ctx.Value("user_key").(string); ok && key != "" {
		userKey = key
	}
	if msp, ok := ctx.Value("user_msp_id").(string); ok && msp != "" {
		userMSPID = msp
	}

	// Get network and contract
	// If user cert is provided, use dynamic gateway connection with user cert
	// Otherwise, use default gateway connection
	var network *client.Network
	var contract *client.Contract
	
	if userCert != "" && userKey != "" && userMSPID != "" {
		// Create dynamic gateway connection with user cert (forward cert to Fabric)
		dynamicGw, err := s.gateway.CreateDynamicGateway(ctx, userCert, userKey, userMSPID)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to create dynamic gateway with user cert: %w", err)
		}
		defer dynamicGw.Close()
		
		network = dynamicGw.GetNetwork(channelName)
		contract = network.GetContract(chaincodeName)
		
		s.gateway.logger.Info("Using dynamic gateway with user cert",
			zap.String("msp_id", userMSPID),
			zap.String("channel", channelName),
		)
	} else {
		// Use default gateway connection (no user cert provided)
		gatewayClient := s.gateway.GetGatewayClient()
		network = gatewayClient.GetNetwork(channelName)
		contract = network.GetContract(chaincodeName)
	}

	// Prepare transient data if provided
	var proposal *client.Proposal
	var err error

	if len(req.Transient) > 0 {
		// Decode base64 transient data
		transientMap := make(map[string][]byte)
		for k, v := range req.Transient {
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				span.RecordError(err)
				return nil, fmt.Errorf("failed to decode transient data for key '%s': %w", k, err)
			}
			transientMap[k] = decoded
		}

		// Create proposal with transient data
		proposal, err = contract.NewProposal(
			req.Function,
			client.WithArguments(req.Args...),
			client.WithTransient(transientMap),
		)
	} else {
		// Create proposal without transient data
		proposal, err = contract.NewProposal(
			req.Function,
			client.WithArguments(req.Args...),
		)
	}

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	// Endorse proposal
	endorsedProposal, err := proposal.Endorse()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to endorse proposal: %w", err)
	}

	// Get result from endorsed proposal (before submit)
	result := endorsedProposal.Result()

	// Submit transaction
	transaction, err := endorsedProposal.Submit()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	// Get transaction ID
	txID := transaction.TransactionID()

	// Get commit status
	commitStatus, err := transaction.Status()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get commit status: %w", err)
	}

	// Parse result as JSON if possible
	var resultData interface{}
	if len(result) > 0 {
		if err := json.Unmarshal(result, &resultData); err != nil {
			// If not JSON, return as string
			resultData = string(result)
		}
	}

	// Build response
	response := &models.InvokeResponse{
		TxID:        txID,
		Result:      resultData,
		Status:      "SUBMITTED",
		BlockNumber: commitStatus.BlockNumber,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	// Update status based on commit status
	if commitStatus.Successful {
		response.Status = "VALID"
	} else {
		response.Status = "INVALID"
	}

	s.gateway.logger.Info("Chaincode invoked successfully",
		zap.String("channel", channelName),
		zap.String("chaincode", chaincodeName),
		zap.String("function", req.Function),
		zap.String("txId", txID),
		zap.Uint64("blockNumber", commitStatus.BlockNumber),
	)

	return response, nil
}

// Query queries a chaincode function (evaluate transaction - read-only)
func (s *ChaincodeService) Query(
	ctx context.Context,
	channelName string,
	chaincodeName string,
	req *models.QueryRequest,
) (*models.QueryResponse, error) {
	ctx, span := s.gateway.tracer.Start(ctx, "ChaincodeService.Query")
	defer span.End()

	span.SetAttributes(
		attribute.String("channel", channelName),
		attribute.String("chaincode", chaincodeName),
		attribute.String("function", req.Function),
		attribute.Int("args.count", len(req.Args)),
	)

	// Use existing EvaluateTransaction method
	// For now, use the existing method (transient data support can be added later)
	result, err := s.gateway.EvaluateTransaction(ctx, req.Function, req.Args...)

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to evaluate transaction: %w", err)
	}

	// Parse result as JSON if possible
	var resultData interface{}
	if len(result) > 0 {
		if err := json.Unmarshal(result, &resultData); err != nil {
			// If not JSON, return as string
			resultData = string(result)
		}
	}

	response := &models.QueryResponse{
		Result:    resultData,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.gateway.logger.Info("Chaincode queried successfully",
		zap.String("channel", channelName),
		zap.String("chaincode", chaincodeName),
		zap.String("function", req.Function),
	)

	return response, nil
}

