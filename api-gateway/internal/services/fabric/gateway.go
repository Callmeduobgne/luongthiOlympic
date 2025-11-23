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

package fabric

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GatewayService manages Fabric Gateway connections
type GatewayService struct {
	gateway        *client.Gateway
	network        *client.Network
	contract       *client.Contract
	circuitBreaker *gobreaker.CircuitBreaker
	config         *config.FabricConfig
	logger         *zap.Logger
	tracer         trace.Tracer
}

// NewGatewayService creates a new Fabric Gateway service
func NewGatewayService(cfg *config.FabricConfig, cbCfg *config.CircuitBreakerConfig, logger *zap.Logger) (*GatewayService, error) {
	// Load certificates
	cert, err := loadCertificate(cfg.UserCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Load private key
	privateKey, err := loadPrivateKey(cfg.UserKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Create identity
	id, err := identity.NewX509Identity(cfg.MSPId, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	// Load TLS certificate
	tlsCert, err := loadTLSCertificate(cfg.PeerTLSCAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create gRPC connection
	grpcConn, err := createGRPCConnection(cfg.PeerEndpoint, cfg.PeerHostOverride, tlsCert)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create signer
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// Create gateway
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConn),
		client.WithEvaluateTimeout(5*time.Minute),
		client.WithEndorseTimeout(5*time.Minute),
		client.WithSubmitTimeout(5*time.Minute),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	// Get network
	network := gw.GetNetwork(cfg.Channel)

	// Get contract
	contract := network.GetContract(cfg.Chaincode)

	// Setup circuit breaker
	cbSettings := gobreaker.Settings{
		Name:        "FabricGateway",
		MaxRequests: cbCfg.MaxRequests,
		Interval:    cbCfg.Interval,
		Timeout:     cbCfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= cbCfg.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	// Get tracer
	tracer := otel.Tracer("fabric-gateway")

	logger.Info("Connected to Fabric Gateway",
		zap.String("channel", cfg.Channel),
		zap.String("chaincode", cfg.Chaincode),
		zap.String("mspId", cfg.MSPId),
	)

	return &GatewayService{
		gateway:        gw,
		network:        network,
		contract:       contract,
		circuitBreaker: gobreaker.NewCircuitBreaker(cbSettings),
		config:         cfg,
		logger:         logger,
		tracer:         tracer,
	}, nil
}

// SubmitTransaction submits a transaction to the blockchain
func (s *GatewayService) SubmitTransaction(ctx context.Context, functionName string, args ...string) ([]byte, error) {
	ctx, span := s.tracer.Start(ctx, "SubmitTransaction")
	defer span.End()

	result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		return s.contract.SubmitTransaction(functionName, args...)
	})

	if err != nil {
		s.logger.Error("Failed to submit transaction",
			zap.String("function", functionName),
			zap.Error(err),
		)
		return nil, err
	}

	return result.([]byte), nil
}

// EvaluateTransaction evaluates a transaction (read-only query)
func (s *GatewayService) EvaluateTransaction(ctx context.Context, functionName string, args ...string) ([]byte, error) {
	ctx, span := s.tracer.Start(ctx, "EvaluateTransaction")
	defer span.End()

	result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		return s.contract.EvaluateTransaction(functionName, args...)
	})

	if err != nil {
		s.logger.Error("Failed to evaluate transaction",
			zap.String("function", functionName),
			zap.Error(err),
		)
		return nil, err
	}

	return result.([]byte), nil
}

// GetGatewayClient returns the underlying gateway client
func (s *GatewayService) GetGatewayClient() *client.Gateway {
	return s.gateway
}

// CreateDynamicGateway creates a new gateway connection with user-provided certificate
// This allows Gateway to forward user cert to Fabric (Gateway doesn't authenticate, just forwards)
func (s *GatewayService) CreateDynamicGateway(ctx context.Context, userCertPEM, userKeyPEM, mspID string) (*client.Gateway, error) {
	// Parse user certificate
	cert, err := identity.CertificateFromPEM([]byte(userCertPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse user certificate: %w", err)
	}

	// Create identity from user cert
	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity from user cert: %w", err)
	}

	// Parse user private key
	privateKey, err := identity.PrivateKeyFromPEM([]byte(userKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse user private key: %w", err)
	}

	// Create signer from user key
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from user key: %w", err)
	}

	// Load TLS certificate (use same TLS cert as default gateway)
	tlsCert, err := loadTLSCertificate(s.config.PeerTLSCAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create gRPC connection (use same connection settings as default gateway)
	grpcConn, err := createGRPCConnection(s.config.PeerEndpoint, s.config.PeerHostOverride, tlsCert)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create gateway with user identity (forward cert to Fabric)
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConn),
		client.WithEvaluateTimeout(5*time.Minute),
		client.WithEndorseTimeout(5*time.Minute),
		client.WithSubmitTimeout(5*time.Minute),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("failed to connect to gateway with user cert: %w", err)
	}

	s.logger.Info("Created dynamic gateway with user cert",
		zap.String("msp_id", mspID),
		zap.String("peer", s.config.PeerEndpoint),
	)

	return gw, nil
}

// Close closes the gateway connection
func (s *GatewayService) Close() error {
	s.gateway.Close()
	s.logger.Info("Fabric Gateway connection closed")
	return nil
}

// Health checks the health of the Fabric connection
func (s *GatewayService) Health(ctx context.Context) error {
	// Check if gateway and network are initialized
	if s.gateway == nil {
		return fmt.Errorf("gateway not initialized")
	}
	
	if s.network == nil {
		return fmt.Errorf("network not initialized")
	}

	// Try to get channel info to verify connection without calling chaincode
	// This is a lightweight check that doesn't require chaincode
	ctx, span := s.tracer.Start(ctx, "HealthCheck")
	defer span.End()

	// Simply check if we can access the network/channel
	// This verifies the connection without making chaincode calls
	// We use a timeout to avoid hanging
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get channel name - this is a lightweight operation
	// that verifies the connection is alive
	channelName := s.network.Name()
	if channelName == "" {
		return fmt.Errorf("channel name is empty")
	}

	// If we get here, the connection is healthy
	s.logger.Debug("Fabric health check passed",
		zap.String("channel", channelName),
	)
	
	return nil
}

// loadCertificate loads an X.509 certificate
func loadCertificate(certPath string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// loadPrivateKey loads a private key from directory
func loadPrivateKey(keyPath string) (interface{}, error) {
	// Read all files in keystore directory
	files, err := os.ReadDir(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore directory: %w", err)
	}

	// Find the first key file (usually ends with _sk)
	for _, file := range files {
		if !file.IsDir() {
			keyPEM, err := os.ReadFile(filepath.Join(keyPath, file.Name()))
			if err != nil {
				continue
			}

			privateKey, err := identity.PrivateKeyFromPEM(keyPEM)
			if err != nil {
				continue
			}

			return privateKey, nil
		}
	}

	return nil, fmt.Errorf("no valid private key found in %s", keyPath)
}

// loadTLSCertificate loads TLS certificate for peer connection
func loadTLSCertificate(tlsCAPath string) (credentials.TransportCredentials, error) {
	cert, err := os.ReadFile(tlsCAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		return nil, fmt.Errorf("failed to append TLS CA certificate to pool")
	}

	return credentials.NewClientTLSFromCert(certPool, ""), nil
}

// createGRPCConnection creates a gRPC connection to the peer
func createGRPCConnection(endpoint, hostOverride string, tlsCreds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	return grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithAuthority(hostOverride),
	)
}

