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

package chaincode

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ibn-network/backend/internal/services/blockchain/transaction"
	"go.uber.org/zap"
)

// GatewayClient interface for Gateway operations
type GatewayClient interface {
	SubmitTransaction(ctx context.Context, req *transaction.GatewayTransactionRequest) (*transaction.GatewayTransactionResponse, error)
	QueryChaincode(ctx context.Context, channelName, chaincodeName, functionName string, args []string) ([]byte, error)
}

// TeaTraceServiceViaGateway provides methods to interact with teaTraceCC chaincode via Gateway
type TeaTraceServiceViaGateway struct {
	gatewayClient GatewayClient
	logger        *zap.Logger
	channel       string
}

// NewTeaTraceServiceViaGateway creates a new TeaTrace service using Gateway
// NOTE: Backend MUST use Gateway for all blockchain operations
func NewTeaTraceServiceViaGateway(gatewayClient GatewayClient, channel string, logger *zap.Logger) *TeaTraceServiceViaGateway {
	if gatewayClient == nil {
		logger.Fatal("Gateway client is required - Backend cannot connect directly to Fabric")
	}
	return &TeaTraceServiceViaGateway{
		gatewayClient: gatewayClient,
		logger:        logger,
		channel:       channel,
	}
}

// CreateBatch creates a new tea batch via Gateway
func (s *TeaTraceServiceViaGateway) CreateBatch(ctx context.Context, batchID, farmName, harvestDate, certification, certificateID string) (string, error) {
	args := []string{batchID, farmName, harvestDate, certification, certificateID}

	// Extract user cert from context (set by handler)
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

	req := &transaction.GatewayTransactionRequest{
		ChannelName:   s.channel,
		ChaincodeName: "teaTraceCC",
		FunctionName:  "createBatch",
		Args:          args,
		UserCert:      userCert,
		UserKey:       userKey,
		MSPID:         userMSPID,
	}

	resp, err := s.gatewayClient.SubmitTransaction(ctx, req)
	if err != nil {
		s.logger.Error("Failed to create batch via Gateway",
			zap.String("batch_id", batchID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to create batch: %w", err)
	}

	s.logger.Info("Tea batch created via Gateway",
		zap.String("batch_id", batchID),
		zap.String("tx_id", resp.TxID),
	)

	return resp.TxID, nil
}

// GetBatch retrieves a tea batch by ID via Gateway
func (s *TeaTraceServiceViaGateway) GetBatch(ctx context.Context, batchID string) (*TeaBatch, error) {
	result, err := s.gatewayClient.QueryChaincode(ctx, s.channel, "teaTraceCC", "getBatchInfo", []string{batchID})
	if err != nil {
		s.logger.Error("Failed to get batch via Gateway",
			zap.String("batch_id", batchID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get batch: %w", err)
	}

	var batch TeaBatch
	if err := json.Unmarshal(result, &batch); err != nil {
		s.logger.Error("Failed to unmarshal batch data", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	return &batch, nil
}

// GetAllBatches retrieves all tea batches via Gateway
func (s *TeaTraceServiceViaGateway) GetAllBatches(ctx context.Context) ([]*TeaBatch, error) {
	result, err := s.gatewayClient.QueryChaincode(ctx, s.channel, "teaTraceCC", "GetAllBatches", []string{})
	if err != nil {
		s.logger.Error("Failed to get all batches via Gateway", zap.Error(err))
		return nil, fmt.Errorf("failed to get all batches: %w", err)
	}

	var batches []*TeaBatch
	if err := json.Unmarshal(result, &batches); err != nil {
		s.logger.Error("Failed to unmarshal batches data", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal batches: %w", err)
	}

	return batches, nil
}

// VerifyBatch verifies a tea batch with a hash via Gateway
func (s *TeaTraceServiceViaGateway) VerifyBatch(ctx context.Context, batchID, verificationHash string) (string, error) {
	args := []string{batchID, verificationHash}

	// Extract user cert from context
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

	req := &transaction.GatewayTransactionRequest{
		ChannelName:   s.channel,
		ChaincodeName: "teaTraceCC",
		FunctionName:  "VerifyBatch",
		Args:          args,
		UserCert:      userCert,
		UserKey:       userKey,
		MSPID:         userMSPID,
	}

	resp, err := s.gatewayClient.SubmitTransaction(ctx, req)
	if err != nil {
		s.logger.Error("Failed to verify batch via Gateway",
			zap.String("batch_id", batchID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to verify batch: %w", err)
	}

	s.logger.Info("Tea batch verified via Gateway",
		zap.String("batch_id", batchID),
		zap.String("tx_id", resp.TxID),
	)

	return resp.TxID, nil
}

// UpdateBatchStatus updates the status of a tea batch via Gateway
func (s *TeaTraceServiceViaGateway) UpdateBatchStatus(ctx context.Context, batchID, status string) (string, error) {
	args := []string{batchID, status}

	// Extract user cert from context
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

	req := &transaction.GatewayTransactionRequest{
		ChannelName:   s.channel,
		ChaincodeName: "teaTraceCC",
		FunctionName:  "updateBatchStatus",
		Args:          args,
		UserCert:      userCert,
		UserKey:       userKey,
		MSPID:         userMSPID,
	}

	resp, err := s.gatewayClient.SubmitTransaction(ctx, req)
	if err != nil {
		s.logger.Error("Failed to update batch status via Gateway",
			zap.String("batch_id", batchID),
			zap.String("status", status),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to update batch status: %w", err)
	}

	s.logger.Info("Tea batch status updated via Gateway",
		zap.String("batch_id", batchID),
		zap.String("status", status),
		zap.String("tx_id", resp.TxID),
	)

	return resp.TxID, nil
}

// CreatePackage creates a new tea package via Gateway
func (s *TeaTraceServiceViaGateway) CreatePackage(ctx context.Context, packageID, batchID string, weight float64, productionDate, expiryDate string) (string, error) {
	args := []string{
		packageID,
		batchID,
		fmt.Sprintf("%.2f", weight),
		productionDate,
	}
	if expiryDate != "" {
		args = append(args, expiryDate)
	}

	// Extract user cert from context
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

	req := &transaction.GatewayTransactionRequest{
		ChannelName:   s.channel,
		ChaincodeName: "teaTraceCC",
		FunctionName:  "createPackage",
		Args:          args,
		UserCert:      userCert,
		UserKey:       userKey,
		MSPID:         userMSPID,
	}

	resp, err := s.gatewayClient.SubmitTransaction(ctx, req)
	if err != nil {
		s.logger.Error("Failed to create package via Gateway",
			zap.String("package_id", packageID),
			zap.String("batch_id", batchID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to create package: %w", err)
	}

	s.logger.Info("Tea package created via Gateway",
		zap.String("package_id", packageID),
		zap.String("batch_id", batchID),
		zap.String("tx_id", resp.TxID),
	)

	return resp.TxID, nil
}

// GetPackage retrieves a tea package by ID via Gateway
func (s *TeaTraceServiceViaGateway) GetPackage(ctx context.Context, packageID string) (*TeaPackage, error) {
	result, err := s.gatewayClient.QueryChaincode(ctx, s.channel, "teaTraceCC", "getPackageInfo", []string{packageID})
	if err != nil {
		s.logger.Error("Failed to get package via Gateway", zap.Error(err), zap.String("package_id", packageID))
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	// Check if package exists (null response)
	if len(result) == 0 || string(result) == "null" {
		return nil, fmt.Errorf("package not found: %s", packageID)
	}

	var pkg TeaPackage
	if err := json.Unmarshal(result, &pkg); err != nil {
		s.logger.Error("Failed to unmarshal package data", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	return &pkg, nil
}

// HealthCheck performs a health check on the chaincode via Gateway
func (s *TeaTraceServiceViaGateway) HealthCheck(ctx context.Context) error {
	_, err := s.gatewayClient.QueryChaincode(ctx, s.channel, "teaTraceCC", "healthCheck", []string{})
	if err != nil {
		s.logger.Error("Chaincode health check failed via Gateway", zap.Error(err))
		return fmt.Errorf("chaincode health check failed: %w", err)
	}

	s.logger.Info("Chaincode health check passed via Gateway")
	return nil
}

