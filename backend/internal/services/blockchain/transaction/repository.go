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
		INSERT INTO blockchain.transactions 
		(id, tx_id, user_id, channel_id, chaincode_id, function_name, args, payload, status, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		tx.ID, tx.TxID, tx.UserID, tx.ChannelID, tx.ChaincodeID,
		tx.FunctionName, tx.Args, tx.Payload, tx.Status, tx.SubmittedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// UpdateTransactionStatus updates the transaction status
func (r *Repository) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string, blockNumber *uint64, txIndex *uint32, responseData string, errorMsg *string, validationCode *int32) error {
	query := `
		UPDATE blockchain.transactions 
		SET status = $2, block_number = $3, tx_index = $4, 
		    response_data = $5, error_message = $6, validation_code = $7,
		    completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status, blockNumber, txIndex, responseData, errorMsg, validationCode)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// GetTransactionByID retrieves a transaction by ID
func (r *Repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	tx := &Transaction{}
	query := `
		SELECT id, tx_id, user_id, channel_id, chaincode_id, function_name, 
		       args, payload, status, block_number, tx_index, response_data, 
		       error_message, submitted_at, completed_at, validation_code
		FROM blockchain.transactions 
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&tx.ID, &tx.TxID, &tx.UserID, &tx.ChannelID, &tx.ChaincodeID,
		&tx.FunctionName, &tx.Args, &tx.Payload, &tx.Status,
		&tx.BlockNumber, &tx.TxIndex, &tx.ResponseData, &tx.ErrorMessage,
		&tx.SubmittedAt, &tx.CompletedAt, &tx.ValidationCode,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

// GetTransactionByTxID retrieves a transaction by blockchain transaction ID
func (r *Repository) GetTransactionByTxID(ctx context.Context, txID string) (*Transaction, error) {
	tx := &Transaction{}
	query := `
		SELECT id, tx_id, user_id, channel_id, chaincode_id, function_name, 
		       args, payload, status, block_number, tx_index, response_data, 
		       error_message, submitted_at, completed_at, validation_code
		FROM blockchain.transactions 
		WHERE tx_id = $1
	`

	err := r.db.QueryRow(ctx, query, txID).Scan(
		&tx.ID, &tx.TxID, &tx.UserID, &tx.ChannelID, &tx.ChaincodeID,
		&tx.FunctionName, &tx.Args, &tx.Payload, &tx.Status,
		&tx.BlockNumber, &tx.TxIndex, &tx.ResponseData, &tx.ErrorMessage,
		&tx.SubmittedAt, &tx.CompletedAt, &tx.ValidationCode,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by tx_id: %w", err)
	}

	return tx, nil
}

// QueryTransactions queries transactions with filters
func (r *Repository) QueryTransactions(ctx context.Context, req *QueryTransactionsRequest) ([]*Transaction, error) {
	query := `
		SELECT id, tx_id, user_id, channel_id, chaincode_id, function_name, 
		       args, payload, status, block_number, tx_index, response_data, 
		       error_message, submitted_at, completed_at, validation_code
		FROM blockchain.transactions
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
		query += fmt.Sprintf(" AND channel_id = $%d", argPos)
		args = append(args, *req.ChannelID)
		argPos++
	}

	if req.ChaincodeID != nil {
		query += fmt.Sprintf(" AND chaincode_id = $%d", argPos)
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

	query += " ORDER BY submitted_at DESC"

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
		err := rows.Scan(
			&tx.ID, &tx.TxID, &tx.UserID, &tx.ChannelID, &tx.ChaincodeID,
			&tx.FunctionName, &tx.Args, &tx.Payload, &tx.Status,
			&tx.BlockNumber, &tx.TxIndex, &tx.ResponseData, &tx.ErrorMessage,
			&tx.SubmittedAt, &tx.CompletedAt, &tx.ValidationCode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// AddStatusHistory adds a status history entry
func (r *Repository) AddStatusHistory(ctx context.Context, history *TransactionStatusHistory) error {
	query := `
		INSERT INTO blockchain.transaction_status_history 
		(id, transaction_id, status, details, timestamp)
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
		SELECT id, transaction_id, status, details, timestamp
		FROM blockchain.transaction_status_history
		WHERE transaction_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status history: %w", err)
	}
	defer rows.Close()

	var history []*TransactionStatusHistory
	for rows.Next() {
		h := &TransactionStatusHistory{}
		err := rows.Scan(&h.ID, &h.TransactionID, &h.Status, &h.Details, &h.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

