package middleware

import "net/http"

func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleVal := r.Context().Value(ContextRole)
			role, ok := roleVal.(string)
			if !ok || role == "" {
				http.Error(w, "Forbidden: no role", http.StatusForbidden)
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden: insufficient role", http.StatusForbidden)
		})
	}
}
