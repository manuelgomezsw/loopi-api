package middleware

import (
	"net/http"
	"os"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := getCorsOrigin(r)

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getCorsOrigin(r *http.Request) string {
	origin := r.Header.Get("Origin")
	// Permitir Angular local
	if strings.HasPrefix(origin, "http://localhost:4200") {
		return origin
	}

	// O usar una variable de entorno como fallback
	if val := os.Getenv("CORS_ALLOWED_ORIGIN"); val != "" {
		return val
	}

	// Default: permitir todo (no recomendado en producción)
	return "*"
}
