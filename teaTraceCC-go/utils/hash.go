package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"teaTraceCC/models"
)

// GenerateBatchHash generates a SHA256 hash from batch input
func GenerateBatchHash(input *models.CreateTeaBatchInput) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%s",
		input.BatchID,
		input.FarmLocation,
		input.HarvestDate,
		input.ProcessingInfo,
		input.QualityCert,
	)
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

// VerifyHash verifies if a hash matches the input
func VerifyHash(storedHash, inputHash string) bool {
	return storedHash == inputHash
}

// GeneratePackageBlockHash generates a unique blockhash for a package
func GeneratePackageBlockHash(packageID, batchID string, weight float64, productionDate, txID, secret string) string {
	payload := fmt.Sprintf("%s|%s|%f|%s|%s", packageID, batchID, weight, productionDate, txID)
	if secret != "" {
		payload += "|" + secret
	}
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}
