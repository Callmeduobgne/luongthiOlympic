package event

import (
	"encoding/json"
	"fmt"

	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// BatchCreatedHandler handles BatchCreated events
func BatchCreatedHandler(logger *zap.Logger) EventHandler {
	return func(event *ChainEvent) error {
		var batch models.TeaBatch
		if err := json.Unmarshal(event.Payload, &batch); err != nil {
			return fmt.Errorf("failed to unmarshal batch: %w", err)
		}

		logger.Info("Batch created event received",
			zap.String("batchId", batch.BatchID),
			zap.String("txId", event.TransactionID),
			zap.Uint64("blockNumber", event.BlockNumber),
		)

		// Handle event (e.g., update cache, send notification, etc.)
		return nil
	}
}

// BatchVerifiedHandler handles BatchVerified events
func BatchVerifiedHandler(logger *zap.Logger) EventHandler {
	return func(event *ChainEvent) error {
		var batch models.TeaBatch
		if err := json.Unmarshal(event.Payload, &batch); err != nil {
			return fmt.Errorf("failed to unmarshal batch: %w", err)
		}

		logger.Info("Batch verified event received",
			zap.String("batchId", batch.BatchID),
			zap.String("status", string(batch.Status)),
			zap.String("txId", event.TransactionID),
		)

		// Handle event (e.g., update analytics, send notification, etc.)
		return nil
	}
}

// BatchStatusUpdatedHandler handles BatchStatusUpdated events
func BatchStatusUpdatedHandler(logger *zap.Logger) EventHandler {
	return func(event *ChainEvent) error {
		var batch models.TeaBatch
		if err := json.Unmarshal(event.Payload, &batch); err != nil {
			return fmt.Errorf("failed to unmarshal batch: %w", err)
		}

		logger.Info("Batch status updated event received",
			zap.String("batchId", batch.BatchID),
			zap.String("status", string(batch.Status)),
			zap.String("txId", event.TransactionID),
		)

		// Handle event
		return nil
	}
}

