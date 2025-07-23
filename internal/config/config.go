package config

import (
	"os"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port  string
	DBUrl string
}

// Load reads configuration from environment variables or uses defaults.
func Load() *Config {
	cfg := &Config{
		Port:  getEnv("PORT", "8080"),
		DBUrl: getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/backend_path?sslmode=disable"),
	}
	return cfg
}

// getEnv returns the value of the environment variable or a default.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
