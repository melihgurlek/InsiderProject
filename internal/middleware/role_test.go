package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRoles(t *testing.T) {
	tests := []struct {
		name       string
		claims     *UserClaims
		allowed    []string
		expectCode int
	}{
		{
			name:       "allowed role",
			claims:     &UserClaims{UserID: "1", Role: "admin"},
			allowed:    []string{"admin"},
			expectCode: http.StatusOK,
		},
		{
			name:       "forbidden role",
			claims:     &UserClaims{UserID: "2", Role: "user"},
			allowed:    []string{"admin"},
			expectCode: http.StatusForbidden,
		},
		{
			name:       "missing claims",
			claims:     nil,
			allowed:    []string{"admin"},
			expectCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Handler that always returns 200 OK if reached
			h := RequireRoles(tc.allowed...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tc.claims != nil {
				req = req.WithContext(WithUserClaims(req.Context(), tc.claims))
			}
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			if rw.Code != tc.expectCode {
				t.Errorf("expected status %d, got %d", tc.expectCode, rw.Code)
			}
		})
	}
}
