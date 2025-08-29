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
	"loopi-api/internal/repository"
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

	// Configurar router
	r := chi.NewRouter()

	// Rutas públicas
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
	})

	r.Route("/franchises", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin", "supervisor"))

		r.Post("/", franchiseHandler.Create)
	})

	r.Route("/employees", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		r.Use(middleware.RequireRoles("admin", "supervisor"))
		r.Use(middleware.RequireFranchiseAccess())

		r.Post("/", franchiseHandler.Create)
	})

	// Rutas protegidas

	// Servidor
	port := os.Getenv("PORT")
	log.Printf("✅ API running at http://localhost:%s", port)
	log.Fatal(nethttp.ListenAndServe(":"+port, r))
}
