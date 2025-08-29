package config

import (
	"log"
	"os"
)

type Secrets struct {
	Port      string
	DbDsn     string
	JWTSecret string
	Env       string
}

var secrets Secrets

func LoadSecrets() {
	secrets = Secrets{
		Port:      getOr("PORT", "8080"),
		DbDsn:     mustGet("DB_DSN"),
		JWTSecret: mustGet("JWT_SECRET"),
		Env:       getOr("ENV", "development"),
	}
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("❌ Missing required environment variable: %s", key)
	}
	return val
}

func getOr(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func GetPort() string {
	return secrets.Port
}

func GetDB() string {
	return secrets.DbDsn
}

func GetJWTSecret() string {
	return secrets.JWTSecret
}

func GetEnv() string {
	return secrets.Env
}
