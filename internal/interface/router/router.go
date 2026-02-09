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
		AllowedOrigins: []string{
			"http://localhost:4200",
			"http://127.0.0.1:4200",
			"https://loopi-c048d.web.app",
			"https://loopi-c048d.firebaseapp.com",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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
	categoryRepo := repository.NewMySQLCategoryRepository(db)
	supplierRepo := repository.NewMySQLSupplierRepository(db)

	// Initialize services
	authService := service.NewAuthService(employeeRepo, jwtManager)
	inventoryService := service.NewInventoryService(
		inventoryRepo,
		inventoryDetailRepo,
		inventoryIssueRepo,
		itemRepo,
	)
	adminService := service.NewAdminService(
		inventoryRepo,
		inventoryDetailRepo,
		employeeRepo,
		itemRepo,
		categoryRepo,
		supplierRepo,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	adminHandler := handler.NewAdminHandler(adminService)

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
				r.Get("/in-progress", inventoryHandler.GetInProgress)
				r.Get("/suggested-schedule", inventoryHandler.GetSuggestedSchedule)
				r.Post("/", inventoryHandler.Create)

				r.Route("/{inventoryID}", func(r chi.Router) {
					r.Get("/items", inventoryHandler.GetItems)
					r.Post("/details", inventoryHandler.SaveDetail)
					r.Get("/discrepancies", inventoryHandler.GetDiscrepancies)
					r.Post("/sales", inventoryHandler.SaveSales)
					r.Get("/summary", inventoryHandler.GetSummary)
					r.Post("/complete", inventoryHandler.Complete)
				})
			})

			// Admin routes (requires admin role)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminOnly)

				// Dashboard
				r.Get("/dashboard", adminHandler.GetDashboard)

			// Admin inventories management
			r.Route("/inventories", func(r chi.Router) {
				r.Get("/", adminHandler.ListInventories)
				r.Get("/active-count", adminHandler.GetActiveInventoriesCount)
				r.Get("/{inventoryID}", adminHandler.GetInventoryDetail)
				r.Put("/{inventoryID}/details/{detailID}", adminHandler.UpdateInventoryDetail)
			})

				// Admin items management
				r.Route("/items", func(r chi.Router) {
					r.Get("/", adminHandler.ListItems)
					r.Post("/", adminHandler.CreateItem)
					r.Get("/{itemID}", adminHandler.GetItem)
					r.Put("/{itemID}", adminHandler.UpdateItem)
					r.Patch("/{itemID}/status", adminHandler.UpdateItemStatus)
				})

				// Admin employees management
				r.Route("/employees", func(r chi.Router) {
					r.Get("/", adminHandler.ListEmployees)
					r.Post("/", adminHandler.CreateEmployee)
					r.Get("/{employeeID}", adminHandler.GetEmployee)
					r.Put("/{employeeID}", adminHandler.UpdateEmployee)
					r.Patch("/{employeeID}/status", adminHandler.UpdateEmployeeStatus)
					r.Post("/{employeeID}/reset-password", adminHandler.ResetEmployeePassword)
				})

				// Admin categories management
				r.Route("/categories", func(r chi.Router) {
					r.Get("/", adminHandler.ListCategories)
					r.Post("/", adminHandler.CreateCategory)
					r.Post("/reorder", adminHandler.ReorderCategories)
					r.Get("/{categoryID}", adminHandler.GetCategory)
					r.Put("/{categoryID}", adminHandler.UpdateCategory)
					r.Patch("/{categoryID}/status", adminHandler.UpdateCategoryStatus)
				})

				// Admin suppliers management
				r.Route("/suppliers", func(r chi.Router) {
					r.Get("/", adminHandler.ListSuppliers)
					r.Get("/active", adminHandler.ListAllActiveSuppliers)
					r.Post("/", adminHandler.CreateSupplier)
					r.Get("/{supplierID}", adminHandler.GetSupplier)
					r.Put("/{supplierID}", adminHandler.UpdateSupplier)
					r.Patch("/{supplierID}/status", adminHandler.UpdateSupplierStatus)
				})
			})
		})
	})

	return r
}
