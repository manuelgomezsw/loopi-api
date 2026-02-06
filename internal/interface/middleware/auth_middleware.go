package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/auth"
)

// ContextKey is the type for context keys.
type ContextKey string

const (
	// EmployeeIDKey is the context key for the employee ID.
	EmployeeIDKey ContextKey = "employee_id"
	// EmployeeRoleKey is the context key for the employee role.
	EmployeeRoleKey ContextKey = "employee_role"
	// EmployeeUsernameKey is the context key for the employee username.
	EmployeeUsernameKey ContextKey = "employee_username"
)

// AuthMiddleware creates a middleware that validates JWT tokens.
func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), EmployeeIDKey, claims.EmployeeID)
			ctx = context.WithValue(ctx, EmployeeRoleKey, claims.Role)
			ctx = context.WithValue(ctx, EmployeeUsernameKey, claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetEmployeeID extracts the employee ID from the context.
func GetEmployeeID(ctx context.Context) (uint16, bool) {
	id, ok := ctx.Value(EmployeeIDKey).(uint16)
	return id, ok
}

// GetEmployeeRole extracts the employee role from the context.
func GetEmployeeRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(EmployeeRoleKey).(string)
	return role, ok
}

// AdminOnly creates a middleware that restricts access to admin users only.
// Must be used after AuthMiddleware.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetEmployeeRole(r.Context())
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if role != "admin" {
			http.Error(w, `{"error":"forbidden: admin access required"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
