package router

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/manuelgomezsw/loopi-api/internal/application/service"
	"github.com/manuelgomezsw/loopi-api/internal/domain/inventory"
	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/auth"
	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/repository"
	"github.com/manuelgomezsw/loopi-api/internal/interface/handler/admin"
	authhandler "github.com/manuelgomezsw/loopi-api/internal/interface/handler/auth"
	"github.com/manuelgomezsw/loopi-api/internal/interface/handler/employee"
	"github.com/manuelgomezsw/loopi-api/internal/interface/middleware"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
)

// New creates and configures the HTTP router with all routes.
func New(db *sql.DB, cfg *config.Config, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimiddleware.Recoverer)

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
	measurementUnitRepo := repository.NewMySQLMeasurementUnitRepository(db)
	inventoryRepo := repository.NewMySQLInventoryRepository(db)
	inventoryDetailRepo := repository.NewMySQLInventoryDetailRepository(db)
	categoryRepo := repository.NewMySQLCategoryRepository(db)
	supplierRepo := repository.NewMySQLSupplierRepository(db)

	inventoryEnricher := inventory.NewEnricher(inventoryRepo, inventoryDetailRepo)

	// Initialize services
	authService := service.NewAuthService(employeeRepo, jwtManager)
	inventoryService := service.NewInventoryService(
		inventoryRepo,
		inventoryDetailRepo,
		itemRepo,
		inventoryEnricher,
	)
	employeeService := service.NewEmployeeService(inventoryService)
	adminService := service.NewAdminService(
		inventoryRepo,
		inventoryDetailRepo,
		employeeRepo,
		itemRepo,
		measurementUnitRepo,
		categoryRepo,
		supplierRepo,
		inventoryEnricher,
	)

	// Initialize handlers: auth, employee (inventory), admin (per domain)
	authHandler := authhandler.NewHandler(authService)
	employeeInventoryHandler := employee.NewInventoryHandler(employeeService)
	adminDashboardHandler := admin.NewDashboardHandler(adminService.Dashboard())
	adminInventoryHandler := admin.NewInventoryHandler(adminService.Inventory())
	adminItemHandler := admin.NewItemHandler(adminService.Item())
	adminEmployeeHandler := admin.NewEmployeeHandler(adminService.Employee())
	adminCategoryHandler := admin.NewCategoryHandler(adminService.Category())
	adminSupplierHandler := admin.NewSupplierHandler(adminService.Supplier())

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

			// Employees routes (session)
			r.Route("/employees", func(r chi.Router) {
				r.Get("/me", authHandler.GetMe)
			})

			// Employee inventories (operativo)
			r.Route("/inventories", func(r chi.Router) {
				r.Get("/latest", employeeInventoryHandler.GetLatest)
				r.Get("/in-progress", employeeInventoryHandler.GetInProgress)
				r.Get("/suggested-schedule", employeeInventoryHandler.GetSuggestedSchedule)
				r.Post("/", employeeInventoryHandler.Create)

				r.Route("/{inventoryID}", func(r chi.Router) {
					r.Get("/items", employeeInventoryHandler.GetItems)
					r.Post("/details", employeeInventoryHandler.SaveDetail)
					r.Get("/discrepancies", employeeInventoryHandler.GetDiscrepancies)
					r.Post("/sales", employeeInventoryHandler.SaveSales)
					r.Get("/summary", employeeInventoryHandler.GetSummary)
					r.Post("/complete", employeeInventoryHandler.Complete)
				})
			})

			// Admin routes (requires admin role)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminOnly)

				r.Get("/dashboard", adminDashboardHandler.GetDashboard)

				r.Route("/inventories", func(r chi.Router) {
					r.Get("/", adminInventoryHandler.ListInventories)
					r.Get("/active-count", adminInventoryHandler.GetActiveInventoriesCount)
					r.Post("/initial", adminInventoryHandler.CreateInitialInventory)
					r.Get("/{inventoryID}", adminInventoryHandler.GetInventoryDetail)
					r.Put("/{inventoryID}/details/{detailID}", adminInventoryHandler.UpdateInventoryDetail)
				})

				r.Get("/measurement-units", adminItemHandler.ListMeasurementUnits)

				r.Route("/items", func(r chi.Router) {
					r.Get("/", adminItemHandler.ListItems)
					r.Post("/", adminItemHandler.CreateItem)
					r.Get("/{itemID}", adminItemHandler.GetItem)
					r.Put("/{itemID}", adminItemHandler.UpdateItem)
					r.Patch("/{itemID}/status", adminItemHandler.UpdateItemStatus)
				})

				r.Route("/employees", func(r chi.Router) {
					r.Get("/", adminEmployeeHandler.ListEmployees)
					r.Get("/active", adminEmployeeHandler.ListAllActiveEmployees)
					r.Post("/", adminEmployeeHandler.CreateEmployee)
					r.Get("/{employeeID}", adminEmployeeHandler.GetEmployee)
					r.Put("/{employeeID}", adminEmployeeHandler.UpdateEmployee)
					r.Patch("/{employeeID}/status", adminEmployeeHandler.UpdateEmployeeStatus)
					r.Post("/{employeeID}/reset-password", adminEmployeeHandler.ResetEmployeePassword)
				})

				r.Route("/categories", func(r chi.Router) {
					r.Get("/", adminCategoryHandler.ListCategories)
					r.Post("/", adminCategoryHandler.CreateCategory)
					r.Post("/reorder", adminCategoryHandler.ReorderCategories)
					r.Get("/{categoryID}", adminCategoryHandler.GetCategory)
					r.Put("/{categoryID}", adminCategoryHandler.UpdateCategory)
					r.Patch("/{categoryID}/status", adminCategoryHandler.UpdateCategoryStatus)
				})

				r.Route("/suppliers", func(r chi.Router) {
					r.Get("/", adminSupplierHandler.ListSuppliers)
					r.Get("/active", adminSupplierHandler.ListAllActiveSuppliers)
					r.Post("/", adminSupplierHandler.CreateSupplier)
					r.Get("/{supplierID}", adminSupplierHandler.GetSupplier)
					r.Put("/{supplierID}", adminSupplierHandler.UpdateSupplier)
					r.Patch("/{supplierID}/status", adminSupplierHandler.UpdateSupplierStatus)
				})
			})
		})
	})

	return r
}
