package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atlet99/pscloud-exporter/internal/client"
	"github.com/atlet99/pscloud-exporter/internal/collector"
	"github.com/atlet99/pscloud-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Version information, set during build
	Version = "dev"
	// Build information, set during build
	Build = "unknown"
)

// displayVersion prints the version information in a formatted way
func displayVersion() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build: %s\n", Build)
}

func findConfigFile(configPath string) (string, error) {
	// If path is explicitly specified, check its existence
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
		return "", fmt.Errorf("config file not found: %s", configPath)
	}

	// Check both extension variants
	configFiles := []string{"config.yml", "config.yaml"}
	for _, file := range configFiles {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
	}

	return "", fmt.Errorf("no config file found. Please create either config.yml or config.yaml")
}

// validateAuth attempts to validate the API token by making a test API call
func validateAuth(c *client.Client) error {
	log.Println("Validating API token...")
	userData, err := c.TestAuth()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	log.Printf("Authentication successful! Logged in as user ID: %d, username: %s",
		userData.Data.User.ID, userData.Data.User.Username)
	return nil
}

func main() {
	// Variable declarations
	var (
		listenAddress = flag.String("listen-address", ":9116", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("metrics-path", "/metrics", "Path under which to expose metrics.")
		configFile    = flag.String("config", "", "Path to configuration file (supports .yml or .yaml)")
		token         = flag.String("token", "", "PS.KZ API token")
		serviceID     = flag.String("service-id", "", "PS.KZ service ID for cloud servers")
		baseURL       = flag.String("base-url", "", "Base URL for PS.KZ API (default: https://console.ps.kz)")
		skipAuth      = flag.Bool("skip-auth-check", false, "Skip authentication validation on startup")
		showVersion   = flag.Bool("version", false, "Show version information and exit")
	)

	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		displayVersion()
		os.Exit(0)
	}

	// Find configuration file
	configPath, err := findConfigFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Using config file: %s", configPath)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Command line arguments take priority
	if *token != "" {
		cfg.Token = *token
	}

	if *serviceID != "" {
		cfg.ServiceID = *serviceID
	}

	// Check if token exists
	if cfg.Token == "" {
		log.Fatal("API token is required. Set it in config file or via -token flag.")
	}

	// Create API client with options
	clientOptions := client.ClientOptions{}

	// Set base URL if provided
	if *baseURL != "" {
		clientOptions.BaseURL = *baseURL
	} else if cfg.BaseURL != "" {
		clientOptions.BaseURL = cfg.BaseURL
	}

	// Create client with options
	c := client.NewWithOptions(cfg.Token, clientOptions)

	// Validate authentication unless skipped
	if !*skipAuth {
		if err := validateAuth(c); err != nil {
			log.Fatal(err)
		}
	}

	// Create a new registry for our metrics
	reg := prometheus.NewRegistry()

	// Create and register our collector
	exporter := collector.New(c, cfg.ServiceID)
	reg.MustRegister(exporter)

	// Create handler for metrics with our registry
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>PSCloud Exporter</title></head>
			<body>
			<h1>PSCloud Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<p>Version: ` + Version + `</p>
			<p>Build: ` + Build + `</p>
			</body>
			</html>`))
		if err != nil {
			log.Printf("Error writing response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	srv := &http.Server{
		Addr: *listenAddress,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	log.Printf("Server listening on %s", *listenAddress)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %s", err)
	}
}
