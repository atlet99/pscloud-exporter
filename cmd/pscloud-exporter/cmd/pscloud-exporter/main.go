package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/atlet99/pscloud-exporter/internal/client"
	"github.com/atlet99/pscloud-exporter/internal/collector"
	"github.com/atlet99/pscloud-exporter/internal/config"
)

func main() {
	cfg := config.Load()
	api := client.New(cfg.Username, cfg.Password)
	exporter := collector.New(api)

	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Exporter running on port:", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
