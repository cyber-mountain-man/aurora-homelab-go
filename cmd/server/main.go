package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/config"
	"github.com/cyber-mountain-man/aurora-homelab-go/internal/handlers"
	"github.com/cyber-mountain-man/aurora-homelab-go/internal/health"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	addr := getEnv("AURORA_ADDR", ":8080")

	// Load configuration (services, etc.).
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("warning: could not load config.yaml: %v", err)
		cfg = &config.Config{}
	}

	// Health checker: run every 30s, 3s timeout per service.
	checker := health.NewChecker(
		cfg.Services,
		30*time.Second, // interval
		3*time.Second,  // HTTP timeout
		2*time.Second,  // TCP timeout
	)

	checker.Start()

	mux := http.NewServeMux()

	// Static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Dashboard handler with services + health checker.
	dh, err := handlers.NewDashboardHandler("./web/templates", cfg.Services, checker)
	if err != nil {
		log.Fatalf("failed to initialize dashboard handler: %v", err)
	}

	mux.HandleFunc("/", dh.Dashboard)
	mux.HandleFunc("/dashboard/partial", dh.DashboardPartial)
	mux.HandleFunc("/services/recheck", dh.RecheckService)

	log.Printf("Aurora Homelab listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
