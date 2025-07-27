package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandlingMiddleware(t *testing.T) {
	// Create a test logger that writes to a buffer
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)
	handler := NewDefaultErrorHandler(testLogger)

	tests := []struct {
		name           string
		handlerFunc    http.HandlerFunc
		expectStatus   int
		expectError    bool
		expectPanic    bool
		expectLogEntry bool
	}{
		{
			name: "normal request",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			expectStatus:   http.StatusOK,
			expectError:    false,
			expectPanic:    false,
			expectLogEntry: false,
		},
		{
			name: "handler returns error",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				// To properly test the middleware, the handler should not write its own
				// error body with http.Error. Instead, it should signal an error to the
				// middleware and let it handle the JSON response formatting.
				w.WriteHeader(http.StatusNotFound)

				// We type-assert the ResponseWriter to our custom type to access WithError.
				// This simulates how a handler compatible with this middleware should behave.
				if rw, ok := w.(*responseWriter); ok {
					rw.WithError(assert.AnError) // assert.AnError is a predefined test error
				}
			},
			expectStatus:   http.StatusNotFound,
			expectError:    true,
			expectPanic:    false,
			expectLogEntry: true,
		},
		{
			name: "handler panics",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectStatus:   http.StatusInternalServerError,
			expectError:    true,
			expectPanic:    true,
			expectLogEntry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log buffer
			logBuffer.Reset()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			mw := ErrorHandlingMiddleware(handler)(tt.handlerFunc)
			mw.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectStatus, w.Code)

			// Check if error response was sent
			if tt.expectError {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectStatus, response.Code)
				assert.NotEmpty(t, response.Error)

				if tt.expectPanic {
					assert.Equal(t, "Internal server error", response.Message)
				}
			}

			// Check if error was logged
			if tt.expectLogEntry {
				assert.Contains(t, logBuffer.String(), "request error")
			} else {
				assert.Empty(t, logBuffer.String())
			}
		})
	}
}

func TestDefaultErrorHandler_HandleError(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)
	handler := NewDefaultErrorHandler(testLogger)

	tests := []struct {
		name         string
		err          error
		statusCode   int
		expectMsg    string
		expectLogMsg string
	}{
		{
			name:         "client error",
			err:          assert.AnError,
			statusCode:   http.StatusBadRequest,
			expectMsg:    assert.AnError.Error(),
			expectLogMsg: "request error",
		},
		{
			name:         "server error",
			err:          assert.AnError,
			statusCode:   http.StatusInternalServerError,
			expectMsg:    "Internal server error",
			expectLogMsg: "request error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			handler.HandleError(w, req, tt.err, tt.statusCode)

			// Check response
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.statusCode, response.Code)
			assert.Equal(t, tt.expectMsg, response.Message)

			// Check logging
			assert.Contains(t, logBuffer.String(), tt.expectLogMsg)
		})
	}
}

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, rw.statusCode)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test Write
	data := []byte("test data")
	n, err := rw.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Nil(t, rw.err)

	// Test WithError
	testErr := assert.AnError
	rw.WithError(testErr)
	assert.Equal(t, testErr, rw.err)

	// Test GetResponseWriter
	assert.Equal(t, w, rw.GetResponseWriter())
}

func TestErrorMiddleware_ConvenienceFunction(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	mw := ErrorMiddleware()(handler)
	mw.ServeHTTP(w, req)

	// Should handle panic and return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "Internal server error", response.Message)
}
