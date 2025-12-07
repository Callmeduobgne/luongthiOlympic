package teatrace

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"github.com/ibn-network/backend/internal/infrastructure/gateway"
	"github.com/ibn-network/backend/internal/services/analytics/metrics"
	"go.uber.org/zap"
)

// Service handles Tea Traceability operations
type Service struct {
	gatewayClient *gateway.Client
	metrics       *metrics.Service
	logger        *zap.Logger
}

// NewService creates a new Tea Traceability service
func NewService(gatewayClient *gateway.Client, metrics *metrics.Service, logger *zap.Logger) *Service {
	return &Service{
		gatewayClient: gatewayClient,
		metrics:       metrics,
		logger:        logger,
	}
}

// VerifyByHash verifies an entity by its hash via Gateway
func (s *Service) VerifyByHash(ctx context.Context, hash string) (map[string]interface{}, error) {
	start := time.Now()
	var err error
	
	// Record metric on exit
	defer func() {
		s.metrics.RecordBlockchainTransaction("verify_by_hash", time.Since(start), err == nil)
	}()

	reqBody := map[string]string{
		"hash": hash,
	}

	respBody, err := s.gatewayClient.Post(ctx, "/api/v1/teatrace/verify-by-hash", reqBody)
	if err != nil {
		s.logger.Error("Failed to verify by hash via Gateway", zap.Error(err))
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	var apiResp struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
		Error   interface{}            `json:"error"`
	}

	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		err = fmt.Errorf("verification failed: %v", apiResp.Error)
		return nil, err
	}

	return apiResp.Data, nil
}
// CreateBatch creates a new tea batch via Gateway
func (s *Service) CreateBatch(ctx context.Context, batchID, farmName, harvestDate, certification, certificateID string) (map[string]interface{}, error) {
	start := time.Now()
	var err error
	
	// Record metric on exit
	defer func() {
		s.metrics.RecordBlockchainTransaction("create_batch", time.Since(start), err == nil)
	}()

	reqBody := map[string]interface{}{
		"function": "createBatch",
		"args": []string{
			batchID,
			farmName,
			harvestDate,
			certification,
			certificateID,
		},
	}

	// Call Gateway to invoke chaincode
	// Endpoint: /api/v1/channels/{channel}/chaincodes/{chaincode}/invoke
	// We need to know the channel and chaincode names. 
	// For now, hardcoding "ibnchannel" and "teaTraceCC" as seen in main.go
	respBody, err := s.gatewayClient.Post(ctx, "/api/v1/channels/ibnchannel/chaincodes/teaTraceCC/invoke", reqBody)
	if err != nil {
		s.logger.Error("Failed to create batch via Gateway", zap.Error(err))
		return nil, fmt.Errorf("create batch failed: %w", err)
	}

	var apiResp struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
		Error   interface{}            `json:"error"`
	}

	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		err = fmt.Errorf("create batch failed: %v", apiResp.Error)
		return nil, err
	}

	return apiResp.Data, nil
}

// GetBatch retrieves a tea batch by ID via Gateway
func (s *Service) GetBatch(ctx context.Context, batchID string) (map[string]interface{}, error) {
	start := time.Now()
	var err error
	
	// Record metric on exit
	defer func() {
		s.metrics.RecordBlockchainTransaction("get_batch", time.Since(start), err == nil)
	}()

	// Call Gateway to query chaincode
	respBody, err := s.gatewayClient.QueryChaincode(ctx, "ibnchannel", "teaTraceCC", "getBatchInfo", []string{batchID})
	if err != nil {
		s.logger.Error("Failed to get batch via Gateway", zap.Error(err))
		return nil, fmt.Errorf("get batch failed: %w", err)
	}

	if len(respBody) == 0 || string(respBody) == "null" {
		return nil, nil
	}

	var batch map[string]interface{}
	if err := json.Unmarshal(respBody, &batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	return batch, nil
}

// CreatePackage creates a new tea package via Gateway
func (s *Service) CreatePackage(ctx context.Context, packageID, batchID string, weight float64, productionDate, expiryDate string) (map[string]interface{}, error) {
	start := time.Now()
	var err error
	
	// Record metric on exit
	defer func() {
		s.metrics.RecordBlockchainTransaction("create_package", time.Since(start), err == nil)
	}()

	reqBody := map[string]interface{}{
		"function": "createPackage",
		"args": []string{
			packageID,
			batchID,
			fmt.Sprintf("%f", weight),
			productionDate,
			expiryDate,
			"", // qrCode
		},
	}

	// Call Gateway to invoke chaincode
	// Endpoint: /api/v1/channels/{channel}/chaincodes/{chaincode}/invoke
	respBody, err := s.gatewayClient.Post(ctx, "/api/v1/channels/ibnchannel/chaincodes/teaTraceCC/invoke", reqBody)
	if err != nil {
		s.logger.Error("Failed to create package via Gateway", zap.Error(err))
		return nil, fmt.Errorf("create package failed: %w", err)
	}

	var apiResp struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
		Error   interface{}            `json:"error"`
	}

	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		err = fmt.Errorf("create package failed: %v", apiResp.Error)
		return nil, err
	}

	return apiResp.Data, nil
}

// GetPackage retrieves a tea package by ID via Gateway
func (s *Service) GetPackage(ctx context.Context, packageID string) (map[string]interface{}, error) {
	start := time.Now()
	var err error
	
	// Record metric on exit
	defer func() {
		s.metrics.RecordBlockchainTransaction("get_package", time.Since(start), err == nil)
	}()

	// Call Gateway to query chaincode
	respBody, err := s.gatewayClient.QueryChaincode(ctx, "ibnchannel", "teaTraceCC", "getPackageInfo", []string{packageID})
	if err != nil {
		s.logger.Error("Failed to get package via Gateway", zap.Error(err))
		return nil, fmt.Errorf("get package failed: %w", err)
	}

	if len(respBody) == 0 || string(respBody) == "null" {
		return nil, nil
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(respBody, &pkg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	return pkg, nil
}

// GetPackagesByBatch retrieves all packages for a specific batch via Gateway
func (s *Service) GetPackagesByBatch(ctx context.Context, batchID string, limit, offset int) ([]map[string]interface{}, int, error) {
	start := time.Now()
	var err error
	
	// Record metric on exit
	defer func() {
		s.metrics.RecordBlockchainTransaction("get_packages_by_batch", time.Since(start), err == nil)
	}()

	// Call Gateway to query chaincode
	// getPackagesByBatch only accepts batchId due to Fabric validation
	// We'll get all packages and filter/paginate in backend
	respBody, err := s.gatewayClient.QueryChaincode(ctx, "ibnchannel", "teaTraceCC", "getPackagesByBatch", []string{
		batchID,
	})
	if err != nil {
		s.logger.Error("Failed to get packages by batch via Gateway", zap.Error(err))
		return nil, 0, fmt.Errorf("get packages by batch failed: %w", err)
	}

	if len(respBody) == 0 || string(respBody) == "null" {
		return []map[string]interface{}{}, 0, nil
	}

	// Parse response - chaincode returns {packages: [...], total: number}
	var result struct {
		Packages []map[string]interface{} `json:"packages"`
		Total    int                      `json:"total"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		s.logger.Error("Failed to unmarshal packages response", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to unmarshal packages: %w", err)
	}

	// Apply pagination in backend
	allPackages := result.Packages
	totalCount := len(allPackages)
	
	// Calculate slice bounds
	sliceStart := offset
	if sliceStart > totalCount {
		sliceStart = totalCount
	}
	
	sliceEnd := sliceStart + limit
	if sliceEnd > totalCount {
		sliceEnd = totalCount
	}
	
	// Slice the results
	pagedPackages := allPackages[sliceStart:sliceEnd]

	return pagedPackages, totalCount, nil
}
