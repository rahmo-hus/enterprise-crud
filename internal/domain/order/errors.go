package order

import (
	"fmt"

	"github.com/google/uuid"
)

// Error types for order operations
type OrderError struct {
	Code    string
	Message string
	Err     error
}

func (e *OrderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *OrderError) Unwrap() error {
	return e.Err
}

// Error constants
const (
	OrderNotFoundErrorCode     = "ORDER_NOT_FOUND"
	EventNotFoundErrorCode     = "EVENT_NOT_FOUND"
	InsufficientTicketsErrorCode = "INSUFFICIENT_TICKETS"
	InvalidQuantityErrorCode   = "INVALID_QUANTITY"
	EventNotActiveErrorCode    = "EVENT_NOT_ACTIVE"
	ValidationErrorCode        = "VALIDATION_ERROR"
	OrderCreationErrorCode     = "ORDER_CREATION_ERROR"
	UnauthorizedErrorCode      = "UNAUTHORIZED"
)

// NewOrderNotFoundError creates a new order not found error
func NewOrderNotFoundError(id uuid.UUID) *OrderError {
	return &OrderError{
		Code:    OrderNotFoundErrorCode,
		Message: fmt.Sprintf("Order with ID %s not found", id),
	}
}

// NewEventNotFoundError creates a new event not found error
func NewEventNotFoundError(eventID uuid.UUID) *OrderError {
	return &OrderError{
		Code:    EventNotFoundErrorCode,
		Message: fmt.Sprintf("Event with ID %s not found", eventID),
	}
}

// NewInsufficientTicketsError creates a new insufficient tickets error
func NewInsufficientTicketsError(requested, available int) *OrderError {
	return &OrderError{
		Code:    InsufficientTicketsErrorCode,
		Message: fmt.Sprintf("Insufficient tickets: requested %d, available %d", requested, available),
	}
}

// NewInvalidQuantityError creates a new invalid quantity error
func NewInvalidQuantityError(quantity int) *OrderError {
	return &OrderError{
		Code:    InvalidQuantityErrorCode,
		Message: fmt.Sprintf("Invalid quantity: %d. Quantity must be greater than 0", quantity),
	}
}

// NewEventNotActiveError creates a new event not active error
func NewEventNotActiveError(eventID uuid.UUID, status string) *OrderError {
	return &OrderError{
		Code:    EventNotActiveErrorCode,
		Message: fmt.Sprintf("Event %s is not active (status: %s)", eventID, status),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *OrderError {
	return &OrderError{
		Code:    ValidationErrorCode,
		Message: message,
	}
}

// NewOrderCreationError creates a new order creation error
func NewOrderCreationError(err error) *OrderError {
	return &OrderError{
		Code:    OrderCreationErrorCode,
		Message: "Failed to create order",
		Err:     err,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *OrderError {
	return &OrderError{
		Code:    UnauthorizedErrorCode,
		Message: message,
	}
}

// Error type checking functions
func IsOrderNotFoundError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == OrderNotFoundErrorCode
	}
	return false
}

func IsEventNotFoundError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == EventNotFoundErrorCode
	}
	return false
}

func IsInsufficientTicketsError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == InsufficientTicketsErrorCode
	}
	return false
}

func IsInvalidQuantityError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == InvalidQuantityErrorCode
	}
	return false
}

func IsEventNotActiveError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == EventNotActiveErrorCode
	}
	return false
}

func IsValidationError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == ValidationErrorCode
	}
	return false
}

func IsOrderCreationError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == OrderCreationErrorCode
	}
	return false
}

func IsUnauthorizedError(err error) bool {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code == UnauthorizedErrorCode
	}
	return false
}

// GetOrderErrorCode extracts the error code from an order error
func GetOrderErrorCode(err error) string {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.Code
	}
	return "UNKNOWN_ERROR"
}