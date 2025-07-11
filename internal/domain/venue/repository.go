package venue

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for venue data operations
type Repository interface {
	// Create creates a new venue
	Create(ctx context.Context, venue *Venue) error

	// GetByID retrieves a venue by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Venue, error)

	// GetAll retrieves all venues
	GetAll(ctx context.Context) ([]*Venue, error)

	// Update updates an existing venue
	Update(ctx context.Context, venue *Venue) error

	// Delete deletes a venue by its ID
	Delete(ctx context.Context, id uuid.UUID) error
}
