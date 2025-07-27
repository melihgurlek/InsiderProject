package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testPayload struct {
	Name string `json:"name"`
}

func TestValidationMiddleware(t *testing.T) {
	validator := &JSONValidator{}
	factory := func() interface{} { return &testPayload{} }

	tests := []struct {
		name        string
		body        string
		contentType string
		expectCode  int
		expectValid bool
	}{
		{"valid json", `{"name":"test"}`, "application/json", http.StatusOK, true},
		{"invalid json", `{"name":}`, "application/json", http.StatusBadRequest, false},
		{"missing content type", `{"name":"test"}`, "", http.StatusBadRequest, false},
		{"wrong content type", `{"name":"test"}`, "text/plain", http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			var gotPayload *testPayload
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				v, ok := GetValidatedBody[*testPayload](r.Context())
				if ok {
					gotPayload = v
				}
				w.WriteHeader(http.StatusOK)
			})

			mw := ValidationMiddleware(validator, factory)(h)
			mw.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("expected status %d, got %d", tt.expectCode, w.Code)
			}
			if tt.expectValid && gotPayload == nil {
				t.Error("expected validated payload in context, got nil")
			}
			if !tt.expectValid && gotPayload != nil {
				t.Error("did not expect validated payload in context, but got one")
			}
		})
	}
}

// TestGetValidatedBody tests the GetValidatedBody function directly
func TestGetValidatedBody(t *testing.T) {
	ctx := context.Background()

	// Test with nil context
	_, ok := GetValidatedBody[*testPayload](ctx)
	if ok {
		t.Error("expected false when no validated body in context")
	}

	// Test with context containing validated body
	testData := &testPayload{Name: "test"}
	ctxWithData := context.WithValue(ctx, validatedBodyKey{}, testData)

	retrieved, ok := GetValidatedBody[*testPayload](ctxWithData)
	if !ok {
		t.Error("expected true when validated body exists in context")
	}
	if retrieved.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", retrieved.Name)
	}
}
