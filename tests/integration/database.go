//go:build integration
// +build integration

package integration

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"enterprise-crud/internal/domain/event"
	"enterprise-crud/internal/domain/order"
	"enterprise-crud/internal/domain/role"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/domain/venue"
	"enterprise-crud/internal/infrastructure/database"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDatabase represents a test database connection
type TestDatabase struct {
	DB     *gorm.DB
	Config *TestConfig
}

// SetupTestDatabase creates a new test database connection and runs migrations
func SetupTestDatabase(t *testing.T) *TestDatabase {
	config := NewTestConfig()

	// Wait for database to be ready
	waitForDatabase(t, config)

	// Connect to database
	db, err := connectToDatabase(config)
	require.NoError(t, err, "Failed to connect to test database")

	// Run GORM auto-migrations for test tables
	runAutoMigrations(t, db)

	return &TestDatabase{
		DB:     db,
		Config: config,
	}
}

// CleanupTestDatabase cleans up the test database
func (td *TestDatabase) Cleanup(t *testing.T) {
	// Clean all tables in reverse order to handle foreign keys
	tables := []string{
		"orders",
		"events",
		"venues",
		"user_roles",
		"users",
		"roles",
	}

	for _, table := range tables {
		err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error
		if err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// Close closes the database connection
func (td *TestDatabase) Close() error {
	sqlDB, err := td.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// waitForDatabase waits for the database to be ready
func waitForDatabase(t *testing.T, config *TestConfig) {
	maxRetries := 30
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		db, err := sql.Open("postgres", config.DatabaseURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				db.Close()
				return
			}
			db.Close()
		}

		if i == maxRetries-1 {
			require.NoError(t, err, "Database not ready after %d retries", maxRetries)
		}

		t.Logf("Waiting for database to be ready... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(retryDelay)
	}
}

// connectToDatabase creates a GORM connection to the test database
func connectToDatabase(config *TestConfig) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	}

	return gorm.Open(postgres.Open(config.DatabaseURL), gormConfig)
}

// runAutoMigrations runs GORM auto-migrations for test entities
func runAutoMigrations(t *testing.T, db *gorm.DB) {
	// Import all domain entities for auto-migration
	err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	require.NoError(t, err, "Failed to create uuid extension")

	// Auto-migrate all tables in correct order
	err = db.AutoMigrate(
		&role.Role{},
		&user.User{},
		&venue.Venue{},
		&event.Event{},
		&order.Order{},
	)
	require.NoError(t, err, "Failed to run auto-migrations")

	t.Log("Database auto-migrations completed successfully")
}

// CreateTestConnection creates a database connection for testing (similar to production)
func CreateTestConnection(config *TestConfig) (*database.Connection, error) {
	db, err := connectToDatabase(config)
	if err != nil {
		return nil, err
	}

	return &database.Connection{DB: db}, nil
}
