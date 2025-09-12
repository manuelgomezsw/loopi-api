package app

import (
	"log"
	"net/http"

	"loopi-api/config"
	"loopi-api/internal/container"
	"loopi-api/internal/router"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Application represents the main application structure
type Application struct {
	Container *container.Container
	Router    http.Handler
}

// Initialize sets up the entire application
func Initialize() (*Application, error) {
	// Load environment variables
	if err := loadEnvironment(); err != nil {
		return nil, err
	}

	// Initialize database connection
	db, err := initializeDatabase()
	if err != nil {
		return nil, err
	}

	// Create dependency container
	appContainer := container.NewContainer(db)

	// Setup routes
	appRouter := router.SetupRoutes(appContainer)

	return &Application{
		Container: appContainer,
		Router:    appRouter,
	}, nil
}

// Start starts the HTTP server
func (app *Application) Start() error {
	port := config.GetPort()
	log.Printf("✅ API running at http://localhost:%s", port)
	return http.ListenAndServe(":"+port, app.Router)
}

// loadEnvironment loads the .env file and configuration
func loadEnvironment() error {
	// Load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, proceeding with system env")
	}

	// Load secrets from environment
	config.LoadSecrets()
	return nil
}

// initializeDatabase creates and configures the database connection
func initializeDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.GetDB()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("✅ Database connection established")
	return db, nil
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown() error {
	// Get the underlying SQL DB to close it
	if app.Container.DB != nil {
		sqlDB, err := app.Container.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
