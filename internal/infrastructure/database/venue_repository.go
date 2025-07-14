package database

import (
	"context"
	"enterprise-crud/internal/domain/venue"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// venueRepository implements the venue.Repository interface
type venueRepository struct {
	db *gorm.DB
}

// NewVenueRepository creates a new venue repository instance
func NewVenueRepository(db *gorm.DB) venue.Repository {
	return &venueRepository{db: db}
}

// Create creates a new venue in the database
func (r *venueRepository) Create(ctx context.Context, v *venue.Venue) error {
	if err := r.db.WithContext(ctx).Create(v).Error; err != nil {
		return venue.NewVenueError(venue.ErrVenueCreationFailed, err)
	}
	return nil
}

// GetByID retrieves a venue by its ID
func (r *venueRepository) GetByID(ctx context.Context, id uuid.UUID) (*venue.Venue, error) {
	var v venue.Venue
	if err := r.db.WithContext(ctx).First(&v, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, venue.NewVenueNotFoundError(id)
		}
		return nil, venue.NewVenueError(venue.ErrVenueRetrievalFailed, err)
	}
	return &v, nil
}

// GetAll retrieves all venues
func (r *venueRepository) GetAll(ctx context.Context) ([]*venue.Venue, error) {
	var venues []*venue.Venue
	if err := r.db.WithContext(ctx).Find(&venues).Error; err != nil {
		return nil, venue.NewVenueError(venue.ErrVenueRetrievalFailed, err)
	}
	return venues, nil
}

// Update updates an existing venue
func (r *venueRepository) Update(ctx context.Context, v *venue.Venue) error {
	if err := r.db.WithContext(ctx).Save(v).Error; err != nil {
		return venue.NewVenueError(venue.ErrVenueUpdateFailed, err)
	}
	return nil
}

// Delete deletes a venue by its ID
func (r *venueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&venue.Venue{}, id)
	if result.Error != nil {
		return venue.NewVenueError(venue.ErrVenueDeletionFailed, result.Error)
	}
	if result.RowsAffected == 0 {
		return venue.NewVenueNotFoundError(id)
	}
	return nil
}
