package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// ErrorHandler defines the interface for custom error handling.
type ErrorHandler interface {
	HandleError(w http.ResponseWriter, r *http.Request, err error, statusCode int)
}

// DefaultErrorHandler is the default implementation of ErrorHandler.
type DefaultErrorHandler struct {
	logger zerolog.Logger
}

// NewDefaultErrorHandler creates a new DefaultErrorHandler with the given logger.
func NewDefaultErrorHandler(logger zerolog.Logger) *DefaultErrorHandler {
	return &DefaultErrorHandler{logger: logger}
}

// HandleError logs the error and sends a JSON error response.
func (h *DefaultErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	// Log the error with request context
	h.logger.Error().
		Err(err).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Str("user_agent", r.UserAgent()).
		Int("status_code", statusCode).
		Msg("request error")

	// Send JSON error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: http.StatusText(statusCode),
		Code:  statusCode,
	}

	// Include error message for client errors (4xx), but not for server errors (5xx)
	if statusCode < 500 && err != nil {
		response.Message = err.Error()
	} else if statusCode >= 500 {
		response.Message = "Internal server error"
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		h.logger.Error().Err(err).Msg("failed to encode error response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ErrorHandlingMiddleware returns a middleware that handles panics and errors.
func ErrorHandlingMiddleware(handler ErrorHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Log the panic with stack trace
					log.Error().
						Interface("panic", rec).
						Str("stack", string(debug.Stack())).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("panic recovered")

					// Create a generic error for the panic
					err := fmt.Errorf("panic: %v", rec)
					handler.HandleError(w, r, err, http.StatusInternalServerError)
				}
			}()

			// Create a custom response writer that can capture errors
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			// Handle any errors that occurred during request processing
			if rw.err != nil {
				handler.HandleError(w, r, rw.err, rw.statusCode)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture errors and status codes.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	err        error
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures any write errors.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	if err != nil {
		rw.err = err
	}
	return n, err
}

// WithError adds an error to the response writer for later handling.
func (rw *responseWriter) WithError(err error) {
	rw.err = err
}

// GetResponseWriter returns the underlying http.ResponseWriter.
func (rw *responseWriter) GetResponseWriter() http.ResponseWriter {
	return rw.ResponseWriter
}

// ErrorMiddleware is a convenience function that creates error handling middleware with default settings.
func ErrorMiddleware() func(http.Handler) http.Handler {
	handler := NewDefaultErrorHandler(log.Logger)
	return ErrorHandlingMiddleware(handler)
}
