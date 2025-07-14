package order

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the contract for order data access
type Repository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByEventID(ctx context.Context, eventID uuid.UUID) ([]*Order, error)

	// Transaction methods
	CreateWithTx(ctx context.Context, tx *gorm.DB, order *Order) error
	GetEventWithTx(ctx context.Context, tx *gorm.DB, eventID uuid.UUID) (*EventInfo, error)
	UpdateEventTicketsWithTx(ctx context.Context, tx *gorm.DB, eventID uuid.UUID, newAvailableTickets int) error
}

// EventInfo represents event information needed for order processing
type EventInfo struct {
	ID               uuid.UUID
	Title            string
	TicketPrice      float64
	AvailableTickets int
	TotalTickets     int
	Status           string
}
