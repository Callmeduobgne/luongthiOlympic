package fabric

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ibn-network/api-gateway/internal/models"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ContractService provides high-level contract operations
type ContractService struct {
	gateway *GatewayService
}

// NewContractService creates a new contract service
func NewContractService(gateway *GatewayService) *ContractService {
	return &ContractService{
		gateway: gateway,
	}
}

// CreateBatch creates a new tea batch on the blockchain
func (s *ContractService) CreateBatch(ctx context.Context, req *models.CreateBatchRequest) (*models.TeaBatch, error) {
	ctx, span := s.gateway.tracer.Start(ctx, "ContractService.CreateBatch")
	defer span.End()

	span.SetAttributes(
		attribute.String("batch.id", req.BatchID),
		attribute.String("farm.location", req.FarmLocation),
	)

	result, err := s.gateway.SubmitTransaction(
		ctx,
		"createBatch",
		req.BatchID,
		req.FarmLocation,
		req.HarvestDate,
		req.ProcessingInfo,
		req.QualityCert,
	)

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	var batch models.TeaBatch
	if err := json.Unmarshal(result, &batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	s.gateway.logger.Info("Batch created successfully",
		zap.String("batch.id", batch.BatchID),
	)

	return &batch, nil
}

// GetBatchInfo retrieves batch information from the blockchain
func (s *ContractService) GetBatchInfo(ctx context.Context, batchID string) (*models.TeaBatch, error) {
	ctx, span := s.gateway.tracer.Start(ctx, "ContractService.GetBatchInfo")
	defer span.End()

	span.SetAttributes(attribute.String("batch.id", batchID))

	result, err := s.gateway.EvaluateTransaction(ctx, "getBatchInfo", batchID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get batch info: %w", err)
	}

	var batch models.TeaBatch
	if err := json.Unmarshal(result, &batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	return &batch, nil
}

// VerifyBatch verifies a batch hash
func (s *ContractService) VerifyBatch(ctx context.Context, batchID, hashInput string) (*models.VerifyBatchResponse, error) {
	ctx, span := s.gateway.tracer.Start(ctx, "ContractService.VerifyBatch")
	defer span.End()

	span.SetAttributes(attribute.String("batch.id", batchID))

	result, err := s.gateway.SubmitTransaction(ctx, "verifyBatch", batchID, hashInput)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to verify batch: %w", err)
	}

	var response models.VerifyBatchResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verify response: %w", err)
	}

	s.gateway.logger.Info("Batch verification completed",
		zap.String("batch.id", batchID),
		zap.Bool("is_valid", response.IsValid),
	)

	return &response, nil
}

// UpdateBatchStatus updates the status of a batch
func (s *ContractService) UpdateBatchStatus(ctx context.Context, batchID, status string) (*models.TeaBatch, error) {
	ctx, span := s.gateway.tracer.Start(ctx, "ContractService.UpdateBatchStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("batch.id", batchID),
		attribute.String("status", status),
	)

	result, err := s.gateway.SubmitTransaction(ctx, "updateBatchStatus", batchID, status)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update batch status: %w", err)
	}

	var batch models.TeaBatch
	if err := json.Unmarshal(result, &batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	s.gateway.logger.Info("Batch status updated",
		zap.String("batch.id", batchID),
		zap.String("new_status", status),
	)

	return &batch, nil
}

