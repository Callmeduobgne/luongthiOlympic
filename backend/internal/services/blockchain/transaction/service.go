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

package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NOTE: FabricClient interface removed - Backend MUST use Gateway for all blockchain operations

// GatewayClient is an interface for API Gateway client operations
type GatewayClient interface {
	SubmitTransaction(ctx context.Context, req *GatewayTransactionRequest) (*GatewayTransactionResponse, error)
	GetTransaction(ctx context.Context, idOrTxID string) (*GatewayTransactionResponse, error)
	QueryChaincode(ctx context.Context, channelName, chaincodeName, functionName string, args []string) ([]byte, error)
}

// GatewayTransactionRequest represents Gateway transaction request
type GatewayTransactionRequest struct {
	ChannelName   string
	ChaincodeName string
	FunctionName  string
	Args          []string
	TransientData map[string]interface{}
	EndorsingOrgs []string
	UserCert      string // User certificate (PEM format) - required for Gateway
	UserKey       string // User private key (PEM format) - required for Gateway
	MSPID         string // MSP ID for the user
}

// GatewayTransactionResponse represents Gateway transaction response
type GatewayTransactionResponse struct {
	ID          string
	TxID        string
	Status      string
	BlockNumber uint64
	Timestamp   time.Time
}

// Service handles transaction business logic
type Service struct {
	repo          *Repository
	gatewayClient GatewayClient // Gateway client (REQUIRED - Backend must use Gateway)
	logger        *zap.Logger
}

// NewService creates a new transaction service
// NOTE: fabricClient is no longer used - Backend MUST use Gateway for all blockchain operations
func NewService(repo *Repository, gatewayClient GatewayClient, logger *zap.Logger) *Service {
	if gatewayClient == nil {
		logger.Fatal("Gateway client is required - Backend cannot connect directly to Fabric")
	}
	return &Service{
		repo:          repo,
		gatewayClient: gatewayClient,
		logger:        logger,
	}
}

// SubmitTransaction submits a transaction to the blockchain
func (s *Service) SubmitTransaction(ctx context.Context, userID uuid.UUID, req *SubmitTransactionRequest) (*Transaction, error) {
	// Validate request
	if req.ChannelID == "" || req.ChaincodeID == "" || req.FunctionName == "" {
		return nil, fmt.Errorf("channel_id, chaincode_id, and function_name are required")
	}

	tx := &Transaction{
		ID:           uuid.New(),
		TxID:         "", // Will be set after submission
		UserID:       userID,
		ChannelID:    req.ChannelID,
		ChaincodeID:  req.ChaincodeID,
		FunctionName: req.FunctionName,
		Args:         req.Args,
		Payload:      req.Payload,
		Status:       StatusPending,
		SubmittedAt:  time.Now(),
	}

	// Save to database
	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to create transaction record", zap.Error(err))
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Add status history
	s.addStatusHistory(ctx, tx.ID, StatusPending, "Transaction created")

	// Submit to Gateway (REQUIRED - Backend cannot connect directly to Fabric)
	// Note: submitToGateway runs in goroutine, so we need to preserve context values
	// but use background context to avoid cancellation when HTTP request ends
	if s.gatewayClient == nil {
		s.logger.Error("Gateway client is required but not configured")
		errMsg := "Gateway client not configured - Backend must use Gateway for blockchain operations"
		s.repo.UpdateTransactionStatus(ctx, tx.ID, StatusFailed, nil, nil, "", &errMsg, nil)
		s.addStatusHistory(ctx, tx.ID, StatusFailed, errMsg)
		return tx, fmt.Errorf("gateway client required")
	}
	
	// Extract context values before passing to goroutine (to avoid cancellation issues)
	// Create background context with preserved values
	bgCtx := context.Background()
	if userCert, ok := ctx.Value("user_cert").(string); ok {
		bgCtx = context.WithValue(bgCtx, "user_cert", userCert)
	}
	if userKey, ok := ctx.Value("user_key").(string); ok {
		bgCtx = context.WithValue(bgCtx, "user_key", userKey)
	}
	if mspID, ok := ctx.Value("user_msp_id").(string); ok {
		bgCtx = context.WithValue(bgCtx, "user_msp_id", mspID)
	}
	
	go s.submitToGateway(bgCtx, tx, req.Transient)

	return tx, nil
}

// submitToGateway submits the transaction via Gateway
func (s *Service) submitToGateway(ctx context.Context, tx *Transaction, transient map[string][]byte) {
	// Create timeout context but preserve original context values (cert, etc.)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.logger.Info("Submitting transaction via Gateway",
		zap.String("tx_id", tx.ID.String()),
		zap.String("channel", tx.ChannelID),
		zap.String("chaincode", tx.ChaincodeID),
	)

	// Get user cert from context (should be set by handler)
	userCert := ""
	userKey := ""
	mspID := ""
	if cert, ok := ctx.Value("user_cert").(string); ok && cert != "" {
		userCert = cert
	}
	if key, ok := ctx.Value("user_key").(string); ok && key != "" {
		userKey = key
	}
	if msp, ok := ctx.Value("user_msp_id").(string); ok && msp != "" {
		mspID = msp
	}

	if userCert == "" {
		s.logger.Error("User certificate not found in context", zap.String("user_id", tx.UserID.String()))
		errMsg := "User certificate required for Gateway submission"
		s.repo.UpdateTransactionStatus(ctx, tx.ID, StatusFailed, nil, nil, "", &errMsg, nil)
		s.addStatusHistory(ctx, tx.ID, StatusFailed, errMsg)
		return
	}

	// Convert transient data
	transientData := make(map[string]interface{})
	if transient != nil {
		for k, v := range transient {
			transientData[k] = string(v)
		}
	}

	// Prepare Gateway request
	gatewayReq := &GatewayTransactionRequest{
		ChannelName:   tx.ChannelID,
		ChaincodeName: tx.ChaincodeID,
		FunctionName:  tx.FunctionName,
		Args:          tx.Args,
		TransientData: transientData,
		UserCert:      userCert,
		UserKey:       userKey,
		MSPID:         mspID,
	}

	// Submit via Gateway
	gatewayResp, err := s.gatewayClient.SubmitTransaction(ctx, gatewayReq)
	if err != nil {
		s.logger.Error("Failed to submit transaction via Gateway", zap.Error(err))
		errMsg := err.Error()
		s.repo.UpdateTransactionStatus(ctx, tx.ID, StatusFailed, nil, nil, "", &errMsg, nil)
		s.addStatusHistory(ctx, tx.ID, StatusFailed, err.Error())
		return
	}

	// Update with Gateway transaction ID
	tx.TxID = gatewayResp.TxID
	status := StatusSubmitted
	if gatewayResp.Status == "VALID" {
		status = StatusCommitted
	} else if gatewayResp.Status == "INVALID" || gatewayResp.Status == "FAILED" {
		status = StatusFailed
	}

	s.repo.UpdateTransactionStatus(ctx, tx.ID, status, &gatewayResp.BlockNumber, nil, "", nil, nil)
	s.addStatusHistory(ctx, tx.ID, status, fmt.Sprintf("Submitted via Gateway with tx_id: %s, status: %s", gatewayResp.TxID, gatewayResp.Status))

	// If already committed, we're done. Otherwise wait for commit
	if status == StatusCommitted {
		s.logger.Info("Transaction committed via Gateway",
			zap.String("tx_id", tx.TxID),
			zap.Uint64("block_number", gatewayResp.BlockNumber),
		)
	} else {
		// Wait for commit (in background)
		go s.waitForCommitViaGateway(tx)
	}
}

// NOTE: submitToFabric removed - Backend MUST use Gateway for all blockchain operations

// waitForCommitViaGateway waits for transaction commit via Gateway
func (s *Service) waitForCommitViaGateway(tx *Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gatewayTx, err := s.gatewayClient.GetTransaction(ctx, tx.TxID)
			if err != nil {
				s.logger.Error("Failed to get transaction status from Gateway", zap.Error(err))
				continue
			}

			if gatewayTx.Status == "VALID" {
				s.repo.UpdateTransactionStatus(ctx, tx.ID, StatusCommitted, &gatewayTx.BlockNumber, nil, "", nil, nil)
				s.addStatusHistory(ctx, tx.ID, StatusCommitted, fmt.Sprintf("Transaction committed via Gateway, block: %d", gatewayTx.BlockNumber))
				return
			} else if gatewayTx.Status == "INVALID" || gatewayTx.Status == "FAILED" {
				errMsg := "Transaction validation failed"
				s.repo.UpdateTransactionStatus(ctx, tx.ID, StatusFailed, nil, nil, "", &errMsg, nil)
				s.addStatusHistory(ctx, tx.ID, StatusFailed, errMsg)
				return
			}

		case <-ctx.Done():
			errMsg := "Transaction commit timeout"
			s.repo.UpdateTransactionStatus(ctx, tx.ID, StatusTimeout, nil, nil, "", &errMsg, nil)
			s.addStatusHistory(ctx, tx.ID, StatusTimeout, errMsg)
			return
		}
	}
}

// waitForCommit waits for the transaction to be committed
// NOTE: waitForCommit removed - Backend MUST use Gateway for all blockchain operations

// QueryTransaction performs a read-only query on the blockchain via Gateway
// NOTE: Backend MUST use Gateway for all blockchain operations
func (s *Service) QueryTransaction(ctx context.Context, req *QueryTransactionRequest) ([]byte, error) {
	if s.gatewayClient == nil {
		return nil, fmt.Errorf("gateway client required - Backend cannot connect directly to Fabric")
	}

	// Use Gateway for query
	result, err := s.gatewayClient.QueryChaincode(ctx, req.ChannelID, req.ChaincodeID, req.FunctionName, req.Args)
	if err != nil {
		s.logger.Error("Failed to query via Gateway", zap.Error(err))
		return nil, fmt.Errorf("failed to query via Gateway: %w", err)
	}
	return result, nil
}

// GetTransaction retrieves a transaction by ID
func (s *Service) GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	return s.repo.GetTransactionByID(ctx, id)
}

// GetTransactionByTxID retrieves a transaction by blockchain transaction ID
func (s *Service) GetTransactionByTxID(ctx context.Context, txID string) (*Transaction, error) {
	return s.repo.GetTransactionByTxID(ctx, txID)
}

// QueryTransactions queries multiple transactions
func (s *Service) QueryTransactions(ctx context.Context, req *QueryTransactionsRequest) ([]*Transaction, error) {
	// Set defaults
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	return s.repo.QueryTransactions(ctx, req)
}

// GetStatusHistory retrieves status history for a transaction
func (s *Service) GetStatusHistory(ctx context.Context, transactionID uuid.UUID) ([]*TransactionStatusHistory, error) {
	return s.repo.GetStatusHistory(ctx, transactionID)
}

// addStatusHistory adds a status history entry
func (s *Service) addStatusHistory(ctx context.Context, transactionID uuid.UUID, status, details string) {
	history := &TransactionStatusHistory{
		ID:            uuid.New(),
		TransactionID: transactionID,
		Status:        status,
		Details:       details,
		Timestamp:     time.Now(),
	}

	if err := s.repo.AddStatusHistory(ctx, history); err != nil {
		s.logger.Error("Failed to add status history", zap.Error(err))
	}
}

