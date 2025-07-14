package venue

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the contract for venue business logic
type Service interface {
	CreateVenue(ctx context.Context, venue *Venue) error
	GetVenueByID(ctx context.Context, id uuid.UUID) (*Venue, error)
	GetAllVenues(ctx context.Context) ([]*Venue, error)
	UpdateVenue(ctx context.Context, venue *Venue) error
	DeleteVenue(ctx context.Context, id uuid.UUID) error
}

// VenueService implements the venue service interface
type VenueService struct {
	repository Repository
}

// NewVenueService creates a new instance of venue service
func NewVenueService(repository Repository) Service {
	return &VenueService{
		repository: repository,
	}
}

// CreateVenue creates a new venue
func (s *VenueService) CreateVenue(ctx context.Context, venue *Venue) error {
	// Validate venue data
	if err := s.validateVenue(venue); err != nil {
		return err
	}

	// Create the venue
	return s.repository.Create(ctx, venue)
}

// GetVenueByID retrieves a venue by its ID
func (s *VenueService) GetVenueByID(ctx context.Context, id uuid.UUID) (*Venue, error) {
	return s.repository.GetByID(ctx, id)
}

// GetAllVenues retrieves all venues
func (s *VenueService) GetAllVenues(ctx context.Context) ([]*Venue, error) {
	return s.repository.GetAll(ctx)
}

// UpdateVenue updates an existing venue
func (s *VenueService) UpdateVenue(ctx context.Context, venue *Venue) error {
	// Validate venue data
	if err := s.validateVenue(venue); err != nil {
		return err
	}

	// Check if venue exists
	_, err := s.repository.GetByID(ctx, venue.ID)
	if err != nil {
		return err
	}

	// Update the venue
	return s.repository.Update(ctx, venue)
}

// DeleteVenue deletes a venue by its ID
func (s *VenueService) DeleteVenue(ctx context.Context, id uuid.UUID) error {
	// Check if venue exists
	_, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

// validateVenue validates venue data
func (s *VenueService) validateVenue(venue *Venue) error {
	if venue.Capacity <= 0 {
		return ErrInvalidVenueCapacity
	}
	
	// Add more validation rules as needed
	return nil
}