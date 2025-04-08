package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
	"go.uber.org/zap"
)

// For a real application, you would implement proper authentication
// This is a simplified version for demonstration purposes

// contextKey is a type for context keys
type contextKey string

// UserIDKey is the context key for storing the user ID
const UserIDKey contextKey = "userID"

// Role represents a user role
type Role string

// Roles
const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

// contextRoleKey is the context key for storing the user role
const contextRoleKey contextKey = "role"

// APIKey is a simple authentication middleware that checks for an API key
func APIKey(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header
			key := r.Header.Get("X-API-Key")
			if key == "" {
				// Try from query parameter
				key = r.URL.Query().Get("api_key")
			}

			// Check API key
			if key != apiKey {
				logging.GetLogger().Warn("Invalid API key",
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("path", r.URL.Path),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusUnauthorized,
					Message: "Invalid API key",
				})
				return
			}

			// API key is valid, continue
			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuth is a simple HTTP basic authentication middleware
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No Authorization header
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusUnauthorized,
					Message: "Authorization required",
				})
				return
			}

			// Parse Authorization header
			authParts := strings.SplitN(authHeader, " ", 2)
			if len(authParts) != 2 || authParts[0] != "Basic" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusUnauthorized,
					Message: "Invalid Authorization header format",
				})
				return
			}

			// Decode credentials
			creds, err := decodeBasicAuth(authParts[1])
			if err != nil || creds[0] != username || creds[1] != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusUnauthorized,
					Message: "Invalid credentials",
				})
				return
			}

			// Authentication successful, continue
			// Add the user ID to the context
			ctx := context.WithValue(r.Context(), UserIDKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// decodeBasicAuth decodes a base64 encoded basic auth string
func decodeBasicAuth(auth string) ([]string, error) {
	// In a real application, implement base64 decoding and parsing
	// For simplicity, we'll just return a simple array
	return []string{"admin", "password"}, nil
}

// RequireRole checks if the user has the required role
func RequireRole(role Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user role from context
			userRole, ok := r.Context().Value(contextRoleKey).(Role)
			if !ok || userRole != role {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusForbidden,
					Message: "Insufficient permissions",
				})
				return
			}

			// User has the required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout middleware aborts requests that take too long
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal when the request is done
			done := make(chan struct{})
			defer close(done)

			// Create a response writer wrapper
			rw := &responseWriter{w: w, status: http.StatusOK}

			// Process request in a goroutine
			go func() {
				next.ServeHTTP(rw, r.WithContext(ctx))
				done <- struct{}{}
			}()

			// Wait for the request to complete or timeout
			select {
			case <-done:
				// Request completed normally
				return
			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded {
					logging.GetLogger().Warn("Request timed out",
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.Duration("timeout", timeout),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusRequestTimeout)
					json.NewEncoder(w).Encode(ErrorResponse{
						Status:  http.StatusRequestTimeout,
						Message: "Request timed out",
					})
				}
				return
			}
		})
	}
}
