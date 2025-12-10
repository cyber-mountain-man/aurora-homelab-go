package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/handlers"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	addr := getEnv("AURORA_ADDR", ":8080")

	mux := http.NewServeMux()

	// Static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Dashboard handler
	dh, err := handlers.NewDashboardHandler("./web/templates")
	if err != nil {
		log.Fatalf("failed to initialize dashboard handler: %v", err)
	}

	mux.HandleFunc("/", dh.Dashboard)

	log.Printf("Aurora Homelab listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
