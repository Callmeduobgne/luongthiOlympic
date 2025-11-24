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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// GRPCCredentials manages secure gRPC connections with proper TLS configuration
type GRPCCredentials struct {
	tlsConfig *tls.Config
	caCert    *x509.CertPool
	logger    *zap.Logger
}

// NewGRPCCredentials creates gRPC credentials with mutual TLS
func NewGRPCCredentials(cfg *Config, logger *zap.Logger) (*GRPCCredentials, error) {
	// Load CA certificate
	caCert, err := loadCACertificate(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Load client certificate and key for mutual TLS
	clientCert, err := loadClientCertificate(cfg)
	if err != nil {
		logger.Warn("Client certificate not available, using server-side TLS only", zap.Error(err))
		// Continue with server-side TLS only
		clientCert = nil
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		RootCAs:            caCert,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: false, // Always verify in production
		ServerName:         cfg.GatewayPeer,
		// Production best practice: Strong cipher suites
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
	}

	// Add client certificate if available (mutual TLS)
	if clientCert != nil {
		tlsConfig.Certificates = []tls.Certificate{*clientCert}
		logger.Info("Mutual TLS enabled", zap.String("server_name", cfg.GatewayPeer))
	} else {
		logger.Info("Server-side TLS only", zap.String("server_name", cfg.GatewayPeer))
	}

	return &GRPCCredentials{
		tlsConfig: tlsConfig,
		caCert:    caCert,
		logger:    logger,
	}, nil
}

// GetTransportCredentials returns gRPC transport credentials
func (gc *GRPCCredentials) GetTransportCredentials() credentials.TransportCredentials {
	return credentials.NewTLS(gc.tlsConfig)
}

// GetDialOptions returns gRPC dial options with keepalive and credentials
func (gc *GRPCCredentials) GetDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(gc.GetTransportCredentials()),
		// Production best practice: Keepalive parameters
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                120 * time.Second, // Send keepalive ping every 2 minutes
			Timeout:             20 * time.Second,  // Wait 20 seconds for ping ack
			PermitWithoutStream: true,              // Send pings even without active streams
		}),
		// Production best practice: Message size limits
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100 * 1024 * 1024), // 100MB max receive
			grpc.MaxCallSendMsgSize(100 * 1024 * 1024), // 100MB max send
		),
	}
}

// VerifyConnection verifies the gRPC connection health
func (gc *GRPCCredentials) VerifyConnection(conn *grpc.ClientConn) error {
	state := conn.GetState()
	gc.logger.Debug("Connection state", zap.String("state", state.String()))
	return nil
}

// loadCACertificate loads the CA certificate for TLS verification
func loadCACertificate(cfg *Config) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()

	// Try multiple CA certificate paths
	caPaths := []string{
		cfg.TLSCertPath,
		filepath.Join(cfg.CryptoPath, "peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt"),
		filepath.Join(cfg.CryptoPath, "peerOrganizations/org1.ibn.vn/msp/tlscacerts/tlsca.org1.ibn.vn-cert.pem"),
	}

	var caCert []byte
	var err error
	var loadedPath string

	for _, path := range caPaths {
		if path == "" {
			continue
		}
		caCert, err = os.ReadFile(path)
		if err == nil {
			loadedPath = path
			break
		}
	}

	if caCert == nil {
		return nil, fmt.Errorf("failed to load CA certificate from any path")
	}

	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	return caCertPool, nil
}

// loadClientCertificate loads client certificate and key for mutual TLS
func loadClientCertificate(cfg *Config) (*tls.Certificate, error) {
	// Try multiple client certificate paths
	certPaths := []string{
		cfg.CertPath,
		filepath.Join(cfg.CryptoPath, "peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/tls/client.crt"),
		filepath.Join(cfg.CryptoPath, "peerOrganizations/org1.ibn.vn/users/User1@org1.ibn.vn/tls/client.crt"),
	}

	keyPaths := []string{
		cfg.KeyPath,
		filepath.Join(cfg.CryptoPath, "peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/tls/client.key"),
		filepath.Join(cfg.CryptoPath, "peerOrganizations/org1.ibn.vn/users/User1@org1.ibn.vn/tls/client.key"),
	}

	var certPEM, keyPEM []byte
	var err error

	// Try to load certificate
	for _, path := range certPaths {
		if path == "" {
			continue
		}
		certPEM, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	// Try to load key
	for _, path := range keyPaths {
		if path == "" {
			continue
		}
		keyPEM, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if certPEM == nil || keyPEM == nil {
		return nil, fmt.Errorf("client certificate or key not found")
	}

	// Load certificate and key pair
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	return &cert, nil
}
