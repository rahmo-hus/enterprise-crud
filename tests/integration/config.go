//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	DatabaseURL      string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
}

// NewTestConfig creates a new test configuration with defaults for Docker Compose setup
func NewTestConfig() *TestConfig {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5434")
	dbName := getEnvOrDefault("TEST_DB_NAME", "enterprise_crud_test")
	user := getEnvOrDefault("TEST_DB_USER", "test_user")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "test_password")

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName)

	return &TestConfig{
		DatabaseURL:      databaseURL,
		DatabaseHost:     host,
		DatabasePort:     port,
		DatabaseName:     dbName,
		DatabaseUser:     user,
		DatabasePassword: password,
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
