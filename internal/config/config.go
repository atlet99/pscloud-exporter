package config

import (
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Username string    `yaml:"username" env:"PSCLOUD_USERNAME"`
	Password string    `yaml:"password" env:"PSCLOUD_PASSWORD"`
	BaseURL  string    `yaml:"baseUrl" env:"PSCLOUD_BASE_URL"`
	UseHTTP  bool      `yaml:"useHttp" env:"PSCLOUD_USE_HTTP"`
	Web      WebConfig `yaml:"web"`
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
		BaseURL: "https://api.ps.kz/v1",
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
	config.Username = getEnvOrDefault("PSCLOUD_USERNAME", config.Username)
	config.Password = getEnvOrDefault("PSCLOUD_PASSWORD", config.Password)
	config.BaseURL = getEnvOrDefault("PSCLOUD_BASE_URL", config.BaseURL)
	config.UseHTTP = getBoolEnv("PSCLOUD_USE_HTTP", config.UseHTTP)

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

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}
