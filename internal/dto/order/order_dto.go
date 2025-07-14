package order

import (
	"time"

	"github.com/google/uuid"
)

// CreateOrderRequest represents the request structure for creating a new order
type CreateOrderRequest struct {
	EventID  uuid.UUID `json:"event_id" binding:"required"`
	Quantity int       `json:"quantity" binding:"required,min=1"`
}

// OrderResponse represents the response structure for order operations
type OrderResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	EventID     uuid.UUID `json:"event_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrderListResponse represents the response structure for listing orders
type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Count  int             `json:"count"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents success response structure
type SuccessResponse struct {
	Message string `json:"message"`
}
