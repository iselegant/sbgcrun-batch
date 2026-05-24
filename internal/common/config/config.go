package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/uma-arai/sbcntr-batch/internal/common/database"
)

type Config struct {
	DB            database.Config
	EnableTracing bool
}

// LoadConfig は設定を読み込みます
func LoadConfig() (*Config, error) {
	cfg := &Config{
		DB: database.Config{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
			UserName: getEnvOrDefault("DB_USERNAME", "sbcntrapp"),
			Password: getEnvOrDefault("DB_PASSWORD", "password"),
			DBName:   getEnvOrDefault("DB_NAME", "sbcntrapp"),
		},
		EnableTracing: isTracingEnabled(),
	}

	return cfg, nil
}

func isTracingEnabled() bool {
	v := os.Getenv("ENABLE_TRACING")
	return strings.ToLower(v) == "true" || v == "1"
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("Environment variable %s is not set, using default value", key)
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
