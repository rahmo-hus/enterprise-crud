package event

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for event data operations
type Repository interface {
	// Create creates a new event
	Create(ctx context.Context, event *Event) error

	// GetByID retrieves an event by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)

	// GetAll retrieves all events
	GetAll(ctx context.Context) ([]*Event, error)

	// GetByOrganizer retrieves events by organizer ID
	GetByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*Event, error)

	// GetByVenue retrieves events by venue ID
	GetByVenue(ctx context.Context, venueID uuid.UUID) ([]*Event, error)

	// Update updates an existing event
	Update(ctx context.Context, event *Event) error

	// Delete deletes an event by its ID
	Delete(ctx context.Context, id uuid.UUID) error
}
