package main

import (
	"log"
	"net/http"

	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/database"
	"github.com/manuelgomezsw/loopi-api/internal/interface/router"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database successfully")

	// Initialize router with dependencies
	r := router.New(db, cfg)

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
