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

package models

// CAEnrollRequest represents a user enrollment request with Fabric CA
type CAEnrollRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// CAEnrollResponse represents enrollment response
type CAEnrollResponse struct {
	Username    string `json:"username"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
	MSPID       string `json:"mspId"`
}

// CARegisterRequest represents a user registration request with Fabric CA
type CARegisterRequest struct {
	Username    string   `json:"username" validate:"required"`
	Type        string   `json:"type" validate:"required,oneof=client peer orderer admin"`
	Affiliation string   `json:"affiliation" validate:"required"`
	Role        string   `json:"role" validate:"required,oneof=member admin"`
	Attributes  []string `json:"attributes,omitempty"`
}

// CARegisterResponse represents registration response
type CARegisterResponse struct {
	Username string `json:"username"`
	Secret   string `json:"secret"`
}

// CARevokeRequest represents a certificate revocation request
type CARevokeRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CAUserInfo represents user information from Fabric CA
type CAUserInfo struct {
	Username    string   `json:"username"`
	Type        string   `json:"type"`
	Affiliation string   `json:"affiliation"`
	Attributes  []string `json:"attributes"`
	Revoked     bool     `json:"revoked"`
}

