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

