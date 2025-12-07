package listener

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/ibn-network/backend/internal/config"
	"github.com/ibn-network/backend/internal/services/blockchain/db"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Service implements the blockchain listener service
type Service struct {
	cfg       *config.FabricConfig
	dbService *db.Service
	logger    *zap.Logger
	network   *client.Network
	gateway   *client.Gateway
	conn      *grpc.ClientConn
	cancel    context.CancelFunc
}

// NewService creates a new listener service
func NewService(cfg *config.FabricConfig, dbService *db.Service, logger *zap.Logger) *Service {
	return &Service{
		cfg:       cfg,
		dbService: dbService,
		logger:    logger,
	}
}

// Start starts the listener
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Initializing Blockchain Listener...", 
		zap.String("msp_id", s.cfg.MSPID),
		zap.String("peer_endpoint", s.cfg.PeerEndpoint),
		zap.String("channel", s.cfg.ChannelName),
	)

	// Create gRPC connection
	if err := s.connectGRPC(); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Create Gateway connection
	if err := s.connectGateway(); err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	// Start listening for events
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go s.listenBlockEvents(ctx)

	s.logger.Info("Blockchain Listener started successfully")
	return nil
}

// Stop stops the listener
func (s *Service) Stop() {
	s.logger.Info("Stopping Blockchain Listener...")
	if s.cancel != nil {
		s.cancel()
	}
	if s.gateway != nil {
		s.gateway.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *Service) connectGRPC() error {
	certificate, err := loadCertificate(s.cfg.TLSCertPath)
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, s.cfg.GatewayPeer)

	conn, err := grpc.Dial(s.cfg.PeerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	s.conn = conn
	return nil
}

func (s *Service) connectGateway() error {
	id, err := newIdentity(s.cfg.CertPath, s.cfg.MSPID)
	if err != nil {
		return err
	}

	sign, err := newSign(s.cfg.KeyPath)
	if err != nil {
		return err
	}

	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(s.conn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	s.gateway = gateway
	s.network = gateway.GetNetwork(s.cfg.ChannelName)

	return nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func newIdentity(certPath, mspID string) (*identity.X509Identity, error) {
	certificatePEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, err
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func newSign(keyPath string) (identity.Sign, error) {
	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	return sign, nil
}
