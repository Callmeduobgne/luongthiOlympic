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

// FabricConfig holds Hyperledger Fabric network configuration
type FabricConfig struct {
	Channel          string   `mapstructure:"channel" validate:"required"`
	Chaincode        string   `mapstructure:"chaincode" validate:"required"`
	MSPId            string   `mapstructure:"msp_id" validate:"required"`
	PeerEndpoint     string   `mapstructure:"peer_endpoint" validate:"required"`
	PeerHostOverride string   `mapstructure:"peer_host_override" validate:"required,hostname"`
	UserCertPath     string   `mapstructure:"user_cert_path" validate:"required,file"`
	UserKeyPath      string   `mapstructure:"user_key_path" validate:"required,dir"`
	PeerTLSCAPath    string   `mapstructure:"peer_tls_ca_path" validate:"required,file"`
	// Additional peers (comma-separated, format: host:port)
	AdditionalPeers  []string `mapstructure:"additional_peers"`
	// Orderers (comma-separated, format: host:port)
	Orderers         []string `mapstructure:"orderers"`
	// Fabric CA endpoints (comma-separated, format: host:port)
	CAEndpoints      []string `mapstructure:"ca_endpoints"`
}

