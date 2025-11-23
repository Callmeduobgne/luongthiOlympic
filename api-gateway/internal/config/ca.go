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

