package cache

import (
	"context"
	"log"

	"enterprise-crud/internal/domain/event"

	"github.com/google/uuid"
)

// CachedEventRepository implements the event.Repository interface with Redis caching
// It uses the cache-aside pattern: check cache first, fallback to database, then populate cache
type CachedEventRepository struct {
	baseRepo event.Repository   // The original database repository
	cache    *EventCacheService // Redis cache service
}

// NewCachedEventRepository creates a new cached event repository
func NewCachedEventRepository(baseRepo event.Repository, cache *EventCacheService) *CachedEventRepository {
	return &CachedEventRepository{
		baseRepo: baseRepo,
		cache:    cache,
	}
}

// Create creates a new event and invalidates related caches
func (r *CachedEventRepository) Create(ctx context.Context, evt *event.Event) error {
	// Create in database first
	if err := r.baseRepo.Create(ctx, evt); err != nil {
		return err
	}

	// Invalidate caches since we have a new event
	if err := r.cache.InvalidateEventRelatedCaches(ctx, evt.ID, evt.VenueID, evt.OrganizerID); err != nil {
		// Log but don't fail the operation - cache invalidation is not critical for data consistency
		log.Printf("Warning: Failed to invalidate cache after event creation: %v", err)
	}

	return nil
}

// GetByID implements cache-aside pattern for single event retrieval
func (r *CachedEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	// 1. Try cache first (cache-aside pattern)
	if cachedEvent, err := r.cache.GetEvent(ctx, id); err != nil {
		log.Printf("Cache error for event %s: %v", id, err)
	} else if cachedEvent != nil {
		// Cache hit!
		return cachedEvent, nil
	}

	// 2. Cache miss - get from database
	evt, err := r.baseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Populate cache for next time (async to avoid blocking)
	go func() {
		if err := r.cache.SetEvent(context.Background(), evt); err != nil {
			log.Printf("Warning: Failed to cache event %s: %v", id, err)
		}
	}()

	return evt, nil
}

// GetAll implements caching for all events
func (r *CachedEventRepository) GetAll(ctx context.Context) ([]*event.Event, error) {
	// 1. Try cache first
	if cachedEvents, err := r.cache.GetAllEvents(ctx); err != nil {
		log.Printf("Cache error for all events: %v", err)
	} else if cachedEvents != nil {
		return cachedEvents, nil
	}

	// 2. Cache miss - get from database
	events, err := r.baseRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Populate cache (async)
	go func() {
		if err := r.cache.SetAllEvents(context.Background(), events); err != nil {
			log.Printf("Warning: Failed to cache all events: %v", err)
		}
	}()

	return events, nil
}

// GetByOrganizer implements caching for events by organizer
func (r *CachedEventRepository) GetByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*event.Event, error) {
	// 1. Try cache first
	if cachedEvents, err := r.cache.GetEventsByOrganizer(ctx, organizerID); err != nil {
		log.Printf("Cache error for organizer %s events: %v", organizerID, err)
	} else if cachedEvents != nil {
		return cachedEvents, nil
	}

	// 2. Cache miss - get from database
	events, err := r.baseRepo.GetByOrganizer(ctx, organizerID)
	if err != nil {
		return nil, err
	}

	// 3. Populate cache (async)
	go func() {
		if err := r.cache.SetEventsByOrganizer(context.Background(), organizerID, events); err != nil {
			log.Printf("Warning: Failed to cache events for organizer %s: %v", organizerID, err)
		}
	}()

	return events, nil
}

// GetByVenue implements caching for events by venue
func (r *CachedEventRepository) GetByVenue(ctx context.Context, venueID uuid.UUID) ([]*event.Event, error) {
	// 1. Try cache first
	if cachedEvents, err := r.cache.GetEventsByVenue(ctx, venueID); err != nil {
		log.Printf("Cache error for venue %s events: %v", venueID, err)
	} else if cachedEvents != nil {
		return cachedEvents, nil
	}

	// 2. Cache miss - get from database
	events, err := r.baseRepo.GetByVenue(ctx, venueID)
	if err != nil {
		return nil, err
	}

	// 3. Populate cache (async)
	go func() {
		if err := r.cache.SetEventsByVenue(context.Background(), venueID, events); err != nil {
			log.Printf("Warning: Failed to cache events for venue %s: %v", venueID, err)
		}
	}()

	return events, nil
}

// Update updates an event and invalidates related caches
func (r *CachedEventRepository) Update(ctx context.Context, evt *event.Event) error {
	// Update in database first
	if err := r.baseRepo.Update(ctx, evt); err != nil {
		return err
	}

	// Invalidate related caches
	if err := r.cache.InvalidateEventRelatedCaches(ctx, evt.ID, evt.VenueID, evt.OrganizerID); err != nil {
		log.Printf("Warning: Failed to invalidate cache after event update: %v", err)
	}

	return nil
}

// Delete deletes an event and removes it from cache
func (r *CachedEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get event details before deletion for cache invalidation
	evt, err := r.baseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database first
	if err := r.baseRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate related caches
	if err := r.cache.InvalidateEventRelatedCaches(ctx, evt.ID, evt.VenueID, evt.OrganizerID); err != nil {
		log.Printf("Warning: Failed to invalidate cache after event deletion: %v", err)
	}

	return nil
}
