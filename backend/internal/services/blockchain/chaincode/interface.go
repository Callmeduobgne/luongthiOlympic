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
	"context"
)

// TeaTraceService is a common interface for TeaTrace services
// This allows both TeaTraceService (direct Fabric) and TeaTraceServiceViaGateway to be used
type TeaTraceService interface {
	CreateBatch(ctx context.Context, batchID, farmName, harvestDate, certification, certificateID string) (string, error)
	GetBatch(ctx context.Context, batchID string) (*TeaBatch, error)
	GetAllBatches(ctx context.Context) ([]*TeaBatch, error)
	VerifyBatch(ctx context.Context, batchID, verificationHash string) (string, error)
	UpdateBatchStatus(ctx context.Context, batchID, status string) (string, error)
	CreatePackage(ctx context.Context, packageID, batchID string, weight float64, productionDate, expiryDate string) (string, error)
	GetPackage(ctx context.Context, packageID string) (*TeaPackage, error)
	HealthCheck(ctx context.Context) error
}



