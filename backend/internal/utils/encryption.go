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

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Encryption constants
	saltLength    = 32
	nonceLength   = 12 // GCM standard nonce size
	keyLength     = 32 // AES-256 key length
	pbkdf2Iterations = 100000 // OWASP recommended iterations
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext format")
	ErrDecryptionFailed  = errors.New("decryption failed")
)

// EncryptPrivateKey encrypts a private key using AES-256-GCM
// Format: base64(salt + nonce + ciphertext + tag)
func EncryptPrivateKey(plaintext, masterKey string) (string, error) {
	if masterKey == "" {
		return "", fmt.Errorf("master key is required")
	}

	// Generate random salt
	salt := make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key from master key using PBKDF2
	key := pbkdf2.Key([]byte(masterKey), salt, pbkdf2Iterations, keyLength, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, nonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Combine: salt + nonce + ciphertext (tag is appended by GCM)
	combined := append(salt, nonce...)
	combined = append(combined, ciphertext...)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptPrivateKey decrypts an encrypted private key using AES-256-GCM
func DecryptPrivateKey(encrypted, masterKey string) (string, error) {
	if masterKey == "" {
		return "", fmt.Errorf("master key is required")
	}

	// Decode from base64
	combined, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check minimum length
	minLength := saltLength + nonceLength + 1 // at least 1 byte of ciphertext
	if len(combined) < minLength {
		return "", ErrInvalidCiphertext
	}

	// Extract components
	salt := combined[:saltLength]
	nonce := combined[saltLength : saltLength+nonceLength]
	ciphertext := combined[saltLength+nonceLength:]

	// Derive encryption key from master key using PBKDF2
	key := pbkdf2.Key([]byte(masterKey), salt, pbkdf2Iterations, keyLength, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// ValidateMasterKey checks if master key is strong enough
func ValidateMasterKey(masterKey string) error {
	if len(masterKey) < 32 {
		return fmt.Errorf("master key must be at least 32 characters")
	}
	return nil
}

