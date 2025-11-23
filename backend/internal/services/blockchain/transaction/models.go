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
	"time"

	"github.com/google/uuid"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	ID            uuid.UUID              `json:"id"`
	TxID          string                 `json:"tx_id"`
	UserID        uuid.UUID              `json:"user_id"`
	ChannelID     string                 `json:"channel_id"`
	ChaincodeID   string                 `json:"chaincode_id"`
	FunctionName  string                 `json:"function_name"`
	Args          []string               `json:"args"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Status        string                 `json:"status"`
	BlockNumber   *uint64                `json:"block_number,omitempty"`
	TxIndex       *uint32                `json:"tx_index,omitempty"`
	ResponseData  string                 `json:"response_data,omitempty"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	SubmittedAt   time.Time              `json:"submitted_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	ValidationCode *int32                `json:"validation_code,omitempty"`
}

// SubmitTransactionRequest represents a transaction submission request
type SubmitTransactionRequest struct {
	ChannelID    string                 `json:"channel_id" validate:"required"`
	ChaincodeID  string                 `json:"chaincode_id" validate:"required"`
	FunctionName string                 `json:"function_name" validate:"required"`
	Args         []string               `json:"args"`
	Transient    map[string][]byte      `json:"transient,omitempty"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
}

// QueryTransactionRequest represents a transaction query request
type QueryTransactionRequest struct {
	ChannelID    string   `json:"channel_id" validate:"required"`
	ChaincodeID  string   `json:"chaincode_id" validate:"required"`
	FunctionName string   `json:"function_name" validate:"required"`
	Args         []string `json:"args"`
}

// QueryTransactionsRequest represents a query for multiple transactions
type QueryTransactionsRequest struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	ChannelID   *string    `json:"channel_id,omitempty"`
	ChaincodeID *string    `json:"chaincode_id,omitempty"`
	Status      *string    `json:"status,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// Transaction status constants
const (
	StatusPending   = "pending"
	StatusSubmitted = "submitted"
	StatusCommitted = "committed"
	StatusFailed    = "failed"
	StatusTimeout   = "timeout"
)

// TransactionStatusHistory tracks status changes
type TransactionStatusHistory struct {
	ID            uuid.UUID  `json:"id"`
	TransactionID uuid.UUID  `json:"transaction_id"`
	Status        string     `json:"status"`
	Details       string     `json:"details,omitempty"`
	Timestamp     time.Time  `json:"timestamp"`
}

