package main

import (
	"fmt"
	"log"

	"pdrest/internal/config"
	"pdrest/internal/data"
	"pdrest/internal/database"
	"pdrest/internal/interfaces/http"
	"pdrest/internal/interfaces/services"

	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create Echo instance
	e := echo.New()

	var repo data.ClientRepository
	var clientService *services.ClientService
	var eventService *services.EventService

	// Try to connect to PostgreSQL database
	db, err := database.New(cfg.GetDatabaseURL(), cfg.Database.MaxConns)
	if err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL: %v", err)
		log.Println("Falling back to in-memory repository")
		repo = data.NewInMemoryClientRepository()
		clientService = services.NewClientService(repo)
		// Event service requires database - will return error if accessed
		eventService = nil
	} else {
		defer db.Close()
		log.Println("Successfully connected to PostgreSQL database")

		// Create PostgreSQL repositories
		postgresRepo := data.NewPostgresClientRepository(db.Pool)
		eventRepo := data.NewPostgresEventRepository(db.Pool)

		repo = postgresRepo

		// Create services
		clientService = services.NewClientService(repo)
		eventService = services.NewEventService(eventRepo)
	}

	// Register HTTP handlers (eventService may be nil if database unavailable)
	http.NewHTTPHandler(e, clientService, eventService)

	// Start server
	addr := cfg.GetAddress()
	fmt.Printf("Server starting on %s\n", addr)
	if err := e.Start(addr); err != nil {
		log.Fatal(err)
	}
}
