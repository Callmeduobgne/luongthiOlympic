package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/utils"
)

// ValidatorMiddleware provides request validation middleware
type ValidatorMiddleware struct {
	validator *utils.Validator
}

// NewValidatorMiddleware creates a new validator middleware
func NewValidatorMiddleware() *ValidatorMiddleware {
	return &ValidatorMiddleware{
		validator: utils.NewValidator(),
	}
}

// ValidateBody validates request body against a struct
func (m *ValidatorMiddleware) ValidateBody(v interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Decode request body
			if err := json.NewDecoder(r.Body).Decode(v); err != nil {
				respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
					models.ErrCodeBadRequest,
					"Invalid request body",
					err.Error(),
				))
				return
			}

			// Validate struct
			if err := m.validator.Validate(v); err != nil {
				respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
					models.ErrCodeBadRequest,
					"Validation failed",
					err.Error(),
				))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

