package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		// Validar presencia de datos
		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Missing user ID", http.StatusUnauthorized)
			return
		}

		// Contexto extendido
		ctx := context.WithValue(r.Context(), ContextUserID, int(userID))

		if email, ok := claims["email"].(string); ok {
			ctx = context.WithValue(ctx, ContextEmail, email)
		}
		if role, ok := claims["role"].(string); ok {
			ctx = context.WithValue(ctx, ContextRole, role)
		}
		if fid, ok := claims["franchise_id"].(float64); ok {
			ctx = context.WithValue(ctx, ContextFranchiseID, int(fid))
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
