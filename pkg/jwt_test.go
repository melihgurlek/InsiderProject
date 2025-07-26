package pkg

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateToken(secret string, claims jwt.MapClaims, method jwt.SigningMethod) string {
	t := jwt.NewWithClaims(method, claims)
	token, _ := t.SignedString([]byte(secret))
	return token
}

func TestJWTValidatorImpl_ValidateToken(t *testing.T) {
	secret := "testsecret"
	validator := NewJWTValidator(secret)

	tests := []struct {
		name       string
		token      string
		expectErr  bool
		expectUID  string
		expectRole string
	}{
		{
			name: "valid token",
			token: generateToken(secret, jwt.MapClaims{
				"user_id": "u1",
				"role":    "admin",
				"exp":     time.Now().Add(time.Hour).Unix(),
			}, jwt.SigningMethodHS256),
			expectErr:  false,
			expectUID:  "u1",
			expectRole: "admin",
		},
		{
			name: "invalid signature",
			token: generateToken("wrongsecret", jwt.MapClaims{
				"user_id": "u2",
				"role":    "user",
				"exp":     time.Now().Add(time.Hour).Unix(),
			}, jwt.SigningMethodHS256),
			expectErr: true,
		},
		{
			name: "missing claims",
			token: generateToken(secret, jwt.MapClaims{
				"exp": time.Now().Add(time.Hour).Unix(),
			}, jwt.SigningMethodHS256),
			expectErr: true,
		},
		{
			name: "wrong signing method",
			token: generateToken(secret, jwt.MapClaims{
				"user_id": "u3",
				"role":    "user",
				"exp":     time.Now().Add(time.Hour).Unix(),
			}, jwt.SigningMethodHS384),
			expectErr: true,
		},
		{
			name: "expired token",
			token: generateToken(secret, jwt.MapClaims{
				"user_id": "u4",
				"role":    "user",
				"exp":     time.Now().Add(-time.Hour).Unix(),
			}, jwt.SigningMethodHS256),
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := validator.ValidateToken(tc.token)
			if tc.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.expectErr {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if claims.UserID != tc.expectUID {
					t.Errorf("expected userID %q, got %q", tc.expectUID, claims.UserID)
				}
				if claims.Role != tc.expectRole {
					t.Errorf("expected role %q, got %q", tc.expectRole, claims.Role)
				}
			}
		})
	}
}
