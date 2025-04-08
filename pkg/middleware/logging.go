package middleware

import (
	"net/http"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
	"go.uber.org/zap"
)

// Logger is a middleware that logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{w: w, status: http.StatusOK}

		// Process request
		next.ServeHTTP(rw, r)

		// Log request details
		duration := time.Since(start)
		logging.GetLogger().Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Int("status", rw.status),
			zap.Duration("duration", duration),
			zap.String("formatted_duration", logging.FormatDuration(duration)),
		)
	})
}

// responseWriter is a wrapper of http.ResponseWriter to capture status code
type responseWriter struct {
	w      http.ResponseWriter
	status int
}

// Header returns the header map to implement http.ResponseWriter
func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

// Write writes the data to the connection as part of an HTTP reply
func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.w.Write(data)
}

// WriteHeader sends an HTTP response header with the provided status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.w.WriteHeader(statusCode)
}
