package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"loopi-api/config"
	"loopi-api/internal/delivery/http"
	"loopi-api/internal/middleware"
	repository "loopi-api/internal/repository/mysql"
	"loopi-api/internal/usecase"
	nethttp "net/http"
	"os"
)

func main() {
	// Cargar configuración
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, proceeding with system env")
	}
	config.LoadSecrets()

	// Instanciar conexión a DB
	db, err := gorm.Open(mysql.Open(config.GetDB()), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	// Instanciar casos de uso
	userRepo := repository.NewUserRepository(db)
	authUseCase := usecase.NewAuthUseCase(userRepo)
	authHandler := http.NewAuthHandler(authUseCase)

	franchiseRepo := repository.NewFranchiseRepository(db)
	franchiseUseCase := usecase.NewFranchiseUseCase(franchiseRepo)
	franchiseHandler := http.NewFranchiseHandler(franchiseUseCase)

	storeRepo := repository.NewStoreRepository(db)
	storeUseCase := usecase.NewStoreUseCase(storeRepo)
	storeHandler := http.NewStoreHandler(storeUseCase)

	employeeUseCase := usecase.NewEmployeeUseCase(userRepo)
	employeeHandler := http.NewEmployeeHandler(employeeUseCase)

	assignedRepo := repository.NewAssignedShiftRepository(db)
	absenceRepo := repository.NewAbsenceRepository(db)
	noveltyRepo := repository.NewNoveltyRepository(db)
	employeeHoursUseCase := usecase.NewEmployeeHoursUseCase(assignedRepo, absenceRepo, noveltyRepo, userRepo)
	employeeHoursHandler := http.NewEmployeeHoursHandler(employeeHoursUseCase)

	shiftRepo := repository.NewShiftRepository(db)
	shiftUseCase := usecase.NewShiftUseCase(shiftRepo)
	shiftHandler := http.NewShiftHandler(shiftUseCase)

	workConfigRepo := repository.NewWorkConfigRepository(db)
	shiftProjectionUseCase := usecase.NewShiftProjectionUseCase(shiftRepo, workConfigRepo)
	shiftProjectionHandler := http.NewShiftProjectionHandler(shiftProjectionUseCase)

	calendarUseCase := usecase.NewCalendarUseCase()
	calendarHandler := http.NewCalendarHandler(calendarUseCase)

	absenceUseCase := usecase.NewAbsenceUseCase(absenceRepo)
	absenceHandler := http.NewAbsenceHandler(absenceUseCase)

	noveltyUseCase := usecase.NewNoveltyUseCase(noveltyRepo)
	noveltyHandler := http.NewNoveltyHandler(noveltyUseCase)

	// Configurar router
	r := chi.NewRouter()
	r.Use(middleware.CORS)

	// Rutas públicas
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTMiddleware)

			r.Post("/context", authHandler.SelectContext)
		})
	})

	// Rutas protegidas
	r.Route("/franchises", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)

		r.Get("/", franchiseHandler.GetAll)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRoles("admin"))

			r.Get("/{id}", franchiseHandler.GetById)
			r.Post("/", franchiseHandler.Create)
		})
	})

	r.Route("/stores", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)

		r.Get("/franchise/{franchiseID}", storeHandler.GetByFranchiseID)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRoles("admin"))

			r.Get("/", storeHandler.GetAll)
			r.Get("/{id}", storeHandler.GetByID)
			r.Post("/", storeHandler.Create)
			r.Put("/{id}", storeHandler.Update)
			r.Delete("/{id}", storeHandler.Delete)
		})
	})

	r.Route("/employees", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))
		r.Use(middleware.RequireFranchiseAccess())

		r.Post("/", employeeHandler.Create)
	})

	r.Route("/employee-hours", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))
		r.Use(middleware.RequireFranchiseAccess())

		r.Get("/{id}", employeeHoursHandler.GetMonthlySummary)
	})

	r.Route("/calendar", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Get("/holidays", calendarHandler.GetHolidays)
		r.Get("/month-summary", calendarHandler.GetMonthSummary)
		r.Post("/clear-cache", calendarHandler.ClearCache)
	})

	r.Route("/shifts", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", shiftHandler.Create)
		r.Get("/", shiftHandler.GetAll)
		r.Get("/single", shiftHandler.Get)
	})

	r.Route("/shift-planning", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/preview", shiftProjectionHandler.Preview)
	})

	r.Route("/absences", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", absenceHandler.Create)
		r.Get("/", absenceHandler.GetByEmployeeAndMonth)
	})

	r.Route("/novelties", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", noveltyHandler.Create)
		r.Get("/", noveltyHandler.GetByEmployeeAndMonth)
	})

	// Servidor
	port := os.Getenv("PORT")
	log.Printf("✅ API running at http://localhost:%s", port)
	log.Fatal(nethttp.ListenAndServe(":"+port, r))
}
