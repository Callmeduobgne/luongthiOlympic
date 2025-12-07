package config

// CAConfig holds Fabric CA configuration (optional - for future CA server integration)
type CAConfig struct {
	URL         string `mapstructure:"url" validate:"omitempty,url"`
	CAName      string `mapstructure:"ca_name" validate:"omitempty"`
	TLSCertPath string `mapstructure:"tls_cert_path" validate:"omitempty,file"`
	MSPDir      string `mapstructure:"msp_dir" validate:"omitempty"`
	AdminUser   string `mapstructure:"admin_user" validate:"omitempty"`
	AdminPass   string `mapstructure:"admin_pass" validate:"omitempty"`
	MSPID       string `mapstructure:"msp_id" validate:"omitempty"`
}

