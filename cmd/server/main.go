package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"loopi-api/config"
	"loopi-api/internal/delivery/http"
	"loopi-api/internal/middleware"
	"loopi-api/internal/usecase"
	nethttp "net/http"
	"os"
)

func main() {
	config.LoadEnv()

	// Instanciar conexión a DB

	// Instanciar casos de uso
	authUseCase := usecase.NewAuthUseCase()
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
