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



