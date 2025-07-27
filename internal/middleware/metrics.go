package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/melihgurlek/backend-path/pkg/metrics"
)

// MetricsMiddleware provides Prometheus metrics collection for HTTP requests
type MetricsMiddleware struct{}

// NewMetricsMiddleware creates a new MetricsMiddleware
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Middleware collects metrics for HTTP requests
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get route pattern
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}

		// Track request in flight
		metrics.HTTPRequestsInFlight.WithLabelValues(r.Method, route).Inc()
		defer metrics.HTTPRequestsInFlight.WithLabelValues(r.Method, route).Dec()

		// Create a response writer wrapper to capture status code
		wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(wrapped.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, route, statusCode).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, route).Observe(duration)
	})
}

// metricsResponseWriter wraps http.ResponseWriter to capture status code for metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
