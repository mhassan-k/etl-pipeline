package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	APIURL        string
	DatabaseURL   string
	FetchInterval int
	ServerPort    string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	fetchInterval, err := strconv.Atoi(getEnv("FETCH_INTERVAL", "30"))
	if err != nil {
		fetchInterval = 30
	}

	return &Config{
		APIURL:        getEnv("API_URL", "https://jsonplaceholder.typicode.com/posts"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://etl_user:etl_password@localhost:5432/etl_db?sslmode=disable"),
		FetchInterval: fetchInterval,
		ServerPort:    getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

