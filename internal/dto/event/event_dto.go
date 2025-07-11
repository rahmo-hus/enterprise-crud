package event

import (
	"time"

	"github.com/google/uuid"
)

// CreateEventRequest represents the request to create a new event
type CreateEventRequest struct {
	VenueID      uuid.UUID `json:"venue_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title        string    `json:"title" binding:"required" example:"Summer Concert"`
	Description  string    `json:"description" example:"An amazing summer concert with live music"`
	EventDate    time.Time `json:"event_date" binding:"required" example:"2024-08-15T20:00:00Z"`
	TicketPrice  float64   `json:"ticket_price" binding:"required,min=0" example:"50.00"`
	TotalTickets int       `json:"total_tickets" binding:"required,min=1" example:"100"`
}

// UpdateEventRequest represents the request to update an existing event
type UpdateEventRequest struct {
	VenueID      uuid.UUID `json:"venue_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title        string    `json:"title" binding:"required" example:"Summer Concert - Updated"`
	Description  string    `json:"description" example:"An amazing summer concert with live music - Updated"`
	EventDate    time.Time `json:"event_date" binding:"required" example:"2024-08-15T20:00:00Z"`
	TicketPrice  float64   `json:"ticket_price" binding:"required,min=0" example:"60.00"`
	TotalTickets int       `json:"total_tickets" binding:"required,min=1" example:"150"`
}

// EventResponse represents the response when returning event data
type EventResponse struct {
	ID               uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	VenueID          uuid.UUID `json:"venue_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizerID      uuid.UUID `json:"organizer_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title            string    `json:"title" example:"Summer Concert"`
	Description      string    `json:"description" example:"An amazing summer concert with live music"`
	EventDate        time.Time `json:"event_date" example:"2024-08-15T20:00:00Z"`
	TicketPrice      float64   `json:"ticket_price" example:"50.00"`
	AvailableTickets int       `json:"available_tickets" example:"75"`
	TotalTickets     int       `json:"total_tickets" example:"100"`
	Status           string    `json:"status" example:"ACTIVE"`
	CreatedAt        time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt        time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// EventListResponse represents the response when returning a list of events
type EventListResponse struct {
	Events []EventResponse `json:"events"`
	Count  int             `json:"count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"validation_error"`
	Message string `json:"message" example:"Invalid input data"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message" example:"Event created successfully"`
}
