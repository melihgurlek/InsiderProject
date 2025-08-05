package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// JWTValidator defines the interface for validating JWT tokens.
type JWTValidator interface {
	ValidateToken(tokenString string) (*UserClaims, error)
}

// UserClaims represents the claims extracted from a valid JWT.
type UserClaims struct {
	UserID string
	Role   string
}

// AuthMiddleware holds dependencies for authentication middleware.
type AuthMiddleware struct {
	validator JWTValidator
}

// NewAuthMiddleware constructs a new AuthMiddleware with the given validator.
func NewAuthMiddleware(validator JWTValidator) *AuthMiddleware {
	return &AuthMiddleware{validator: validator}
}

// Middleware is the HTTP middleware function for authentication.
func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		fmt.Printf("Validating token: %s\n", tokenString[:10]+"...") // First 10 chars

		claims, err := a.validator.ValidateToken(tokenString)
		if err != nil {
			fmt.Printf("Token validation failed: %v\n", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		fmt.Printf("Token validated successfully for user: %s, role: %s\n", claims.UserID, claims.Role)

		_ = claims // will be used in the next step
		ctx := WithUserClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Context helpers for extracting user claims can be added here.

// contextKey is a private type to avoid context key collisions.
type contextKey string

const userClaimsKey contextKey = "userClaims"

// WithUserClaims injects UserClaims into the context.
func WithUserClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

// UserClaimsFromContext retrieves UserClaims from the context, if present.
func UserClaimsFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	return claims, ok
}
