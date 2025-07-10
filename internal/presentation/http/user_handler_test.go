package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-crud/internal/domain/user"
	userDTO "enterprise-crud/internal/dto/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of user.Service interface
// Used for testing HTTP handlers without service layer dependencies
type MockUserService struct {
	mock.Mock
}

// CreateUser mocks the CreateUser method of Service interface
// Returns user and error based on test scenario configuration
func (m *MockUserService) CreateUser(ctx context.Context, email, username, password string) (*user.User, error) {
	args := m.Called(ctx, email, username, password)
	return args.Get(0).(*user.User), args.Error(1)
}

// GetUserByEmail mocks the GetUserByEmail method of Service interface
// Returns user and error based on test scenario configuration
func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*user.User), args.Error(1)
}

// setupTestRouter creates a test Gin router with user routes
// Returns configured router for testing HTTP endpoints
func setupTestRouter(userService user.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create handler and register routes
	userHandler := NewUserHandler(userService)
	v1 := router.Group("/api/v1")
	userHandler.RegisterRoutes(v1)
	
	return router
}

// TestUserHandler_CreateUser tests the CreateUser HTTP handler
// Covers successful creation, validation errors, and service errors
func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string                    // Test case name
		requestBody    interface{}               // Request body to send
		mockFunc       func(*MockUserService)    // Mock service setup function
		expectedStatus int                       // Expected HTTP status code
		expectedBody   string                    // Expected response body content
	}{
		{
			name: "successful user creation",
			requestBody: userDTO.CreateUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "password123",
			},
			mockFunc: func(m *MockUserService) {
				// Mock CreateUser to return successful user creation
				createdUser := &user.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
				}
				m.On("CreateUser", mock.Anything, "test@example.com", "testuser", "password123").Return(createdUser, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"email":"test@example.com"`,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"email": "invalid-email", // Invalid email format
			},
			mockFunc: func(m *MockUserService) {
				// No mock expectations needed for validation errors
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error":"Invalid request"`,
		},
		{
			name: "user already exists",
			requestBody: userDTO.CreateUserRequest{
				Email:    "existing@example.com",
				Username: "existinguser",
				Password: "password123",
			},
			mockFunc: func(m *MockUserService) {
				// Mock CreateUser to return user already exists error
				m.On("CreateUser", mock.Anything, "existing@example.com", "existinguser", "password123").Return((*user.User)(nil), errors.New("user with email existing@example.com already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `"error":"User already exists"`,
		},
		{
			name: "internal server error",
			requestBody: userDTO.CreateUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "password123",
			},
			mockFunc: func(m *MockUserService) {
				// Mock CreateUser to return internal error
				m.On("CreateUser", mock.Anything, "test@example.com", "testuser", "password123").Return((*user.User)(nil), errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error":"Failed to create user"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(MockUserService)
			tt.mockFunc(mockService)

			// Create test router
			router := setupTestRouter(mockService)

			// Prepare request body
			bodyBytes, _ := json.Marshal(tt.requestBody)
			
			// Create HTTP request
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// TestUserHandler_GetUserByEmail tests the GetUserByEmail HTTP handler
// Covers successful retrieval, missing parameter, and not found scenarios
func TestUserHandler_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name           string                    // Test case name
		email          string                    // Email parameter in URL
		mockFunc       func(*MockUserService)    // Mock service setup function
		expectedStatus int                       // Expected HTTP status code
		expectedBody   string                    // Expected response body content
	}{
		{
			name:  "successful user retrieval",
			email: "test@example.com",
			mockFunc: func(m *MockUserService) {
				// Mock GetUserByEmail to return user
				foundUser := &user.User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
				}
				m.On("GetUserByEmail", mock.Anything, "test@example.com").Return(foundUser, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"email":"test@example.com"`,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			mockFunc: func(m *MockUserService) {
				// Mock GetUserByEmail to return not found error
				m.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return((*user.User)(nil), errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"error":"User not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(MockUserService)
			tt.mockFunc(mockService)

			// Create test router
			router := setupTestRouter(mockService)

			// Create HTTP request
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/"+tt.email, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// TestUserHandler_GetUserByEmail_EmptyEmail tests empty email parameter
// Covers validation of required email parameter
func TestUserHandler_GetUserByEmail_EmptyEmail(t *testing.T) {
	// Setup mock service (no expectations needed)
	mockService := new(MockUserService)

	// Create test router
	router := setupTestRouter(mockService)

	// Create HTTP request with empty email parameter
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response - should return 404 for route not found
	assert.Equal(t, http.StatusNotFound, w.Code)
}