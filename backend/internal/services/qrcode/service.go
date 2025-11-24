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
type PackageQRData struct {
	PackageID string `json:"packageId"`
	BlockHash string `json:"blockHash"`
	VerifyURL string `json:"verifyUrl,omitempty"` // Optional: direct verification URL
	TxID      string `json:"txId,omitempty"`      // Optional: transaction ID
}

// BatchQRData represents the data structure encoded in QR code for batches
type BatchQRData struct {
	BatchID         string `json:"batchId"`
	VerificationHash string `json:"verificationHash"`
	VerifyURL       string `json:"verifyUrl,omitempty"` // Optional: direct verification URL
	TxID            string `json:"txId,omitempty"`      // Optional: transaction ID
}

// GeneratePackageQRCode generates QR code image (PNG) from package data
// Returns PNG image bytes
func (s *Service) GeneratePackageQRCode(packageID, blockHash, txID string) ([]byte, error) {
	// Create QR data structure
	qrData := PackageQRData{
		PackageID: packageID,
		BlockHash: blockHash,
		VerifyURL: fmt.Sprintf("%s/packages/%s", s.verifyBaseURL, packageID),
		TxID:      txID,
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
		VerifyURL: fmt.Sprintf("%s/packages/%s", s.verifyBaseURL, packageID),
		TxID:      txID,
	}
}

// GenerateBatchQRCode generates QR code image (PNG) from batch data
// Returns PNG image bytes
func (s *Service) GenerateBatchQRCode(batchID, verificationHash, txID string) ([]byte, error) {
	// Create QR data structure
	qrData := BatchQRData{
		BatchID:          batchID,
		VerificationHash: verificationHash,
		VerifyURL:        fmt.Sprintf("%s/batches/%s", s.verifyBaseURL, batchID),
		TxID:             txID,
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
		VerifyURL:        fmt.Sprintf("%s/batches/%s", s.verifyBaseURL, batchID),
		TxID:             txID,
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

