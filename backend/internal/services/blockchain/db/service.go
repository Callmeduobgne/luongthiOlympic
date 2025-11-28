package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Transaction represents a blockchain transaction from DB
type Transaction struct {
	ID            string    `json:"id"`
	TxID          string    `json:"txId"`
	ChannelName   string    `json:"channelName"`
	ChaincodeName string    `json:"chaincodeName"`
	FunctionName  string    `json:"functionName"`
	Status        string    `json:"status"`
	BlockNumber   uint64    `json:"blockNumber"`
	BlockHash     string    `json:"blockHash"`
	Timestamp     time.Time `json:"timestamp"`
	Args          []string  `json:"args"`
	NfcTagId      string    `json:"nfcTagId"`
}

// BlockInfo represents block information from DB
type BlockInfo struct {
	Height            uint64 `json:"height"`
	CurrentBlockHash  string `json:"currentBlockHash"`
	PreviousBlockHash string `json:"previousBlockHash"`
}

// Service handles database queries for blockchain data
type Service struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewService creates a new blockchain database service
func NewService(db *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// ListTransactions returns a list of transactions
func (s *Service) ListTransactions(ctx context.Context, limit, offset int) ([]Transaction, int64, error) {
	// Get total count
	var total int64
	err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM transactions").Scan(&total)
	if err != nil {
		s.logger.Error("Failed to count transactions", zap.Error(err))
		return nil, 0, err
	}

	// Get transactions
	rows, err := s.db.Query(ctx, `
		SELECT id, tx_id, channel_name, chaincode_name, function_name, status, 
		       COALESCE(block_number, 0), COALESCE(block_hash, ''), 
		       COALESCE(timestamp, created_at) as timestamp,
		       COALESCE(args, '[]'::jsonb)
		FROM transactions
		ORDER BY COALESCE(timestamp, created_at) DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list transactions", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		var idStr string
		var argsJSON []byte
		err := rows.Scan(
			&idStr, &tx.TxID, &tx.ChannelName, &tx.ChaincodeName, &tx.FunctionName,
			&tx.Status, &tx.BlockNumber, &tx.BlockHash, &tx.Timestamp,
			&argsJSON,
		)
		if err != nil {
			s.logger.Error("Failed to scan transaction", zap.Error(err))
			continue
		}
		
		// Parse args JSON
		if len(argsJSON) > 0 {
			if err := json.Unmarshal(argsJSON, &tx.Args); err != nil {
				s.logger.Warn("Failed to unmarshal args", zap.Error(err))
				tx.Args = []string{}
			}
		} else {
			tx.Args = []string{}
		}
		
		tx.ID = idStr
		txs = append(txs, tx)
	}

	return txs, total, nil
}

// GetTransaction returns a transaction by ID
func (s *Service) GetTransaction(ctx context.Context, txID string) (*Transaction, error) {
	var tx Transaction
	var idStr string
	
	err := s.db.QueryRow(ctx, `
		SELECT id, tx_id, channel_name, chaincode_name, function_name, status, 
		       COALESCE(block_number, 0), COALESCE(block_hash, ''),
		       COALESCE(timestamp, created_at) as timestamp
		FROM transactions
		WHERE tx_id = $1
	`, txID).Scan(
		&idStr, &tx.TxID, &tx.ChannelName, &tx.ChaincodeName, &tx.FunctionName,
		&tx.Status, &tx.BlockNumber, &tx.BlockHash, &tx.Timestamp,
	)
	
	if err != nil {
		s.logger.Error("Failed to get transaction", zap.String("txID", txID), zap.Error(err))
		return nil, err
	}
	
	tx.ID = idStr
	return &tx, nil
}

// GetTransactionByNfcId retrieves a transaction by its NFC Tag ID
func (s *Service) GetTransactionByNfcId(ctx context.Context, nfcId string) (*Transaction, error) {
	var tx Transaction
	var idStr string
	var argsJSON []byte

	err := s.db.QueryRow(ctx, `
		SELECT id, tx_id, channel_name, chaincode_name, function_name, status, 
		       COALESCE(block_number, 0), COALESCE(block_hash, ''),
		       COALESCE(timestamp, created_at) as timestamp,
		       COALESCE(args, '[]'::jsonb), COALESCE(nfc_tag_id, '')
		FROM transactions
		WHERE nfc_tag_id = $1
		LIMIT 1
	`, nfcId).Scan(
		&idStr, &tx.TxID, &tx.ChannelName, &tx.ChaincodeName, &tx.FunctionName,
		&tx.Status, &tx.BlockNumber, &tx.BlockHash, &tx.Timestamp, &argsJSON, &tx.NfcTagId,
	)
	if err != nil {
		s.logger.Error("Failed to get transaction by NFC ID", zap.String("nfcId", nfcId), zap.Error(err))
		return nil, err
	}
	tx.ID = idStr

	if len(argsJSON) > 0 {
		if err := json.Unmarshal(argsJSON, &tx.Args); err != nil {
			s.logger.Warn("Failed to unmarshal args for transaction by NFC ID", zap.Error(err))
			tx.Args = []string{}
		}
	} else {
		tx.Args = []string{}
	}

	return &tx, nil
}

// GetLatestBlock returns the latest block info from transactions
func (s *Service) GetLatestBlock(ctx context.Context) (*BlockInfo, error) {
	var info BlockInfo
	
	// Get max block number and its hash
	err := s.db.QueryRow(ctx, `
		SELECT block_number, COALESCE(block_hash, '')
		FROM transactions
		WHERE block_number IS NOT NULL
		ORDER BY block_number DESC
		LIMIT 1
	`).Scan(&info.Height, &info.CurrentBlockHash)
	
	if err != nil {
		// If no blocks found, return empty info (not an error)
		if err.Error() == "no rows in result set" {
			return &BlockInfo{Height: 0}, nil
		}
		s.logger.Error("Failed to get latest block", zap.Error(err))
		return nil, err
	}

	return &info, nil
}

// Batch represents a tea batch
type Batch struct {
	BatchID     string    `json:"batch_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Quantity    string    `json:"quantity"`
	TxID        string    `json:"tx_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// ListBatches returns a list of batches from CreateBatch transactions
func (s *Service) ListBatches(ctx context.Context, limit, offset int) ([]Batch, int64, error) {
	// Get total count
	var total int64
	err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM transactions WHERE function_name = 'CreateBatch'").Scan(&total)
	if err != nil {
		s.logger.Error("Failed to count batches", zap.Error(err))
		return nil, 0, err
	}

	// Get transactions
	rows, err := s.db.Query(ctx, `
		SELECT tx_id, args, COALESCE(timestamp, created_at) as timestamp 
		FROM transactions 
		WHERE function_name = 'CreateBatch'
		ORDER BY COALESCE(timestamp, created_at) DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list batches", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var batches []Batch
	for rows.Next() {
		var txID string
		var argsJSON []byte
		var timestamp time.Time
		
		err := rows.Scan(&txID, &argsJSON, &timestamp)
		if err != nil {
			s.logger.Error("Failed to scan batch transaction", zap.Error(err))
			continue
		}
		
		// Parse args
		var args []string
		if err := json.Unmarshal(argsJSON, &args); err != nil {
			s.logger.Error("Failed to unmarshal args", zap.Error(err))
			continue
		}
		
		if len(args) >= 1 {
			batch := Batch{
				BatchID:   args[0],
				TxID:      txID,
				Timestamp: timestamp,
			}
			if len(args) >= 2 { batch.Name = args[1] }
			if len(args) >= 3 { batch.Description = args[2] }
			if len(args) >= 4 { batch.Quantity = args[3] }
			
			batches = append(batches, batch)
		}
	}
	
	return batches, total, nil
}

// SaveTransaction saves a transaction to the database
func (s *Service) SaveTransaction(ctx context.Context, tx *Transaction) error {
	argsJSON, _ := json.Marshal(tx.Args)
	
	_, err := s.db.Exec(ctx, `
		INSERT INTO transactions (
			tx_id, channel_name, chaincode_name, function_name, 
			status, block_number, block_hash, timestamp, args
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tx_id) DO NOTHING
	`, tx.TxID, tx.ChannelName, tx.ChaincodeName, tx.FunctionName, 
	   tx.Status, tx.BlockNumber, tx.BlockHash, tx.Timestamp, argsJSON)
	
	if err != nil {
		s.logger.Error("Failed to save transaction", zap.Error(err))
		return err
	}
	return nil
	return nil
}

// UpdateTransactionNfcId updates the NFC Tag ID for a transaction
func (s *Service) UpdateTransactionNfcId(ctx context.Context, txId string, nfcId string) error {
	result, err := s.db.Exec(ctx, `
		UPDATE transactions 
		SET nfc_tag_id = $1
		WHERE tx_id = $2
	`, nfcId, txId)
	
	if err != nil {
		s.logger.Error("Failed to update transaction NFC ID", zap.String("txId", txId), zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		// Try public schema if blockchain schema fails (fallback)
		result, err = s.db.Exec(ctx, `
			UPDATE transactions 
			SET nfc_tag_id = $1
			WHERE tx_id = $2
		`, nfcId, txId)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return errors.New("transaction not found")
		}
	}

	return nil
}
