package router

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
)

// New creates and configures the HTTP router with all routes.
func New(db *sql.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			// TODO: Add auth handler
			r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotImplemented)
				w.Write([]byte(`{"message":"login not implemented yet"}`))
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// TODO: Add auth middleware
			// r.Use(authMiddleware)

			// Employees routes
			r.Route("/employees", func(r chi.Router) {
				r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotImplemented)
					w.Write([]byte(`{"message":"get profile not implemented yet"}`))
				})
			})

			// Inventories routes
			r.Route("/inventories", func(r chi.Router) {
				r.Get("/latest", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotImplemented)
					w.Write([]byte(`{"message":"get latest inventory not implemented yet"}`))
				})

				r.Get("/suggested-schedule", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotImplemented)
					w.Write([]byte(`{"message":"get suggested schedule not implemented yet"}`))
				})

				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotImplemented)
					w.Write([]byte(`{"message":"create inventory not implemented yet"}`))
				})

				r.Route("/{inventoryID}", func(r chi.Router) {
					r.Get("/items", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotImplemented)
						w.Write([]byte(`{"message":"get inventory items not implemented yet"}`))
					})

					r.Post("/details", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotImplemented)
						w.Write([]byte(`{"message":"save inventory detail not implemented yet"}`))
					})

					r.Get("/summary", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotImplemented)
						w.Write([]byte(`{"message":"get inventory summary not implemented yet"}`))
					})

					r.Post("/complete", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotImplemented)
						w.Write([]byte(`{"message":"complete inventory not implemented yet"}`))
					})
				})
			})
		})
	})

	return r
}
