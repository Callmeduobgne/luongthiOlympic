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

package qrcode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ibn-network/backend/internal/utils/merkle"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
)

// Service handles QR code and NFC generation
type Service struct {
	logger        *zap.Logger
	verifyBaseURL string // Base URL for verification page (e.g., https://verify.ibn.network)
}

// NewService creates a new QR code service
func NewService(logger *zap.Logger, verifyBaseURL string) *Service {
	if verifyBaseURL == "" {
		verifyBaseURL = "https://verify.ibn.network" // Default
	}
	return &Service{
		logger:        logger,
		verifyBaseURL: verifyBaseURL,
	}
}

// PackageQRData represents the data structure encoded in QR code
// Now includes Merkle proof for cryptographic verification
type PackageQRData struct {
	PackageID   string              `json:"packageId"`
	BlockHash   string              `json:"blockHash"`
	TxID        string              `json:"txId,omitempty"`        // Optional: transaction ID
	BlockNumber uint64              `json:"blockNumber,omitempty"` // Optional: block number for verification
	MerkleProof []merkle.ProofStep  `json:"merkleProof,omitempty"` // Optional: Merkle proof for fast verification
	MerkleRoot  string              `json:"merkleRoot,omitempty"`  // Optional: Merkle root hash
	VerifyURL   string              `json:"verifyUrl,omitempty"`   // Optional: direct verification URL
	Timestamp   int64               `json:"timestamp,omitempty"`   // Optional: QR generation timestamp
}

// BatchQRData represents the data structure encoded in QR code for batches
// Now includes Merkle proof for cryptographic verification
type BatchQRData struct {
	BatchID          string             `json:"batchId"`
	VerificationHash string             `json:"verificationHash"`
	TxID             string             `json:"txId,omitempty"`        // Optional: transaction ID
	BlockNumber      uint64             `json:"blockNumber,omitempty"` // Optional: block number for verification
	MerkleProof      []merkle.ProofStep `json:"merkleProof,omitempty"` // Optional: Merkle proof for fast verification
	MerkleRoot       string             `json:"merkleRoot,omitempty"`  // Optional: Merkle root hash
	VerifyURL        string             `json:"verifyUrl,omitempty"`   // Optional: direct verification URL
	Timestamp        int64              `json:"timestamp,omitempty"`   // Optional: QR generation timestamp
}

// GeneratePackageQRCode generates QR code image (PNG) from package data
// Returns PNG image bytes
// Note: MerkleProof and MerkleRoot fields are left empty for now (backward compatible)
// Use GeneratePackageQRCodeWithProof() to include Merkle proof
func (s *Service) GeneratePackageQRCode(packageID, blockHash, txID string) ([]byte, error) {
	// Create QR data structure
	qrData := PackageQRData{
		PackageID: packageID,
		BlockHash: blockHash,
		TxID:      txID,
		VerifyURL:  fmt.Sprintf("%s/packages/%s", s.verifyBaseURL, packageID),
		Timestamp: time.Now().Unix(),
		// MerkleProof and MerkleRoot are omitted (nil/empty) for backward compatibility
	}

	// Convert to JSON string
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QR data: %w", err)
	}

	// Generate QR code (size: 256x256, recovery level: Medium)
	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Convert to PNG bytes
	pngBytes, err := qr.PNG(256)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code as PNG: %w", err)
	}

	s.logger.Info("QR code generated",
		zap.String("package_id", packageID),
		zap.Int("size_bytes", len(pngBytes)),
	)

	return pngBytes, nil
}

// GeneratePackageQRCodeToWriter writes QR code PNG to writer
func (s *Service) GeneratePackageQRCodeToWriter(w io.Writer, packageID, blockHash, txID string) error {
	pngBytes, err := s.GeneratePackageQRCode(packageID, blockHash, txID)
	if err != nil {
		return err
	}

	_, err = w.Write(pngBytes)
	return err
}

// GeneratePackageQRCodeString returns QR code data as string (for embedding in HTML)
// Returns base64-encoded PNG data URI
func (s *Service) GeneratePackageQRCodeString(packageID, blockHash, txID string) (string, error) {
	pngBytes, err := s.GeneratePackageQRCode(packageID, blockHash, txID)
	if err != nil {
		return "", err
	}

	// Convert to base64 data URI
	dataURI := fmt.Sprintf("data:image/png;base64,%s", encodeBase64(pngBytes))
	return dataURI, nil
}

// GetPackageQRData returns the QR code data structure (without generating image)
func (s *Service) GetPackageQRData(packageID, blockHash, txID string) *PackageQRData {
	return &PackageQRData{
		PackageID: packageID,
		BlockHash: blockHash,
		TxID:      txID,
		VerifyURL: fmt.Sprintf("%s/packages/%s", s.verifyBaseURL, packageID),
		Timestamp: time.Now().Unix(),
	}
}

// GenerateBatchQRCode generates QR code image (PNG) from batch data
// Returns PNG image bytes
// Note: MerkleProof and MerkleRoot fields are left empty for now (backward compatible)
// Use GenerateBatchQRCodeWithProof() to include Merkle proof
func (s *Service) GenerateBatchQRCode(batchID, verificationHash, txID string) ([]byte, error) {
	// Create QR data structure
	qrData := BatchQRData{
		BatchID:          batchID,
		VerificationHash: verificationHash,
		TxID:             txID,
		VerifyURL:        fmt.Sprintf("%s/batches/%s", s.verifyBaseURL, batchID),
		Timestamp:        time.Now().Unix(),
		// MerkleProof and MerkleRoot are omitted (nil/empty) for backward compatibility
	}

	// Convert to JSON string
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QR data: %w", err)
	}

	// Generate QR code (size: 256x256, recovery level: Medium)
	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Convert to PNG bytes
	pngBytes, err := qr.PNG(256)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code as PNG: %w", err)
	}

	s.logger.Info("QR code generated for batch",
		zap.String("batch_id", batchID),
		zap.Int("size_bytes", len(pngBytes)),
	)

	return pngBytes, nil
}

// GenerateBatchQRCodeString returns QR code data as string (for embedding in HTML)
// Returns base64-encoded PNG data URI
func (s *Service) GenerateBatchQRCodeString(batchID, verificationHash, txID string) (string, error) {
	pngBytes, err := s.GenerateBatchQRCode(batchID, verificationHash, txID)
	if err != nil {
		return "", err
	}

	// Convert to base64 data URI
	dataURI := fmt.Sprintf("data:image/png;base64,%s", encodeBase64(pngBytes))
	return dataURI, nil
}

// GetBatchQRData returns the QR code data structure for batch (without generating image)
func (s *Service) GetBatchQRData(batchID, verificationHash, txID string) *BatchQRData {
	return &BatchQRData{
		BatchID:          batchID,
		VerificationHash: verificationHash,
		TxID:             txID,
		VerifyURL:        fmt.Sprintf("%s/batches/%s", s.verifyBaseURL, batchID),
		Timestamp:        time.Now().Unix(),
	}
}

// NFCData represents NFC tag data format
type NFCData struct {
	Type        string `json:"type"`        // "tea_package"
	PackageID   string `json:"packageId"`
	BlockHash   string `json:"blockHash"`
	VerifyURL   string `json:"verifyUrl"`
	TxID        string `json:"txId,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// GenerateNFCPayload generates NFC tag payload (NDEF format)
// Returns JSON string that can be written to NFC tag
func (s *Service) GenerateNFCPayload(packageID, blockHash, txID string) (string, error) {
	nfcData := NFCData{
		Type:      "tea_package",
		PackageID: packageID,
		BlockHash: blockHash,
		VerifyURL: fmt.Sprintf("%s/packages/%s", s.verifyBaseURL, packageID),
		TxID:      txID,
	}

	jsonData, err := json.Marshal(nfcData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal NFC data: %w", err)
	}

	return string(jsonData), nil
}

// encodeBase64 encodes data to base64 string
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// ============================================================================
// MERKLE PROOF FUNCTIONS (For future use - Phase 2b)
// ============================================================================

// GeneratePackageQRCodeWithProof generates QR code with Merkle proof
// This function will be used when block transactions data is available
// For now, it's a placeholder that calls the regular function
func (s *Service) GeneratePackageQRCodeWithProof(
	packageID string,
	blockHash string,
	txID string,
	blockNumber uint64,
	blockTransactions []string, // All transaction IDs in the block
) ([]byte, error) {
	// TODO: Implement Merkle proof generation when block data is available
	// For now, generate QR without proof (backward compatible)
	
	if len(blockTransactions) == 0 {
		// No block data available, use regular function
		s.logger.Warn("Block transactions not provided, generating QR without Merkle proof",
			zap.String("package_id", packageID),
		)
		return s.GeneratePackageQRCode(packageID, blockHash, txID)
	}

	// Generate Merkle tree from block transactions
	tree, err := merkle.GenerateMerkleTree(blockTransactions)
	if err != nil {
		s.logger.Error("Failed to generate Merkle tree",
			zap.Error(err),
			zap.String("package_id", packageID),
		)
		// Fallback to regular QR generation
		return s.GeneratePackageQRCode(packageID, blockHash, txID)
	}

	// Generate Merkle proof for this transaction
	proof, err := merkle.GenerateMerkleProof(tree, txID)
	if err != nil {
		s.logger.Error("Failed to generate Merkle proof",
			zap.Error(err),
			zap.String("tx_id", txID),
		)
		// Fallback to regular QR generation
		return s.GeneratePackageQRCode(packageID, blockHash, txID)
	}

	// Create QR data with Merkle proof
	qrData := PackageQRData{
		PackageID:   packageID,
		BlockHash:   blockHash,
		TxID:        txID,
		BlockNumber: blockNumber,
		MerkleProof: proof,
		MerkleRoot:  tree.Root.Hash,
		VerifyURL:   fmt.Sprintf("%s/packages/%s", s.verifyBaseURL, packageID),
		Timestamp:   time.Now().Unix(),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QR data with proof: %w", err)
	}

	// Generate QR code
	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	pngBytes, err := qr.PNG(256)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code as PNG: %w", err)
	}

	s.logger.Info("QR code with Merkle proof generated",
		zap.String("package_id", packageID),
		zap.Int("size_bytes", len(pngBytes)),
		zap.Int("proof_steps", len(proof)),
	)

	return pngBytes, nil
}

// GenerateBatchQRCodeWithProof generates batch QR code with Merkle proof
// This function will be used when block transactions data is available
func (s *Service) GenerateBatchQRCodeWithProof(
	batchID string,
	verificationHash string,
	txID string,
	blockNumber uint64,
	blockTransactions []string,
) ([]byte, error) {
	// TODO: Implement Merkle proof generation when block data is available
	// For now, generate QR without proof (backward compatible)
	
	if len(blockTransactions) == 0 {
		s.logger.Warn("Block transactions not provided, generating QR without Merkle proof",
			zap.String("batch_id", batchID),
		)
		return s.GenerateBatchQRCode(batchID, verificationHash, txID)
	}

	// Generate Merkle tree
	tree, err := merkle.GenerateMerkleTree(blockTransactions)
	if err != nil {
		s.logger.Error("Failed to generate Merkle tree",
			zap.Error(err),
			zap.String("batch_id", batchID),
		)
		return s.GenerateBatchQRCode(batchID, verificationHash, txID)
	}

	// Generate proof
	proof, err := merkle.GenerateMerkleProof(tree, txID)
	if err != nil {
		s.logger.Error("Failed to generate Merkle proof",
			zap.Error(err),
			zap.String("tx_id", txID),
		)
		return s.GenerateBatchQRCode(batchID, verificationHash, txID)
	}

	// Create QR data with proof
	qrData := BatchQRData{
		BatchID:          batchID,
		VerificationHash: verificationHash,
		TxID:             txID,
		BlockNumber:      blockNumber,
		MerkleProof:      proof,
		MerkleRoot:       tree.Root.Hash,
		VerifyURL:        fmt.Sprintf("%s/batches/%s", s.verifyBaseURL, batchID),
		Timestamp:        time.Now().Unix(),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QR data with proof: %w", err)
	}

	// Generate QR code
	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	pngBytes, err := qr.PNG(256)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code as PNG: %w", err)
	}

	s.logger.Info("Batch QR code with Merkle proof generated",
		zap.String("batch_id", batchID),
		zap.Int("size_bytes", len(pngBytes)),
		zap.Int("proof_steps", len(proof)),
	)

	return pngBytes, nil
}

