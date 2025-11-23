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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/middleware"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/repository/db"
	"github.com/ibn-network/api-gateway/internal/services/cache"
	"github.com/ibn-network/api-gateway/internal/services/identity"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Account lockout configuration
	MaxFailedLoginAttempts = 5
	AccountLockoutDuration = 15 * time.Minute
)

// Service handles authentication operations
type Service struct {
	db              *pgxpool.Pool
	queries         *db.Queries
	authMW          *middleware.AuthMiddleware
	jwtConfig       *config.JWTConfig
	logger          *zap.Logger
	identityService *identity.Service // For Fabric identity verification
	cache           *cache.Service    // For tracking failed login attempts
}

// NewService creates a new auth service
func NewService(dbPool *pgxpool.Pool, authMW *middleware.AuthMiddleware, jwtConfig *config.JWTConfig, identityService *identity.Service, cacheService *cache.Service, logger *zap.Logger) *Service {
	return &Service{
		db:              dbPool,
		queries:         db.New(),
		authMW:          authMW,
		jwtConfig:       jwtConfig,
		logger:          logger,
		identityService: identityService,
		cache:           cacheService,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *models.RegisterRequest) (*models.UserInfo, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Set default MSPID if not provided
	mspID := req.MSPID
	if mspID == "" {
		mspID = "Org1MSP" // Default MSP
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, s.db, db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		MspID:        mspID,
		Role:         role,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert UUID to string
	userIDStr := uuidToString(user.ID)

	s.logger.Info("User registered successfully",
		zap.String("email", req.Email),
		zap.String("role", role),
	)

	return &models.UserInfo{
		ID:        userIDStr,
		Email:     user.Email,
		MSPId:     user.MspID,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}

// checkAccountLockout checks if account is locked and returns lockout status
func (s *Service) checkAccountLockout(ctx context.Context, userID string) (bool, time.Time, error) {
	if s.cache == nil {
		return false, time.Time{}, nil
	}

	key := fmt.Sprintf("account_lockout:%s", userID)
	lockedUntilStr, err := s.cache.Get(ctx, key)
	if err != nil {
		s.logger.Warn("Failed to check account lockout", zap.Error(err))
		return false, time.Time{}, nil // Don't block on cache errors
	}

	if lockedUntilStr == "" {
		return false, time.Time{}, nil
	}

	lockedUntil, err := time.Parse(time.RFC3339, lockedUntilStr)
	if err != nil {
		s.logger.Warn("Failed to parse lockout timestamp", zap.Error(err))
		return false, time.Time{}, nil
	}

	if time.Now().Before(lockedUntil) {
		return true, lockedUntil, nil
	}

	// Lockout expired, remove from cache
	_ = s.cache.Delete(ctx, key)
	return false, time.Time{}, nil
}

// incrementFailedAttempts increments failed login attempts and locks account if threshold reached
func (s *Service) incrementFailedAttempts(ctx context.Context, userID string) error {
	if s.cache == nil {
		return nil
	}

	key := fmt.Sprintf("failed_attempts:%s", userID)
	attemptsStr, err := s.cache.Get(ctx, key)
	if err != nil {
		s.logger.Warn("Failed to get failed attempts", zap.Error(err))
		return nil
	}

	attempts := 0
	if attemptsStr != "" {
		fmt.Sscanf(attemptsStr, "%d", &attempts)
	}

	attempts++

	// Store attempts with expiry (reset after lockout duration)
	if err := s.cache.Set(ctx, key, fmt.Sprintf("%d", attempts), AccountLockoutDuration); err != nil {
		s.logger.Warn("Failed to store failed attempts", zap.Error(err))
	}

	// Lock account if threshold reached
	if attempts >= MaxFailedLoginAttempts {
		lockoutKey := fmt.Sprintf("account_lockout:%s", userID)
		lockedUntil := time.Now().Add(AccountLockoutDuration)
		if err := s.cache.Set(ctx, lockoutKey, lockedUntil.Format(time.RFC3339), AccountLockoutDuration); err != nil {
			s.logger.Warn("Failed to lock account", zap.Error(err))
		}

		s.logger.Warn("Account locked due to too many failed login attempts",
			zap.String("user_id", userID),
			zap.Int("attempts", attempts),
			zap.Time("locked_until", lockedUntil),
		)
	}

	return nil
}

// resetFailedAttempts resets failed login attempts on successful login
func (s *Service) resetFailedAttempts(ctx context.Context, userID string) {
	if s.cache == nil {
		return
	}

	key := fmt.Sprintf("failed_attempts:%s", userID)
	_ = s.cache.Delete(ctx, key)

	lockoutKey := fmt.Sprintf("account_lockout:%s", userID)
	_ = s.cache.Delete(ctx, lockoutKey)
}

// Login authenticates a user and returns JWT tokens
func (s *Service) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Log login attempt
	s.logger.Info("Login attempt",
		zap.String("username", req.Username),
		zap.Bool("has_password", len(req.Password) > 0),
	)

	// Find user by username or email
	user, err := s.queries.GetUserByUsernameOrEmail(ctx, s.db, req.Username)
	if err != nil {
		s.logger.Warn("User not found",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		// Don't reveal if user exists - return generic error
		return nil, fmt.Errorf("invalid credentials")
	}

	userIDStr := uuidToString(user.ID)

	// Check if account is locked
	isLocked, lockedUntil, err := s.checkAccountLockout(ctx, userIDStr)
	if err != nil {
		s.logger.Warn("Failed to check account lockout", zap.Error(err))
		// Continue with login attempt if check fails
	} else if isLocked {
		remainingTime := time.Until(lockedUntil)
		s.logger.Warn("Login attempt for locked account",
			zap.String("email", user.Email),
			zap.Time("locked_until", lockedUntil),
			zap.Duration("remaining", remainingTime),
		)
		return nil, fmt.Errorf("account is locked due to too many failed login attempts. Please try again after %v", remainingTime.Round(time.Minute))
	}

	s.logger.Info("User found",
		zap.String("email", user.Email),
		zap.String("role", user.Role),
		zap.Bool("is_active", user.IsActive.Bool),
	)

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		s.logger.Warn("Password verification failed",
			zap.String("email", user.Email),
			zap.Error(err),
		)
		// Increment failed attempts
		_ = s.incrementFailedAttempts(ctx, userIDStr)
		return nil, fmt.Errorf("invalid credentials")
	}

	s.logger.Info("Password verified successfully", zap.String("email", user.Email))

	// Reset failed attempts on successful login
	s.resetFailedAttempts(ctx, userIDStr)

	// Check if user is active
	if !user.IsActive.Bool {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify user identity in Fabric blockchain
	// Check if user has valid certificate in Fabric network
	if s.identityService != nil {
		// Try to get user from Fabric identity service
		// Use email as username for Fabric identity lookup
		fabricUser, err := s.identityService.GetUser(ctx, user.Email)
		if err != nil {
			s.logger.Warn("User not found in Fabric network",
				zap.String("email", user.Email),
				zap.Error(err),
			)
			// For now, we allow login even if Fabric identity doesn't exist
			// In production, you might want to enforce this check:
			// return nil, fmt.Errorf("user identity not found in blockchain network")
		} else {
			s.logger.Info("User verified in Fabric network",
				zap.String("email", user.Email),
				zap.String("fabric_username", fabricUser.Username),
				zap.Bool("revoked", fabricUser.Revoked),
			)
			// Check if certificate is revoked
			if fabricUser.Revoked {
				return nil, fmt.Errorf("user certificate has been revoked in blockchain network")
			}
		}
	}

	// Generate access token
	expiry := s.jwtConfig.Expiry
	accessToken, err := s.authMW.GenerateToken(
		userIDStr,
		user.Email,
		user.MspID,
		user.Role,
		expiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.authMW.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token for storage
	refreshTokenHash := s.hashToken(refreshToken)

	// Store refresh token (7 days expiry)
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	_, err = s.queries.CreateRefreshToken(ctx, s.db, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: pgtype.Timestamp{Time: refreshExpiry, Valid: true},
	})
	if err != nil {
		s.logger.Warn("Failed to store refresh token", zap.Error(err))
		// Continue anyway - refresh token is optional
	}

	s.logger.Info("User logged in successfully",
		zap.String("email", user.Email),
		zap.String("role", user.Role),
	)

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(expiry.Seconds()),
		TokenType:    "Bearer",
		User: models.UserInfo{
			ID:        userIDStr,
			Email:     user.Email,
			MSPId:     user.MspID,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

// RefreshToken generates a new access token from refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// Hash provided token
	tokenHash := s.hashToken(refreshToken)

	// Find refresh token
	rt, err := s.queries.GetRefreshToken(ctx, s.db, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check expiration
	if !rt.ExpiresAt.Valid || rt.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Get user
	user, err := s.queries.GetUser(ctx, s.db, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive.Bool {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Generate new access token
	expiry := s.jwtConfig.Expiry
	userIDStr := uuidToString(user.ID)
	accessToken, err := s.authMW.GenerateToken(
		userIDStr,
		user.Email,
		user.MspID,
		user.Role,
		expiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token (rotate)
	newRefreshToken, err := s.authMW.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Revoke old refresh token
	err = s.queries.RevokeRefreshToken(ctx, s.db, tokenHash)
	if err != nil {
		s.logger.Warn("Failed to revoke old refresh token", zap.Error(err))
	}

	// Store new refresh token
	newTokenHash := s.hashToken(newRefreshToken)
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	_, err = s.queries.CreateRefreshToken(ctx, s.db, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: pgtype.Timestamp{Time: refreshExpiry, Valid: true},
	})
	if err != nil {
		s.logger.Warn("Failed to store new refresh token", zap.Error(err))
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(expiry.Seconds()),
		TokenType:    "Bearer",
		User: models.UserInfo{
			ID:        userIDStr,
			Email:     user.Email,
			MSPId:     user.MspID,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

// GenerateAPIKey creates a new API key for a user
func (s *Service) GenerateAPIKey(ctx context.Context, userID string, req *models.APIKeyRequest) (*models.APIKeyResponse, error) {
	// Parse user ID
	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Generate random API key (32 bytes = 256 bits)
	apiKey, err := s.generateRandomKey(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash key for storage
	keyHash := s.hashToken(apiKey)

	// Prepare permissions (JSONB)
	var permissionsJSON []byte
	if len(req.Permissions) > 0 {
		permissionsJSON, err = json.Marshal(req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal permissions: %w", err)
		}
	}

	// Set rate limit (default 1000)
	rateLimit := int32(1000)
	if req.RateLimit > 0 {
		rateLimit = int32(req.RateLimit)
	}

	// Set expiration
	var expiresAt pgtype.Timestamptz
	if req.ExpiresIn > 0 {
		expiresAt = pgtype.Timestamptz{
			Time:  time.Now().Add(time.Duration(req.ExpiresIn) * 24 * time.Hour),
			Valid: true,
		}
	}

	// Store in database
	apiKeyRecord, err := s.queries.CreateAPIKey(ctx, s.db, db.CreateAPIKeyParams{
		UserID:      userUUID,
		KeyHash:     keyHash,
		Name:        pgtype.Text{String: req.Name, Valid: true},
		Permissions: permissionsJSON,
		RateLimit:   pgtype.Int4{Int32: rateLimit, Valid: true},
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Format API key as "ibn_<base64>" for better UX
	formattedKey := fmt.Sprintf("ibn_%s", apiKey)

	keyIDStr := uuidToString(apiKeyRecord.ID)
	var expiresAtTime time.Time
	if apiKeyRecord.ExpiresAt.Valid {
		expiresAtTime = apiKeyRecord.ExpiresAt.Time
	}

	return &models.APIKeyResponse{
		Key:       formattedKey,
		KeyID:     keyIDStr,
		Name:      req.Name,
		ExpiresAt: expiresAtTime,
		CreatedAt: apiKeyRecord.CreatedAt.Time,
	}, nil
}

// ValidateAPIKey validates an API key
func (s *Service) ValidateAPIKey(ctx context.Context, apiKey string) (*models.UserInfo, error) {
	// Remove prefix if present
	key := apiKey
	if len(apiKey) > 4 && apiKey[:4] == "ibn_" {
		key = apiKey[4:]
	}

	// Hash provided key
	keyHash := s.hashToken(key)

	// Find in database
	apiKeyRecord, err := s.queries.GetAPIKeyByHash(ctx, s.db, keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if active
	if !apiKeyRecord.IsActive.Bool {
		return nil, fmt.Errorf("API key is inactive")
	}

	// Check expiration
	if apiKeyRecord.ExpiresAt.Valid && apiKeyRecord.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("API key expired")
	}

	// Update last_used_at
	err = s.queries.UpdateAPIKeyLastUsed(ctx, s.db, apiKeyRecord.ID)
	if err != nil {
		s.logger.Warn("Failed to update API key last used", zap.Error(err))
	}

	// Get user
	user, err := s.queries.GetUser(ctx, s.db, apiKeyRecord.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive.Bool {
		return nil, fmt.Errorf("user account is inactive")
	}

	return &models.UserInfo{
		ID:        uuidToString(user.ID),
		Email:     user.Email,
		MSPId:     user.MspID,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}

// ListAPIKeys lists all API keys for a user
func (s *Service) ListAPIKeys(ctx context.Context, userID string) ([]*models.APIKeyInfo, error) {
	// Parse user ID
	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get API keys
	apiKeys, err := s.queries.ListAPIKeys(ctx, s.db, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Convert to APIKeyInfo
	result := make([]*models.APIKeyInfo, 0, len(apiKeys))
	for _, key := range apiKeys {
		keyInfo := &models.APIKeyInfo{
			KeyID:     uuidToString(key.ID),
			Name:      key.Name.String,
			CreatedAt: key.CreatedAt.Time,
		}

		if key.LastUsedAt.Valid {
			keyInfo.LastUsedAt = key.LastUsedAt.Time
		}

		if key.ExpiresAt.Valid {
			keyInfo.ExpiresAt = key.ExpiresAt.Time
		}

		result = append(result, keyInfo)
	}

	return result, nil
}

// RevokeAPIKey revokes an API key
func (s *Service) RevokeAPIKey(ctx context.Context, userID, keyID string) error {
	// Parse IDs
	userUUID, err := parseUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	keyUUID, err := parseUUID(keyID)
	if err != nil {
		return fmt.Errorf("invalid API key ID: %w", err)
	}

	// Get API key to verify ownership
	apiKey, err := s.queries.GetAPIKey(ctx, s.db, keyUUID)
	if err != nil {
		return fmt.Errorf("API key not found")
	}

	// Verify ownership
	if apiKey.UserID != userUUID {
		return fmt.Errorf("API key not found")
	}

	// Revoke by setting is_active to false
	_, err = s.queries.UpdateAPIKey(ctx, s.db, db.UpdateAPIKeyParams{
		ID:       keyUUID,
		IsActive: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	return nil
}

// GetProfile returns profile information for the authenticated user
func (s *Service) GetProfile(ctx context.Context, userID string) (*models.UserInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.queries.GetUser(ctx, s.db, userUUID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	createdAt := time.Now()
	if user.CreatedAt.Valid {
		createdAt = user.CreatedAt.Time
	}

	return &models.UserInfo{
		ID:        uuidToString(user.ID),
		Email:     user.Email,
		MSPId:     user.MspID,
		Role:      user.Role,
		CreatedAt: createdAt,
	}, nil
}

// Helper functions

func (s *Service) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *Service) generateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}

// uuidToString converts pgtype.UUID to string
func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid.Bytes[0:4],
		uuid.Bytes[4:6],
		uuid.Bytes[6:8],
		uuid.Bytes[8:10],
		uuid.Bytes[10:16],
	)
}
