package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
	"go.uber.org/zap"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Recovery is a middleware that recovers from panics and logs them
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				logging.GetLogger().Error("panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)

				// Create error response
				errorMsg := "Internal server error"
				if logging.IsDevelopment() {
					errorMsg = fmt.Sprintf("panic: %v", err)
				}

				// Send error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusInternalServerError,
					Message: "An unexpected error occurred",
					Error:   errorMsg,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// NotFound is a handler for 404 errors
func NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Resource not found",
		})
	})
}

// MethodNotAllowed is a handler for 405 errors
func MethodNotAllowed() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusMethodNotAllowed,
			Message: "Method not allowed",
		})
	})
}

// HandleError sends a standardized error response
func HandleError(w http.ResponseWriter, err error, status int, message string) {
	// Log the error
	logging.GetLogger().Error(message,
		zap.Error(err),
	)

	// Create error message
	errorMsg := err.Error()
	if !logging.IsDevelopment() {
		errorMsg = ""
	}

	// Send error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Status:  status,
		Message: message,
		Error:   errorMsg,
	})
}
