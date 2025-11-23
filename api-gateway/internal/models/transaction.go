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

package models

import "time"

// TransactionStatus represents transaction status
type TransactionStatus string

const (
	TransactionStatusSubmitted TransactionStatus = "SUBMITTED"
	TransactionStatusValid     TransactionStatus = "VALID"
	TransactionStatusInvalid   TransactionStatus = "INVALID"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	ID            string                 `json:"id"`
	TxID          string                 `json:"txId"`
	ChannelName   string                 `json:"channelName"`
	ChaincodeName string                 `json:"chaincodeName"`
	FunctionName  string                 `json:"functionName"`
	Args          []string               `json:"args,omitempty"`
	TransientData map[string]interface{} `json:"transientData,omitempty"`
	UserID        string                 `json:"userId,omitempty"`
	APIKeyID      string                 `json:"apiKeyId,omitempty"`
	Status        TransactionStatus      `json:"status"`
	BlockNumber   uint64                 `json:"blockNumber,omitempty"`
	BlockHash     string                 `json:"blockHash,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	EndorsingOrgs []string               `json:"endorsingOrgs,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

// TransactionRequest represents a transaction submission request
type TransactionRequest struct {
	ChannelName   string                 `json:"channelName" validate:"required"`
	ChaincodeName string                 `json:"chaincodeName" validate:"required"`
	FunctionName  string                 `json:"functionName" validate:"required"`
	Args          []string               `json:"args,omitempty"`
	TransientData map[string]interface{} `json:"transientData,omitempty"`
	EndorsingOrgs []string               `json:"endorsingOrgs,omitempty"`
}

// TransactionResponse represents transaction submission response
type TransactionResponse struct {
	ID          string            `json:"id"`
	TxID        string            `json:"txId"`
	Status      TransactionStatus `json:"status"`
	BlockNumber uint64            `json:"blockNumber,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// TransactionReceipt represents transaction receipt
type TransactionReceipt struct {
	TxID          string            `json:"txId"`
	Status        TransactionStatus `json:"status"`
	BlockNumber   uint64            `json:"blockNumber"`
	BlockHash     string            `json:"blockHash,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	ChannelName   string            `json:"channelName"`
	ChaincodeName string            `json:"chaincodeName"`
	FunctionName  string            `json:"functionName"`
	Result        interface{}       `json:"result,omitempty"`
	ErrorMessage  string            `json:"errorMessage,omitempty"`
}

// TransactionListQuery represents query parameters for listing transactions
type TransactionListQuery struct {
	ChannelName   string            `json:"channelName,omitempty"`
	ChaincodeName string            `json:"chaincodeName,omitempty"`
	Status        TransactionStatus `json:"status,omitempty"`
	UserID        string            `json:"userId,omitempty"`
	Limit         int               `json:"limit,omitempty"`
	Offset        int               `json:"offset,omitempty"`
	StartTime     *time.Time        `json:"startTime,omitempty"`
	EndTime       *time.Time         `json:"endTime,omitempty"`
}



