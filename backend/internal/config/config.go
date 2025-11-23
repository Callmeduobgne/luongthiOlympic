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
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	Fabric        FabricConfig
	Gateway       GatewayConfig
	AdminService  AdminServiceConfig
	JWT           JWTConfig
	OPA           OPAConfig
	RateLimit     RateLimitConfig
	CORS          CORSConfig
	CircuitBreaker CircuitBreakerConfig
	Logging       LoggingConfig
	OTEL          OTELConfig
	Encryption    EncryptionConfig
	Loki          LokiConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string
	Port int
	Env  string // development, staging, production
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MinConns     int
	MaxConns     int
	IdleTimeout  time.Duration
	MaxLifetime  time.Duration
	ReadReplicas []DatabaseReplicaConfig
}

// DatabaseReplicaConfig holds read replica configuration
type DatabaseReplicaConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// FabricConfig holds Hyperledger Fabric configuration
type FabricConfig struct {
	MSPID          string
	CryptoPath     string
	CertPath       string
	KeyPath        string
	TLSCertPath    string
	PeerEndpoint   string
	GatewayPeer    string
	ChannelName    string
}

// GatewayConfig holds API Gateway client configuration
// NOTE: Backend MUST use Gateway for all blockchain operations (Enabled must be true)
type GatewayConfig struct {
	BaseURL string
	APIKey  string // Optional API key for service-to-service auth
	Timeout time.Duration
	Enabled bool // MUST be true - Backend cannot connect directly to Fabric
}

// AdminServiceConfig holds Admin Service client configuration
type AdminServiceConfig struct {
	BaseURL string
	APIKey  string // Required API key for admin service authentication
	Timeout time.Duration
	Enabled bool
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret           string
	Issuer           string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

// OPAConfig holds Open Policy Agent configuration
type OPAConfig struct {
	Enabled bool
	BaseURL string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled       bool
	RequestsPerHour int
	BurstSize     int
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxRequests    uint32
	Interval       time.Duration
	Timeout        time.Duration
	FailureRatio   float64
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, console
}

// OTELConfig holds OpenTelemetry configuration
type OTELConfig struct {
	Enabled  bool
	Endpoint string
	ServiceName string
}

// EncryptionConfig holds encryption configuration for sensitive data
type EncryptionConfig struct {
	MasterKey string // Master key for encrypting private keys (MUST be at least 32 chars)
}

// LokiConfig holds Loki configuration
type LokiConfig struct {
	BaseURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnvAsInt("DB_PORT", 5432),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", "postgres"),
			Database:    getEnv("DB_NAME", "api_gateway"),
			SSLMode:     getEnv("DB_SSLMODE", "disable"),
			MinConns:    getEnvAsInt("DB_MIN_CONNS", 5),
			MaxConns:    getEnvAsInt("DB_MAX_CONNS", 25),
			IdleTimeout: getEnvAsDuration("DB_IDLE_TIMEOUT", 5*time.Minute),
			MaxLifetime: getEnvAsDuration("DB_MAX_LIFETIME", 1*time.Hour),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Fabric: FabricConfig{
			MSPID:        getEnv("FABRIC_MSP_ID", "Org1MSP"),
			CryptoPath:   getEnv("FABRIC_CRYPTO_PATH", "../core/organizations"),
			CertPath:     getEnv("FABRIC_CERT_PATH", ""),
			KeyPath:      getEnv("FABRIC_KEY_PATH", ""),
			TLSCertPath:  getEnv("FABRIC_TLS_CERT_PATH", ""),
			PeerEndpoint: getEnv("FABRIC_PEER_ENDPOINT", "localhost:7051"),
			GatewayPeer:  getEnv("FABRIC_GATEWAY_PEER", "peer0.org1.example.com"),
			ChannelName:  getEnv("FABRIC_CHANNEL_NAME", "mychannel"),
		},
		Gateway: GatewayConfig{
			BaseURL: getEnv("GATEWAY_BASE_URL", "http://api-gateway-nginx:80"),
			APIKey:  getEnv("GATEWAY_API_KEY", ""),
			Timeout: getEnvAsDuration("GATEWAY_TIMEOUT", 30*time.Second),
			Enabled: getEnvAsBool("GATEWAY_ENABLED", true),
		},
		AdminService: AdminServiceConfig{
			BaseURL: getEnv("ADMIN_SERVICE_BASE_URL", "http://admin-service:8090"),
			APIKey:  getEnv("ADMIN_SERVICE_API_KEY", ""),
			Timeout: getEnvAsDuration("ADMIN_SERVICE_TIMEOUT", 60*time.Second),
			Enabled: getEnvAsBool("ADMIN_SERVICE_ENABLED", true),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key-change-this"),
			Issuer:          getEnv("JWT_ISSUER", "ibn-network"),
			AccessTokenTTL:  getEnvAsDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvAsDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
		OPA: OPAConfig{
			Enabled: getEnvAsBool("OPA_ENABLED", true),
			BaseURL: getEnv("OPA_BASE_URL", "http://opa:8181"),
		},
		RateLimit: RateLimitConfig{
			Enabled:         getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerHour: getEnvAsInt("RATE_LIMIT_REQUESTS", 1000),
			BurstSize:       getEnvAsInt("RATE_LIMIT_BURST", 50),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "Idempotency-Key"},
			MaxAge:         300,
		},
		CircuitBreaker: CircuitBreakerConfig{
			MaxRequests:  uint32(getEnvAsInt("CB_MAX_REQUESTS", 3)),
			Interval:     getEnvAsDuration("CB_INTERVAL", 10*time.Second),
			Timeout:      getEnvAsDuration("CB_TIMEOUT", 60*time.Second),
			FailureRatio: getEnvAsFloat("CB_FAILURE_RATIO", 0.6),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		OTEL: OTELConfig{
			Enabled:     getEnvAsBool("OTEL_ENABLED", false),
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "ibn-backend"),
		},
		Encryption: EncryptionConfig{
			MasterKey: getEnv("ENCRYPTION_MASTER_KEY", ""), // MUST be set in production
		},
		Loki: LokiConfig{
			BaseURL: getEnv("LOKI_URL", "http://loki:3100"),
		},
	}

	return cfg, nil
}

// Address returns the full database connection string
func (d *DatabaseConfig) Address() string {
	return fmt.Sprintf("%s:%d", d.Host, d.Port)
}

// DSN returns the PostgreSQL DSN
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode)
}

// Address returns the Redis address
func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// Address returns the server address
func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

