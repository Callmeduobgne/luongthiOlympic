package config

// FabricConfig holds Hyperledger Fabric network configuration
type FabricConfig struct {
	Channel          string `mapstructure:"channel" validate:"required"`
	Chaincode        string `mapstructure:"chaincode" validate:"required"`
	MSPId            string `mapstructure:"msp_id" validate:"required"`
	PeerEndpoint     string `mapstructure:"peer_endpoint" validate:"required"`
	PeerHostOverride string `mapstructure:"peer_host_override" validate:"required,hostname"`
	UserCertPath     string `mapstructure:"user_cert_path" validate:"required,file"`
	UserKeyPath      string `mapstructure:"user_key_path" validate:"required,dir"`
	PeerTLSCAPath    string `mapstructure:"peer_tls_ca_path" validate:"required,file"`
}

