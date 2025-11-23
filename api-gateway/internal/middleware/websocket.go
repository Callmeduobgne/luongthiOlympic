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
	"net/http"
	"strings"
)

// WebSocketMiddleware is a middleware that skips unnecessary processing for WebSocket connections
// Production best practice: WebSocket connections should bypass compression, audit logging, etc.
func WebSocketMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a WebSocket upgrade request
		if isWebSocketRequest(r) {
			// For WebSocket, we need to preserve the original response writer
			// to maintain http.Hijacker capability
			next.ServeHTTP(w, r)
			return
		}
		// For non-WebSocket requests, continue with normal processing
		next.ServeHTTP(w, r)
	})
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func isWebSocketRequest(r *http.Request) bool {
	upgrade := r.Header.Get("Upgrade")
	connection := r.Header.Get("Connection")
	
	// Check for WebSocket upgrade header
	if strings.ToLower(upgrade) == "websocket" {
		return true
	}
	
	// Check for upgrade in Connection header
	if connection != "" {
		connectionLower := strings.ToLower(connection)
		if strings.Contains(connectionLower, "upgrade") {
			return true
		}
	}
	
	return false
}

// SkipAuditForWebSocket is a helper to skip audit logging for WebSocket connections
// This should be used in AuditMiddleware to skip WebSocket connections
func SkipAuditForWebSocket(r *http.Request) bool {
	return isWebSocketRequest(r)
}

