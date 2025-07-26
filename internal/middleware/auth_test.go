package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockValidator struct {
	validateFunc func(token string) (*UserClaims, error)
}

func (m *mockValidator) ValidateToken(token string) (*UserClaims, error) {
	return m.validateFunc(token)
}

func TestAuthMiddleware_Middleware(t *testing.T) {
	tests := []struct {
		name           string
		header         string
		validateFunc   func(token string) (*UserClaims, error)
		expectStatus   int
		expectNextCall bool
	}{
		{
			name:           "missing header",
			header:         "",
			validateFunc:   nil,
			expectStatus:   http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:           "malformed header",
			header:         "Token abcdefg",
			validateFunc:   nil,
			expectStatus:   http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:   "invalid token",
			header: "Bearer invalidtoken",
			validateFunc: func(token string) (*UserClaims, error) {
				return nil, http.ErrNoCookie
			},
			expectStatus:   http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:   "valid token",
			header: "Bearer validtoken",
			validateFunc: func(token string) (*UserClaims, error) {
				return &UserClaims{UserID: "123", Role: "user"}, nil
			},
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validator := &mockValidator{validateFunc: tc.validateFunc}
			mw := NewAuthMiddleware(validator)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				claims, ok := UserClaimsFromContext(r.Context())
				if tc.expectNextCall && (!ok || claims.UserID != "123") {
					t.Errorf("expected claims in context, got: %v", claims)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rw := httptest.NewRecorder()

			mw.Middleware(next).ServeHTTP(rw, req)

			if rw.Code != tc.expectStatus {
				t.Errorf("expected status %d, got %d", tc.expectStatus, rw.Code)
			}
			if nextCalled != tc.expectNextCall {
				t.Errorf("expected next handler called: %v, got %v", tc.expectNextCall, nextCalled)
			}
		})
	}
}
