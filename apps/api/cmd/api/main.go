package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func main() {
	cfg := config.FromEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/api/v1/health", healthHandler)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api listening on :%s", cfg.AppPort)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:    "ok",
		Service:   "api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
