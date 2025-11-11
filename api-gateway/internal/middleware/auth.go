package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/utils"
	"go.uber.org/zap"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	jwtSecret []byte
	issuer    string
	logger    *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(cfg *config.JWTConfig, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(cfg.Secret),
		issuer:    cfg.Issuer,
		logger:    logger,
	}
}

// Authenticate validates JWT token
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Missing authorization header",
				nil,
			))
			return
		}

		// Check if it starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid authorization header format",
				nil,
			))
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid or expired token",
				nil,
			))
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid token claims",
				nil,
			))
			return
		}

		// Validate issuer
		if iss, ok := claims["iss"].(string); !ok || iss != m.issuer {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid token issuer",
				nil,
			))
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "userID", claims["userId"])
		ctx = context.WithValue(ctx, "email", claims["email"])
		ctx = context.WithValue(ctx, "mspId", claims["mspId"])
		ctx = context.WithValue(ctx, "role", claims["role"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth validates token if present, but allows request to continue if not
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// If header is present, validate it
		m.Authenticate(next).ServeHTTP(w, r)
	})
}

// GenerateToken generates a new JWT token
func (m *AuthMiddleware) GenerateToken(userID, email, mspID, role string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"mspId":  mspID,
		"role":   role,
		"iss":    m.issuer,
		"iat":    now.Unix(),
		"exp":    now.Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token
func (m *AuthMiddleware) GenerateRefreshToken() (string, error) {
	return utils.GenerateRandomString(32)
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Note: In production, use a proper JSON encoder
}

