package middleware

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/melihgurlek/backend-path/pkg/cache"
)

// CacheMiddleware provides HTTP response caching
type CacheMiddleware struct {
	cache *cache.RedisCache
	ttl   time.Duration
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(cache *cache.RedisCache, ttl time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		cache: cache,
		ttl:   ttl,
	}
}

// Middleware caches HTTP responses
func (m *CacheMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Skip caching for certain paths
		if shouldSkipCache(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Generate cache key
		cacheKey := m.generateCacheKey(r)

		// Try to get from cache
		var cachedResponse CachedResponse
		if found, err := m.cache.Get(r.Context(), cacheKey, &cachedResponse); err == nil && found {
			// Return cached response
			w.Header().Set("Content-Type", cachedResponse.ContentType)
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(cachedResponse.StatusCode)
			w.Write(cachedResponse.Body)
			return
		}

		// Cache miss, capture response
		responseWriter := &cacheResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           []byte{},
		}

		next.ServeHTTP(responseWriter, r)

		// Cache successful responses
		if responseWriter.statusCode >= 200 && responseWriter.statusCode < 300 {
			cachedResponse := CachedResponse{
				StatusCode:  responseWriter.statusCode,
				ContentType: responseWriter.Header().Get("Content-Type"),
				Body:        responseWriter.body,
				Timestamp:   time.Now(),
			}

			if err := m.cache.Set(r.Context(), cacheKey, cachedResponse, m.ttl); err != nil {
				// Log cache set error but don't fail the request
				fmt.Printf("Failed to cache response: %v\n", err)
			}
		}

		// Add cache miss header
		w.Header().Set("X-Cache", "MISS")
	})
}

// generateCacheKey creates a unique cache key for the request
func (m *CacheMiddleware) generateCacheKey(r *http.Request) string {
	// Include method, path, and query parameters
	key := fmt.Sprintf("%s:%s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

	// Create MD5 hash for consistent key length
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("http_cache:%x", hash)
}

// shouldSkipCache determines if a request should skip caching
func shouldSkipCache(path string) bool {
	skipPaths := []string{
		"/metrics",
		"/api/v1/test/health",
		"/api/v1/test/panic",
		"/api/v1/test/error",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int       `json:"status_code"`
	ContentType string    `json:"content_type"`
	Body        []byte    `json:"body"`
	Timestamp   time.Time `json:"timestamp"`
}

// cacheResponseWriter captures the response for caching
type cacheResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *cacheResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *cacheResponseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}
