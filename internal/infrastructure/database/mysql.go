package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
)

// NewMySQL creates a new MySQL database connection using the default timezone.
func NewMySQL(cfg config.DatabaseConfig) (*sql.DB, error) {
	return NewMySQLWithTimezone(cfg, config.DefaultTimezone)
}

// NewMySQLWithTimezone creates a new MySQL database connection with a specific timezone.
func NewMySQLWithTimezone(cfg config.DatabaseConfig, timezone string) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSNWithTimezone(timezone))
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
