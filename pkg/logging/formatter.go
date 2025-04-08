package logging

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// HTTPRequest formats an HTTP request for logging
func HTTPRequest(r *http.Request) []zapcore.Field {
	return []zapcore.Field{
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("proto", r.Proto),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("referer", r.Referer()),
	}
}

// HTTPResponse formats an HTTP response for logging
func HTTPResponse(status int, latency time.Duration) []zapcore.Field {
	return []zapcore.Field{
		zap.Int("status", status),
		zap.Duration("latency", latency),
	}
}

// Error formats an error for logging
func LogError(err error) zapcore.Field {
	return zap.Error(err)
}

// RequestResponseLogger is a middleware for logging HTTP requests and responses
func RequestResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{w: w, status: http.StatusOK}

		// Log request
		Info("Request received", HTTPRequest(r)...)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log response
		latency := time.Since(start)
		Info("Response sent", append(
			HTTPResponse(rw.status, latency),
			zap.String("path", r.URL.Path),
		)...)
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
	return rw.w.Write(data)
}

// WriteHeader sends an HTTP response header with the provided status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.w.WriteHeader(statusCode)
}

// FormatDuration formats a duration as a readable string
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d μs", d.Microseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Milliseconds()))
	} else if d < time.Minute {
		return fmt.Sprintf("%.2f s", d.Seconds())
	} else {
		return d.String()
	}
}
