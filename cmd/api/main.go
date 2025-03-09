package main

import (
	"log"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/api"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/config"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/db"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	database, err := db.Init(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Auto migrate database models
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// Initialize and start the API server
	server := api.NewServer(cfg, database)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
