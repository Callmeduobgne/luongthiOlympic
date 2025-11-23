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

// ChannelInfo represents channel information
type ChannelInfo struct {
	Name        string   `json:"name"`
	Peers       []string `json:"peers"`
	Orderers    []string `json:"orderers"`
	Chaincodes  []string `json:"chaincodes"`
	BlockHeight uint64   `json:"blockHeight"`
}

// ChannelConfig represents channel configuration
type ChannelConfig struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Consortium      string                 `json:"consortium"`
	Organizations   []string               `json:"organizations"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	Policies        map[string]interface{} `json:"policies"`
	OrdererConfig   map[string]interface{} `json:"ordererConfig"`
	ApplicationConfig map[string]interface{} `json:"applicationConfig"`
}

// PeerInfo represents peer information
type PeerInfo struct {
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	MSPID        string   `json:"mspId"`
	Channels     []string `json:"channels"`
	Chaincodes   []string `json:"chaincodes"`
	Status       string   `json:"status"`
	BlockHeight  uint64   `json:"blockHeight,omitempty"`
}

// OrdererInfo represents orderer information
type OrdererInfo struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	MSPID     string `json:"mspId"`
	Status    string `json:"status"`
	IsLeader  bool   `json:"isLeader"`
}

// NetworkInfo represents overall network information
type NetworkInfo struct {
	Channels  []ChannelInfo  `json:"channels"`
	Peers     []PeerInfo     `json:"peers"`
	Orderers  []OrdererInfo  `json:"orderers"`
	MSPs      []string         `json:"msps"`
}

// BlockInfo represents block information
type BlockInfo struct {
	Number       uint64   `json:"number"`
	Hash         string   `json:"hash"`
	PreviousHash string   `json:"previousHash"`
	DataHash     string   `json:"dataHash"`
	Timestamp    string   `json:"timestamp"`
	Transactions []string `json:"transactions"`
	TransactionCount int  `json:"transactionCount"`
}

// TransactionInfo represents transaction information
type TransactionInfo struct {
	TxID        string                 `json:"txId"`
	ChannelID   string                 `json:"channelId"`
	Type        string                 `json:"type"`
	Timestamp   string                 `json:"timestamp"`
	Creator     string                 `json:"creator"`
	ChaincodeID string                 `json:"chaincodeId,omitempty"`
	Function    string                 `json:"function,omitempty"`
	Args        []string               `json:"args,omitempty"`
	ReadSet     map[string]interface{} `json:"readSet,omitempty"`
	WriteSet    map[string]interface{} `json:"writeSet,omitempty"`
	ValidationCode int32               `json:"validationCode"`
	BlockNumber  uint64                `json:"blockNumber"`
}

// ChaincodeInfoOnChannel represents chaincode information on a channel
type ChaincodeInfoOnChannel struct {
	Name                string   `json:"name"`
	Version             string   `json:"version"`
	Sequence            int64    `json:"sequence"`
	EndorsementPlugin   string   `json:"endorsementPlugin"`
	ValidationPlugin    string   `json:"validationPlugin"`
	InitRequired        bool     `json:"initRequired"`
	ApprovedOrganizations []string `json:"approvedOrganizations"`
}

// CAInfo represents Fabric CA information
type CAInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	MSPID   string `json:"mspId"`
	Status  string `json:"status"`
}

// NetworkTopology represents network topology
type NetworkTopology struct {
	Peers    []*PeerInfo    `json:"peers"`
	Orderers []*OrdererInfo `json:"orderers"`
	CAs      []*CAInfo      `json:"cas"`
	Channels []string       `json:"channels"`
	MSPs     []string       `json:"msps"`
}

// HealthStatus represents health status of a component
type HealthStatus struct {
	Component string `json:"component"`
	ID        string `json:"id"`
	Status    string `json:"status"` // healthy, unhealthy, unknown
	Timestamp string `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

