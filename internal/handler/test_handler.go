package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// TestHandler provides test endpoints for middleware validation.
type TestHandler struct{}

// NewTestHandler creates a new TestHandler.
func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

// EchoRequest represents the request body for the echo endpoint.
type EchoRequest struct {
	Message string `json:"message"`
	Number  int    `json:"number"`
}

// EchoResponse represents the response from the echo endpoint.
type EchoResponse struct {
	Message string `json:"message"`
	Number  int    `json:"number"`
	Echoed  bool   `json:"echoed"`
}

// Echo handles POST /api/v1/test/echo - validates JSON and echoes back.
func (h *TestHandler) Echo(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request body
	var req EchoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := EchoResponse{
		Message: req.Message,
		Number:  req.Number,
		Echoed:  true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Panic handles GET /api/v1/test/panic - triggers a panic to test error handling.
func (h *TestHandler) Panic(w http.ResponseWriter, r *http.Request) {
	panic("test panic from handler")
}

// Error handles GET /api/v1/test/error?code=500 - triggers an error with specified status code.
func (h *TestHandler) Error(w http.ResponseWriter, r *http.Request) {
	codeStr := r.URL.Query().Get("code")
	if codeStr == "" {
		codeStr = "500"
	}

	code, err := strconv.Atoi(codeStr)
	if err != nil {
		code = 500
	}

	http.Error(w, "test error", code)
}

// Slow handles GET /api/v1/test/slow - intentionally slow to test performance monitoring.
func (h *TestHandler) Slow(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	response := map[string]interface{}{
		"message": "slow response",
		"delay":   "100ms",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Health handles GET /api/v1/test/health - health check endpoint for Docker and load balancers.
func (h *TestHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "backend-path-api",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CacheTest handles GET /api/v1/test/cache - demonstrates caching with timestamp
func (h *TestHandler) CacheTest(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "This response should be cached for 5 minutes",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"cache_key": "cache_test",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers test routes to the router.
func (h *TestHandler) RegisterRoutes(r chi.Router) {
	r.Post("/echo", h.Echo)
	r.Get("/panic", h.Panic)
	r.Get("/error", h.Error)
	r.Get("/slow", h.Slow)
	r.Get("/health", h.Health)
	r.Get("/cache", h.CacheTest)
}
