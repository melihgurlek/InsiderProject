package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// PerformanceMetrics holds performance data for a request.
type PerformanceMetrics struct {
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration_ms"`
	Size       int64         `json:"response_size"`
	RemoteAddr string        `json:"remote_addr"`
	UserAgent  string        `json:"user_agent"`
}

// PerformanceMonitor defines the interface for performance monitoring.
type PerformanceMonitor interface {
	RecordMetrics(metrics PerformanceMetrics)
}

// DefaultPerformanceMonitor is the default implementation using zerolog.
type DefaultPerformanceMonitor struct {
	logger zerolog.Logger
}

// NewDefaultPerformanceMonitor creates a new DefaultPerformanceMonitor.
func NewDefaultPerformanceMonitor(logger zerolog.Logger) *DefaultPerformanceMonitor {
	return &DefaultPerformanceMonitor{logger: logger}
}

// RecordMetrics logs performance metrics using structured logging.
func (m *DefaultPerformanceMonitor) RecordMetrics(metrics PerformanceMetrics) {
	level := zerolog.InfoLevel

	// Use warn level for slow requests (>1s) or server errors
	if metrics.Duration > time.Second || metrics.StatusCode >= 500 {
		level = zerolog.WarnLevel
	}

	// Use error level for very slow requests (>5s)
	if metrics.Duration > 5*time.Second {
		level = zerolog.ErrorLevel
	}

	m.logger.WithLevel(level).
		Str("method", metrics.Method).
		Str("path", metrics.Path).
		Int("status_code", metrics.StatusCode).
		Int64("duration_ms", metrics.Duration.Milliseconds()).
		Int64("response_size", metrics.Size).
		Str("remote_addr", metrics.RemoteAddr).
		Str("user_agent", metrics.UserAgent).
		Msg("request performance")
}

// performanceResponseWriter wraps http.ResponseWriter to capture response data.
type performanceResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

// WriteHeader captures the status code.
func (rw *performanceResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size.
func (rw *performanceResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += int64(n)
	return n, err
}

// NewPerformanceMiddleware returns a middleware that monitors request performance.
func NewPerformanceMiddleware(monitor PerformanceMonitor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timing BEFORE any processing
			start := time.Now()

			// Create a custom response writer to capture metrics
			rw := &performanceResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status code
			}

			// Process the request through the entire middleware chain
			next.ServeHTTP(rw, r)

			// Calculate metrics AFTER response is written
			duration := time.Since(start)
			metrics := PerformanceMetrics{
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rw.statusCode,
				Duration:   duration,
				Size:       rw.size,
				RemoteAddr: r.RemoteAddr,
				UserAgent:  r.UserAgent(),
			}

			// Record the metrics
			monitor.RecordMetrics(metrics)
		})
	}
}

// PerformanceMiddlewareWithOptions returns a middleware with custom options.
func PerformanceMiddlewareWithOptions(monitor PerformanceMonitor, options ...PerformanceOption) func(http.Handler) http.Handler {
	config := &PerformanceConfig{
		Monitor: monitor,
		Logger:  log.Logger,
	}

	for _, option := range options {
		option(config)
	}

	return NewPerformanceMiddleware(config.Monitor)
}

// PerformanceConfig holds configuration for performance monitoring.
type PerformanceConfig struct {
	Monitor PerformanceMonitor
	Logger  zerolog.Logger
}

// PerformanceOption is a function that configures performance monitoring.
type PerformanceOption func(*PerformanceConfig)

// WithCustomLogger sets a custom logger for performance monitoring.
func WithCustomLogger(logger zerolog.Logger) PerformanceOption {
	return func(config *PerformanceConfig) {
		config.Logger = logger
		if config.Monitor == nil {
			config.Monitor = NewDefaultPerformanceMonitor(logger)
		}
	}
}

// WithCustomMonitor sets a custom performance monitor.
func WithCustomMonitor(monitor PerformanceMonitor) PerformanceOption {
	return func(config *PerformanceConfig) {
		config.Monitor = monitor
	}
}

// DefaultPerformanceMiddleware is a convenience function that creates performance monitoring middleware with default settings.
func DefaultPerformanceMiddleware() func(http.Handler) http.Handler {
	monitor := NewDefaultPerformanceMonitor(log.Logger)
	return NewPerformanceMiddleware(monitor)
}
