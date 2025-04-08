package config

import (
	"os"
)

type Config struct {
	Username string
	Password string
	Port     string
}

func Load() *Config {
	return &Config{
		Username: getEnv("PSCLOUD_USERNAME", ""),
		Password: getEnv("PSCLOUD_PASSWORD", ""),
		Port:     getEnv("PSCLOUD_EXPORTER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
