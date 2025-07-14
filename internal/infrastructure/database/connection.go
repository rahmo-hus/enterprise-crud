// Package database provides database connection and repository implementations.
// It handles PostgreSQL database connectivity using GORM and implements repository patterns
// for all domain entities including users, events, venues, orders, and tickets.
package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connection manages database connectivity
// Handles PostgreSQL database connection and configuration
type Connection struct {
	DB *gorm.DB // GORM database instance
}

// NewConnection creates a new database connection
// Returns Connection instance with established database connection
func NewConnection() (*Connection, error) {
	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Fallback to default connection string
		databaseURL = "postgres://postgres:postgres@localhost:5433/enterprise_crud?sslmode=disable"
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return &Connection{DB: db}, nil
}

// Close closes the database connection
// Should be called when application shuts down
func (c *Connection) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}
