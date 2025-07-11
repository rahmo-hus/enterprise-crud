package event

import (
	"time"

	"github.com/google/uuid"
)

// Event represents an event that can be attended
type Event struct {
	// ID is the unique identifier for each event
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`

	// VenueID is the ID of the venue where the event takes place
	VenueID uuid.UUID `gorm:"not null;type:uuid" json:"venue_id" binding:"required"`

	// OrganizerID is the ID of the user who organized the event
	OrganizerID uuid.UUID `gorm:"not null;type:uuid" json:"organizer_id"`

	// Title is the event title
	Title string `gorm:"not null;size:255" json:"title" binding:"required"`

	// Description provides additional information about the event
	Description string `gorm:"type:text" json:"description"`

	// EventDate is when the event takes place
	EventDate time.Time `gorm:"not null" json:"event_date" binding:"required"`

	// TicketPrice is the price per ticket
	TicketPrice float64 `gorm:"not null;type:decimal(10,2);check:ticket_price >= 0" json:"ticket_price" binding:"required,min=0"`

	// AvailableTickets is the number of tickets still available
	AvailableTickets int `gorm:"not null;check:available_tickets >= 0" json:"available_tickets"`

	// TotalTickets is the total number of tickets for the event
	TotalTickets int `gorm:"not null;check:total_tickets > 0" json:"total_tickets" binding:"required,min=1"`

	// Status indicates the current state of the event
	Status string `gorm:"default:'ACTIVE';size:20;check:status IN ('ACTIVE', 'CANCELLED', 'COMPLETED')" json:"status"`

	// Timestamps track when the event was created and last updated
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event status constants
const (
	StatusActive    = "ACTIVE"
	StatusCancelled = "CANCELLED"
	StatusCompleted = "COMPLETED"
)

// TableName tells GORM what table to use for this model
func (Event) TableName() string {
	return "events"
}

// IsActive checks if the event is active
func (e *Event) IsActive() bool {
	return e.Status == StatusActive
}

// IsCancelled checks if the event is cancelled
func (e *Event) IsCancelled() bool {
	return e.Status == StatusCancelled
}

// IsCompleted checks if the event is completed
func (e *Event) IsCompleted() bool {
	return e.Status == StatusCompleted
}

// HasAvailableTickets checks if there are tickets available
func (e *Event) HasAvailableTickets() bool {
	return e.AvailableTickets > 0
}

// CanSellTickets checks if tickets can be sold for this event
func (e *Event) CanSellTickets() bool {
	return e.IsActive() && e.HasAvailableTickets()
}
