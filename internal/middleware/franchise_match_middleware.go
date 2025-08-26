package middleware

import (
	"net/http"
)

func RequireFranchiseAccess() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			franchiseID := GetFranchiseID(r.Context())
			if franchiseID == 0 {
				http.Error(w, "Missing franchise ID in token", http.StatusForbidden)
				return
			}

			// ❗ Aquí podrías aplicar lógica extra (ej: validación temporal, estado activo...)

			next.ServeHTTP(w, r)
		})
	}
}
