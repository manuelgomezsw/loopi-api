package main

import (
	"log"
	"loopi-api/internal/app"
)

func main() {
	// Initialize application
	application, err := app.Initialize()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application: %v", err)
	}

	// Start server
	if err := application.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
