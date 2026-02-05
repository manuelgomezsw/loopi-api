package router

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/auth"
	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/repository"
	"github.com/manuelgomezsw/loopi-api/internal/interface/handler"
	"github.com/manuelgomezsw/loopi-api/internal/interface/middleware"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
)

// New creates and configures the HTTP router with all routes.
func New(db *sql.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://127.0.0.1:4200"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize infrastructure
	jwtManager := auth.NewJWTManager(cfg.JWT)

	// Initialize repositories
	employeeRepo := repository.NewMySQLEmployeeRepository(db)
	itemRepo := repository.NewMySQLItemRepository(db)
	inventoryRepo := repository.NewMySQLInventoryRepository(db)
	inventoryDetailRepo := repository.NewMySQLInventoryDetailRepository(db)
	inventoryIssueRepo := repository.NewMySQLInventoryIssueRepository(db)

	// Initialize services
	authService := service.NewAuthService(employeeRepo, jwtManager)
	inventoryService := service.NewInventoryService(
		inventoryRepo,
		inventoryDetailRepo,
		inventoryIssueRepo,
		itemRepo,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

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
			r.Post("/login", authHandler.Login)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(jwtManager))

			// Employees routes
			r.Route("/employees", func(r chi.Router) {
				r.Get("/me", authHandler.GetMe)
			})

			// Inventories routes
			r.Route("/inventories", func(r chi.Router) {
				r.Get("/latest", inventoryHandler.GetLatest)
				r.Get("/suggested-schedule", inventoryHandler.GetSuggestedSchedule)
				r.Post("/", inventoryHandler.Create)

				r.Route("/{inventoryID}", func(r chi.Router) {
					r.Get("/items", inventoryHandler.GetItems)
					r.Post("/details", inventoryHandler.SaveDetail)
					r.Get("/summary", inventoryHandler.GetSummary)
					r.Post("/complete", inventoryHandler.Complete)
				})
			})
		})
	})

	return r
}
