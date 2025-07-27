package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestPerformanceMiddleware(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)
	monitor := NewDefaultPerformanceMonitor(testLogger)

	tests := []struct {
		name           string
		handlerFunc    http.HandlerFunc
		expectDuration bool
		expectLog      bool
		expectStatus   int
	}{
		{
			name: "normal request",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			expectDuration: true,
			expectLog:      true,
			expectStatus:   http.StatusOK,
		},
		{
			name: "slow request",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(10 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("slow response"))
			},
			expectDuration: true,
			expectLog:      true,
			expectStatus:   http.StatusOK,
		},
		{
			name: "error request",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error"))
			},
			expectDuration: true,
			expectLog:      true,
			expectStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			mw := NewPerformanceMiddleware(monitor)(tt.handlerFunc)
			mw.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectStatus, w.Code)

			// Check logging
			if tt.expectLog {
				logOutput := logBuffer.String()
				assert.Contains(t, logOutput, "request performance")
				assert.Contains(t, logOutput, `"method":"GET"`)
				assert.Contains(t, logOutput, `"path":"/test"`)
				assert.Contains(t, logOutput, `"status_code":`+strconv.Itoa(tt.expectStatus))
				assert.Contains(t, logOutput, `"duration_ms":`)
				assert.Contains(t, logOutput, `"response_size":`)
			}
		})
	}
}

func TestPerformanceResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &performanceResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, rw.statusCode)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test Write
	data := []byte("test data")
	n, err := rw.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, int64(len(data)), rw.size)

	// Test multiple writes
	moreData := []byte("more data")
	n, err = rw.Write(moreData)
	assert.NoError(t, err)
	assert.Equal(t, len(moreData), n)
	assert.Equal(t, int64(len(data)+len(moreData)), rw.size)
}

func TestDefaultPerformanceMonitor(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)
	monitor := NewDefaultPerformanceMonitor(testLogger)

	tests := []struct {
		name           string
		metrics        PerformanceMetrics
		expectLogLevel string
	}{
		{
			name: "normal request",
			metrics: PerformanceMetrics{
				Method:     "GET",
				Path:       "/test",
				StatusCode: 200,
				Duration:   50 * time.Millisecond,
				Size:       100,
				RemoteAddr: "127.0.0.1",
				UserAgent:  "test-agent",
			},
			expectLogLevel: "info",
		},
		{
			name: "slow request",
			metrics: PerformanceMetrics{
				Method:     "POST",
				Path:       "/slow",
				StatusCode: 200,
				Duration:   2 * time.Second,
				Size:       200,
				RemoteAddr: "127.0.0.1",
				UserAgent:  "test-agent",
			},
			expectLogLevel: "warn",
		},
		{
			name: "server error",
			metrics: PerformanceMetrics{
				Method:     "GET",
				Path:       "/error",
				StatusCode: 500,
				Duration:   100 * time.Millisecond,
				Size:       50,
				RemoteAddr: "127.0.0.1",
				UserAgent:  "test-agent",
			},
			expectLogLevel: "warn",
		},
		{
			name: "very slow request",
			metrics: PerformanceMetrics{
				Method:     "POST",
				Path:       "/very-slow",
				StatusCode: 200,
				Duration:   6 * time.Second,
				Size:       300,
				RemoteAddr: "127.0.0.1",
				UserAgent:  "test-agent",
			},
			expectLogLevel: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()

			monitor.RecordMetrics(tt.metrics)

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, "request performance")
			assert.Contains(t, logOutput, `"method":"`+tt.metrics.Method+`"`)
			assert.Contains(t, logOutput, `"path":"`+tt.metrics.Path+`"`)
			assert.Contains(t, logOutput, `"status_code":`+strconv.Itoa(tt.metrics.StatusCode))
			assert.Contains(t, logOutput, `"duration_ms":`)
			assert.Contains(t, logOutput, `"response_size":`+strconv.FormatInt(tt.metrics.Size, 10))
			assert.Contains(t, logOutput, `"remote_addr":"`+tt.metrics.RemoteAddr+`"`)
			assert.Contains(t, logOutput, `"user_agent":"`+tt.metrics.UserAgent+`"`)

			// Check log level
			if tt.expectLogLevel == "warn" {
				assert.Contains(t, logOutput, `"level":"warn"`)
			} else if tt.expectLogLevel == "error" {
				assert.Contains(t, logOutput, `"level":"error"`)
			} else {
				assert.Contains(t, logOutput, `"level":"info"`)
			}
		})
	}
}

func TestDefaultPerformanceMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	mw := DefaultPerformanceMiddleware()(handler)
	mw.ServeHTTP(w, req)

	// Should not panic and should return 200
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Body.String())
}
