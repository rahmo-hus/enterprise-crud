package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"enterprise-crud/internal/domain/event"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// EventCacheService provides caching functionality for events
// It implements a cache-aside pattern with automatic TTL management
type EventCacheService struct {
	client   *redis.Client
	cacheTTL time.Duration
}

// NewEventCacheService creates a new event cache service
func NewEventCacheService(redisClient *RedisClient) *EventCacheService {
	return &EventCacheService{
		client:   redisClient.GetClient(),
		cacheTTL: redisClient.GetConfig().CacheTTL,
	}
}

// Cache Keys - Educational: Good practice to centralize cache key generation
const (
	eventByIDKeyPrefix     = "event:id:"
	eventsByVenueKeyPrefix = "events:venue:"
	eventsByOrgKeyPrefix   = "events:organizer:"
	allEventsKey           = "events:all"
)

// GetEvent retrieves an event from cache by ID
// Returns nil if not found in cache (cache miss)
func (s *EventCacheService) GetEvent(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	key := eventByIDKeyPrefix + id.String()

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get event from cache: %w", err)
	}

	var cachedEvent event.Event
	if err := json.Unmarshal([]byte(data), &cachedEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached event: %w", err)
	}

	return &cachedEvent, nil
}

// SetEvent stores an event in cache with TTL
func (s *EventCacheService) SetEvent(ctx context.Context, evt *event.Event) error {
	key := eventByIDKeyPrefix + evt.ID.String()

	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event for cache: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set event in cache: %w", err)
	}

	return nil
}

// DeleteEvent removes an event from cache
func (s *EventCacheService) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	key := eventByIDKeyPrefix + id.String()

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete event from cache: %w", err)
	}

	return nil
}

// GetEventsByVenue retrieves cached events for a specific venue
func (s *EventCacheService) GetEventsByVenue(ctx context.Context, venueID uuid.UUID) ([]*event.Event, error) {
	key := eventsByVenueKeyPrefix + venueID.String()

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get events by venue from cache: %w", err)
	}

	var events []*event.Event
	if err := json.Unmarshal([]byte(data), &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached events: %w", err)
	}

	return events, nil
}

// SetEventsByVenue stores events for a venue in cache
func (s *EventCacheService) SetEventsByVenue(ctx context.Context, venueID uuid.UUID, events []*event.Event) error {
	key := eventsByVenueKeyPrefix + venueID.String()

	data, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal events for cache: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set events by venue in cache: %w", err)
	}

	return nil
}

// GetEventsByOrganizer retrieves cached events for a specific organizer
func (s *EventCacheService) GetEventsByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*event.Event, error) {
	key := eventsByOrgKeyPrefix + organizerID.String()

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get events by organizer from cache: %w", err)
	}

	var events []*event.Event
	if err := json.Unmarshal([]byte(data), &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached events: %w", err)
	}

	return events, nil
}

// SetEventsByOrganizer stores events for an organizer in cache
func (s *EventCacheService) SetEventsByOrganizer(ctx context.Context, organizerID uuid.UUID, events []*event.Event) error {
	key := eventsByOrgKeyPrefix + organizerID.String()

	data, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal events for cache: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set events by organizer in cache: %w", err)
	}

	return nil
}

// GetAllEvents retrieves all cached events
func (s *EventCacheService) GetAllEvents(ctx context.Context) ([]*event.Event, error) {
	data, err := s.client.Get(ctx, allEventsKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get all events from cache: %w", err)
	}

	var events []*event.Event
	if err := json.Unmarshal([]byte(data), &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached events: %w", err)
	}

	return events, nil
}

// SetAllEvents stores all events in cache
func (s *EventCacheService) SetAllEvents(ctx context.Context, events []*event.Event) error {
	data, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal events for cache: %w", err)
	}

	if err := s.client.Set(ctx, allEventsKey, data, s.cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set all events in cache: %w", err)
	}

	return nil
}

// InvalidateEventCaches removes all event-related caches
// This is called when events are modified to ensure cache consistency
func (s *EventCacheService) InvalidateEventCaches(ctx context.Context) error {
	// Use a pipeline for efficient batch operations
	pipe := s.client.Pipeline()

	// Delete pattern-based keys (requires Redis SCAN)
	iter := s.client.Scan(ctx, 0, "event:*", 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan event keys: %w", err)
	}

	// Add all keys to deletion pipeline
	for _, key := range keys {
		pipe.Del(ctx, key)
	}

	// Delete specific keys
	pipe.Del(ctx, allEventsKey)

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute cache invalidation pipeline: %w", err)
	}

	return nil
}

// InvalidateEventRelatedCaches invalidates caches related to a specific event
// This is more granular than full cache invalidation
func (s *EventCacheService) InvalidateEventRelatedCaches(ctx context.Context, eventID, venueID, organizerID uuid.UUID) error {
	pipe := s.client.Pipeline()

	// Delete specific event cache
	pipe.Del(ctx, eventByIDKeyPrefix+eventID.String())

	// Delete venue-related cache
	pipe.Del(ctx, eventsByVenueKeyPrefix+venueID.String())

	// Delete organizer-related cache
	pipe.Del(ctx, eventsByOrgKeyPrefix+organizerID.String())

	// Delete all events cache
	pipe.Del(ctx, allEventsKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute event-related cache invalidation: %w", err)
	}

	return nil
}
