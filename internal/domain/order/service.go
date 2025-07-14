package order

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service defines the contract for order business logic
type Service interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, quantity int) (*Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	GetOrdersByEventID(ctx context.Context, eventID uuid.UUID) ([]*Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error
	DeleteOrder(ctx context.Context, id uuid.UUID) error
}

// OrderService implements the order service interface
type OrderService struct {
	repository Repository
	db         *gorm.DB
}

// NewOrderService creates a new instance of order service
func NewOrderService(repository Repository, db *gorm.DB) Service {
	return &OrderService{
		repository: repository,
		db:         db,
	}
}

// CreateOrder creates a new order with transaction support
func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, quantity int) (*Order, error) {
	// Validate input
	if quantity <= 0 {
		return nil, NewInvalidQuantityError(quantity)
	}

	var createdOrder *Order
	var err error

	// Execute within transaction to ensure atomicity
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get event information within transaction
		eventInfo, err := s.repository.GetEventWithTx(ctx, tx, eventID)
		if err != nil {
			return err
		}

		// Validate event is active
		if eventInfo.Status != "ACTIVE" {
			return NewEventNotActiveError(eventID, eventInfo.Status)
		}

		// Check if sufficient tickets are available
		if eventInfo.AvailableTickets < quantity {
			return NewInsufficientTicketsError(quantity, eventInfo.AvailableTickets)
		}

		// Calculate total amount
		totalAmount := eventInfo.TicketPrice * float64(quantity)

		// Create order entity
		newOrder := &Order{
			ID:          uuid.New(),
			UserID:      userID,
			EventID:     eventID,
			Quantity:    quantity,
			TotalAmount: totalAmount,
			Status:      StatusPending,
			CreatedAt:   time.Now(),
		}

		// Create order within transaction
		if err := s.repository.CreateWithTx(ctx, tx, newOrder); err != nil {
			return NewOrderCreationError(err)
		}

		// Update event available tickets within transaction
		newAvailableTickets := eventInfo.AvailableTickets - quantity
		if err := s.repository.UpdateEventTicketsWithTx(ctx, tx, eventID, newAvailableTickets); err != nil {
			return NewOrderCreationError(err)
		}

		createdOrder = newOrder
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}

// GetOrderByID retrieves an order by its ID
func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	return s.repository.GetByID(ctx, id)
}

// GetOrdersByUserID retrieves all orders for a specific user
func (s *OrderService) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error) {
	return s.repository.GetByUserID(ctx, userID)
}

// GetOrdersByEventID retrieves all orders for a specific event
func (s *OrderService) GetOrdersByEventID(ctx context.Context, eventID uuid.UUID) ([]*Order, error) {
	return s.repository.GetByEventID(ctx, eventID)
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	// Validate status
	if !isValidStatus(status) {
		return NewValidationError("Invalid order status: " + status)
	}

	// Get existing order
	existingOrder, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update status
	existingOrder.Status = status
	return s.repository.Update(ctx, existingOrder)
}

// DeleteOrder deletes an order
func (s *OrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	// Check if order exists
	_, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

// isValidStatus checks if the provided status is valid
func isValidStatus(status string) bool {
	validStatuses := []string{StatusPending, StatusCompleted, StatusFailed}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}