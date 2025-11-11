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

// Close closes the gateway connection
func (s *GatewayService) Close() error {
	s.gateway.Close()
	s.logger.Info("Fabric Gateway connection closed")
	return nil
}

// Health checks the health of the Fabric connection
func (s *GatewayService) Health(ctx context.Context) error {
	// Try to evaluate a simple query
	_, err := s.EvaluateTransaction(ctx, "getBatchInfo", "health-check")
	if err != nil {
		// It's ok if batch doesn't exist, we just want to check connection
		return nil
	}
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

