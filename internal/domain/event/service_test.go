package event

import (
	"context"
	"testing"
	"time"

	"enterprise-crud/internal/domain/venue"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventRepository is a mock implementation of Repository interface
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Event), args.Error(1)
}

func (m *MockEventRepository) GetAll(ctx context.Context) ([]*Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Event), args.Error(1)
}

func (m *MockEventRepository) GetByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*Event, error) {
	args := m.Called(ctx, organizerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Event), args.Error(1)
}

func (m *MockEventRepository) GetByVenue(ctx context.Context, venueID uuid.UUID) ([]*Event, error) {
	args := m.Called(ctx, venueID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Event), args.Error(1)
}

func (m *MockEventRepository) Update(ctx context.Context, event *Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockVenueRepository is a mock implementation of venue.Repository interface
type MockVenueRepository struct {
	mock.Mock
}

func (m *MockVenueRepository) Create(ctx context.Context, venue *venue.Venue) error {
	args := m.Called(ctx, venue)
	return args.Error(0)
}

func (m *MockVenueRepository) GetByID(ctx context.Context, id uuid.UUID) (*venue.Venue, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*venue.Venue), args.Error(1)
}

func (m *MockVenueRepository) GetAll(ctx context.Context) ([]*venue.Venue, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*venue.Venue), args.Error(1)
}

func (m *MockVenueRepository) Update(ctx context.Context, venue *venue.Venue) error {
	args := m.Called(ctx, venue)
	return args.Error(0)
}

func (m *MockVenueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestEventService_CreateEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       *Event
		setupMocks  func(*MockEventRepository, *MockVenueRepository)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "successful event creation",
			event: &Event{
				ID:           uuid.New(),
				VenueID:      uuid.New(),
				OrganizerID:  uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				venueRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&venue.Venue{
					ID:       uuid.New(),
					Name:     "Test Venue",
					Capacity: 200,
				}, nil)
				eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "venue not found",
			event: &Event{
				ID:           uuid.New(),
				VenueID:      uuid.New(),
				OrganizerID:  uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				venueRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, venue.ErrVenueNotFound)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsVenueNotFoundError(err)
			},
		},
		{
			name: "event date in past",
			event: &Event{
				ID:           uuid.New(),
				VenueID:      uuid.New(),
				OrganizerID:  uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(-24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				venueRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&venue.Venue{
					ID:       uuid.New(),
					Name:     "Test Venue",
					Capacity: 200,
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err == ErrEventDateInPast
			},
		},
		{
			name: "tickets exceed venue capacity",
			event: &Event{
				ID:           uuid.New(),
				VenueID:      uuid.New(),
				OrganizerID:  uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 300,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				venueRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&venue.Venue{
					ID:       uuid.New(),
					Name:     "Test Venue",
					Capacity: 200,
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsValidationError(err)
			},
		},
		{
			name: "repository create error",
			event: &Event{
				ID:           uuid.New(),
				VenueID:      uuid.New(),
				OrganizerID:  uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				venueRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&venue.Venue{
					ID:       uuid.New(),
					Name:     "Test Venue",
					Capacity: 200,
				}, nil)
				eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*event.Event")).Return(ErrEventCreationFailed)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err == ErrEventCreationFailed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := new(MockEventRepository)
			venueRepo := new(MockVenueRepository)

			tt.setupMocks(eventRepo, venueRepo)

			service := NewService(eventRepo, venueRepo)
			err := service.CreateEvent(context.Background(), tt.event)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, StatusActive, tt.event.Status)
				assert.Equal(t, tt.event.TotalTickets, tt.event.AvailableTickets)
			}

			eventRepo.AssertExpectations(t)
			venueRepo.AssertExpectations(t)
		})
	}
}

func TestEventService_GetEventByID(t *testing.T) {
	tests := []struct {
		name        string
		eventID     uuid.UUID
		setupMocks  func(*MockEventRepository, *MockVenueRepository)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:    "successful event retrieval",
			eventID: uuid.New(),
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&Event{
					ID:    uuid.New(),
					Title: "Test Event",
				}, nil)
			},
			expectError: false,
		},
		{
			name:    "event not found",
			eventID: uuid.New(),
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, ErrEventNotFound)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsEventNotFoundError(err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := new(MockEventRepository)
			venueRepo := new(MockVenueRepository)

			tt.setupMocks(eventRepo, venueRepo)

			service := NewService(eventRepo, venueRepo)
			event, err := service.GetEventByID(context.Background(), tt.eventID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, event)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, event)
			}

			eventRepo.AssertExpectations(t)
			venueRepo.AssertExpectations(t)
		})
	}
}

func TestEventService_CancelEvent(t *testing.T) {
	organizerID := uuid.New()
	eventID := uuid.New()

	tests := []struct {
		name        string
		eventID     uuid.UUID
		organizerID uuid.UUID
		setupMocks  func(*MockEventRepository, *MockVenueRepository)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "successful event cancellation",
			eventID:     eventID,
			organizerID: organizerID,
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, eventID).Return(&Event{
					ID:          eventID,
					OrganizerID: organizerID,
					Status:      StatusActive,
				}, nil)
				eventRepo.On("Update", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "unauthorized cancellation",
			eventID:     eventID,
			organizerID: uuid.New(),
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, eventID).Return(&Event{
					ID:          eventID,
					OrganizerID: organizerID,
					Status:      StatusActive,
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsUnauthorizedError(err)
			},
		},
		{
			name:        "event already cancelled",
			eventID:     eventID,
			organizerID: organizerID,
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, eventID).Return(&Event{
					ID:          eventID,
					OrganizerID: organizerID,
					Status:      StatusCancelled,
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err == ErrEventAlreadyCancelled
			},
		},
		{
			name:        "cannot cancel completed event",
			eventID:     eventID,
			organizerID: organizerID,
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, eventID).Return(&Event{
					ID:          eventID,
					OrganizerID: organizerID,
					Status:      StatusCompleted,
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err == ErrCannotCancelCompleted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := new(MockEventRepository)
			venueRepo := new(MockVenueRepository)

			tt.setupMocks(eventRepo, venueRepo)

			service := NewService(eventRepo, venueRepo)
			err := service.CancelEvent(context.Background(), tt.eventID, tt.organizerID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}

			eventRepo.AssertExpectations(t)
			venueRepo.AssertExpectations(t)
		})
	}
}

func TestEventService_UpdateEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       *Event
		setupMocks  func(*MockEventRepository, *MockVenueRepository)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "successful event update",
			event: &Event{
				ID:               uuid.New(),
				VenueID:          uuid.New(),
				OrganizerID:      uuid.New(),
				Title:            "Updated Event",
				EventDate:        time.Now().Add(24 * time.Hour),
				TotalTickets:     150,
				AvailableTickets: 100,
				Status:           StatusActive,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&Event{
					ID:               uuid.New(),
					Status:           StatusActive,
					TotalTickets:     100,
					AvailableTickets: 50,
				}, nil)
				venueRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&venue.Venue{
					ID:       uuid.New(),
					Capacity: 200,
				}, nil)
				eventRepo.On("Update", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "cannot update cancelled event",
			event: &Event{
				ID:     uuid.New(),
				Status: StatusActive,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&Event{
					ID:     uuid.New(),
					Status: StatusCancelled,
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err == ErrCannotUpdateCancelled
			},
		},
		{
			name: "invalid ticket reduction",
			event: &Event{
				ID:           uuid.New(),
				TotalTickets: 30,
				Status:       StatusActive,
			},
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&Event{
					ID:               uuid.New(),
					Status:           StatusActive,
					TotalTickets:     100,
					AvailableTickets: 50, // 50 tickets sold
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsValidationError(err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := new(MockEventRepository)
			venueRepo := new(MockVenueRepository)

			tt.setupMocks(eventRepo, venueRepo)

			service := NewService(eventRepo, venueRepo)
			err := service.UpdateEvent(context.Background(), tt.event)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}

			eventRepo.AssertExpectations(t)
			venueRepo.AssertExpectations(t)
		})
	}
}

func TestEventService_DeleteEvent(t *testing.T) {
	organizerID := uuid.New()
	eventID := uuid.New()

	tests := []struct {
		name        string
		eventID     uuid.UUID
		organizerID uuid.UUID
		setupMocks  func(*MockEventRepository, *MockVenueRepository)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "successful event deletion",
			eventID:     eventID,
			organizerID: organizerID,
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, eventID).Return(&Event{
					ID:               eventID,
					OrganizerID:      organizerID,
					TotalTickets:     100,
					AvailableTickets: 100, // No tickets sold
				}, nil)
				eventRepo.On("Delete", mock.Anything, eventID).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "cannot delete event with sold tickets",
			eventID:     eventID,
			organizerID: organizerID,
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, eventID).Return(&Event{
					ID:               eventID,
					OrganizerID:      organizerID,
					TotalTickets:     100,
					AvailableTickets: 50, // 50 tickets sold
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err == ErrCannotDeleteWithTickets
			},
		},
		{
			name:        "unauthorized deletion",
			eventID:     eventID,
			organizerID: uuid.New(),
			setupMocks: func(eventRepo *MockEventRepository, venueRepo *MockVenueRepository) {
				eventRepo.On("GetByID", mock.Anything, eventID).Return(&Event{
					ID:          eventID,
					OrganizerID: organizerID,
				}, nil)
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsUnauthorizedError(err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := new(MockEventRepository)
			venueRepo := new(MockVenueRepository)

			tt.setupMocks(eventRepo, venueRepo)

			service := NewService(eventRepo, venueRepo)
			err := service.DeleteEvent(context.Background(), tt.eventID, tt.organizerID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}

			eventRepo.AssertExpectations(t)
			venueRepo.AssertExpectations(t)
		})
	}
}
