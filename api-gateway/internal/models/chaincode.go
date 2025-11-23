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

// ChaincodeInfo represents chaincode information
type ChaincodeInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
	Package string `json:"package,omitempty"`
}

// InstalledChaincode represents an installed chaincode
type InstalledChaincode struct {
	PackageID string `json:"packageId"`
	Label     string `json:"label"`
	Chaincode ChaincodeInfo `json:"chaincode"`
}

// CommittedChaincode represents a committed chaincode on a channel
type CommittedChaincode struct {
	Name                string   `json:"name"`
	Version             string   `json:"version"`
	Sequence            int64    `json:"sequence"`
	EndorsementPlugin   string   `json:"endorsementPlugin"`
	ValidationPlugin    string   `json:"validationPlugin"`
	InitRequired        bool     `json:"initRequired"`
	Collections         []string `json:"collections,omitempty"`
	ApprovedOrganizations []string `json:"approvedOrganizations"`
}

// InstallChaincodeRequest represents a chaincode installation request
type InstallChaincodeRequest struct {
	PackagePath string `json:"packagePath" validate:"required"`
	Label       string `json:"label,omitempty"`
}

// ApproveChaincodeRequest represents a chaincode approval request
type ApproveChaincodeRequest struct {
	ChannelName         string   `json:"channelName" validate:"required"` // Channel name for approval
	Name                string   `json:"name" validate:"required"`
	Version             string   `json:"version" validate:"required"`
	Sequence            int64    `json:"sequence" validate:"required,min=1"`
	PackageID           string   `json:"packageId,omitempty"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

// CommitChaincodeRequest represents a chaincode commit request
type CommitChaincodeRequest struct {
	ChannelName         string   `json:"channelName" validate:"required"` // Channel name for commit
	Name                string   `json:"name" validate:"required"`
	Version             string   `json:"version" validate:"required"`
	Sequence            int64    `json:"sequence" validate:"required,min=1"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

// UpgradeChaincodeRequest represents a chaincode upgrade request
type UpgradeChaincodeRequest struct {
	Name     string `json:"name" validate:"required"`
	Version  string `json:"version" validate:"required"`
	Sequence int64  `json:"sequence" validate:"required,min=1"`
}

// InvokeRequest represents a generic chaincode invocation request
type InvokeRequest struct {
	Function      string            `json:"function" validate:"required"`
	Args          []string          `json:"args,omitempty"`
	Transient     map[string]string `json:"transient,omitempty"` // base64 encoded values
	EndorsingOrgs []string          `json:"endorsingOrgs,omitempty"` // Future: custom endorsing organizations
}

// QueryRequest represents a generic chaincode query request
type QueryRequest struct {
	Function  string            `json:"function" validate:"required"`
	Args      []string          `json:"args,omitempty"`
	Transient map[string]string `json:"transient,omitempty"` // base64 encoded values
}

// InvokeResponse represents chaincode invocation response
type InvokeResponse struct {
	TxID        string      `json:"txId"`
	Result      interface{} `json:"result,omitempty"`
	Status      string      `json:"status"`
	BlockNumber uint64      `json:"blockNumber,omitempty"`
	Timestamp   string      `json:"timestamp,omitempty"`
}

// QueryResponse represents chaincode query response
type QueryResponse struct {
	Result    interface{} `json:"result,omitempty"`
	Timestamp string      `json:"timestamp"`
}

