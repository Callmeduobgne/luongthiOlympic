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

package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// Config holds all configuration for Admin Service
type Config struct {
	Server ServerConfig `validate:"required"`
	Fabric FabricConfig `validate:"required"`
	Auth   AuthConfig   `validate:"required"`
	Logging LoggingConfig `validate:"required"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int    `validate:"required,min=1024,max=65535"`
	Host string `validate:"required"`
	Env  string `validate:"required,oneof=development staging production"`
}

// FabricConfig holds Fabric network configuration
type FabricConfig struct {
	Channel        string `validate:"required"`
	MSPId          string `validate:"required"`
	PeerEndpoint   string `validate:"required"`
	UserCertPath   string `validate:"required"`
	UserKeyPath    string `validate:"required"`
	PeerTLSCAPath  string `validate:"required"`
	OrdererTLSCAPath string `validate:"required"`
	PeerHostOverride string
}

// AuthConfig holds authentication configuration for internal API
type AuthConfig struct {
	APIKey string `validate:"required,min=32"`
	Enabled bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `validate:"required,oneof=debug info warn error fatal"`
	Format string `validate:"required,oneof=json text"`
	Output string `validate:"required,oneof=stdout stderr file"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("ADMIN_SERVER_PORT", 8090),
			Host: getEnv("ADMIN_SERVER_HOST", "0.0.0.0"),
			Env:  getEnv("ADMIN_SERVER_ENV", "production"),
		},
		Fabric: FabricConfig{
			Channel:         getEnv("ADMIN_FABRIC_CHANNEL", ""),
			MSPId:           getEnv("ADMIN_FABRIC_MSP_ID", ""),
			PeerEndpoint:    getEnv("ADMIN_FABRIC_PEER_ENDPOINT", ""),
			UserCertPath:    getEnv("ADMIN_FABRIC_USER_CERT_PATH", ""),
			UserKeyPath:     getEnv("ADMIN_FABRIC_USER_KEY_PATH", ""),
			PeerTLSCAPath:   getEnv("ADMIN_FABRIC_PEER_TLS_CA_PATH", ""),
			OrdererTLSCAPath: getEnv("ADMIN_FABRIC_ORDERER_TLS_CA_PATH", ""),
			PeerHostOverride: getEnv("ADMIN_FABRIC_PEER_HOST_OVERRIDE", ""),
		},
		Auth: AuthConfig{
			APIKey:  getEnv("ADMIN_AUTH_API_KEY", ""),
			Enabled: getEnvAsBool("ADMIN_AUTH_ENABLED", true),
		},
		Logging: LoggingConfig{
			Level:  getEnv("ADMIN_LOGGING_LEVEL", "info"),
			Format: getEnv("ADMIN_LOGGING_FORMAT", "json"),
			Output: getEnv("ADMIN_LOGGING_OUTPUT", "stdout"),
		},
	}

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Additional validation: check if required files exist
	if err := validateFileExists(cfg.Fabric.UserCertPath); err != nil {
		return nil, fmt.Errorf("invalid user cert path: %w", err)
	}
	if err := validateFileExists(cfg.Fabric.PeerTLSCAPath); err != nil {
		return nil, fmt.Errorf("invalid peer TLS CA path: %w", err)
	}
	if err := validateFileExists(cfg.Fabric.OrdererTLSCAPath); err != nil {
		return nil, fmt.Errorf("invalid orderer TLS CA path: %w", err)
	}

	return cfg, nil
}

// Helper functions to read environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// validateFileExists checks if a file exists (for certificate paths)
func validateFileExists(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}
	// Note: We can't check file existence at config load time in Docker
	// because the volume might not be mounted yet. This validation is done
	// at runtime when the service actually uses the certificates.
	// For now, we just check that the path is not empty.
	return nil
}
