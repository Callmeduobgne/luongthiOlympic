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
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/repository/db"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	"go.uber.org/zap"
)

// Service handles transaction operations
type Service struct {
	db              *pgxpool.Pool
	queries         *db.Queries
	chaincodeService *fabric.ChaincodeService
	logger          *zap.Logger
}

// NewService creates a new transaction service
func NewService(dbPool *pgxpool.Pool, chaincodeService *fabric.ChaincodeService, logger *zap.Logger) *Service {
	return &Service{
		db:              dbPool,
		queries:         db.New(),
		chaincodeService: chaincodeService,
		logger:          logger,
	}
}

// SubmitTransaction submits a transaction and tracks it in the database
func (s *Service) SubmitTransaction(
	ctx context.Context,
	req *models.TransactionRequest,
	userID string,
	apiKeyID string,
) (*models.TransactionResponse, error) {
	// Convert transient data from map[string]interface{} to map[string]string (base64)
	transientMap := make(map[string]string)
	if req.TransientData != nil {
		for k, v := range req.TransientData {
			// Convert value to base64 string
			var strValue string
			switch val := v.(type) {
			case string:
				strValue = val
			default:
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal transient data for key '%s': %w", k, err)
				}
				strValue = string(jsonBytes)
			}
			transientMap[k] = strValue
		}
	}

	// Prepare invoke request
	invokeReq := &models.InvokeRequest{
		Function:      req.FunctionName,
		Args:          req.Args,
		Transient:     transientMap,
		EndorsingOrgs: req.EndorsingOrgs,
	}

	// Extract user certificate from context (sent by Backend via header)
	// Context already has cert values from handler, no need to set again

	// Submit transaction via chaincode service (will use user cert if provided)
	invokeResp, err := s.chaincodeService.Invoke(ctx, req.ChannelName, req.ChaincodeName, invokeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke chaincode: %w", err)
	}

	// Prepare transaction data for database
	var userUUID, apiKeyUUID pgtype.UUID
	if userID != "" {
		if err := userUUID.Scan(userID); err != nil {
			s.logger.Warn("Failed to parse user ID", zap.String("userID", userID), zap.Error(err))
		}
	}
	if apiKeyID != "" {
		if err := apiKeyUUID.Scan(apiKeyID); err != nil {
			s.logger.Warn("Failed to parse API key ID", zap.String("apiKeyID", apiKeyID), zap.Error(err))
		}
	}

	// Marshal args and transient data to JSONB
	argsJSON, err := json.Marshal(req.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal args: %w", err)
	}

	transientJSON, err := json.Marshal(req.TransientData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transient data: %w", err)
	}

	// Determine initial status
	status := "SUBMITTED"
	if invokeResp.Status == "VALID" {
		status = "VALID"
	} else if invokeResp.Status == "INVALID" {
		status = "INVALID"
	}

	// Create transaction record
	txRecord, err := s.queries.CreateTransaction(ctx, s.db, db.CreateTransactionParams{
		TxID:          invokeResp.TxID,
		ChannelName:   req.ChannelName,
		ChaincodeName: req.ChaincodeName,
		FunctionName:  req.FunctionName,
		Args:          argsJSON,
		TransientData: transientJSON,
		UserID:        userUUID,
		ApiKeyID:      apiKeyUUID,
		Status:        status,
		BlockNumber:   pgtype.Int8{Int64: int64(invokeResp.BlockNumber), Valid: invokeResp.BlockNumber > 0},
		EndorsingOrgs: req.EndorsingOrgs,
	})
	if err != nil {
		s.logger.Error("Failed to create transaction record", zap.Error(err))
		// Continue anyway - transaction was submitted successfully
	} else {
		// Add status history
		_, err = s.queries.AddTransactionStatusHistory(ctx, s.db, db.AddTransactionStatusHistoryParams{
			TransactionID: txRecord.ID,
			Status:        status,
			BlockNumber:   pgtype.Int8{Int64: int64(invokeResp.BlockNumber), Valid: invokeResp.BlockNumber > 0},
			Details:       nil,
		})
		if err != nil {
			s.logger.Warn("Failed to add transaction status history", zap.Error(err))
		}
	}

	// Convert timestamp
	timestamp, err := time.Parse(time.RFC3339, invokeResp.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	return &models.TransactionResponse{
		ID:          uuidToString(txRecord.ID),
		TxID:        invokeResp.TxID,
		Status:      models.TransactionStatus(status),
		BlockNumber: invokeResp.BlockNumber,
		Timestamp:   timestamp,
	}, nil
}

// GetTransaction retrieves a transaction by ID or TxID
func (s *Service) GetTransaction(ctx context.Context, idOrTxID string) (*models.Transaction, error) {
	// Try as UUID first
	var txRecord db.Transaction
	var err error

	txUUID, parseErr := parseUUID(idOrTxID)
	if parseErr == nil {
		// Try as UUID
		txRecord, err = s.queries.GetTransactionByID(ctx, s.db, txUUID)
		if err == nil {
			return s.convertTransaction(&txRecord), nil
		}
	}

	// Try as TxID
	txRecord, err = s.queries.GetTransactionByTxID(ctx, s.db, idOrTxID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found")
	}

	return s.convertTransaction(&txRecord), nil
}

// GetTransactionStatus retrieves current transaction status
func (s *Service) GetTransactionStatus(ctx context.Context, idOrTxID string) (models.TransactionStatus, error) {
	tx, err := s.GetTransaction(ctx, idOrTxID)
	if err != nil {
		return "", err
	}

	return tx.Status, nil
}

// GetTransactionReceipt retrieves transaction receipt with block information
func (s *Service) GetTransactionReceipt(ctx context.Context, idOrTxID string) (*models.TransactionReceipt, error) {
	tx, err := s.GetTransaction(ctx, idOrTxID)
	if err != nil {
		return nil, err
	}

	return &models.TransactionReceipt{
		TxID:          tx.TxID,
		Status:        tx.Status,
		BlockNumber:   tx.BlockNumber,
		BlockHash:     tx.BlockHash,
		Timestamp:     tx.Timestamp,
		ChannelName:   tx.ChannelName,
		ChaincodeName: tx.ChaincodeName,
		FunctionName:  tx.FunctionName,
		ErrorMessage:  tx.ErrorMessage,
	}, nil
}

// ListTransactions lists transactions with filters
func (s *Service) ListTransactions(ctx context.Context, query *models.TransactionListQuery) ([]*models.Transaction, int64, error) {
	// Prepare parameters
	var channelName, chaincodeName, status pgtype.Text
	var userID pgtype.UUID
	var startTime, endTime pgtype.Timestamp

	if query.ChannelName != "" {
		channelName = pgtype.Text{String: query.ChannelName, Valid: true}
	}
	if query.ChaincodeName != "" {
		chaincodeName = pgtype.Text{String: query.ChaincodeName, Valid: true}
	}
	if query.Status != "" {
		status = pgtype.Text{String: string(query.Status), Valid: true}
	}
	if query.UserID != "" {
		if err := userID.Scan(query.UserID); err != nil {
			s.logger.Warn("Failed to parse user ID", zap.String("userID", query.UserID), zap.Error(err))
		}
	}
	if query.StartTime != nil {
		startTime = pgtype.Timestamp{Time: *query.StartTime, Valid: true}
	}
	if query.EndTime != nil {
		endTime = pgtype.Timestamp{Time: *query.EndTime, Valid: true}
	}

	limit := int32(50) // Default limit
	if query.Limit > 0 {
		limit = int32(query.Limit)
	}
	offset := int32(0)
	if query.Offset > 0 {
		offset = int32(query.Offset)
	}

	// List transactions
	txRecords, err := s.queries.ListTransactions(ctx, s.db, db.ListTransactionsParams{
		Column1: channelName.String,
		Column2: chaincodeName.String,
		Column3: status.String,
		Column4: userID,
		Column5: startTime,
		Column6: endTime,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Count total
	count, err := s.queries.CountTransactions(ctx, s.db, db.CountTransactionsParams{
		Column1: channelName.String,
		Column2: chaincodeName.String,
		Column3: status.String,
		Column4: userID,
		Column5: startTime,
		Column6: endTime,
	})
	if err != nil {
		s.logger.Warn("Failed to count transactions", zap.Error(err))
		count = int64(len(txRecords))
	}

	// Convert to models
	result := make([]*models.Transaction, 0, len(txRecords))
	for _, tx := range txRecords {
		result = append(result, s.convertTransaction(&tx))
	}

	return result, count, nil
}

// UpdateTransactionStatus updates transaction status (called by event listener)
func (s *Service) UpdateTransactionStatus(
	ctx context.Context,
	txID string,
	newStatus models.TransactionStatus,
	blockNumber uint64,
	blockHash string,
	errorMessage string,
) error {
	// Get transaction by TxID
	txRecord, err := s.queries.GetTransactionByTxID(ctx, s.db, txID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Update status
	_, err = s.queries.UpdateTransactionStatus(ctx, s.db, db.UpdateTransactionStatusParams{
		ID:          txRecord.ID,
		Status:      string(newStatus),
		BlockNumber: pgtype.Int8{Int64: int64(blockNumber), Valid: blockNumber > 0},
		BlockHash:   pgtype.Text{String: blockHash, Valid: blockHash != ""},
		ErrorMessage: pgtype.Text{String: errorMessage, Valid: errorMessage != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Add status history
	_, err = s.queries.AddTransactionStatusHistory(ctx, s.db, db.AddTransactionStatusHistoryParams{
		TransactionID: txRecord.ID,
		Status:        string(newStatus),
		BlockNumber:   pgtype.Int8{Int64: int64(blockNumber), Valid: blockNumber > 0},
		Details:       nil,
	})
	if err != nil {
		s.logger.Warn("Failed to add transaction status history", zap.Error(err))
	}

	return nil
}

// Helper functions

func (s *Service) convertTransaction(tx *db.Transaction) *models.Transaction {
	// Unmarshal args
	var args []string
	if len(tx.Args) > 0 {
		json.Unmarshal(tx.Args, &args)
	}

	// Unmarshal transient data
	var transientData map[string]interface{}
	if len(tx.TransientData) > 0 {
		json.Unmarshal(tx.TransientData, &transientData)
	}

	return &models.Transaction{
		ID:            uuidToString(tx.ID),
		TxID:          tx.TxID,
		ChannelName:   tx.ChannelName,
		ChaincodeName: tx.ChaincodeName,
		FunctionName:  tx.FunctionName,
		Args:          args,
		TransientData: transientData,
		UserID:        uuidToString(tx.UserID),
		APIKeyID:      uuidToString(tx.ApiKeyID),
		Status:        models.TransactionStatus(tx.Status),
		BlockNumber:   uint64(tx.BlockNumber.Int64),
		BlockHash:     tx.BlockHash.String,
		Timestamp:     tx.Timestamp.Time,
		ErrorMessage:  tx.ErrorMessage.String,
		EndorsingOrgs: tx.EndorsingOrgs,
		CreatedAt:     tx.CreatedAt.Time,
		UpdatedAt:     tx.UpdatedAt.Time,
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}

func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid.Bytes[0:4],
		uuid.Bytes[4:6],
		uuid.Bytes[6:8],
		uuid.Bytes[8:10],
		uuid.Bytes[10:16],
	)
}

