package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// DefaultTimezone is the default timezone for the application (Colombia).
const DefaultTimezone = "America/Bogota"

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	Tracing  TracingConfig
	Timezone string
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string // debug | info | warn | error
	Format string // text (dev) | json (production + Cloud Logging)
}

// TracingConfig holds OpenTelemetry / Cloud Trace configuration.
type TracingConfig struct {
	ProjectID      string
	ServiceName    string
	ServiceVersion string
	Env            string
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database-related configuration.
type DatabaseConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	Name               string
	InstanceConnection string // For Cloud SQL Unix socket (App Engine)
}

// JWTConfig holds JWT-related configuration.
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	_ = godotenv.Load()

	appEnv, err := getEnvRequired("APP_ENV")
	if err != nil {
		return nil, err
	}

	dbName, err := resolveDBName(appEnv)
	if err != nil {
		return nil, err
	}

	jwtExpHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION_HOURS: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		Database: DatabaseConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnv("DB_PORT", "3306"),
			User:               getEnv("DB_USER", "loopi"),
			Password:           getEnv("DB_PASSWORD", ""),
			Name:               dbName,
			InstanceConnection: getEnv("DB_INSTANCE_CONNECTION", ""),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "dev-secret-key-change-in-production"),
			ExpirationHours: jwtExpHours,
		},
		Tracing: TracingConfig{
			ProjectID:      getEnv("GOOGLE_CLOUD_PROJECT", ""),
			ServiceName:    getEnv("SERVICE_NAME", "loopi-api"),
			ServiceVersion: getEnv("SERVICE_VERSION", "dev"),
			Env:            getEnv("SERVICE_ENV", "development"),
		},
		Timezone: getEnv("TZ", DefaultTimezone),
	}, nil
}

// DSN returns the database connection string.
func (c *DatabaseConfig) DSN() string {
	return c.DSNWithTimezone(DefaultTimezone)
}

// DSNWithTimezone returns the database connection string with a specific timezone.
func (c *DatabaseConfig) DSNWithTimezone(timezone string) string {
	encodedTZ := url.QueryEscape(timezone)
	if c.InstanceConnection != "" {
		return fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
			c.User, c.Password, c.InstanceConnection, c.Name, encodedTZ)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, encodedTZ)
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvRequired retrieves an environment variable and returns an error if it is not set or empty.
func getEnvRequired(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return value, nil
}

// resolveDBName maps an APP_ENV value to the corresponding MySQL schema name.
func resolveDBName(appEnv string) (string, error) {
	switch appEnv {
	case "development", "test":
		return "loopi_dev", nil
	case "production":
		return "loopi_prod", nil
	default:
		return "", fmt.Errorf("unknown APP_ENV value %q: must be one of: development, test, production", appEnv)
	}
}
