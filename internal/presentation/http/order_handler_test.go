package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"enterprise-crud/internal/domain/order"
	orderDto "enterprise-crud/internal/dto/order"
	"enterprise-crud/internal/infrastructure/auth"
	httpHandlers "enterprise-crud/internal/presentation/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderService is a mock implementation of order.Service
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

func setupOrderHandlerTest() (*gin.Engine, *MockOrderService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(MockOrderService)
	mockJWTService := &auth.JWTService{} // Mock JWT service

	handler := httpHandlers.NewOrderHandler(mockService, mockJWTService)

	// Add a test route with auth middleware mock
	router.POST("/orders", func(c *gin.Context) {
		// Mock user authentication
		claims := &auth.JWTClaims{
			UserID: uuid.New(),
			Roles:  []string{"USER"},
		}
		c.Set("user", claims)
		c.Next()
	}, handler.CreateOrder)

	router.GET("/orders/:id", func(c *gin.Context) {
		// Mock user authentication
		claims := &auth.JWTClaims{
			UserID: uuid.New(),
			Roles:  []string{"USER"},
		}
		c.Set("user", claims)
		c.Next()
	}, handler.GetOrder)

	router.GET("/orders/my-orders", func(c *gin.Context) {
		// Mock user authentication
		claims := &auth.JWTClaims{
			UserID: uuid.New(),
			Roles:  []string{"USER"},
		}
		c.Set("user", claims)
		c.Next()
	}, handler.GetMyOrders)

	return router, mockService
}

func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	eventID := uuid.New()
	userID := uuid.New()

	requestBody := orderDto.CreateOrderRequest{
		EventID:  eventID,
		Quantity: 2,
	}

	expectedOrder := &order.Order{
		ID:          uuid.New(),
		UserID:      userID,
		EventID:     eventID,
		Quantity:    2,
		TotalAmount: 100.0,
		Status:      order.StatusPending,
		CreatedAt:   time.Now(),
	}

	mockService.On("CreateOrder", mock.Anything, mock.AnythingOfType("uuid.UUID"), eventID, 2).Return(expectedOrder, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response orderDto.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder.ID, response.ID)
	assert.Equal(t, expectedOrder.EventID, response.EventID)
	assert.Equal(t, expectedOrder.Quantity, response.Quantity)
	assert.Equal(t, expectedOrder.TotalAmount, response.TotalAmount)
	assert.Equal(t, expectedOrder.Status, response.Status)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_ValidationError(t *testing.T) {
	// Arrange
	router, _ := setupOrderHandlerTest()

	// Invalid request body (missing event_id)
	requestBody := map[string]interface{}{
		"quantity": 2,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response orderDto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_error", response.Error)
}

func TestOrderHandler_CreateOrder_InvalidQuantity(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	eventID := uuid.New()

	requestBody := orderDto.CreateOrderRequest{
		EventID:  eventID,
		Quantity: 1, // Valid quantity for JSON binding
	}

	mockService.On("CreateOrder", mock.Anything, mock.AnythingOfType("uuid.UUID"), eventID, 1).Return((*order.Order)(nil), order.NewInvalidQuantityError(1))

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response orderDto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, order.InvalidQuantityErrorCode, response.Error)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_EventNotFound(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	eventID := uuid.New()

	requestBody := orderDto.CreateOrderRequest{
		EventID:  eventID,
		Quantity: 2,
	}

	mockService.On("CreateOrder", mock.Anything, mock.AnythingOfType("uuid.UUID"), eventID, 2).Return((*order.Order)(nil), order.NewEventNotFoundError(eventID))

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response orderDto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, order.EventNotFoundErrorCode, response.Error)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_InsufficientTickets(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	eventID := uuid.New()

	requestBody := orderDto.CreateOrderRequest{
		EventID:  eventID,
		Quantity: 10,
	}

	mockService.On("CreateOrder", mock.Anything, mock.AnythingOfType("uuid.UUID"), eventID, 10).Return((*order.Order)(nil), order.NewInsufficientTicketsError(10, 5))

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response orderDto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, order.InsufficientTicketsErrorCode, response.Error)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_Success(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	orderID := uuid.New()
	userID := uuid.New()

	expectedOrder := &order.Order{
		ID:          orderID,
		UserID:      userID,
		EventID:     uuid.New(),
		Quantity:    2,
		TotalAmount: 100.0,
		Status:      order.StatusPending,
		CreatedAt:   time.Now(),
	}

	mockService.On("GetOrderByID", mock.Anything, orderID).Return(expectedOrder, nil)

	// Update the router to use the same user ID
	router = gin.New()
	router.GET("/orders/:id", func(c *gin.Context) {
		claims := &auth.JWTClaims{
			UserID: userID, // Same user ID as the order
			Roles:  []string{"USER"},
		}
		c.Set("user", claims)
		c.Next()
	}, httpHandlers.NewOrderHandler(mockService, &auth.JWTService{}).GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response orderDto.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder.ID, response.ID)
	assert.Equal(t, expectedOrder.UserID, response.UserID)
	assert.Equal(t, expectedOrder.EventID, response.EventID)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	orderID := uuid.New()

	mockService.On("GetOrderByID", mock.Anything, orderID).Return((*order.Order)(nil), order.NewOrderNotFoundError(orderID))

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response orderDto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, order.OrderNotFoundErrorCode, response.Error)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_Forbidden(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	orderID := uuid.New()
	orderUserID := uuid.New()
	requestUserID := uuid.New() // Different user

	expectedOrder := &order.Order{
		ID:          orderID,
		UserID:      orderUserID,
		EventID:     uuid.New(),
		Quantity:    2,
		TotalAmount: 100.0,
		Status:      order.StatusPending,
		CreatedAt:   time.Now(),
	}

	mockService.On("GetOrderByID", mock.Anything, orderID).Return(expectedOrder, nil)

	// Update the router to use different user ID
	router = gin.New()
	router.GET("/orders/:id", func(c *gin.Context) {
		claims := &auth.JWTClaims{
			UserID: requestUserID, // Different user ID
			Roles:  []string{"USER"},
		}
		c.Set("user", claims)
		c.Next()
	}, httpHandlers.NewOrderHandler(mockService, &auth.JWTService{}).GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response orderDto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "forbidden", response.Error)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetMyOrders_Success(t *testing.T) {
	// Arrange
	router, mockService := setupOrderHandlerTest()

	userID := uuid.New()

	expectedOrders := []*order.Order{
		{
			ID:          uuid.New(),
			UserID:      userID,
			EventID:     uuid.New(),
			Quantity:    2,
			TotalAmount: 100.0,
			Status:      order.StatusPending,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			EventID:     uuid.New(),
			Quantity:    1,
			TotalAmount: 50.0,
			Status:      order.StatusCompleted,
			CreatedAt:   time.Now(),
		},
	}

	mockService.On("GetOrdersByUserID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(expectedOrders, nil)

	req := httptest.NewRequest(http.MethodGet, "/orders/my-orders", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response orderDto.OrderListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Orders, 2)
	assert.Equal(t, 2, response.Count)
	assert.Equal(t, expectedOrders[0].ID, response.Orders[0].ID)
	assert.Equal(t, expectedOrders[1].ID, response.Orders[1].ID)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_InvalidID(t *testing.T) {
	// Arrange
	router, _ := setupOrderHandlerTest()

	invalidID := "invalid-uuid"
	req := httptest.NewRequest(http.MethodGet, "/orders/"+invalidID, nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response orderDto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_id", response.Error)
}
