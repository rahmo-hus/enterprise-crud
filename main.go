// @title Enterprise CRUD API
// @version 1.0.0
// @description A RESTful API for user management and event ticketing system with CRUD operations and JWT authentication
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"log"

	"enterprise-crud/internal/app"
	"enterprise-crud/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Method 1: Using manual dependency injection
	deps, err := app.NewDependencies(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Create application with dependencies
	application := app.NewWireApp(cfg, deps.DBConn, deps.RedisClient, deps.UserHandler, deps.EventHandler, deps.OrderHandler, deps.VenueHandler)

	// Run application (handles startup and graceful shutdown)
	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
