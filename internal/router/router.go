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

		r.Get("/franchise/{franchiseID}", container.Handlers.Store.GetByFranchiseID)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRoles("admin"))

			r.Get("/", container.Handlers.Store.GetAll)
			r.Get("/{id}", container.Handlers.Store.GetByID)
			r.Post("/", container.Handlers.Store.Create)
			r.Put("/{id}", container.Handlers.Store.Update)
			r.Delete("/{id}", container.Handlers.Store.Delete)
		})
	})
}

// setupEmployeeRoutes configures employee routes
func setupEmployeeRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/employees", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))
		r.Use(middleware.RequireFranchiseAccess())

		r.Get("/", container.Handlers.Employee.GetAll)
		r.Get("/{id}", container.Handlers.Employee.FindByID)
		r.Post("/", container.Handlers.Employee.Create)
		r.Put("/{id}", container.Handlers.Employee.Update)
		r.Delete("/{id}", container.Handlers.Employee.Delete)

		r.Get("/store/{store_id}", container.Handlers.Employee.GetByStore)
	})
}

// setupEmployeeHoursRoutes configures employee hours routes
func setupEmployeeHoursRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/employee-hours", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))
		r.Use(middleware.RequireFranchiseAccess())

		r.Get("/{id}", container.Handlers.EmployeeHours.GetMonthlySummary)
	})
}

// setupCalendarRoutes configures calendar routes
func setupCalendarRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/calendar", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Get("/holidays", container.Handlers.Calendar.GetHolidays)
		r.Get("/month-summary", container.Handlers.Calendar.GetMonthSummary)
		r.Post("/clear-cache", container.Handlers.Calendar.ClearCache)
	})
}

// setupShiftRoutes configures shift routes
func setupShiftRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/shifts", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", container.Handlers.Shift.Create)
		r.Get("/", container.Handlers.Shift.GetAll)
		r.Get("/single", container.Handlers.Shift.Get)
		r.Get("/store/{store_id}", container.Handlers.Shift.GetByStore)
	})
}

// setupShiftPlanningRoutes configures shift planning routes
func setupShiftPlanningRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/shift-planning", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/preview", container.Handlers.ShiftProjection.Preview)
	})
}

// setupAbsenceRoutes configures absence routes
func setupAbsenceRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/absences", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", container.Handlers.Absence.Create)
		r.Get("/", container.Handlers.Absence.GetByEmployeeAndMonth)
	})
}

// setupNoveltyRoutes configures novelty routes
func setupNoveltyRoutes(r *chi.Mux, container *container.Container) {
	r.Route("/novelties", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", container.Handlers.Novelty.Create)
		r.Get("/", container.Handlers.Novelty.GetByEmployeeAndMonth)
	})
}
