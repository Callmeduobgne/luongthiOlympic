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

