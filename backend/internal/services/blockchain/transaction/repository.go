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
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles transaction data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new transaction repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateTransaction creates a new transaction record
func (r *Repository) CreateTransaction(ctx context.Context, tx *Transaction) error {
	query := `
		INSERT INTO transactions 
		(id, tx_id, user_id, channel_name, chaincode_name, function_name, args, transient_data, status, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// Convert Payload to transient_data (JSONB)
	var transientData interface{}
	if tx.Payload != nil {
		transientData = tx.Payload
	}

	_, err := r.db.Exec(ctx, query,
		tx.ID, tx.TxID, tx.UserID, tx.ChannelID, tx.ChaincodeID,
		tx.FunctionName, tx.Args, transientData, tx.Status, tx.SubmittedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// UpdateTransactionStatus updates the transaction status
// Note: Database schema only has: status, block_number, error_message, committed_at
func (r *Repository) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string, blockNumber *uint64, txIndex *uint32, responseData string, errorMsg *string, validationCode *int32) error {
	query := `
		UPDATE transactions 
		SET status = $2, block_number = $3, error_message = $4,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status, blockNumber, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// GetTransactionByID retrieves a transaction by ID
func (r *Repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	tx := &Transaction{}
	query := `
		SELECT id, tx_id, user_id, channel_name, chaincode_name, function_name, 
		       args, transient_data, status, block_number, 
		       error_message, timestamp, updated_at
		FROM transactions 
		WHERE id = $1
	`


	var timestamp, updatedAt sql.NullTime
	var argsJSON, payloadJSON sql.NullString
	var blockNumber sql.NullInt64
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tx.ID, &tx.TxID, &tx.UserID, &tx.ChannelID, &tx.ChaincodeID,
		&tx.FunctionName, &argsJSON, &payloadJSON, &tx.Status,
		&blockNumber, &tx.ErrorMessage,
		&timestamp, &updatedAt,
	)
	
	// Handle NULL block_number
	if blockNumber.Valid {
		blockNum := uint64(blockNumber.Int64)
		tx.BlockNumber = &blockNum
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	
	// Handle NULL timestamp - use zero time if NULL
	if timestamp.Valid {
		tx.SubmittedAt = timestamp.Time
	} else {
		tx.SubmittedAt = time.Time{}
	}
	
	// Parse args JSONB
	if argsJSON.Valid && argsJSON.String != "" {
		var args []string
		if err := json.Unmarshal([]byte(argsJSON.String), &args); err == nil {
			tx.Args = args
		}
	}
	
	// Parse payload/transient_data JSONB
	if payloadJSON.Valid && payloadJSON.String != "" {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(payloadJSON.String), &payload); err == nil {
			tx.Payload = payload
		}
	}
	
	// Handle NULL updated_at as completed time
	if updatedAt.Valid {
		tx.CompletedAt = &updatedAt.Time
	}

	return tx, nil
}

// GetTransactionByTxID retrieves a transaction by blockchain transaction ID
func (r *Repository) GetTransactionByTxID(ctx context.Context, txID string) (*Transaction, error) {
	tx := &Transaction{}
	query := `
		SELECT id, tx_id, user_id, channel_name, chaincode_name, function_name, 
		       args, transient_data, status, block_number, 
		       error_message, timestamp, updated_at
		FROM transactions 
		WHERE tx_id = $1
	`


	var timestamp, updatedAt sql.NullTime
	var argsJSON, payloadJSON sql.NullString
	var blockNumber sql.NullInt64
	err := r.db.QueryRow(ctx, query, txID).Scan(
		&tx.ID, &tx.TxID, &tx.UserID, &tx.ChannelID, &tx.ChaincodeID,
		&tx.FunctionName, &argsJSON, &payloadJSON, &tx.Status,
		&blockNumber, &tx.ErrorMessage,
		&timestamp, &updatedAt,
	)
	
	// Handle NULL block_number
	if blockNumber.Valid {
		blockNum := uint64(blockNumber.Int64)
		tx.BlockNumber = &blockNum
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by tx_id: %w", err)
	}
	
	// Handle NULL timestamp - use zero time if NULL
	if timestamp.Valid {
		tx.SubmittedAt = timestamp.Time
	} else {
		tx.SubmittedAt = time.Time{}
	}
	
	// Parse args JSONB
	if argsJSON.Valid && argsJSON.String != "" {
		var args []string
		if err := json.Unmarshal([]byte(argsJSON.String), &args); err == nil {
			tx.Args = args
		}
	}
	
	// Parse payload/transient_data JSONB
	if payloadJSON.Valid && payloadJSON.String != "" {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(payloadJSON.String), &payload); err == nil {
			tx.Payload = payload
		}
	}
	
	// Handle NULL updated_at as completed time
	if updatedAt.Valid {
		tx.CompletedAt = &updatedAt.Time
	}

	return tx, nil
}

// QueryTransactions queries transactions with filters
func (r *Repository) QueryTransactions(ctx context.Context, req *QueryTransactionsRequest) ([]*Transaction, error) {
	query := `
		SELECT id, tx_id, user_id, channel_name, chaincode_name, function_name, 
		       args, transient_data, status, block_number, 
		       error_message, timestamp, updated_at
		FROM transactions
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if req.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *req.UserID)
		argPos++
	}

	if req.ChannelID != nil {
		query += fmt.Sprintf(" AND channel_name = $%d", argPos)
		args = append(args, *req.ChannelID)
		argPos++
	}

	if req.ChaincodeID != nil {
		query += fmt.Sprintf(" AND chaincode_name = $%d", argPos)
		args = append(args, *req.ChaincodeID)
		argPos++
	}

	if req.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *req.Status)
		argPos++
	}

	if req.StartDate != nil {
		query += fmt.Sprintf(" AND submitted_at >= $%d", argPos)
		args = append(args, *req.StartDate)
		argPos++
	}

	if req.EndDate != nil {
		query += fmt.Sprintf(" AND submitted_at <= $%d", argPos)
		args = append(args, *req.EndDate)
		argPos++
	}

	query += " ORDER BY COALESCE(submitted_at, created_at) DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, req.Limit)
		argPos++
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, req.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		var timestamp, updatedAt sql.NullTime
		var argsJSON, payloadJSON sql.NullString
		var blockNumber sql.NullInt64
		
		err := rows.Scan(
			&tx.ID, &tx.TxID, &tx.UserID, &tx.ChannelID, &tx.ChaincodeID,
			&tx.FunctionName, &argsJSON, &payloadJSON, &tx.Status,
			&blockNumber, &tx.ErrorMessage,
			&timestamp, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		
		// Handle NULL block_number
		if blockNumber.Valid {
			blockNum := uint64(blockNumber.Int64)
			tx.BlockNumber = &blockNum
		}
		
		// Handle NULL timestamp - use zero time if NULL
		if timestamp.Valid {
			tx.SubmittedAt = timestamp.Time
		} else {
			// If NULL, use zero time (will be serialized as "0001-01-01T00:00:00Z" in JSON)
			tx.SubmittedAt = time.Time{}
		}
		
		// Parse args JSONB
		if argsJSON.Valid && argsJSON.String != "" {
			var args []string
			if err := json.Unmarshal([]byte(argsJSON.String), &args); err == nil {
				tx.Args = args
			}
		}
		
		// Parse payload/transient_data JSONB
		if payloadJSON.Valid && payloadJSON.String != "" {
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(payloadJSON.String), &payload); err == nil {
				tx.Payload = payload
			}
		}
		
		// Handle NULL updated_at as completed time
		if updatedAt.Valid {
			tx.CompletedAt = &updatedAt.Time
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// AddStatusHistory adds a status history entry
// Note: Database schema uses: previous_status, new_status, error_message, metadata, created_at
func (r *Repository) AddStatusHistory(ctx context.Context, history *TransactionStatusHistory) error {
	query := `
		INSERT INTO transaction_status_history 
		(id, transaction_id, new_status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		history.ID, history.TransactionID, history.Status, history.Details, history.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to add status history: %w", err)
	}

	return nil
}

// GetStatusHistory retrieves status history for a transaction
func (r *Repository) GetStatusHistory(ctx context.Context, transactionID uuid.UUID) ([]*TransactionStatusHistory, error) {
	query := `
		SELECT id, transaction_id, new_status, error_message, created_at
		FROM transaction_status_history
		WHERE transaction_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status history: %w", err)
	}
	defer rows.Close()

	var history []*TransactionStatusHistory
	for rows.Next() {
		h := &TransactionStatusHistory{}
		var errorMsg *string
		err := rows.Scan(&h.ID, &h.TransactionID, &h.Status, &errorMsg, &h.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status history: %w", err)
		}
		if errorMsg != nil {
			h.Details = *errorMsg
		}
		history = append(history, h)
	}

	return history, nil
}

