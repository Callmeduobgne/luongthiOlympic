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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// WebhookClient handles webhook deliveries
type WebhookClient struct {
	httpClient *http.Client
	logger     *zap.Logger
}

// NewWebhookClient creates a new webhook client
func NewWebhookClient(logger *zap.Logger) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Deliver sends webhook event to endpoint
func (c *WebhookClient) Deliver(subscription *models.EventSubscription, event *models.ChaincodeEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequest("POST", subscription.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Type", "chaincode-event")
	req.Header.Set("X-Subscription-ID", subscription.ID)
	req.Header.Set("X-Channel-Name", event.ChannelName)
	req.Header.Set("X-Chaincode-Name", event.ChaincodeName)
	req.Header.Set("X-Event-Name", event.EventName)
	req.Header.Set("X-Transaction-ID", event.TransactionID)

	// Add signature if secret is provided
	if subscription.WebhookSecret != "" {
		signature := c.generateSignature(payload, subscription.WebhookSecret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (limited to 1KB)
	responseBody := make([]byte, 1024)
	n, _ := resp.Body.Read(responseBody)
	responseBodyStr := string(responseBody[:n])

	// Log delivery result
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.logger.Info("Webhook delivered successfully",
			zap.String("subscription_id", subscription.ID),
			zap.String("url", subscription.WebhookURL),
			zap.Int("status_code", resp.StatusCode),
		)
	} else {
		c.logger.Warn("Webhook delivery failed",
			zap.String("subscription_id", subscription.ID),
			zap.String("url", subscription.WebhookURL),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", responseBodyStr),
		)
	}

	return nil
}

// generateSignature generates HMAC-SHA256 signature for webhook payload
func (c *WebhookClient) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}


