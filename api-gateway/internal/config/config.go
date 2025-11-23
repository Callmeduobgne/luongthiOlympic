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
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Redis          RedisConfig          `mapstructure:"redis"`
	Fabric         FabricConfig         `mapstructure:"fabric"`
	CA             CAConfig             `mapstructure:"ca"`
	JWT            JWTConfig            `mapstructure:"jwt"`
	RateLimit      RateLimitConfig      `mapstructure:"ratelimit"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuitbreaker"`
	OpenTelemetry  OpenTelemetryConfig  `mapstructure:"otel"`
	Logging        LoggingConfig        `mapstructure:"logging"`
	CORS           CORSConfig           `mapstructure:"cors"`
	WebSocket      WebSocketConfig      `mapstructure:"websocket"`
	Upstream       UpstreamConfig       `mapstructure:"upstream"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int    `mapstructure:"port" validate:"required,min=1024,max=65535"`
	Host string `mapstructure:"host" validate:"required"`
	Env  string `mapstructure:"env" validate:"required,oneof=development staging production"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host" validate:"required,hostname"`
	Port            int           `mapstructure:"port" validate:"required,min=1024,max=65535"`
	Database        string        `mapstructure:"database" validate:"required"`
	User            string        `mapstructure:"user" validate:"required"`
	Password        string        `mapstructure:"password" validate:"required,min=8"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"required,min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" validate:"required"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host" validate:"required,hostname"`
	Port     int    `mapstructure:"port" validate:"required,min=1024,max=65535"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0,max=15"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string        `mapstructure:"secret" validate:"required,min=32"`
	Expiry time.Duration `mapstructure:"expiry" validate:"required"`
	Issuer string        `mapstructure:"issuer" validate:"required"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Requests int           `mapstructure:"requests" validate:"required,min=1"`
	Window   time.Duration `mapstructure:"window" validate:"required"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxRequests  uint32        `mapstructure:"max_requests" validate:"required,min=1"`
	Interval     time.Duration `mapstructure:"interval" validate:"required"`
	Timeout      time.Duration `mapstructure:"timeout" validate:"required"`
	FailureRatio float64       `mapstructure:"failure_ratio" validate:"required,min=0,max=1"`
}

// OpenTelemetryConfig holds OpenTelemetry configuration
type OpenTelemetryConfig struct {
	Enabled          bool    `mapstructure:"enabled"`
	ExporterEndpoint string  `mapstructure:"exporter_endpoint" validate:"required_if=Enabled true"`
	ServiceName      string  `mapstructure:"service_name" validate:"required"`
	TraceSampleRate  float64 `mapstructure:"trace_sample_rate" validate:"min=0,max=1"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error fatal"`
	Format string `mapstructure:"format" validate:"required,oneof=json text"`
	Output string `mapstructure:"output" validate:"required,oneof=stdout stderr file"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins" validate:"required,min=1"`
	AllowedMethods   []string `mapstructure:"allowed_methods" validate:"required,min=1"`
	AllowedHeaders   []string `mapstructure:"allowed_headers" validate:"required,min=1"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	MaxAge           int      `mapstructure:"max_age" validate:"min=0"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	AllowedOrigins        []string      `mapstructure:"allowed_origins"`
	MaxConnections        int           `mapstructure:"max_connections" validate:"min=1"`
	MaxConnectionsPerIP   int           `mapstructure:"max_connections_per_ip" validate:"min=1"`
	MaxConnectionsPerUser int           `mapstructure:"max_connections_per_user" validate:"min=1"`
	PingInterval          time.Duration `mapstructure:"ping_interval" validate:"required"`
	PongTimeout           time.Duration `mapstructure:"pong_timeout" validate:"required"`
	EnableCompression     bool          `mapstructure:"enable_compression"`
	RateLimitEnabled      bool          `mapstructure:"rate_limit_enabled"`
	RateLimitMessages     int           `mapstructure:"rate_limit_messages" validate:"min=1"`
	RateLimitWindow       time.Duration `mapstructure:"rate_limit_window" validate:"required"`
}

// UpstreamConfig holds upstream service configuration
type UpstreamConfig struct {
	BackendBaseURL string `mapstructure:"backend_base_url" validate:"required,url"`
}

// DSN returns the database connection string
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Database,
	)
}

// Address returns the Redis address
func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// Load loads configuration from environment variables and validates it
func Load() (*Config, error) {
	// Set config file
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Read from environment variables
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (optional - skip if not found, use env vars instead)
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional when running in Docker (using env vars)
		// Only return error if it's not a "file not found" error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Log warning but continue with environment variables
			// This allows Docker containers to work without .env file
		}
	}

	// Unmarshal config
	var config Config

	// Server
	config.Server.Port = viper.GetInt("GATEWAY_PORT")
	config.Server.Host = viper.GetString("GATEWAY_HOST")
	config.Server.Env = viper.GetString("GATEWAY_ENV")

	// Database
	config.Database.Host = viper.GetString("POSTGRES_HOST")
	config.Database.Port = viper.GetInt("POSTGRES_PORT")
	config.Database.Database = viper.GetString("POSTGRES_DB")
	config.Database.User = viper.GetString("POSTGRES_USER")
	config.Database.Password = viper.GetString("POSTGRES_PASSWORD")
	config.Database.MaxOpenConns = viper.GetInt("POSTGRES_MAX_OPEN_CONNS")
	config.Database.MaxIdleConns = viper.GetInt("POSTGRES_MAX_IDLE_CONNS")
	config.Database.ConnMaxLifetime = viper.GetDuration("POSTGRES_CONN_MAX_LIFETIME")
	config.Database.ConnMaxIdleTime = viper.GetDuration("POSTGRES_CONN_MAX_IDLE_TIME")

	// Redis
	config.Redis.Host = viper.GetString("REDIS_HOST")
	config.Redis.Port = viper.GetInt("REDIS_PORT")
	config.Redis.Password = viper.GetString("REDIS_PASSWORD")
	config.Redis.DB = viper.GetInt("REDIS_DB")

	// Fabric
	config.Fabric.Channel = viper.GetString("FABRIC_CHANNEL")
	config.Fabric.Chaincode = viper.GetString("FABRIC_CHAINCODE")
	config.Fabric.MSPId = viper.GetString("FABRIC_MSP_ID")
	config.Fabric.PeerEndpoint = viper.GetString("FABRIC_PEER_ENDPOINT")
	config.Fabric.PeerHostOverride = viper.GetString("FABRIC_PEER_HOST_OVERRIDE")
	config.Fabric.UserCertPath = viper.GetString("FABRIC_USER_CERT_PATH")
	config.Fabric.UserKeyPath = viper.GetString("FABRIC_USER_KEY_PATH")
	config.Fabric.PeerTLSCAPath = viper.GetString("FABRIC_PEER_TLS_CA_PATH")
	config.Fabric.AdditionalPeers = viper.GetStringSlice("FABRIC_ADDITIONAL_PEERS")
	config.Fabric.Orderers = viper.GetStringSlice("FABRIC_ORDERERS")
	config.Fabric.CAEndpoints = viper.GetStringSlice("FABRIC_CA_ENDPOINTS")

	// CA (optional - for future CA server integration)
	config.CA.URL = viper.GetString("FABRIC_CA_URL")
	config.CA.CAName = viper.GetString("FABRIC_CA_NAME")
	config.CA.TLSCertPath = viper.GetString("FABRIC_CA_TLS_CERT_PATH")
	config.CA.MSPDir = viper.GetString("FABRIC_CA_MSP_DIR")
	config.CA.AdminUser = viper.GetString("FABRIC_CA_ADMIN_USER")
	config.CA.AdminPass = viper.GetString("FABRIC_CA_ADMIN_PASS")
	config.CA.MSPID = viper.GetString("FABRIC_MSP_ID")

	// JWT
	config.JWT.Secret = viper.GetString("JWT_SECRET")
	config.JWT.Expiry = viper.GetDuration("JWT_EXPIRY")
	config.JWT.Issuer = viper.GetString("JWT_ISSUER")

	// Rate Limit
	config.RateLimit.Enabled = viper.GetBool("RATE_LIMIT_ENABLED")
	config.RateLimit.Requests = viper.GetInt("RATE_LIMIT_REQUESTS")
	config.RateLimit.Window = viper.GetDuration("RATE_LIMIT_WINDOW")

	// Circuit Breaker
	config.CircuitBreaker.MaxRequests = uint32(viper.GetInt("CIRCUIT_BREAKER_MAX_REQUESTS"))
	config.CircuitBreaker.Interval = viper.GetDuration("CIRCUIT_BREAKER_INTERVAL")
	config.CircuitBreaker.Timeout = viper.GetDuration("CIRCUIT_BREAKER_TIMEOUT")
	config.CircuitBreaker.FailureRatio = viper.GetFloat64("CIRCUIT_BREAKER_FAILURE_RATIO")

	// OpenTelemetry
	config.OpenTelemetry.Enabled = viper.GetBool("OTEL_ENABLED")
	config.OpenTelemetry.ExporterEndpoint = viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")
	config.OpenTelemetry.ServiceName = viper.GetString("OTEL_SERVICE_NAME")
	config.OpenTelemetry.TraceSampleRate = viper.GetFloat64("OTEL_TRACE_SAMPLE_RATE")

	// Logging
	config.Logging.Level = viper.GetString("LOG_LEVEL")
	config.Logging.Format = viper.GetString("LOG_FORMAT")
	config.Logging.Output = viper.GetString("LOG_OUTPUT")

	// CORS
	config.CORS.AllowedOrigins = viper.GetStringSlice("CORS_ALLOWED_ORIGINS")
	config.CORS.AllowedMethods = viper.GetStringSlice("CORS_ALLOWED_METHODS")
	config.CORS.AllowedHeaders = viper.GetStringSlice("CORS_ALLOWED_HEADERS")
	config.CORS.ExposedHeaders = viper.GetStringSlice("CORS_EXPOSED_HEADERS")
	config.CORS.MaxAge = viper.GetInt("CORS_MAX_AGE")
	config.CORS.AllowCredentials = viper.GetBool("CORS_ALLOW_CREDENTIALS")

	// WebSocket
	config.WebSocket.AllowedOrigins = viper.GetStringSlice("WEBSOCKET_ALLOWED_ORIGINS")
	config.WebSocket.MaxConnections = viper.GetInt("WEBSOCKET_MAX_CONNECTIONS")
	config.WebSocket.MaxConnectionsPerIP = viper.GetInt("WEBSOCKET_MAX_CONNECTIONS_PER_IP")
	config.WebSocket.MaxConnectionsPerUser = viper.GetInt("WEBSOCKET_MAX_CONNECTIONS_PER_USER")
	config.WebSocket.PingInterval = viper.GetDuration("WEBSOCKET_PING_INTERVAL")
	config.WebSocket.PongTimeout = viper.GetDuration("WEBSOCKET_PONG_TIMEOUT")
	config.WebSocket.EnableCompression = viper.GetBool("WEBSOCKET_ENABLE_COMPRESSION")
	config.WebSocket.RateLimitEnabled = viper.GetBool("WEBSOCKET_RATE_LIMIT_ENABLED")
	config.WebSocket.RateLimitMessages = viper.GetInt("WEBSOCKET_RATE_LIMIT_MESSAGES")
	config.WebSocket.RateLimitWindow = viper.GetDuration("WEBSOCKET_RATE_LIMIT_WINDOW")

	// Upstream services
	config.Upstream.BackendBaseURL = viper.GetString("BACKEND_BASE_URL")

	// Validate config
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Security warnings for production
	if config.Server.Env == "production" {
		// Check JWT_SECRET strength
		weakSecrets := []string{
			"your-super-secret-jwt-key-change-this-in-production-min-32-chars",
			"ibn-network-production-secret-2024-change-this",
			"change-this",
			"secret",
			"default",
		}
		for _, weak := range weakSecrets {
			if config.JWT.Secret == weak || len(config.JWT.Secret) < 32 {
				return nil, fmt.Errorf("SECURITY WARNING: JWT_SECRET is too weak or uses default value. Please use a strong, random secret (min 32 characters) in production")
			}
		}
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("GATEWAY_PORT", 8080)
	viper.SetDefault("GATEWAY_HOST", "0.0.0.0")
	viper.SetDefault("GATEWAY_ENV", "development")

	// Database defaults
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_DB", "ibn_gateway")
	viper.SetDefault("POSTGRES_USER", "gateway")
	viper.SetDefault("POSTGRES_MAX_OPEN_CONNS", 25)
	viper.SetDefault("POSTGRES_MAX_IDLE_CONNS", 10)
	viper.SetDefault("POSTGRES_CONN_MAX_LIFETIME", "5m")
	viper.SetDefault("POSTGRES_CONN_MAX_IDLE_TIME", "5m")

	// Redis defaults
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_DB", 0)

	// JWT defaults
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("JWT_ISSUER", "ibn-api-gateway")

	// Rate limit defaults
	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_REQUESTS", 1000)
	viper.SetDefault("RATE_LIMIT_WINDOW", "1h")

	// Circuit breaker defaults
	viper.SetDefault("CIRCUIT_BREAKER_MAX_REQUESTS", 3)
	viper.SetDefault("CIRCUIT_BREAKER_INTERVAL", "10s")
	viper.SetDefault("CIRCUIT_BREAKER_TIMEOUT", "60s")
	viper.SetDefault("CIRCUIT_BREAKER_FAILURE_RATIO", 0.6)

	// OpenTelemetry defaults
	viper.SetDefault("OTEL_ENABLED", false)
	viper.SetDefault("OTEL_SERVICE_NAME", "ibn-api-gateway")
	viper.SetDefault("OTEL_TRACE_SAMPLE_RATE", 1.0)

	// Logging defaults
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("LOG_OUTPUT", "stdout")

	// CORS defaults
	viper.SetDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"})
	viper.SetDefault("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"})
	viper.SetDefault("CORS_EXPOSED_HEADERS", []string{"Link"})
	viper.SetDefault("CORS_MAX_AGE", 300)
	viper.SetDefault("CORS_ALLOW_CREDENTIALS", true)

	// WebSocket defaults
	viper.SetDefault("WEBSOCKET_ALLOWED_ORIGINS", []string{}) // Empty = allow all (development)
	viper.SetDefault("WEBSOCKET_MAX_CONNECTIONS", 1000)
	viper.SetDefault("WEBSOCKET_MAX_CONNECTIONS_PER_IP", 10)
	viper.SetDefault("WEBSOCKET_MAX_CONNECTIONS_PER_USER", 5)
	viper.SetDefault("WEBSOCKET_PING_INTERVAL", "30s")
	viper.SetDefault("WEBSOCKET_PONG_TIMEOUT", "60s")
	viper.SetDefault("WEBSOCKET_ENABLE_COMPRESSION", true)
	viper.SetDefault("WEBSOCKET_RATE_LIMIT_ENABLED", true)
	viper.SetDefault("WEBSOCKET_RATE_LIMIT_MESSAGES", 100)
	viper.SetDefault("WEBSOCKET_RATE_LIMIT_WINDOW", "1m")

	// Upstream defaults
	viper.SetDefault("BACKEND_BASE_URL", "http://ibn-backend:8080")
}
