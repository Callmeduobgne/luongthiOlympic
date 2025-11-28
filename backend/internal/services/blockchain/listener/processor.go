package listener

import (
	"context"
	"time"

	"github.com/ibn-network/backend/internal/services/blockchain/db"
	"go.uber.org/zap"
)

func (s *Service) listenBlockEvents(ctx context.Context) {
	s.logger.Info("Starting block event listener loop")

	// Retry loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := s.processEvents(ctx); err != nil {
				s.logger.Error("Error processing block events, retrying in 5s...", zap.Error(err))
				
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}
		}
	}
}

func (s *Service) processEvents(ctx context.Context) error {
	// Start listening for block events
	events, err := s.network.BlockEvents(ctx)
	if err != nil {
		return err
	}

	for event := range events {
		blockNumber := event.BlockNumber()
		s.logger.Info("Received new block", zap.Uint64("block_number", blockNumber))

		for _, tx := range event.Transactions() {
			// Only process valid transactions
			if tx.ValidationCode() != 0 {
				continue
			}

			// Extract transaction details
			txID := tx.TransactionID()
			
			// In a full implementation, we would parse the transaction payload here
			// using fabric-protos-go to extract Function Name and Args.
			// For now, we save the essential metadata.
			
			s.processTransaction(ctx, txID, blockNumber)
		}
	}
	return nil
}

func (s *Service) processTransaction(ctx context.Context, txID string, blockNumber uint64) {
	tx := &db.Transaction{
		TxID:        txID,
		ChannelName: s.cfg.ChannelName,
		Status:      "VALID",
		BlockNumber: blockNumber,
		Timestamp:   time.Now(), // Approximate timestamp
		FunctionName: "SyncedTransaction", // Placeholder
		Args:        []string{},
	}
	
	if err := s.dbService.SaveTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to save transaction", zap.String("tx_id", txID), zap.Error(err))
	} else {
		s.logger.Info("Transaction synced to DB", zap.String("tx_id", txID))
	}
}
