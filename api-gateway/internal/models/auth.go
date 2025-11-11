package models

import "time"

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	MSPId    string `json:"mspId" validate:"required,oneof=Org1MSP Org2MSP Org3MSP"`
	Role     string `json:"role" validate:"required,oneof=farmer verifier admin"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
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

