package middleware

import (
	"net/http"
	"strconv"
)

// RequireRoles returns a middleware that authorizes requests based on user roles.
// Usage: r.With(RequireRoles("admin")).Get("/admin", handler)
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := UserClaimsFromContext(r.Context())
			if !ok || claims == nil {
				http.Error(w, "Unauthorized: missing user claims", http.StatusUnauthorized)
				return
			}
			if _, ok := roleSet[claims.Role]; !ok {
				http.Error(w, "Forbidden: insufficient role", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IsAdminOrSelf returns true if the claims belong to an admin, or if the claims' user ID matches the target user ID.
// Use this for endpoints where a user can act on their own resource, or an admin can act on any resource.
func IsAdminOrSelf(claims *UserClaims, targetUserID int) bool {
	if claims == nil {
		return false
	}
	if claims.Role == "admin" {
		return true
	}
	// Convert claims.UserID to int for comparison
	if uid, err := strconv.Atoi(claims.UserID); err == nil && uid == targetUserID {
		return true
	}
	return false
}
