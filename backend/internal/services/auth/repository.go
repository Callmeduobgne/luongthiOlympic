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
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles auth data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new auth repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO public.users (id, email, username, password_hash, role, msp_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash, user.Role,
		user.MSPID, user.IsActive,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// IsUsernameTaken checks if a username already exists
func (r *Repository) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	query := `
		SELECT 1
		FROM public.users
		WHERE username = $1
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check username: %w", err)
	}

	return true, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, role, msp_id, is_active, 
		       NULL::TIMESTAMPTZ AS last_login_at, created_at, updated_at
		FROM public.users
		WHERE email = $1
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Role,
		&user.MSPID, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, role, msp_id, is_active,
		       NULL::TIMESTAMPTZ AS last_login_at, created_at, updated_at
		FROM public.users
		WHERE id = $1
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Role,
		&user.MSPID, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateLastLogin updates user's last login time
func (r *Repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE public.users
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// CreateAPIKey creates a new API key
func (r *Repository) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	query := `
		INSERT INTO public.api_keys (id, user_id, key_hash, name, description, permissions, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		apiKey.ID, apiKey.UserID, apiKey.KeyHash, apiKey.Name,
		apiKey.Description, apiKey.Permissions, apiKey.IsActive, apiKey.ExpiresAt,
	).Scan(&apiKey.CreatedAt, &apiKey.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetAPIKeyByHash retrieves an API key by hash
func (r *Repository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, description, permissions, is_active,
		       last_used_at, expires_at, created_at, updated_at
		FROM public.api_keys
		WHERE key_hash = $1
	`

	apiKey := &APIKey{}
	var deletedAt *time.Time
	err := r.db.QueryRow(ctx, query, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.Name,
		&apiKey.Description, &apiKey.Permissions, &apiKey.IsActive,
		&apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	apiKey.DeletedAt = deletedAt

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// UpdateAPIKeyLastUsed updates API key's last used time
func (r *Repository) UpdateAPIKeyLastUsed(ctx context.Context, keyID uuid.UUID) error {
	query := `
		UPDATE public.api_keys
		SET last_used_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), keyID)
	if err != nil {
		return fmt.Errorf("failed to update API key last used: %w", err)
	}

	return nil
}

// ListAPIKeysByUser lists all API keys for a user
func (r *Repository) ListAPIKeysByUser(ctx context.Context, userID uuid.UUID) ([]*APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, description, permissions, is_active,
		       last_used_at, expires_at, created_at, updated_at
		FROM public.api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		apiKey := &APIKey{}
		var deletedAt *time.Time
		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.Name,
			&apiKey.Description, &apiKey.Permissions, &apiKey.IsActive,
			&apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		apiKey.DeletedAt = deletedAt
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// CreateRefreshToken creates a new refresh token
func (r *Repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO public.refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetRefreshTokenByHash retrieves a refresh token by hash
func (r *Repository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, is_revoked,
		       expires_at, created_at, revoked_at
		FROM public.refresh_tokens
		WHERE token_hash = $1
	`

	token := &RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.IsRevoked, &token.ExpiresAt,
		&token.CreatedAt, &token.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE public.refresh_tokens
		SET is_revoked = true, revoked_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// DeleteExpiredRefreshTokens deletes expired refresh tokens
func (r *Repository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	query := `
		DELETE FROM public.refresh_tokens
		WHERE expires_at < $1
	`

	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}

// UpdateUserAvatar updates user's avatar URL
func (r *Repository) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	query := `
		UPDATE public.users
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
	}

	return nil
}
