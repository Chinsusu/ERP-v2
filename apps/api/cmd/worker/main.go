package main

import (
	"log"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func main() {
	cfg := config.FromEnv()
	log.Printf("worker started env=%s", cfg.AppEnv)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Print("worker heartbeat")
	}
}
