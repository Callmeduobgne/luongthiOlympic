package chaincode

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"teaTraceCC/models"
	"teaTraceCC/utils"
)

// TeaTraceContract implements the tea traceability chaincode
type TeaTraceContract struct {
	contractapi.Contract
	mspConfig models.MSPConfig
}

// NewTeaTraceContract creates a new TeaTraceContract
func NewTeaTraceContract() *TeaTraceContract {
	return &TeaTraceContract{
		mspConfig: models.DefaultMSPConfig(),
	}
}

// ==================== Batch Functions ====================

// CreateBatch creates a new tea batch
func (c *TeaTraceContract) CreateBatch(ctx contractapi.TransactionContextInterface,
	batchID, farmLocation, harvestDate, processingInfo, qualityCert string) (*models.TeaBatch, error) {

	// Input validation
	if err := utils.ValidateBatchID(batchID); err != nil {
		return nil, err
	}
	if err := utils.ValidateString(farmLocation, "Farm location", 200); err != nil {
		return nil, err
	}
	if err := utils.ValidateDate(harvestDate); err != nil {
		return nil, err
	}
	if err := utils.ValidateString(processingInfo, "Processing info", 1000); err != nil {
		return nil, err
	}
	if err := utils.ValidateString(qualityCert, "Quality certificate", 100); err != nil {
		return nil, err
	}

	// Check authorization
	if err := c.ensureOrg(ctx, []string{c.mspConfig.Farmer}, "create batches"); err != nil {
		return nil, err
	}

	// Check if batch already exists
	exists, err := c.batchExists(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("batch with id '%s' already exists", batchID)
	}

	// Get owner from client identity
	owner, err := c.getOwner(ctx)
	if err != nil {
		return nil, err
	}

	// Generate hash
	input := &models.CreateTeaBatchInput{
		BatchID:        batchID,
		FarmLocation:   farmLocation,
		HarvestDate:    harvestDate,
		ProcessingInfo: processingInfo,
		QualityCert:    qualityCert,
	}
	hashValue := utils.GenerateBatchHash(input)

	// Create batch
	batch := &models.TeaBatch{
		DocType:        "batch",
		BatchID:        batchID,
		FarmLocation:   farmLocation,
		HarvestDate:    harvestDate,
		ProcessingInfo: processingInfo,
		QualityCert:    qualityCert,
		HashValue:      hashValue,
		Owner:          owner,
		Timestamp:      c.getCurrentTimestamp(ctx),
		Status:         models.BatchCreated,
	}

	// Store batch
	batchJSON, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch: %v", err)
	}

	if err := ctx.GetStub().PutState(batchID, batchJSON); err != nil {
		return nil, fmt.Errorf("failed to store batch: %v", err)
	}

	return batch, nil
}

// VerifyBatch verifies a batch by comparing hashes
func (c *TeaTraceContract) VerifyBatch(ctx contractapi.TransactionContextInterface,
	batchID, hashInput string) (*VerifyBatchResult, error) {

	if err := utils.ValidateBatchID(batchID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(hashInput) == "" {
		return nil, fmt.Errorf("hash input cannot be empty")
	}

	// Check authorization
	if err := c.ensureOrg(ctx, []string{c.mspConfig.Verifier, c.mspConfig.Admin, c.mspConfig.Farmer}, "verify batches"); err != nil {
		return nil, err
	}

	batch, err := c.getBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}

	isValid := utils.VerifyHash(batch.HashValue, hashInput)

	// Update status if valid and not already verified
	if isValid && batch.Status != models.BatchVerified {
		batch.Status = models.BatchVerified
		batch.Timestamp = c.getCurrentTimestamp(ctx)

		batchJSON, err := json.Marshal(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal batch: %v", err)
		}

		if err := ctx.GetStub().PutState(batchID, batchJSON); err != nil {
			return nil, fmt.Errorf("failed to update batch: %v", err)
		}
	}

	return &VerifyBatchResult{
		IsValid: isValid,
		Batch:   batch,
	}, nil
}

// VerifyBatchResult is the result of VerifyBatch
type VerifyBatchResult struct {
	IsValid bool             `json:"isValid"`
	Batch   *models.TeaBatch `json:"batch"`
}

// GetBatchInfo retrieves batch information
func (c *TeaTraceContract) GetBatchInfo(ctx contractapi.TransactionContextInterface, batchID string) (*models.TeaBatch, error) {
	if err := utils.ValidateBatchID(batchID); err != nil {
		return nil, err
	}

	batchBytes, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch: %v", err)
	}
	if batchBytes == nil {
		return nil, nil // Not found
	}

	var batch models.TeaBatch
	if err := json.Unmarshal(batchBytes, &batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %v", err)
	}

	return &batch, nil
}

// GetAllBatches retrieves all batches with pagination
func (c *TeaTraceContract) GetAllBatches(ctx contractapi.TransactionContextInterface,
	limitStr, offsetStr string) (*BatchListResult, error) {

	limit, offset, err := c.parsePagination(limitStr, offsetStr)
	if err != nil {
		return nil, err
	}

	// CouchDB query
	queryString := fmt.Sprintf(`{
		"selector": {"docType": "batch"},
		"sort": [{"timestamp": "desc"}],
		"limit": %d,
		"skip": %d
	}`, limit, offset)

	return c.queryBatches(ctx, queryString)
}

// GetBatchesByStatus retrieves batches by status with pagination
func (c *TeaTraceContract) GetBatchesByStatus(ctx contractapi.TransactionContextInterface,
	status, limitStr, offsetStr string) (*BatchListResult, error) {

	normalizedStatus := strings.ToUpper(status)
	if !models.IsValidBatchStatus(normalizedStatus) {
		return nil, fmt.Errorf("invalid status '%s'. Allowed: CREATED, VERIFIED, EXPIRED", status)
	}

	limit, offset, err := c.parsePagination(limitStr, offsetStr)
	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{
		"selector": {"docType": "batch", "status": "%s"},
		"sort": [{"timestamp": "desc"}],
		"limit": %d,
		"skip": %d
	}`, normalizedStatus, limit, offset)

	return c.queryBatches(ctx, queryString)
}

// GetBatchesByOwner retrieves batches by owner with pagination
func (c *TeaTraceContract) GetBatchesByOwner(ctx contractapi.TransactionContextInterface,
	owner, limitStr, offsetStr string) (*BatchListResult, error) {

	if strings.TrimSpace(owner) == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}

	limit, offset, err := c.parsePagination(limitStr, offsetStr)
	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{
		"selector": {"docType": "batch", "owner": "%s"},
		"sort": [{"timestamp": "desc"}],
		"limit": %d,
		"skip": %d
	}`, owner, limit, offset)

	return c.queryBatches(ctx, queryString)
}

// GetBatchHistory retrieves the history of a batch
func (c *TeaTraceContract) GetBatchHistory(ctx contractapi.TransactionContextInterface, batchID string) ([]*models.TeaBatch, error) {
	if err := utils.ValidateBatchID(batchID); err != nil {
		return nil, err
	}

	historyIterator, err := ctx.GetStub().GetHistoryForKey(batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %v", err)
	}
	defer historyIterator.Close()

	var history []*models.TeaBatch
	for historyIterator.HasNext() {
		modification, err := historyIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate history: %v", err)
		}

		if modification.IsDelete {
			continue
		}

		var batch models.TeaBatch
		if err := json.Unmarshal(modification.Value, &batch); err != nil {
			return nil, fmt.Errorf("failed to unmarshal history: %v", err)
		}
		history = append(history, &batch)
	}

	// Reverse to get oldest first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// UpdateBatchStatus updates the status of a batch
func (c *TeaTraceContract) UpdateBatchStatus(ctx contractapi.TransactionContextInterface,
	batchID, status string) (*models.TeaBatch, error) {

	if err := utils.ValidateBatchID(batchID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(status) == "" {
		return nil, fmt.Errorf("status cannot be empty")
	}

	if err := c.ensureOrg(ctx, []string{c.mspConfig.Farmer, c.mspConfig.Admin}, "update batch status"); err != nil {
		return nil, err
	}

	normalizedStatus := strings.ToUpper(status)
	if !models.IsValidBatchStatus(normalizedStatus) {
		return nil, fmt.Errorf("invalid status '%s'. Allowed: CREATED, VERIFIED, EXPIRED", status)
	}

	batch, err := c.getBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}

	batch.Status = models.TeaBatchStatus(normalizedStatus)
	batch.Timestamp = c.getCurrentTimestamp(ctx)

	batchJSON, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch: %v", err)
	}

	if err := ctx.GetStub().PutState(batchID, batchJSON); err != nil {
		return nil, fmt.Errorf("failed to update batch: %v", err)
	}

	return batch, nil
}

// BatchListResult is the result of batch list queries
type BatchListResult struct {
	Batches []*models.TeaBatch `json:"batches"`
	Total   int                `json:"total"`
}

// ==================== Package Functions ====================

// CreatePackage creates a new tea package from a batch
func (c *TeaTraceContract) CreatePackage(ctx contractapi.TransactionContextInterface,
	packageID, batchID, weightStr, productionDate, expiryDate, qrCode string) (*models.TeaPackage, error) {

	// Input validation
	if err := utils.ValidatePackageID(packageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateBatchID(batchID); err != nil {
		return nil, err
	}

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return nil, fmt.Errorf("weight must be a valid number")
	}
	if err := utils.ValidateWeight(weight); err != nil {
		return nil, err
	}

	if err := utils.ValidateDate(productionDate); err != nil {
		return nil, err
	}

	if expiryDate != "" {
		if err := utils.ValidateDate(expiryDate); err != nil {
			return nil, err
		}
		if err := utils.ValidateDateRange(productionDate, expiryDate); err != nil {
			return nil, err
		}
	}

	if len(qrCode) > 500 {
		return nil, fmt.Errorf("QR code must be less than 500 characters")
	}

	// Check authorization
	if err := c.ensureOrg(ctx, []string{c.mspConfig.Farmer, c.mspConfig.Admin}, "create packages"); err != nil {
		return nil, err
	}

	// Verify batch exists
	if _, err := c.getBatch(ctx, batchID); err != nil {
		return nil, err
	}

	// Check package doesn't exist
	exists, err := c.packageExists(ctx, packageID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("package with id '%s' already exists", packageID)
	}

	// Get transaction ID
	txID := ctx.GetStub().GetTxID()

	// Get hash secret from environment
	hashSecret := os.Getenv("HASH_SECRET")

	// Generate blockHash
	blockHash := utils.GeneratePackageBlockHash(packageID, batchID, weight, productionDate, txID, hashSecret)

	// Get owner
	owner, err := c.getOwner(ctx)
	if err != nil {
		return nil, err
	}

	// Determine hash version
	hashVersion := "v1"
	if hashSecret != "" {
		hashVersion = "v2"
	}

	pkg := &models.TeaPackage{
		DocType:        "package",
		PackageID:      packageID,
		BatchID:        batchID,
		BlockHash:      blockHash,
		HashVersion:    hashVersion,
		TxID:           txID,
		Weight:         weight,
		ProductionDate: productionDate,
		ExpiryDate:     expiryDate,
		QRCode:         qrCode,
		Status:         models.PackageCreated,
		Owner:          owner,
		Timestamp:      c.getCurrentTimestamp(ctx),
	}

	pkgJSON, err := json.Marshal(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %v", err)
	}

	// Store with simple key
	if err := ctx.GetStub().PutState(packageID, pkgJSON); err != nil {
		return nil, fmt.Errorf("failed to store package: %v", err)
	}

	// Store with composite key for efficient batch queries
	compositeKey, err := ctx.GetStub().CreateCompositeKey("PACKAGE", []string{batchID, packageID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}
	if err := ctx.GetStub().PutState(compositeKey, pkgJSON); err != nil {
		return nil, fmt.Errorf("failed to store package with composite key: %v", err)
	}

	return pkg, nil
}

// VerifyPackage verifies a package by comparing blockhash
func (c *TeaTraceContract) VerifyPackage(ctx contractapi.TransactionContextInterface,
	packageID, providedBlockHash string) (*VerifyPackageResult, error) {

	if err := utils.ValidatePackageID(packageID); err != nil {
		return nil, err
	}

	pkg, err := c.getPackage(ctx, packageID)
	if err != nil {
		return nil, err
	}

	// If no blockHash provided, just return package info
	if providedBlockHash == "" {
		return &VerifyPackageResult{
			IsValid: true,
			Package: pkg,
		}, nil
	}

	// Verify hash
	isValid := false
	hashSecret := os.Getenv("HASH_SECRET")

	if pkg.HashVersion == "v2" {
		regeneratedHash := utils.GeneratePackageBlockHash(
			pkg.PackageID, pkg.BatchID, pkg.Weight, pkg.ProductionDate, pkg.TxID, hashSecret)
		isValid = regeneratedHash == providedBlockHash
	} else {
		regeneratedHash := utils.GeneratePackageBlockHash(
			pkg.PackageID, pkg.BatchID, pkg.Weight, pkg.ProductionDate, pkg.TxID, "")
		isValid = regeneratedHash == providedBlockHash
	}

	// Update status if valid
	if isValid && pkg.Status == models.PackageCreated {
		pkg.Status = models.PackageVerified
		pkg.Timestamp = c.getCurrentTimestamp(ctx)

		pkgJSON, err := json.Marshal(pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal package: %v", err)
		}

		if err := ctx.GetStub().PutState(packageID, pkgJSON); err != nil {
			return nil, fmt.Errorf("failed to update package: %v", err)
		}

		compositeKey, err := ctx.GetStub().CreateCompositeKey("PACKAGE", []string{pkg.BatchID, packageID})
		if err != nil {
			return nil, fmt.Errorf("failed to create composite key: %v", err)
		}
		if err := ctx.GetStub().PutState(compositeKey, pkgJSON); err != nil {
			return nil, fmt.Errorf("failed to update package with composite key: %v", err)
		}
	}

	return &VerifyPackageResult{
		IsValid: isValid,
		Package: pkg,
	}, nil
}

// VerifyPackageResult is the result of VerifyPackage
type VerifyPackageResult struct {
	IsValid bool               `json:"isValid"`
	Package *models.TeaPackage `json:"package"`
}

// GetPackageInfo retrieves package information
func (c *TeaTraceContract) GetPackageInfo(ctx contractapi.TransactionContextInterface, packageID string) (*models.TeaPackage, error) {
	if err := utils.ValidatePackageID(packageID); err != nil {
		return nil, err
	}

	pkgBytes, err := ctx.GetStub().GetState(packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %v", err)
	}
	if pkgBytes == nil {
		return nil, nil // Not found
	}

	var pkg models.TeaPackage
	if err := json.Unmarshal(pkgBytes, &pkg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %v", err)
	}

	return &pkg, nil
}

// GetAllPackages retrieves all packages with pagination
func (c *TeaTraceContract) GetAllPackages(ctx contractapi.TransactionContextInterface,
	limitStr, offsetStr string) (*PackageListResult, error) {

	limit, offset, err := c.parsePagination(limitStr, offsetStr)
	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{
		"selector": {"docType": "package"},
		"sort": [{"timestamp": "desc"}],
		"limit": %d,
		"skip": %d
	}`, limit, offset)

	return c.queryPackages(ctx, queryString)
}

// GetPackagesByBatch retrieves packages by batch ID using composite key
func (c *TeaTraceContract) GetPackagesByBatch(ctx contractapi.TransactionContextInterface,
	batchID, limitStr, offsetStr string) (*PackageListResult, error) {

	if err := utils.ValidateBatchID(batchID); err != nil {
		return nil, err
	}

	// Verify batch exists
	if _, err := c.getBatch(ctx, batchID); err != nil {
		return nil, err
	}

	limit, offset, err := c.parsePagination(limitStr, offsetStr)
	if err != nil {
		return nil, err
	}

	// Use composite key for efficient querying
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey("PACKAGE", []string{batchID})
	if err != nil {
		return nil, fmt.Errorf("failed to get packages by batch: %v", err)
	}
	defer iterator.Close()

	var packages []*models.TeaPackage
	total := 0
	skipped := 0

	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate packages: %v", err)
		}

		total++

		if skipped < offset {
			skipped++
			continue
		}

		if len(packages) < limit {
			var pkg models.TeaPackage
			if err := json.Unmarshal(result.Value, &pkg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal package: %v", err)
			}
			packages = append(packages, &pkg)
		}
	}

	return &PackageListResult{
		Packages: packages,
		Total:    total,
	}, nil
}

// GetPackagesByStatus retrieves packages by status with pagination
func (c *TeaTraceContract) GetPackagesByStatus(ctx contractapi.TransactionContextInterface,
	status, limitStr, offsetStr string) (*PackageListResult, error) {

	normalizedStatus := strings.ToUpper(status)
	if !models.IsValidPackageStatus(normalizedStatus) {
		return nil, fmt.Errorf("invalid status '%s'. Allowed: CREATED, VERIFIED, SOLD, EXPIRED", status)
	}

	limit, offset, err := c.parsePagination(limitStr, offsetStr)
	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{
		"selector": {"docType": "package", "status": "%s"},
		"sort": [{"timestamp": "desc"}],
		"limit": %d,
		"skip": %d
	}`, normalizedStatus, limit, offset)

	return c.queryPackages(ctx, queryString)
}

// GetPackageHistory retrieves the history of a package
func (c *TeaTraceContract) GetPackageHistory(ctx contractapi.TransactionContextInterface, packageID string) ([]*models.TeaPackage, error) {
	if err := utils.ValidatePackageID(packageID); err != nil {
		return nil, err
	}

	historyIterator, err := ctx.GetStub().GetHistoryForKey(packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %v", err)
	}
	defer historyIterator.Close()

	var history []*models.TeaPackage
	for historyIterator.HasNext() {
		modification, err := historyIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate history: %v", err)
		}

		if modification.IsDelete {
			continue
		}

		var pkg models.TeaPackage
		if err := json.Unmarshal(modification.Value, &pkg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal history: %v", err)
		}
		history = append(history, &pkg)
	}

	// Reverse to get oldest first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// UpdatePackageStatus updates the status of a package
func (c *TeaTraceContract) UpdatePackageStatus(ctx contractapi.TransactionContextInterface,
	packageID, status string) (*models.TeaPackage, error) {

	if err := utils.ValidatePackageID(packageID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(status) == "" {
		return nil, fmt.Errorf("status cannot be empty")
	}

	if err := c.ensureOrg(ctx, []string{c.mspConfig.Farmer, c.mspConfig.Admin}, "update package status"); err != nil {
		return nil, err
	}

	normalizedStatus := strings.ToUpper(status)
	if !models.IsValidPackageStatus(normalizedStatus) {
		return nil, fmt.Errorf("invalid status '%s'. Allowed: CREATED, VERIFIED, SOLD, EXPIRED", status)
	}

	pkg, err := c.getPackage(ctx, packageID)
	if err != nil {
		return nil, err
	}

	pkg.Status = models.TeaPackageStatus(normalizedStatus)
	pkg.Timestamp = c.getCurrentTimestamp(ctx)

	pkgJSON, err := json.Marshal(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %v", err)
	}

	if err := ctx.GetStub().PutState(packageID, pkgJSON); err != nil {
		return nil, fmt.Errorf("failed to update package: %v", err)
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey("PACKAGE", []string{pkg.BatchID, packageID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}
	if err := ctx.GetStub().PutState(compositeKey, pkgJSON); err != nil {
		return nil, fmt.Errorf("failed to update package with composite key: %v", err)
	}

	return pkg, nil
}

// PackageListResult is the result of package list queries
type PackageListResult struct {
	Packages []*models.TeaPackage `json:"packages"`
	Total    int                  `json:"total"`
}

// ==================== Helper Functions ====================

func (c *TeaTraceContract) ensureOrg(ctx contractapi.TransactionContextInterface, allowedMsps []string, action string) error {
	clientMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	for _, msp := range allowedMsps {
		if clientMSP == msp {
			return nil
		}
	}

	return fmt.Errorf("MSP '%s' is not authorized to %s. Allowed MSPs: %s", clientMSP, action, strings.Join(allowedMsps, ", "))
}

func (c *TeaTraceContract) getOwner(ctx contractapi.TransactionContextInterface) (string, error) {
	// Try to get owner attribute first
	owner, found, err := ctx.GetClientIdentity().GetAttributeValue("owner")
	if err == nil && found && owner != "" {
		return owner, nil
	}

	// Try organization attribute
	org, found, err := ctx.GetClientIdentity().GetAttributeValue("organization")
	if err == nil && found && org != "" {
		return org, nil
	}

	// Fall back to MSP ID
	return ctx.GetClientIdentity().GetMSPID()
}

func (c *TeaTraceContract) getCurrentTimestamp(ctx contractapi.TransactionContextInterface) string {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).UTC().Format(time.RFC3339)
}

func (c *TeaTraceContract) batchExists(ctx contractapi.TransactionContextInterface, batchID string) (bool, error) {
	batchBytes, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return false, fmt.Errorf("failed to check batch existence: %v", err)
	}
	return batchBytes != nil, nil
}

func (c *TeaTraceContract) getBatch(ctx contractapi.TransactionContextInterface, batchID string) (*models.TeaBatch, error) {
	batchBytes, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch: %v", err)
	}
	if batchBytes == nil {
		return nil, fmt.Errorf("batch with id '%s' does not exist", batchID)
	}

	var batch models.TeaBatch
	if err := json.Unmarshal(batchBytes, &batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %v", err)
	}

	return &batch, nil
}

func (c *TeaTraceContract) packageExists(ctx contractapi.TransactionContextInterface, packageID string) (bool, error) {
	pkgBytes, err := ctx.GetStub().GetState(packageID)
	if err != nil {
		return false, fmt.Errorf("failed to check package existence: %v", err)
	}
	return pkgBytes != nil, nil
}

func (c *TeaTraceContract) getPackage(ctx contractapi.TransactionContextInterface, packageID string) (*models.TeaPackage, error) {
	pkgBytes, err := ctx.GetStub().GetState(packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %v", err)
	}
	if pkgBytes == nil {
		return nil, fmt.Errorf("package with id '%s' does not exist", packageID)
	}

	var pkg models.TeaPackage
	if err := json.Unmarshal(pkgBytes, &pkg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %v", err)
	}

	return &pkg, nil
}

func (c *TeaTraceContract) parsePagination(limitStr, offsetStr string) (int, int, error) {
	limit := 100
	offset := 0

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid limit: %v", err)
		}
		limit = l
	}

	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid offset: %v", err)
		}
		offset = o
	}

	if err := utils.ValidatePagination(limit, offset); err != nil {
		return 0, 0, err
	}

	return limit, offset, nil
}

func (c *TeaTraceContract) queryBatches(ctx contractapi.TransactionContextInterface, queryString string) (*BatchListResult, error) {
	iterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer iterator.Close()

	var batches []*models.TeaBatch
	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate results: %v", err)
		}

		var batch models.TeaBatch
		if err := json.Unmarshal(result.Value, &batch); err != nil {
			return nil, fmt.Errorf("failed to unmarshal batch: %v", err)
		}
		batches = append(batches, &batch)
	}

	return &BatchListResult{
		Batches: batches,
		Total:   len(batches),
	}, nil
}

func (c *TeaTraceContract) queryPackages(ctx contractapi.TransactionContextInterface, queryString string) (*PackageListResult, error) {
	iterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer iterator.Close()

	var packages []*models.TeaPackage
	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate results: %v", err)
		}

		var pkg models.TeaPackage
		if err := json.Unmarshal(result.Value, &pkg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal package: %v", err)
		}
		packages = append(packages, &pkg)
	}

	return &PackageListResult{
		Packages: packages,
		Total:    len(packages),
	}, nil
}
