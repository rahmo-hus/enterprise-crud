package database

import (
	"context"
	"enterprise-crud/internal/domain/event"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// eventRepository implements the event.Repository interface
type eventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new event repository instance
func NewEventRepository(db *gorm.DB) event.Repository {
	return &eventRepository{db: db}
}

// Create creates a new event in the database
func (r *eventRepository) Create(ctx context.Context, e *event.Event) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return event.NewEventError(event.ErrEventCreationFailed, err)
	}
	return nil
}

// GetByID retrieves an event by its ID
func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	var e event.Event
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, event.NewEventNotFoundError(id)
		}
		return nil, event.NewEventError(event.ErrEventRetrievalFailed, err)
	}
	return &e, nil
}

// GetAll retrieves all events
func (r *eventRepository) GetAll(ctx context.Context) ([]*event.Event, error) {
	var events []*event.Event
	if err := r.db.WithContext(ctx).Order("event_date ASC").Find(&events).Error; err != nil {
		return nil, event.NewEventError(event.ErrEventRetrievalFailed, err)
	}
	return events, nil
}

// GetByOrganizer retrieves events by organizer ID
func (r *eventRepository) GetByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*event.Event, error) {
	var events []*event.Event
	if err := r.db.WithContext(ctx).Where("organizer_id = ?", organizerID).Order("event_date ASC").Find(&events).Error; err != nil {
		return nil, event.NewEventError(event.ErrEventRetrievalFailed, err)
	}
	return events, nil
}

// GetByVenue retrieves events by venue ID
func (r *eventRepository) GetByVenue(ctx context.Context, venueID uuid.UUID) ([]*event.Event, error) {
	var events []*event.Event
	if err := r.db.WithContext(ctx).Where("venue_id = ?", venueID).Order("event_date ASC").Find(&events).Error; err != nil {
		return nil, event.NewEventError(event.ErrEventRetrievalFailed, err)
	}
	return events, nil
}

// Update updates an existing event
func (r *eventRepository) Update(ctx context.Context, e *event.Event) error {
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return event.NewEventError(event.ErrEventUpdateFailed, err)
	}
	return nil
}

// Delete deletes an event by its ID
func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&event.Event{}, id).Error; err != nil {
		return event.NewEventError(event.ErrEventDeletionFailed, err)
	}
	return nil
}