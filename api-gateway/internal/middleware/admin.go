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

package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// AdminOnly middleware ensures only admin users can access the route
func AdminOnly(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (set by auth middleware)
			role, ok := r.Context().Value("role").(string)
			if !ok || role != "admin" {
				logger.Warn("Admin access denied",
					zap.String("path", r.URL.Path),
					zap.String("role", role),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(models.NewErrorResponse(
					models.ErrCodeForbidden,
					"Admin privileges required",
					nil,
				))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

