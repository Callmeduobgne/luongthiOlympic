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

package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	PasswordHash  string     `json:"-"` // Never expose password hash
	FullName      *string    `json:"full_name,omitempty"`
	Role          string     `json:"role"`
	MSPID         *string    `json:"msp_id,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	IsActive      bool       `json:"is_active"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// APIKey represents an API key
type APIKey struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	KeyHash     string                 `json:"-"` // Never expose key hash
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	IsActive    bool                   `json:"is_active"`
	LastUsedAt  *time.Time             `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  string     `json:"-"` // Never expose token hash
	DeviceInfo *string    `json:"device_info,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	IsRevoked  bool       `json:"is_revoked"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user"`
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	FullName *string `json:"full_name,omitempty"`
	Role     string  `json:"role" validate:"required,oneof=user admin operator"`
	MSPID    *string `json:"msp_id,omitempty"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// CreateAPIKeyRequest represents API key creation request
type CreateAPIKeyRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Description *string                `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse represents API key creation response
type CreateAPIKeyResponse struct {
	APIKey   *APIKey `json:"api_key"`
	PlainKey string  `json:"plain_key"` // Only returned on creation
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	MSPID     *string   `json:"msp_id,omitempty"`
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
	Issuer    string    `json:"iss"`
}

// UserRole constants
const (
	RoleUser     = "user"
	RoleOperator = "operator"
	RoleAdmin    = "admin"
)

// IsValidRole checks if role is valid
func IsValidRole(role string) bool {
	return role == RoleUser || role == RoleOperator || role == RoleAdmin
}
