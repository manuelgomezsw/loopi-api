package router

import (
	"loopi-api/internal/container"
	"loopi-api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// SetupRoutes configures all application routes
func SetupRoutes(container *container.Container) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.CORS)

	// Setup route groups
	setupAuthRoutes(r, container)
	setupFranchiseRoutes(r, container)
	setupStoreRoutes(r, container)
	setupEmployeeRoutes(r, container)
	setupEmployeeHoursRoutes(r, container)
	setupCalendarRoutes(r, container)
	setupShiftRoutes(r, container)
	setupShiftPlanningRoutes(r, container)
	setupAbsenceRoutes(r, container)
	setupNoveltyRoutes(r, container)

	return r
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", container.Handlers.Auth.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTMiddleware)
			r.Post("/context", container.Handlers.Auth.SelectContext)
		})
	})
}

// setupFranchiseRoutes configures franchise routes
func setupFranchiseRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/franchises", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)

		r.Get("/", container.Handlers.Franchise.GetAll)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRoles("admin"))

			r.Get("/{id}", container.Handlers.Franchise.GetById)
			r.Post("/", container.Handlers.Franchise.Create)
		})
	})
}

// setupStoreRoutes configures store routes
func setupStoreRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/stores", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)

		// Public franchise-based routes
		r.Get("/franchise/{franchiseID}", container.Handlers.Store.GetByFranchiseID)
		r.Get("/franchise/{franchiseID}/active", container.Handlers.Store.GetActiveStoresByFranchise)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRoles("admin"))

			// Standard CRUD routes
			r.Get("/", container.Handlers.Store.GetAll)
			r.Get("/{id}", container.Handlers.Store.GetByID)
			r.Post("/", container.Handlers.Store.Create)
			r.Put("/{id}", container.Handlers.Store.Update)
			r.Delete("/{id}", container.Handlers.Store.Delete)

			// Business-specific routes
			r.Get("/with-employee-count", container.Handlers.Store.GetStoresWithEmployeeCount)
			r.Get("/statistics", container.Handlers.Store.GetStatistics)
		})
	})
}

// setupEmployeeRoutes configures employee routes
func setupEmployeeRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/employees", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))
		r.Use(middleware.RequireFranchiseAccess())

		// Standard CRUD routes
		r.Get("/", container.Handlers.Employee.GetAll)
		r.Get("/{id}", container.Handlers.Employee.FindByID)
		r.Post("/", container.Handlers.Employee.Create)
		r.Put("/{id}", container.Handlers.Employee.Update)
		r.Delete("/{id}", container.Handlers.Employee.Delete)

		// Business-specific routes
		r.Get("/active", container.Handlers.Employee.GetActiveEmployees)
		r.Get("/store/{store_id}", container.Handlers.Employee.GetByStore)
		r.Get("/store/{store_id}/active", container.Handlers.Employee.GetByStoreAndActive)
	})
}

// setupEmployeeHoursRoutes configures employee hours routes
func setupEmployeeHoursRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/employee-hours", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))
		r.Use(middleware.RequireFranchiseAccess())

		// Standard summary routes
		r.Get("/{id}/monthly", container.Handlers.EmployeeHours.GetMonthlySummary)
		r.Get("/{id}/daily", container.Handlers.EmployeeHours.GetDailySummary)
		r.Get("/{id}/yearly", container.Handlers.EmployeeHours.GetYearlySummary)

		// Business calculation routes
		r.Get("/{id}/working-days", container.Handlers.EmployeeHours.GetWorkingDays)

		// Legacy route for backward compatibility
		r.Get("/{id}", container.Handlers.EmployeeHours.GetMonthlySummary)
	})
}

// setupCalendarRoutes configures calendar routes
func setupCalendarRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/calendar", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		// Standard calendar routes
		r.Get("/holidays", container.Handlers.Calendar.GetHolidays)
		r.Get("/month-summary", container.Handlers.Calendar.GetMonthSummary)
		r.Get("/enhanced-summary", container.Handlers.Calendar.GetEnhancedSummary)

		// Business calculation routes
		r.Get("/working-days", container.Handlers.Calendar.GetWorkingDays)

		// Utility routes
		r.Post("/clear-cache", container.Handlers.Calendar.ClearCache)
	})
}

// setupShiftRoutes configures shift routes
func setupShiftRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/shifts", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		// Standard CRUD routes
		r.Post("/", container.Handlers.Shift.Create)
		r.Get("/", container.Handlers.Shift.GetAll)
		r.Get("/{id}", container.Handlers.Shift.Get)
		r.Put("/{id}", container.Handlers.Shift.Update)
		r.Delete("/{id}", container.Handlers.Shift.Delete)

		// Store-specific routes
		r.Get("/store/{store_id}", container.Handlers.Shift.GetByStore)
		r.Get("/store/{store_id}/statistics", container.Handlers.Shift.GetStatistics)

		// Business-specific routes
	})
}

// setupShiftPlanningRoutes configures shift planning routes
func setupShiftPlanningRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/shift-planning", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		// Standard projection routes
		r.Post("/preview", container.Handlers.ShiftProjection.Preview)
		r.Get("/summary", container.Handlers.ShiftProjection.GetProjectionSummary)

		// Business calculation routes
		r.Get("/projected-days", container.Handlers.ShiftProjection.GetProjectedDays)
	})
}

// setupAbsenceRoutes configures absence routes
func setupAbsenceRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/absences", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		// Standard routes
		r.Post("/", container.Handlers.Absence.Create)
		r.Get("/monthly", container.Handlers.Absence.GetByEmployeeAndMonth)

		// Business-specific routes
		r.Get("/date-range", container.Handlers.Absence.GetByEmployeeAndDateRange)
		r.Get("/total-hours", container.Handlers.Absence.GetTotalHours)

		// Legacy route for backward compatibility
		r.Get("/", container.Handlers.Absence.GetByEmployeeAndMonth)
	})
}

// setupNoveltyRoutes configures novelty routes
func setupNoveltyRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/novelties", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		// Standard routes
		r.Post("/", container.Handlers.Novelty.Create)
		r.Get("/monthly", container.Handlers.Novelty.GetByEmployeeAndMonth)

		// Business-specific routes
		r.Get("/date-range", container.Handlers.Novelty.GetByEmployeeAndDateRange)
		r.Get("/total-hours-by-type", container.Handlers.Novelty.GetTotalHoursByType)
		r.Get("/types-summary", container.Handlers.Novelty.GetTypesSummary)

		// Legacy route for backward compatibility
		r.Get("/", container.Handlers.Novelty.GetByEmployeeAndMonth)
	})
}
