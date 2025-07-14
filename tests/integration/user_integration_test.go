//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-crud/internal/app"
	"enterprise-crud/internal/dto/user"
	"enterprise-crud/internal/infrastructure/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserIntegration(t *testing.T) {
	// Setup test database
	testDB := SetupTestDatabase(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	// Create test fixtures
	fixtures := NewTestFixtures(testDB)
	userRole, _, _ := fixtures.StandardRoles(t)

	// Create database connection for the app
	dbConn := &database.Connection{DB: testDB.DB}

	// Create the application dependencies
	cfg := CreateTestConfig()

	// Create dependencies
	deps, err := CreateTestDependencies(cfg, dbConn)
	require.NoError(t, err, "Failed to create dependencies")

	// Create the application instance
	application := app.NewWireApp(cfg, dbConn, deps.UserHandler, deps.EventHandler, deps.OrderHandler)

	// Create HTTP handler
	router := application.SetupRouter()

	t.Run("POST /users - Create User", func(t *testing.T) {
		// Prepare request payload
		createUserReq := user.CreateUserRequest{
			Email:    "newuser@test.com",
			Username: "newuser",
			Password: "password123",
		}

		payload, err := json.Marshal(createUserReq)
		require.NoError(t, err)

		// Create HTTP request
		req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusCreated, w.Code)

		var response user.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, createUserReq.Email, response.Email)
		assert.Equal(t, createUserReq.Username, response.Username)
		assert.NotEmpty(t, response.ID)
		assert.NotEmpty(t, response.Roles)
	})

	t.Run("GET /users/:email - Get User by Email (No Auth Required)", func(t *testing.T) {
		// Note: This route requires authentication in the actual API
		// For integration testing, we'll test that it returns 401 without auth

		// Create test user
		testUser := fixtures.CreateUser(t, "testuser@test.com", "testuser", "password123", userRole)

		// Create HTTP request without authorization
		req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", testUser.Email), nil)
		require.NoError(t, err)

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response - should be 401 without auth token
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GET /users/:email - Multiple Users Without Auth", func(t *testing.T) {
		// Create multiple test users with unique usernames
		user1 := fixtures.CreateUser(t, "user1@test.com", "user1_unique", "password123", userRole)
		user2 := fixtures.CreateUser(t, "user2@test.com", "user2_unique", "password123", userRole)

		// Test getting first user by email - should return 401 without auth
		req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", user1.Email), nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Test getting second user by email - should return 401 without auth
		req, err = http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", user2.Email), nil)
		require.NoError(t, err)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST /users - Create User with Invalid Data", func(t *testing.T) {
		// Test with invalid email
		createUserReq := user.CreateUserRequest{
			Email:    "invalid-email",
			Username: "testuser",
			Password: "short",
		}

		payload, err := json.Marshal(createUserReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /users - Create User with Duplicate Email", func(t *testing.T) {
		// Create initial user with unique username
		fixtures.CreateUser(t, "duplicate@test.com", "user1_duplicate", "password123", userRole)

		// Try to create another user with same email but different username
		createUserReq := user.CreateUserRequest{
			Email:    "duplicate@test.com",
			Username: "user2_duplicate",
			Password: "password123",
		}

		payload, err := json.Marshal(createUserReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}
