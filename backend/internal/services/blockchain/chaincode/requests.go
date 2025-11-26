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

import (
	"time"
	
	"github.com/ibn-network/backend/internal/utils/merkle"
)

// CreateBatchRequest represents a request to create a tea batch
type CreateBatchRequest struct {
	BatchID       string `json:"batch_id" validate:"required,min=1,max=50"`
	FarmName      string `json:"farm_name" validate:"required,min=1,max=200"`
	HarvestDate   string `json:"harvest_date" validate:"required"`
	Certification string `json:"certification,omitempty"`
	CertificateID string `json:"certificate_id,omitempty"`
}

// VerifyBatchRequest represents a request to verify a tea batch
type VerifyBatchRequest struct {
	BatchID          string `json:"batch_id" validate:"required"`
	VerificationHash string `json:"verification_hash" validate:"required,min=64,max=64"`
}

// UpdateBatchStatusRequest represents a request to update batch status
type UpdateBatchStatusRequest struct {
	BatchID string `json:"batch_id" validate:"required"`
	Status  string `json:"status" validate:"required,oneof=CREATED VERIFIED PROCESSED SHIPPED DELIVERED"`
}

// VerifyByHashRequest represents a request to verify product by hash/blockhash
// Now supports Merkle proof for fast verification
type VerifyByHashRequest struct {
	Hash        string              `json:"hash" validate:"required,min=1,max=128"`
	MerkleProof []merkle.ProofStep  `json:"merkleProof,omitempty"` // Optional: Merkle proof for fast verification
	MerkleRoot  string              `json:"merkleRoot,omitempty"`  // Optional: Merkle root hash
	BlockNumber uint64              `json:"blockNumber,omitempty"` // Optional: Block number for verification
}


// VerifyByHashResponse represents the response for hash verification
// Extended with Merkle proof metadata
type VerifyByHashResponse struct {
	IsValid            bool      `json:"is_valid"`
	Message            string    `json:"message"`
	TransactionID      string    `json:"transaction_id,omitempty"`
	BatchID            string    `json:"batch_id,omitempty"`
	PackageID          string    `json:"package_id,omitempty"`
	BlockNumber        uint64    `json:"block_number,omitempty"`        // NEW: Block number for verification
	EntityType         string    `json:"entity_type,omitempty"`         // "transaction", "batch", "package"
	VerifiedAt         time.Time `json:"verified_at"`                   // NEW: Verification timestamp
	VerificationMethod string          `json:"verification_method,omitempty"` // NEW: "merkle_proof" or "blockchain_query"
	ProductDetails     *ProductDetails `json:"product_details,omitempty"`     // NEW: Full product details
}

// ProductDetails contains detailed information about the verified product
type ProductDetails struct {
	FarmLocation   string  `json:"farm_location,omitempty"`
	HarvestDate    string  `json:"harvest_date,omitempty"`
	ProductionDate string  `json:"production_date,omitempty"`
	ExpiryDate     string  `json:"expiry_date,omitempty"`
	Weight         float64 `json:"weight,omitempty"`
	ProcessingInfo string  `json:"processing_info,omitempty"`
	QualityCert    string  `json:"quality_cert,omitempty"`
	Status         string  `json:"status,omitempty"`
}


