package config

import (
	"log"
	"os"
)

// Config holds application configuration.
type Config struct {
	Port      string
	DBUrl     string
	JWTSecret string
}

// Load reads configuration from environment variables.
func Load() *Config {
	// Get JWT_SECRET or exit
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set.")
	}

	// Get DB_URL or exit
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("FATAL: DB_URL environment variable is not set.")
	}

	cfg := &Config{
		Port:      getEnv("PORT", "8080"), // A default port is fine
		DBUrl:     dbURL,
		JWTSecret: jwtSecret,
	}
	return cfg
}

// getEnv returns an env value or a default. Only use for non-sensitive data.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
