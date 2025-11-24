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

package chaincode

import "time"

// TeaBatch represents the tea batch data model from teaTraceCC
type TeaBatch struct {
	BatchID         string    `json:"batchId"`
	FarmName        string    `json:"farmName"`
	HarvestDate     string    `json:"harvestDate"`
	Certification   string    `json:"certification"`
	CertificateID   string    `json:"certificateId"`
	Status          string    `json:"status"`
	VerificationHash string   `json:"verificationHash"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// TeaPackage represents the tea package data model from teaTraceCC
type TeaPackage struct {
	PackageID     string     `json:"packageId"`
	BatchID       string     `json:"batchId"`
	BlockHash     string     `json:"blockHash"`
	TxID          string     `json:"txId"`
	Weight        float64    `json:"weight"`
	ProductionDate string    `json:"productionDate"`
	ExpiryDate    *string    `json:"expiryDate,omitempty"`
	QRCode        *string    `json:"qrCode,omitempty"`
	Status        string     `json:"status"`
	Owner         string     `json:"owner"`
	Timestamp     time.Time  `json:"timestamp"`
}

