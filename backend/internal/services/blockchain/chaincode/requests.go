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

