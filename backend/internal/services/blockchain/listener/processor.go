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
		// event is *common.Block from fabric-protos-go
		blockNumber := event.Header.Number
		s.logger.Info("Received new block", 
			zap.Uint64("block_number", blockNumber),
			zap.Int("transaction_count", len(event.Data.Data)),
		)

		// Process transactions from block data
		// block.Data.Data is []*common.Envelope
		// Note: Extracting transaction ID from envelope requires parsing protobuf payload
		// For now, we log the block but skip detailed transaction processing
		// as it requires complex protobuf parsing (common.Envelope -> common.Payload -> common.ChannelHeader)
		// This listener is primarily for monitoring block events
		// Detailed transaction data should come from Gateway transaction submission responses
		
		// Log block received (transactions will be saved when submitted via Gateway)
		s.logger.Debug("Block event received",
			zap.Uint64("block_number", blockNumber),
			zap.Int("envelope_count", len(event.Data.Data)),
		)
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
