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

package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ibn-network/admin-service/internal/handlers"
	adminMiddleware "github.com/ibn-network/admin-service/internal/middleware"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes for Admin Service
func SetupRoutes(
	chaincodeHandler *handlers.ChaincodeHandler,
	apiKey string,
	logger *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	// Global timeout: 5 minutes (chaincode install needs extended timeout)
	// Route-specific timeouts can override this
	r.Use(middleware.Timeout(5 * time.Minute))

	// CORS - Only allow internal services
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, restrict to specific origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check (no auth required) - Support both GET and HEAD for Docker healthcheck
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" {
			w.Write([]byte(`{"status":"healthy","service":"admin-service"}`))
		}
	})
	r.Head("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == "GET" {
			w.Write([]byte(`{"status":"ready","service":"admin-service"}`))
		}
	})
	r.Head("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Protected routes - require API key
	r.Group(func(r chi.Router) {
		r.Use(adminMiddleware.APIKeyAuth(apiKey, logger))

		// Chaincode lifecycle routes
		r.Route("/api/v1/chaincode", func(r chi.Router) {
			// File upload endpoint (multipart/form-data)
			r.Post("/upload", chaincodeHandler.UploadPackage)

			// List operations
			r.Get("/installed", chaincodeHandler.ListInstalled)
			r.Get("/committed", chaincodeHandler.ListCommitted)
			r.Get("/committed/info", chaincodeHandler.GetCommittedInfo)

			// Lifecycle operations
			// Install: extended timeout (5 minutes) for peer CLI execution
			r.With(middleware.Timeout(5 * time.Minute)).Post("/install", chaincodeHandler.Install)
			// Approve: extended timeout (2 minutes)
			r.With(middleware.Timeout(2 * time.Minute)).Post("/approve", chaincodeHandler.Approve)
			// Commit: extended timeout (2 minutes)
			r.With(middleware.Timeout(2 * time.Minute)).Post("/commit", chaincodeHandler.Commit)
		})
	})

	return r
}
