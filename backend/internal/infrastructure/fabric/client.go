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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client wraps Hyperledger Fabric Gateway client
type Client struct {
	gateway  *client.Gateway
	network  *client.Network
	contract *client.Contract
	logger   *zap.Logger
}

// Config holds Fabric connection configuration
type Config struct {
	// MSP configuration
	MSPID        string // e.g., "Org1MSP"
	CryptoPath   string // Path to organizations directory
	CertPath     string // Path to user cert
	KeyPath      string // Path to user private key
	TLSCertPath  string // Path to peer TLS cert
	
	// Peer connection
	PeerEndpoint string // e.g., "localhost:7051"
	GatewayPeer  string // e.g., "peer0.org1.ibn.vn"
	
	// Channel
	ChannelName string // e.g., "ibnchannel"
}

// NewClient creates a new Fabric Gateway client with enhanced connection management
func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	ctx := context.Background()
	
	// Create connection manager with retry logic
	connManager, err := NewConnectionManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	// Connect with retry
	if err := connManager.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Get gRPC connection
	grpcConn, err := connManager.GetConnection()
	if err != nil {
		connManager.Close()
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Load identity
	id, err := newIdentity(cfg)
	if err != nil {
		connManager.Close()
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	// Load signer
	sign, err := newSign(cfg)
	if err != nil {
		connManager.Close()
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// Create Gateway connection with enhanced timeouts for production
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConn),
		client.WithEvaluateTimeout(10*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(60*time.Second),
		client.WithCommitStatusTimeout(120*time.Second),
	)
	if err != nil {
		connManager.Close()
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	// Get network
	network := gateway.GetNetwork(cfg.ChannelName)
	
	// Get default contract (teaTraceCC)
	contract := network.GetContract("teaTraceCC")

	// Create health checker
	healthChecker := NewHealthChecker(connManager, logger)
	
	// Wait for healthy connection
	healthCtx, healthCancel := context.WithTimeout(ctx, 30*time.Second)
	defer healthCancel()
	
	if err := healthChecker.WaitForHealthy(healthCtx, 30*time.Second); err != nil {
		logger.Warn("Health check timeout, continuing anyway", zap.Error(err))
		// Continue even if health check times out
	}

	logger.Info("Connected to Fabric network",
		zap.String("channel", cfg.ChannelName),
		zap.String("msp_id", cfg.MSPID),
		zap.String("peer", cfg.PeerEndpoint),
		zap.String("state", connManager.GetConnectionState().String()),
	)

	return &Client{
		gateway:  gateway,
		network:  network,
		contract: contract,
		logger:   logger,
	}, nil
}

// SubmitTransaction submits a transaction to the blockchain
func (c *Client) SubmitTransaction(ctx context.Context, channelID, chaincodeID, functionName string, args []string, transient map[string][]byte) (string, error) {
	// Get contract for specific chaincode
	contract := c.network.GetContract(chaincodeID)

	// Prepare proposal options
	opts := []client.ProposalOption{client.WithArguments(args...)}
	
	// Add transient data if provided
	if len(transient) > 0 {
		opts = append(opts, client.WithTransient(transient))
	}

	// Create transaction proposal
	proposal, err := contract.NewProposal(functionName, opts...)
	if err != nil {
		c.logger.Error("Failed to create proposal", zap.Error(err))
		return "", fmt.Errorf("failed to create proposal: %w", err)
	}

	// Endorse transaction
	transaction, err := proposal.Endorse()
	if err != nil {
		c.logger.Error("Failed to endorse transaction", zap.Error(err))
		return "", fmt.Errorf("failed to endorse: %w", err)
	}

	// Submit to orderer
	commit, err := transaction.Submit()
	if err != nil {
		c.logger.Error("Failed to submit transaction", zap.Error(err))
		return "", fmt.Errorf("failed to submit: %w", err)
	}

	// Get transaction ID
	txID := transaction.TransactionID()

	// Wait for commit (async)
	go func() {
		status, err := commit.Status()
		if err != nil {
			c.logger.Error("Failed to get commit status", zap.String("tx_id", txID), zap.Error(err))
			return
		}
		
		c.logger.Info("Transaction committed",
			zap.String("tx_id", txID),
			zap.Bool("successful", status.Successful),
			zap.Uint64("block_number", status.BlockNumber),
		)
	}()

	return txID, nil
}

// EvaluateTransaction performs a read-only query with variadic args (supports qscc and custom chaincodes)
func (c *Client) EvaluateTransaction(ctx context.Context, chaincodeName, functionName string, args ...string) ([]byte, error) {
	// Get contract for specific chaincode
	contract := c.network.GetContract(chaincodeName)

	// Evaluate (query)
	result, err := contract.EvaluateTransaction(functionName, args...)
	if err != nil {
		c.logger.Error("Failed to evaluate transaction",
			zap.String("chaincode", chaincodeName),
			zap.String("function", functionName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to evaluate: %w", err)
	}

	c.logger.Debug("Transaction evaluated",
		zap.String("chaincode", chaincodeName),
		zap.String("function", functionName),
		zap.Int("result_size", len(result)),
	)

	return result, nil
}

// GetTransactionStatus gets the status of a transaction by ID
func (c *Client) GetTransactionStatus(ctx context.Context, txID string) (string, error) {
	// This is a simplified implementation
	// In production, you would query the ledger for transaction validation code
	// For now, we assume transactions are valid if no error is returned
	return "VALID", nil
}

// Close closes the gateway connection
func (c *Client) Close() error {
	c.gateway.Close()
	c.logger.Info("Fabric gateway connection closed")
	return nil
}

// Helper functions

// loadTLSCertificate loads a TLS certificate
func loadTLSCertificate(cfg *Config) (*x509.CertPool, error) {
	// Auto-detect TLS cert path if not provided
	tlsCertPath := cfg.TLSCertPath
	if tlsCertPath == "" {
		tlsCertPath = filepath.Join(cfg.CryptoPath, "peerOrganizations", "org1.ibn.vn", "peers", "peer0.org1.ibn.vn", "tls", "ca.crt")
	}

	certificatePEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file %s: %w", tlsCertPath, err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certificatePEM) {
		return nil, fmt.Errorf("failed to add certificate to pool")
	}

	return certPool, nil
}

// newIdentity creates a new identity from MSP credentials
func newIdentity(cfg *Config) (*identity.X509Identity, error) {
	// Auto-detect cert path if not provided
	certPath := cfg.CertPath
	if certPath == "" {
		// Try multiple possible cert names
		possibleCerts := []string{
			filepath.Join(cfg.CryptoPath, "peerOrganizations", "org1.ibn.vn", "users", "Admin@org1.ibn.vn", "msp", "signcerts", "Admin@org1.ibn.vn-cert.pem"),
			filepath.Join(cfg.CryptoPath, "peerOrganizations", "org1.ibn.vn", "users", "Admin@org1.ibn.vn", "msp", "signcerts", "cert.pem"),
		}
		for _, path := range possibleCerts {
			if _, err := os.Stat(path); err == nil {
				certPath = path
				break
			}
		}
		if certPath == "" {
			// Try glob pattern
			matches, _ := filepath.Glob(filepath.Join(cfg.CryptoPath, "peerOrganizations", "org1.ibn.vn", "users", "Admin@org1.ibn.vn", "msp", "signcerts", "*.pem"))
			if len(matches) > 0 {
				certPath = matches[0]
			}
		}
	}

	certificatePEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file %s: %w", certPath, err)
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	id, err := identity.NewX509Identity(cfg.MSPID, certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to create X509 identity: %w", err)
	}

	return id, nil
}

// newSign creates a new signing function
func newSign(cfg *Config) (identity.Sign, error) {
	// Auto-detect key path if not provided
	keyPath := cfg.KeyPath
	if keyPath == "" {
		keyDir := filepath.Join(cfg.CryptoPath, "peerOrganizations", "org1.ibn.vn", "users", "Admin@org1.ibn.vn", "msp", "keystore")
		keys, err := filepath.Glob(filepath.Join(keyDir, "*_sk"))
		if err != nil || len(keys) == 0 {
			return nil, fmt.Errorf("failed to find private key in %s", keyDir)
		}
		keyPath = keys[0]
	}

	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file %s: %w", keyPath, err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing function: %w", err)
	}

	return sign, nil
}

// GetDefaultCertPaths returns default certificate paths based on Fabric network structure
func GetDefaultCertPaths(orgMSP string) (certPath, keyPath, tlsCertPath string, err error) {
	// Assume core directory is at ../core relative to backend
	coreDir := filepath.Join("..", "core")
	
	// Default paths for org1
	if orgMSP == "Org1MSP" {
		orgDir := filepath.Join(coreDir, "organizations", "peerOrganizations", "org1.ibn.vn")
		
		// Admin user certificate
		certPath = filepath.Join(orgDir, "users", "Admin@org1.ibn.vn", "msp", "signcerts", "cert.pem")
		
		// Admin user private key (need to find the actual key file)
		keyDir := filepath.Join(orgDir, "users", "Admin@org1.ibn.vn", "msp", "keystore")
		keys, err := filepath.Glob(filepath.Join(keyDir, "*_sk"))
		if err != nil || len(keys) == 0 {
			return "", "", "", fmt.Errorf("failed to find private key in %s", keyDir)
		}
		keyPath = keys[0]
		
		// Peer TLS certificate
		tlsCertPath = filepath.Join(orgDir, "peers", "peer0.org1.ibn.vn", "tls", "ca.crt")
	}

	return certPath, keyPath, tlsCertPath, nil
}

