package event

import (
	"context"
	"enterprise-crud/internal/domain/venue"
	"time"

	"github.com/google/uuid"
)

// Service defines the business logic interface for event operations
type Service interface {
	// CreateEvent creates a new event
	CreateEvent(ctx context.Context, event *Event) error

	// GetEventByID retrieves an event by its ID
	GetEventByID(ctx context.Context, id uuid.UUID) (*Event, error)

	// GetAllEvents retrieves all events
	GetAllEvents(ctx context.Context) ([]*Event, error)

	// GetEventsByOrganizer retrieves events by organizer ID
	GetEventsByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*Event, error)

	// UpdateEvent updates an existing event
	UpdateEvent(ctx context.Context, event *Event) error

	// CancelEvent cancels an event
	CancelEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error

	// DeleteEvent deletes an event (only if no tickets sold)
	DeleteEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error
}

// serviceImpl implements the Service interface
type serviceImpl struct {
	eventRepo Repository
	venueRepo venue.Repository
}

// NewService creates a new event service instance
func NewService(eventRepo Repository, venueRepo venue.Repository) Service {
	return &serviceImpl{
		eventRepo: eventRepo,
		venueRepo: venueRepo,
	}
}

// CreateEvent creates a new event
func (s *serviceImpl) CreateEvent(ctx context.Context, event *Event) error {
	// Validate event data
	if err := s.validateEvent(ctx, event); err != nil {
		return err
	}

	// Set default values
	event.Status = StatusActive
	event.AvailableTickets = event.TotalTickets

	// Create the event
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return err // Repository already returns custom error
	}

	return nil
}

// GetEventByID retrieves an event by its ID
func (s *serviceImpl) GetEventByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err // Repository already returns custom error
	}
	return event, nil
}

// GetAllEvents retrieves all events
func (s *serviceImpl) GetAllEvents(ctx context.Context) ([]*Event, error) {
	events, err := s.eventRepo.GetAll(ctx)
	if err != nil {
		return nil, err // Repository already returns custom error
	}
	return events, nil
}

// GetEventsByOrganizer retrieves events by organizer ID
func (s *serviceImpl) GetEventsByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*Event, error) {
	events, err := s.eventRepo.GetByOrganizer(ctx, organizerID)
	if err != nil {
		return nil, err // Repository already returns custom error
	}
	return events, nil
}

// UpdateEvent updates an existing event
func (s *serviceImpl) UpdateEvent(ctx context.Context, event *Event) error {
	// Get existing event
	existingEvent, err := s.eventRepo.GetByID(ctx, event.ID)
	if err != nil {
		return err // Repository already returns custom error
	}

	// Validate business rules
	if err := s.validateEventUpdate(existingEvent, event); err != nil {
		return err
	}

	// Validate the updated event
	if err := s.validateEvent(ctx, event); err != nil {
		return err
	}

	// Update the event
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return err // Repository already returns custom error
	}

	return nil
}

// CancelEvent cancels an event
func (s *serviceImpl) CancelEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err // Repository already returns custom error
	}

	// Check if user is the organizer
	if event.OrganizerID != organizerID {
		return NewUnauthorizedAccessError("cancel this event")
	}

	// Check if event can be cancelled
	if event.IsCancelled() {
		return ErrEventAlreadyCancelled
	}

	if event.IsCompleted() {
		return ErrCannotCancelCompleted
	}

	// Cancel the event
	event.Status = StatusCancelled
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return err // Repository already returns custom error
	}

	return nil
}

// DeleteEvent deletes an event (only if no tickets sold)
func (s *serviceImpl) DeleteEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err // Repository already returns custom error
	}

	// Check if user is the organizer
	if event.OrganizerID != organizerID {
		return NewUnauthorizedAccessError("delete this event")
	}

	// Check if any tickets have been sold
	if event.AvailableTickets < event.TotalTickets {
		return ErrCannotDeleteWithTickets
	}

	// Delete the event
	if err := s.eventRepo.Delete(ctx, eventID); err != nil {
		return err // Repository already returns custom error
	}

	return nil
}

// validateEvent validates event data
func (s *serviceImpl) validateEvent(ctx context.Context, event *Event) error {
	// Check if venue exists and get venue details
	venue, err := s.venueRepo.GetByID(ctx, event.VenueID)
	if err != nil {
		// Convert venue repository error to event error
		return NewVenueNotFoundError(event.VenueID)
	}

	// Check if event date is in the future
	if event.EventDate.Before(time.Now()) {
		return ErrEventDateInPast
	}

	// Check if total tickets doesn't exceed venue capacity
	if event.TotalTickets > venue.Capacity {
		return NewTicketsExceedCapacityError(event.TotalTickets, venue.Capacity)
	}

	return nil
}

// validateEventUpdate validates event update rules
func (s *serviceImpl) validateEventUpdate(existing *Event, updated *Event) error {
	// Cannot update cancelled or completed events
	if existing.IsCancelled() {
		return ErrCannotUpdateCancelled
	}

	if existing.IsCompleted() {
		return ErrCannotUpdateCompleted
	}

	// Cannot reduce total tickets below sold tickets
	soldTickets := existing.TotalTickets - existing.AvailableTickets
	if updated.TotalTickets < soldTickets {
		return NewInvalidTicketReductionError(updated.TotalTickets, soldTickets)
	}

	// Update available tickets if total tickets changed
	if updated.TotalTickets != existing.TotalTickets {
		ticketDifference := updated.TotalTickets - existing.TotalTickets
		updated.AvailableTickets = existing.AvailableTickets + ticketDifference
	}

	return nil
}
