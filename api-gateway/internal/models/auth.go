package models

import "time"

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"` // Can be username or email
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`
	MSPID    string `json:"mspId,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int64     `json:"expiresIn"` // seconds
	TokenType    string    `json:"tokenType"` // "Bearer"
	User         UserInfo  `json:"user"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	MSPId     string    `json:"mspId"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	MSPId  string `json:"mspId"`
	Role   string `json:"role"`
}

// APIKeyRequest represents an API key creation request
type APIKeyRequest struct {
	Name        string   `json:"name" validate:"required"`
	Permissions []string `json:"permissions,omitempty"`
	RateLimit   int      `json:"rateLimit,omitempty"`
	ExpiresIn   int      `json:"expiresIn,omitempty"` // days
}

// APIKeyResponse represents an API key response
type APIKeyResponse struct {
	Key       string    `json:"key"` // Only shown once on creation
	KeyID     string    `json:"keyId"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// APIKeyInfo represents API key information (without the key itself)
type APIKeyInfo struct {
	KeyID      string    `json:"keyId"`
	Name       string    `json:"name"`
	LastUsedAt time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt  time.Time `json:"expiresAt,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

