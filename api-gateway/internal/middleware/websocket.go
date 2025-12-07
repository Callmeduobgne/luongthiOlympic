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

