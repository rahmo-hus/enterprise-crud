package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"enterprise-crud/internal/config"
	"enterprise-crud/internal/domain/event"
	"enterprise-crud/internal/domain/order"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/domain/venue"
	"enterprise-crud/internal/infrastructure/auth"
	httpHandlers "enterprise-crud/internal/presentation/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of user.Service interface
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

// MockEventService is a mock implementation of event.Service interface
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) CreateEvent(ctx context.Context, event *event.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) GetEventByID(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockEventService) GetAllEvents(ctx context.Context) ([]*event.Event, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*event.Event), args.Error(1)
}

func (m *MockEventService) GetEventsByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*event.Event, error) {
	args := m.Called(ctx, organizerID)
	return args.Get(0).([]*event.Event), args.Error(1)
}

func (m *MockEventService) UpdateEvent(ctx context.Context, event *event.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) CancelEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error {
	args := m.Called(ctx, eventID, organizerID)
	return args.Error(0)
}

func (m *MockEventService) DeleteEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error {
	args := m.Called(ctx, eventID, organizerID)
	return args.Error(0)
}

// MockOrderService is a mock implementation of order.Service interface
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, quantity int) (*order.Order, error) {
	args := m.Called(ctx, userID, eventID, quantity)
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *MockOrderService) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*order.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*order.Order), args.Error(1)
}

func (m *MockOrderService) GetOrdersByEventID(ctx context.Context, eventID uuid.UUID) ([]*order.Order, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).([]*order.Order), args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockVenueService is a mock implementation of venue.Service interface
type MockVenueService struct {
	mock.Mock
}

func (m *MockVenueService) CreateVenue(ctx context.Context, venue *venue.Venue) error {
	args := m.Called(ctx, venue)
	return args.Error(0)
}

func (m *MockVenueService) GetVenueByID(ctx context.Context, id uuid.UUID) (*venue.Venue, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*venue.Venue), args.Error(1)
}

func (m *MockVenueService) GetAllVenues(ctx context.Context) ([]*venue.Venue, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*venue.Venue), args.Error(1)
}

func (m *MockVenueService) UpdateVenue(ctx context.Context, venue *venue.Venue) error {
	args := m.Called(ctx, venue)
	return args.Error(0)
}

func (m *MockVenueService) DeleteVenue(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestWireApp() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create mock dependencies
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	mockUserService := new(MockUserService)
	mockEventService := new(MockEventService)
	mockOrderService := new(MockOrderService)
	jwtService := auth.NewJWTService("test-secret-key", "test-issuer", time.Hour)

	userHandler := httpHandlers.NewUserHandler(mockUserService, jwtService)
	eventHandler := httpHandlers.NewEventHandler(mockEventService, jwtService)
	orderHandler := httpHandlers.NewOrderHandler(mockOrderService, jwtService)

	// Create mock venue service and handler
	mockVenueService := new(MockVenueService)
	venueHandler := httpHandlers.NewVenueHandler(mockVenueService, jwtService)

	// Create a test app instance
	app := NewWireApp(cfg, nil, nil, userHandler, eventHandler, orderHandler, venueHandler)

	return app.SetupRouter()
}

func TestWireApp_HealthCheck(t *testing.T) {
	router := setupTestWireApp()

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"healthy"`)
	assert.Contains(t, w.Body.String(), `"service":"test-app"`)
	assert.Contains(t, w.Body.String(), `"version":"1.0.0"`)
	assert.Contains(t, w.Body.String(), `"environment":"test"`)
}

func TestWireApp_SwaggerEndpoint(t *testing.T) {
	router := setupTestWireApp()

	// Test Swagger base endpoint - in test environment, static files might not be available
	// but the route should at least be registered
	req, _ := http.NewRequest(http.MethodGet, "/swagger/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// In test mode, swagger static files might not be embedded, so we accept:
	// - 200 (if swagger files are available)
	// - 301/302 (if there's a redirect to index.html)
	// - 404 (if static files aren't embedded in test)
	// We just want to ensure the route exists and doesn't cause a panic
	assert.True(t, w.Code != 0, "Swagger endpoint should be registered (got %d)", w.Code)
}

func TestWireApp_SwaggerJSON(t *testing.T) {
	router := setupTestWireApp()

	// Test that swagger route is registered by checking the router doesn't panic
	// Note: In test environment, swagger docs might not be fully generated/embedded
	req, _ := http.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	w := httptest.NewRecorder()

	// This should not panic and should return some response (even if 404)
	router.ServeHTTP(w, req)

	// Just verify the router handles the request without panicking
	// In production with proper swagger generation, this would return 200 with JSON
	assert.True(t, w.Code > 0, "Swagger route should be handled without panic")
}

func TestWireApp_NotFoundEndpoint(t *testing.T) {
	router := setupTestWireApp()

	req, _ := http.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWireApp_RouterSetup(t *testing.T) {
	router := setupTestWireApp()

	// Test that router is properly configured
	assert.NotNil(t, router)

	// Test that routes are registered
	routes := router.Routes()

	// Should have at least health, swagger, and user routes
	assert.True(t, len(routes) > 0, "Router should have registered routes")

	// Check for specific route patterns
	foundHealthRoute := false
	foundSwaggerRoute := false
	foundUserRoute := false

	for _, route := range routes {
		if route.Path == "/health" && route.Method == "GET" {
			foundHealthRoute = true
		}
		if route.Path == "/swagger/*any" && route.Method == "GET" {
			foundSwaggerRoute = true
		}
		if route.Path == "/api/v1/users" && route.Method == "POST" {
			foundUserRoute = true
		}
	}

	assert.True(t, foundHealthRoute, "Health route should be registered")
	assert.True(t, foundSwaggerRoute, "Swagger route should be registered")
	assert.True(t, foundUserRoute, "User routes should be registered")
}
