package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"enterprise-crud/internal/config"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/infrastructure/auth"
	httpHandlers "enterprise-crud/internal/presentation/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "enterprise-crud/docs"
)

// MockUserService for integration testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, email, username, password string) (*user.User, error) {
	args := m.Called(ctx, email, username, password)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, email, password string) (*user.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*user.User), args.Error(1)
}

func setupTestServer() *httptest.Server {
	gin.SetMode(gin.TestMode)
	
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "enterprise-crud-test",
			Version:     "1.0.0",
			Environment: "test",
		},
	}
	
	// Create mock user service
	mockUserService := new(MockUserService)
	jwtService := auth.NewJWTService("test-secret-key", "test-issuer", time.Hour)
	userHandler := httpHandlers.NewUserHandler(mockUserService, jwtService)
	
	// Setup router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"service":     cfg.App.Name,
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
		})
	})
	
	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		userHandler.RegisterRoutes(v1)
	}
	
	return httptest.NewServer(router)
}

func testHealthEndpoint(serverURL string) {
	fmt.Println("🏥 Testing Health Endpoint...")
	
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		log.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Health endpoint returned status %d", resp.StatusCode)
	}
	
	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		log.Fatalf("Failed to decode health response: %v", err)
	}
	
	if health["status"] != "healthy" {
		log.Fatalf("Health endpoint returned unhealthy status: %v", health["status"])
	}
	
	fmt.Println("✅ Health endpoint test passed")
}

func testSwaggerEndpoints(serverURL string) {
	fmt.Println("📚 Testing Swagger Endpoints...")
	
	// Test Swagger JSON endpoint
	resp, err := http.Get(serverURL + "/swagger/doc.json")
	if err != nil {
		log.Fatalf("Failed to call swagger JSON endpoint: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Swagger JSON endpoint returned status %d", resp.StatusCode)
	}
	
	var swaggerDoc map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&swaggerDoc); err != nil {
		log.Fatalf("Failed to decode swagger JSON: %v", err)
	}
	
	// Verify essential Swagger fields
	if swaggerDoc["swagger"] == nil {
		log.Fatalf("Swagger document missing 'swagger' field")
	}
	
	if swaggerDoc["info"] == nil {
		log.Fatalf("Swagger document missing 'info' field")
	}
	
	if swaggerDoc["paths"] == nil {
		log.Fatalf("Swagger document missing 'paths' field")
	}
	
	// Check for our API paths
	paths := swaggerDoc["paths"].(map[string]interface{})
	if paths["/api/v1/users"] == nil {
		log.Fatalf("Swagger document missing '/api/v1/users' path")
	}
	
	if paths["/health"] == nil {
		log.Fatalf("Swagger document missing '/health' path")
	}
	
	fmt.Println("✅ Swagger JSON endpoint test passed")
	
	// Test Swagger UI endpoint
	resp, err = http.Get(serverURL + "/swagger/index.html")
	if err != nil {
		log.Fatalf("Failed to call swagger UI endpoint: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Swagger UI endpoint returned status %d", resp.StatusCode)
	}
	
	fmt.Println("✅ Swagger UI endpoint test passed")
}

func testNotFoundEndpoint(serverURL string) {
	fmt.Println("🔍 Testing Not Found Endpoint...")
	
	resp, err := http.Get(serverURL + "/nonexistent")
	if err != nil {
		log.Fatalf("Failed to call nonexistent endpoint: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNotFound {
		log.Fatalf("Nonexistent endpoint should return 404, got %d", resp.StatusCode)
	}
	
	fmt.Println("✅ Not found endpoint test passed")
}

func main() {
	fmt.Println("🚀 Starting Swagger Integration Tests")
	fmt.Println("=====================================")
	
	// Setup test server
	server := setupTestServer()
	defer server.Close()
	
	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	
	// Run tests
	testHealthEndpoint(server.URL)
	testSwaggerEndpoints(server.URL)
	testNotFoundEndpoint(server.URL)
	
	fmt.Println("")
	fmt.Println("🎉 All Swagger integration tests passed!")
	fmt.Println("")
	fmt.Println("📋 Test Results:")
	fmt.Println("===============")
	fmt.Println("✅ Health endpoint accessible")
	fmt.Println("✅ Swagger JSON endpoint accessible")
	fmt.Println("✅ Swagger UI endpoint accessible")
	fmt.Println("✅ API documentation properly generated")
	fmt.Println("✅ Not found handling works correctly")
	fmt.Println("")
	fmt.Println("🔗 Test server was running at:", server.URL)
	fmt.Println("📚 Swagger UI would be available at:", server.URL+"/swagger/index.html")
	fmt.Println("📄 Swagger JSON would be available at:", server.URL+"/swagger/doc.json")
}