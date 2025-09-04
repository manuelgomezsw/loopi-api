package middleware

import "net/http"

func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleVal := r.Context().Value(ContextRole)
			if roleVal == nil {
				http.Error(w, "missing roles", http.StatusForbidden)
				return
			}

			roles, ok := roleVal.([]string)
			if !ok {
				http.Error(w, "invalid role format", http.StatusForbidden)
				return
			}

			for _, userRole := range roles {
				if allowed[userRole] {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
		})
	}
}
