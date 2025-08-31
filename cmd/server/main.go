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

	franchiseUseCase := usecase.NewFranchiseUseCase()
	franchiseHandler := http.NewFranchiseHandler(franchiseUseCase)

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
	shiftProjectionHanlder := http.NewShiftProjectionHandler(shiftProjectionUseCase)

	calendarUseCase := usecase.NewCalendarUseCase()
	calendarHandler := http.NewCalendarHandler(calendarUseCase)

	absenceUseCase := usecase.NewAbsenceUseCase(absenceRepo)
	absenceHanlder := http.NewAbsenceHandler(absenceUseCase)

	noveltyUseCase := usecase.NewNoveltyUseCase(noveltyRepo)
	noveltyHanlder := http.NewNoveltyHandler(noveltyUseCase)

	// Configurar router
	r := chi.NewRouter()

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
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", franchiseHandler.Create)
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

		r.Post("/preview", shiftProjectionHanlder.Preview)
	})

	r.Route("/absences", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", absenceHanlder.Create)
		r.Get("/", absenceHanlder.GetByEmployeeAndMonth)
	})

	r.Route("/novelties", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin"))

		r.Post("/", noveltyHanlder.Create)
		r.Get("/", noveltyHanlder.GetByEmployeeAndMonth)
	})

	// Servidor
	port := os.Getenv("PORT")
	log.Printf("✅ API running at http://localhost:%s", port)
	log.Fatal(nethttp.ListenAndServe(":"+port, r))
}
