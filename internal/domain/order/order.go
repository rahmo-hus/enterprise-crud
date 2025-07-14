package order

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a ticket purchase order
type Order struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID      uuid.UUID `gorm:"not null;type:uuid" json:"user_id"`
	EventID     uuid.UUID `gorm:"not null;type:uuid" json:"event_id"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	TotalAmount float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      string    `gorm:"size:20;not null;default:'PENDING'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// Order status constants
const (
	StatusPending   = "PENDING"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
)

// TableName tells GORM what table to use for this model
func (Order) TableName() string {
	return "orders"
}

// IsPending checks if the order is pending
func (o *Order) IsPending() bool {
	return o.Status == StatusPending
}

// IsCompleted checks if the order is completed
func (o *Order) IsCompleted() bool {
	return o.Status == StatusCompleted
}

// IsFailed checks if the order has failed
func (o *Order) IsFailed() bool {
	return o.Status == StatusFailed
}
