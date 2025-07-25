package config

import (
	"os"
)

type Config struct {
	Port         string
	DatabasePath string
	LogLevel     string
	Environment  string
	NgrokToken   string
	NgrokDomain  string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "relay.db"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		NgrokToken:   getEnv("NGROK_TOKEN", ""),
		NgrokDomain:  getEnv("NGROK_DOMAIN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 