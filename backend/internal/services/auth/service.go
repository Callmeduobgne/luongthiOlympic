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
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/config"
	"github.com/ibn-network/backend/internal/infrastructure/cache"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication business logic
type Service struct {
	repo   *Repository
	cache  *cache.MultiLayerCache
	cfg    *config.JWTConfig
	logger *zap.Logger
}

// NewService creates a new auth service
func NewService(
	repo *Repository,
	cache *cache.MultiLayerCache,
	cfg *config.JWTConfig,
	logger *zap.Logger,
) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		cfg:    cfg,
		logger: logger,
	}
}

// Register registers a new user
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	// Check if user already exists
	existingUser, _ := s.repo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Validate role
	if !IsValidRole(req.Role) {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	username, err := s.generateUniqueUsername(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate username: %w", err)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:            uuid.New(),
		Email:         req.Email,
		Username:      username,
		PasswordHash:  string(passwordHash),
		FullName:      req.FullName,
		Role:          req.Role,
		MSPID:         req.MSPID,
		IsActive:      true,
		EmailVerified: false,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User registered",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
		zap.String("role", user.Role),
	)

	// Remove password hash before returning
	user.PasswordHash = ""

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Login attempt for non-existent user", zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Login attempt for inactive user", zap.String("email", req.Email))
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logger.Warn("Login attempt with invalid password", zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, expiresAt, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Warn("Failed to update last login", zap.Error(err))
	}

	// Cache user data
	if err := s.cacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache user", zap.Error(err))
	}

	s.logger.Info("User logged in",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	// Remove password hash before returning
	user.PasswordHash = ""

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

// RefreshAccessToken generates a new access token from refresh token
func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*LoginResponse, error) {
	// Hash refresh token
	tokenHash := s.hashToken(refreshTokenStr)

	// Get refresh token from database
	refreshToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token is revoked
	if refreshToken.IsRevoked {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	// Check if token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Generate new access token
	accessToken, expiresAt, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Remove password hash before returning
	user.PasswordHash = ""

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr, // Return same refresh token
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

// GenerateAccessToken generates a JWT access token
func (s *Service) GenerateAccessToken(user *User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.AccessTokenTTL)

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"msp_id":  user.MSPID,
		"iat":     time.Now().Unix(),
		"exp":     expiresAt.Unix(),
		"iss":     s.cfg.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken(ctx context.Context, user *User) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	plainToken := hex.EncodeToString(tokenBytes)

	// Hash token for storage
	tokenHash := s.hashToken(plainToken)

	// Create refresh token record
	refreshToken := &RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
	}

	if err := s.repo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return plainToken, nil
}

// VerifyAccessToken verifies and parses a JWT access token
func (s *Service) VerifyAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Parse claims
	userIDStr, _ := claims["user_id"].(string)
	if userIDStr == "" {
		// Fallback to camelCase (tokens issued by API Gateway)
		userIDStr, _ = claims["userId"].(string)
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	var mspID *string
	if mspIDVal, ok := claims["msp_id"].(string); ok && mspIDVal != "" {
		mspID = &mspIDVal
	} else if mspIDVal, ok := claims["mspId"].(string); ok && mspIDVal != "" {
		mspID = &mspIDVal
	}

	return &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		MSPID:  mspID,
	}, nil
}

// CreateAPIKey creates a new API key for a user
func (s *Service) CreateAPIKey(ctx context.Context, userID uuid.UUID, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	// Generate API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	plainKey := "ibn_" + hex.EncodeToString(keyBytes)

	// Hash key for storage
	keyHash := s.hashToken(plainKey)

	// Create API key record
	apiKey := &APIKey{
		ID:          uuid.New(),
		UserID:      userID,
		KeyHash:     keyHash,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		IsActive:    true,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := s.repo.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	s.logger.Info("API key created",
		zap.String("api_key_id", apiKey.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("name", apiKey.Name),
	)

	return &CreateAPIKeyResponse{
		APIKey:   apiKey,
		PlainKey: plainKey,
	}, nil
}

// VerifyAPIKey verifies an API key and returns the associated user
func (s *Service) VerifyAPIKey(ctx context.Context, plainKey string) (*User, error) {
	// Hash key
	keyHash := s.hashToken(plainKey)

	// Try cache first
	cacheKey := fmt.Sprintf("api_key:%s", keyHash)
	var cachedUser User
	err := s.cache.Get(ctx, cacheKey, &cachedUser, func(ctx context.Context) (interface{}, error) {
		// Get API key from database
		apiKey, err := s.repo.GetAPIKeyByHash(ctx, keyHash)
		if err != nil {
			return nil, err
		}

		// Check if API key is active
		if !apiKey.IsActive {
			return nil, fmt.Errorf("API key is inactive")
		}

		// Check if API key is expired
		if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
			return nil, fmt.Errorf("API key has expired")
		}

		// Get user
		user, err := s.repo.GetUserByID(ctx, apiKey.UserID)
		if err != nil {
			return nil, err
		}

		// Update last used (async)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.repo.UpdateAPIKeyLastUsed(ctx, apiKey.ID); err != nil {
				s.logger.Warn("Failed to update API key last used", zap.Error(err))
			}
		}()

		return user, nil
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 1 * time.Hour,
	})

	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	return &cachedUser, nil
}

// GetUserByID retrieves a user by ID (with caching)
func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	cacheKey := fmt.Sprintf("user:%s", userID.String())

	var user User
	err := s.cache.Get(ctx, cacheKey, &user, func(ctx context.Context) (interface{}, error) {
		return s.repo.GetUserByID(ctx, userID)
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 30 * time.Minute,
	})

	if err != nil {
		return nil, err
	}

	// Remove password hash
	user.PasswordHash = ""

	return &user, nil
}

// UpdateAvatar updates user's avatar URL
func (s *Service) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	// Update in database
	if err := s.repo.UpdateUserAvatar(ctx, userID, avatarURL); err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", userID.String())
	s.cache.Delete(ctx, cacheKey)

	s.logger.Info("Avatar updated",
		zap.String("user_id", userID.String()),
		zap.String("avatar_url", avatarURL),
	)

	return nil
}

// Logout revokes refresh token
func (s *Service) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := s.hashToken(refreshTokenStr)

	refreshToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		// Token not found, already logged out
		return nil
	}

	if err := s.repo.RevokeRefreshToken(ctx, refreshToken.ID); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	s.logger.Info("User logged out", zap.String("user_id", refreshToken.UserID.String()))

	return nil
}

// cacheUser caches user data
func (s *Service) cacheUser(ctx context.Context, user *User) error {
	cacheKey := fmt.Sprintf("user:%s", user.ID.String())

	// Don't cache password hash
	cachedUser := *user
	cachedUser.PasswordHash = ""

	return s.cache.Set(ctx, cacheKey, &cachedUser, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 30 * time.Minute,
	})
}

// hashToken creates SHA-256 hash of a token
func (s *Service) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *Service) generateUniqueUsername(ctx context.Context, email string) (string, error) {
	base := baseUsernameFromEmail(email)
	candidate := base

	for attempts := 0; attempts < 5; attempts++ {
		taken, err := s.repo.IsUsernameTaken(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}

		candidate = fmt.Sprintf("%s-%s", base, uuid.New().String()[:4])
	}

	return "", fmt.Errorf("could not generate unique username")
}

func baseUsernameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	base := email
	if len(parts) > 0 && parts[0] != "" {
		base = parts[0]
	}
	return sanitizeUsername(base)
}

func sanitizeUsername(input string) string {
	input = strings.ToLower(input)
	var builder strings.Builder
	lastDash := false

	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash && builder.Len() > 0 {
			builder.WriteRune('-')
			lastDash = true
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return fmt.Sprintf("user-%s", uuid.New().String()[:6])
	}
	return result
}
