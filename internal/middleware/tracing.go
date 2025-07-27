package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/melihgurlek/backend-path/pkg/tracing"
)

// TracingMiddleware provides OpenTelemetry tracing for HTTP requests
type TracingMiddleware struct{}

// NewTracingMiddleware creates a new TracingMiddleware
func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{}
}

// Middleware traces HTTP requests
func (m *TracingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Start span
		spanName := r.Method + " " + r.URL.Path
		ctx, span := tracing.StartSpan(ctx, spanName)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.user_agent", r.UserAgent()),
			attribute.String("http.remote_addr", r.RemoteAddr),
		)

		// Create response writer wrapper to capture status code
		wrapped := &tracingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Inject trace context into response headers
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(wrapped.Header()))

		// Process request
		next.ServeHTTP(wrapped, r)

		// Set final span attributes
		span.SetAttributes(
			attribute.Int("http.status_code", wrapped.statusCode),
		)

		// Record span events for different status codes
		if wrapped.statusCode >= 400 {
			span.AddEvent("http.error", trace.WithAttributes(
				attribute.Int("http.status_code", wrapped.statusCode),
			))
		}
	})
}

// tracingResponseWriter wraps http.ResponseWriter to capture status code for tracing
type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *tracingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *tracingResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
