// Copyright 2025 IBN Network (ICTU Blockchain Network)
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

// TeaPackageStatus represents the status of a tea package
type TeaPackageStatus string

const (
	PackageCreated  TeaPackageStatus = "CREATED"
	PackageVerified TeaPackageStatus = "VERIFIED"
	PackageSold     TeaPackageStatus = "SOLD"
	PackageExpired  TeaPackageStatus = "EXPIRED"
)

// IsValidPackageStatus checks if a status is valid
func IsValidPackageStatus(status string) bool {
	switch TeaPackageStatus(status) {
	case PackageCreated, PackageVerified, PackageSold, PackageExpired:
		return true
	}
	return false
}

// TeaPackage represents a tea package in the ledger
type TeaPackage struct {
	DocType        string           `json:"docType"`
	PackageID      string           `json:"packageId"`
	BatchID        string           `json:"batchId"`
	BlockHash      string           `json:"blockHash"`
	HashVersion    string           `json:"hashVersion,omitempty"`
	TxID           string           `json:"txId"`
	Weight         float64          `json:"weight"`
	ProductionDate string           `json:"productionDate"`
	ExpiryDate     string           `json:"expiryDate,omitempty"`
	QRCode         string           `json:"qrCode,omitempty"`
	Status         TeaPackageStatus `json:"status"`
	Owner          string           `json:"owner"`
	Timestamp      string           `json:"timestamp"`
}

// CreateTeaPackageInput represents input for creating a tea package
type CreateTeaPackageInput struct {
	PackageID      string  `json:"packageId"`
	BatchID        string  `json:"batchId"`
	Weight         float64 `json:"weight"`
	ProductionDate string  `json:"productionDate"`
	ExpiryDate     string  `json:"expiryDate,omitempty"`
	QRCode         string  `json:"qrCode,omitempty"`
}
