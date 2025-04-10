package config

import (
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Token     string    `yaml:"token" env:"PSCLOUD_TOKEN"`
	ServiceID string    `yaml:"serviceId" env:"PSCLOUD_SERVICE_ID"`
	BaseURL   string    `yaml:"baseUrl" env:"PSCLOUD_BASE_URL"`
	Web       WebConfig `yaml:"web"`
}

// WebConfig represents the web server configuration
type WebConfig struct {
	ListenAddress string `yaml:"listenAddress" env:"WEB_LISTEN_ADDRESS"`
	MetricsPrefix string `yaml:"metricsPrefix" env:"WEB_METRICS_PREFIX"`
	TelemetryPath string `yaml:"telemetryPath" env:"WEB_TELEMETRY_PATH"`
}

// LoadConfig loads the configuration from a YAML file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		BaseURL: "https://console.ps.kz",
		Web: WebConfig{
			ListenAddress: ":9116",
			MetricsPrefix: "pskz",
			TelemetryPath: "/metrics",
		},
	}

	// Load .env file if it exists
	envFiles := []string{".env", ".env.local"}
	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			if err := godotenv.Load(envFile); err != nil {
				return nil, err
			}
		}
	}

	// If config file is provided, load from it
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	config.Token = getEnvOrDefault("PSCLOUD_TOKEN", config.Token)
	config.ServiceID = getEnvOrDefault("PSCLOUD_SERVICE_ID", config.ServiceID)
	config.BaseURL = getEnvOrDefault("PSCLOUD_BASE_URL", config.BaseURL)

	// Web configuration
	config.Web.ListenAddress = getEnvOrDefault("WEB_LISTEN_ADDRESS", config.Web.ListenAddress)
	config.Web.MetricsPrefix = getEnvOrDefault("WEB_METRICS_PREFIX", config.Web.MetricsPrefix)
	config.Web.TelemetryPath = getEnvOrDefault("WEB_TELEMETRY_PATH", config.Web.TelemetryPath)

	return config, nil
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
