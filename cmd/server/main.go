package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"loopi-api/config"
	"loopi-api/internal/delivery/http"
	"loopi-api/internal/middleware"
	"loopi-api/internal/usecase/auth"
	nethttp "net/http"
	"os"
)

func main() {
	config.LoadEnv()

	// Instanciar conexión a DB

	// Instanciar casos de uso
	authUseCase := auth.NewAuthUseCase()
	authHandle := http.NewAuthHandler(authUseCase)

	// Crear handlers

	// Configurar router
	r := chi.NewRouter()

	// Rutas públicas
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandle.Login)
	})

	r.Route("/employees", func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)
		//r.Get("/employees", auth.GetEmployee)
	})

	// Rutas protegidas

	// Servidor
	port := os.Getenv("PORT")
	log.Printf("✅ API running at http://localhost:%s", port)
	log.Fatal(nethttp.ListenAndServe(":"+port, r))
}
