package main

import (
	"log"

	"enterprise-crud/internal/config"
	"enterprise-crud/internal/wire"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize application with Wire (requires wire generate)
	// Uncomment the lines below after running: wire generate ./internal/wire
	
	// app, err := wire.InitializeApp(cfg)
	// if err != nil {
	//     log.Fatalf("Failed to initialize app: %v", err)
	// }

	// Run application (handles startup and graceful shutdown)
	// if err := app.Run(); err != nil {
	//     log.Fatalf("Application failed: %v", err)
	// }
	
	log.Println("Wire-based main.go template created")
	log.Println("To use Wire:")
	log.Println("1. Install wire: go install github.com/google/wire/cmd/wire@latest")
	log.Println("2. Generate wire code: wire generate ./internal/wire")
	log.Println("3. Uncomment the code above")
}