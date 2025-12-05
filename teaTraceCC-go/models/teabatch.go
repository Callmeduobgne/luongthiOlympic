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

// TeaBatchStatus represents the status of a tea batch
type TeaBatchStatus string

const (
	BatchCreated  TeaBatchStatus = "CREATED"
	BatchVerified TeaBatchStatus = "VERIFIED"
	BatchExpired  TeaBatchStatus = "EXPIRED"
)

// IsValidBatchStatus checks if a status is valid
func IsValidBatchStatus(status string) bool {
	switch TeaBatchStatus(status) {
	case BatchCreated, BatchVerified, BatchExpired:
		return true
	}
	return false
}

// TeaBatch represents a tea batch in the ledger
type TeaBatch struct {
	DocType        string         `json:"docType"`
	BatchID        string         `json:"batchId"`
	FarmLocation   string         `json:"farmLocation"`
	HarvestDate    string         `json:"harvestDate"`
	ProcessingInfo string         `json:"processingInfo"`
	QualityCert    string         `json:"qualityCert"`
	HashValue      string         `json:"hashValue"`
	Owner          string         `json:"owner"`
	Timestamp      string         `json:"timestamp"`
	Status         TeaBatchStatus `json:"status"`
}

// CreateTeaBatchInput represents input for creating a tea batch
type CreateTeaBatchInput struct {
	BatchID        string `json:"batchId"`
	FarmLocation   string `json:"farmLocation"`
	HarvestDate    string `json:"harvestDate"`
	ProcessingInfo string `json:"processingInfo"`
	QualityCert    string `json:"qualityCert"`
}
