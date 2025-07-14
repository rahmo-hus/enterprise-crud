package order_test

import (
	"context"
	"testing"
	"time"

	"enterprise-crud/internal/domain/order"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockOrderRepository is a mock implementation of order.Repository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, orderEntity *order.Order) error {
	args := m.Called(ctx, orderEntity)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, orderEntity *order.Order) error {
	args := m.Called(ctx, tx, orderEntity)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*order.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*order.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*order.Order, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).([]*order.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, orderEntity *order.Order) error {
	args := m.Called(ctx, orderEntity)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) GetEventWithTx(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) (*order.EventInfo, error) {
	args := m.Called(ctx, tx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.EventInfo), args.Error(1)
}

func (m *MockOrderRepository) UpdateEventTicketsWithTx(ctx context.Context, tx *gorm.DB, eventID uuid.UUID, newAvailableTickets int) error {
	args := m.Called(ctx, tx, eventID, newAvailableTickets)
	return args.Error(0)
}

// TestOrderService_CreateOrder_InvalidQuantity tests quantity validation
func TestOrderService_CreateOrder_InvalidQuantity(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil) // DB not used for validation

	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()
	quantity := 0

	// Act
	createdOrder, err := service.CreateOrder(ctx, userID, eventID, quantity)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, createdOrder)
	assert.True(t, order.IsInvalidQuantityError(err))
}

// TestOrderService_GetOrderByID_Success tests successful order retrieval
func TestOrderService_GetOrderByID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil)

	ctx := context.Background()
	orderID := uuid.New()

	expectedOrder := &order.Order{
		ID:          orderID,
		UserID:      uuid.New(),
		EventID:     uuid.New(),
		Quantity:    2,
		TotalAmount: 100.0,
		Status:      order.StatusPending,
		CreatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, orderID).Return(expectedOrder, nil)

	// Act
	foundOrder, err := service.GetOrderByID(ctx, orderID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, foundOrder)
	assert.Equal(t, expectedOrder.ID, foundOrder.ID)
	assert.Equal(t, expectedOrder.UserID, foundOrder.UserID)
	assert.Equal(t, expectedOrder.EventID, foundOrder.EventID)

	mockRepo.AssertExpectations(t)
}

// TestOrderService_GetOrderByID_NotFound tests order not found scenario
func TestOrderService_GetOrderByID_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil)

	ctx := context.Background()
	orderID := uuid.New()

	mockRepo.On("GetByID", ctx, orderID).Return(nil, order.NewOrderNotFoundError(orderID))

	// Act
	foundOrder, err := service.GetOrderByID(ctx, orderID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, foundOrder)
	assert.True(t, order.IsOrderNotFoundError(err))

	mockRepo.AssertExpectations(t)
}

// TestOrderService_GetOrdersByUserID_Success tests successful user orders retrieval
func TestOrderService_GetOrdersByUserID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil)

	ctx := context.Background()
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

	mockRepo.On("GetByUserID", ctx, userID).Return(expectedOrders, nil)

	// Act
	foundOrders, err := service.GetOrdersByUserID(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, foundOrders)
	assert.Len(t, foundOrders, 2)
	assert.Equal(t, expectedOrders[0].ID, foundOrders[0].ID)
	assert.Equal(t, expectedOrders[1].ID, foundOrders[1].ID)

	mockRepo.AssertExpectations(t)
}

// TestOrderService_UpdateOrderStatus_Success tests successful order status update
func TestOrderService_UpdateOrderStatus_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil)

	ctx := context.Background()
	orderID := uuid.New()
	newStatus := order.StatusCompleted

	existingOrder := &order.Order{
		ID:          orderID,
		UserID:      uuid.New(),
		EventID:     uuid.New(),
		Quantity:    2,
		TotalAmount: 100.0,
		Status:      order.StatusPending,
		CreatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, orderID).Return(existingOrder, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*order.Order")).Return(nil)

	// Act
	err := service.UpdateOrderStatus(ctx, orderID, newStatus)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestOrderService_UpdateOrderStatus_InvalidStatus tests invalid status validation
func TestOrderService_UpdateOrderStatus_InvalidStatus(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil)

	ctx := context.Background()
	orderID := uuid.New()
	invalidStatus := "INVALID_STATUS"

	// Act
	err := service.UpdateOrderStatus(ctx, orderID, invalidStatus)

	// Assert
	assert.Error(t, err)
	assert.True(t, order.IsValidationError(err))
}

// TestOrderService_DeleteOrder_Success tests successful order deletion
func TestOrderService_DeleteOrder_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil)

	ctx := context.Background()
	orderID := uuid.New()

	existingOrder := &order.Order{
		ID:          orderID,
		UserID:      uuid.New(),
		EventID:     uuid.New(),
		Quantity:    2,
		TotalAmount: 100.0,
		Status:      order.StatusPending,
		CreatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, orderID).Return(existingOrder, nil)
	mockRepo.On("Delete", ctx, orderID).Return(nil)

	// Act
	err := service.DeleteOrder(ctx, orderID)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestOrderService_DeleteOrder_NotFound tests order not found for deletion
func TestOrderService_DeleteOrder_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockOrderRepository)
	service := order.NewOrderService(mockRepo, nil)

	ctx := context.Background()
	orderID := uuid.New()

	mockRepo.On("GetByID", ctx, orderID).Return(nil, order.NewOrderNotFoundError(orderID))

	// Act
	err := service.DeleteOrder(ctx, orderID)

	// Assert
	assert.Error(t, err)
	assert.True(t, order.IsOrderNotFoundError(err))

	mockRepo.AssertExpectations(t)
}

// Note: Transaction-related tests (CreateOrder with business logic) are skipped
// because they require integration testing with a real database for GORM transactions
// These tests should be implemented in integration test files.
