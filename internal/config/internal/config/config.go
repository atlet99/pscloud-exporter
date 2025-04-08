package config

import "os"

type Config struct {
	Username string
	Password string
	Port     string
}

func Load() *Config {
	return &Config{
		Username: os.Getenv("PSCLOUD_USERNAME"),
		Password: os.Getenv("PSCLOUD_PASSWORD"),
		Port:     os.Getenv("EXPORTER_PORT"),
	}
}
