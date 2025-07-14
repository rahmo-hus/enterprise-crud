package database

import (
	"context"
	"errors"

	"enterprise-crud/internal/domain/event"
	"enterprise-crud/internal/domain/order"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderRepository implements the order repository interface
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository instance
func NewOrderRepository(db *gorm.DB) order.Repository {
	return &OrderRepository{db: db}
}

// Create creates a new order in the database
func (r *OrderRepository) Create(ctx context.Context, orderEntity *order.Order) error {
	if err := r.db.WithContext(ctx).Create(orderEntity).Error; err != nil {
		return err
	}
	return nil
}

// CreateWithTx creates a new order within a transaction
func (r *OrderRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, orderEntity *order.Order) error {
	if err := tx.WithContext(ctx).Create(orderEntity).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves an order by its ID
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	var orderEntity order.Order
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&orderEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, order.NewOrderNotFoundError(id)
		}
		return nil, err
	}
	return &orderEntity, nil
}

// GetByUserID retrieves all orders for a specific user
func (r *OrderRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*order.Order, error) {
	var orders []*order.Order
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetByEventID retrieves all orders for a specific event
func (r *OrderRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*order.Order, error) {
	var orders []*order.Order
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// Update updates an existing order
func (r *OrderRepository) Update(ctx context.Context, orderEntity *order.Order) error {
	if err := r.db.WithContext(ctx).Save(orderEntity).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes an order by its ID
func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&order.Order{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return order.NewOrderNotFoundError(id)
	}
	return nil
}

// GetEventWithTx retrieves event information within a transaction
func (r *OrderRepository) GetEventWithTx(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) (*order.EventInfo, error) {
	var eventEntity event.Event
	if err := tx.WithContext(ctx).Where("id = ?", eventID).First(&eventEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, order.NewEventNotFoundError(eventID)
		}
		return nil, err
	}

	return &order.EventInfo{
		ID:               eventEntity.ID,
		Title:            eventEntity.Title,
		TicketPrice:      eventEntity.TicketPrice,
		AvailableTickets: eventEntity.AvailableTickets,
		TotalTickets:     eventEntity.TotalTickets,
		Status:           eventEntity.Status,
	}, nil
}

// UpdateEventTicketsWithTx updates event available tickets within a transaction
func (r *OrderRepository) UpdateEventTicketsWithTx(ctx context.Context, tx *gorm.DB, eventID uuid.UUID, newAvailableTickets int) error {
	result := tx.WithContext(ctx).Model(&event.Event{}).
		Where("id = ?", eventID).
		Update("available_tickets", newAvailableTickets)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return order.NewEventNotFoundError(eventID)
	}

	return nil
}
