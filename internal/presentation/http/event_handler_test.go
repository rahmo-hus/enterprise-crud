package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"enterprise-crud/internal/domain/event"
	eventDto "enterprise-crud/internal/dto/event"
	"enterprise-crud/internal/infrastructure/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventService is a mock implementation of event.Service interface
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) CreateEvent(ctx context.Context, event *event.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) GetEventByID(ctx context.Context, id uuid.UUID) (*event.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockEventService) GetAllEvents(ctx context.Context) ([]*event.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*event.Event), args.Error(1)
}

func (m *MockEventService) GetEventsByOrganizer(ctx context.Context, organizerID uuid.UUID) ([]*event.Event, error) {
	args := m.Called(ctx, organizerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*event.Event), args.Error(1)
}

func (m *MockEventService) UpdateEvent(ctx context.Context, event *event.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventService) CancelEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error {
	args := m.Called(ctx, eventID, organizerID)
	return args.Error(0)
}

func (m *MockEventService) DeleteEvent(ctx context.Context, eventID uuid.UUID, organizerID uuid.UUID) error {
	args := m.Called(ctx, eventID, organizerID)
	return args.Error(0)
}

func TestEventHandler_CreateEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockEventService)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful event creation",
			requestBody: eventDto.CreateEventRequest{
				VenueID:      uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks: func(mockService *MockEventService) {
				mockService.On("CreateEvent", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: uuid.New(),
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"title": "", // Invalid: empty title
			},
			setupMocks:     func(mockService *MockEventService) {},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: uuid.New(),
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name: "user not authenticated",
			requestBody: eventDto.CreateEventRequest{
				VenueID:      uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks:     func(mockService *MockEventService) {},
			setupAuth:      func(c *gin.Context) {}, // No auth setup
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "venue not found",
			requestBody: eventDto.CreateEventRequest{
				VenueID:      uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks: func(mockService *MockEventService) {
				mockService.On("CreateEvent", mock.Anything, mock.AnythingOfType("*event.Event")).Return(event.ErrVenueNotFound)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: uuid.New(),
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "VENUE_NOT_FOUND",
		},
		{
			name: "validation error",
			requestBody: eventDto.CreateEventRequest{
				VenueID:      uuid.New(),
				Title:        "Test Event",
				Description:  "Test Description",
				EventDate:    time.Now().Add(24 * time.Hour),
				TicketPrice:  50.0,
				TotalTickets: 100,
			},
			setupMocks: func(mockService *MockEventService) {
				mockService.On("CreateEvent", mock.Anything, mock.AnythingOfType("*event.Event")).Return(event.ErrEventDateInPast)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: uuid.New(),
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "EVENT_DATE_INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEventService)
			tt.setupMocks(mockService)

			handler := NewEventHandler(mockService, auth.NewJWTService("test-secret", "test-issuer", time.Hour))

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			// Setup auth
			tt.setupAuth(c)
			
			// Call handler
			handler.CreateEvent(c)
			
			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var errorResponse eventDto.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Error)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestEventHandler_GetEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		eventID        string
		setupMocks     func(*MockEventService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "successful event retrieval",
			eventID: uuid.New().String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("GetEventByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&event.Event{
					ID:    uuid.New(),
					Title: "Test Event",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "invalid event ID",
			eventID: "invalid-uuid",
			setupMocks: func(mockService *MockEventService) {
				// No mocks needed as handler should return error before calling service
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_id",
		},
		{
			name:    "event not found",
			eventID: uuid.New().String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("GetEventByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, event.ErrEventNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "EVENT_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEventService)
			tt.setupMocks(mockService)

			handler := NewEventHandler(mockService, auth.NewJWTService("test-secret", "test-issuer", time.Hour))

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/events/"+tt.eventID, nil)
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.eventID}}
			
			// Call handler
			handler.GetEvent(c)
			
			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var errorResponse eventDto.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Error)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestEventHandler_GetAllEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMocks     func(*MockEventService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful events retrieval",
			setupMocks: func(mockService *MockEventService) {
				events := []*event.Event{
					{ID: uuid.New(), Title: "Event 1"},
					{ID: uuid.New(), Title: "Event 2"},
				}
				mockService.On("GetAllEvents", mock.Anything).Return(events, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "empty events list",
			setupMocks: func(mockService *MockEventService) {
				mockService.On("GetAllEvents", mock.Anything).Return([]*event.Event{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "service error",
			setupMocks: func(mockService *MockEventService) {
				mockService.On("GetAllEvents", mock.Anything).Return(nil, event.ErrEventRetrievalFailed)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEventService)
			tt.setupMocks(mockService)

			handler := NewEventHandler(mockService, auth.NewJWTService("test-secret", "test-issuer", time.Hour))

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/events", nil)
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			// Call handler
			handler.GetAllEvents(c)
			
			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedStatus == http.StatusOK {
				var response eventDto.EventListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, response.Count)
				assert.Len(t, response.Events, tt.expectedCount)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestEventHandler_CancelEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventID := uuid.New()
	organizerID := uuid.New()

	tests := []struct {
		name           string
		eventID        string
		setupMocks     func(*MockEventService)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "successful event cancellation",
			eventID: eventID.String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("CancelEvent", mock.Anything, eventID, organizerID).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: organizerID,
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "invalid event ID",
			eventID: "invalid-uuid",
			setupMocks: func(mockService *MockEventService) {
				// No mocks needed as handler should return error before calling service
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: organizerID,
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_id",
		},
		{
			name:    "unauthorized cancellation",
			eventID: eventID.String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("CancelEvent", mock.Anything, eventID, mock.AnythingOfType("uuid.UUID")).Return(event.ErrUnauthorizedAccess)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: uuid.New(), // Different user
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "UNAUTHORIZED_ACCESS",
		},
		{
			name:    "event already cancelled",
			eventID: eventID.String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("CancelEvent", mock.Anything, eventID, organizerID).Return(event.ErrEventAlreadyCancelled)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: organizerID,
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "EVENT_ALREADY_CANCELLED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEventService)
			tt.setupMocks(mockService)

			handler := NewEventHandler(mockService, auth.NewJWTService("test-secret", "test-issuer", time.Hour))

			// Create request
			req := httptest.NewRequest(http.MethodPatch, "/events/"+tt.eventID+"/cancel", nil)
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.eventID}}
			
			// Setup auth
			tt.setupAuth(c)
			
			// Call handler
			handler.CancelEvent(c)
			
			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var errorResponse eventDto.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Error)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestEventHandler_DeleteEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventID := uuid.New()
	organizerID := uuid.New()

	tests := []struct {
		name           string
		eventID        string
		setupMocks     func(*MockEventService)
		setupAuth      func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "successful event deletion",
			eventID: eventID.String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("DeleteEvent", mock.Anything, eventID, organizerID).Return(nil)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: organizerID,
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "cannot delete event with sold tickets",
			eventID: eventID.String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("DeleteEvent", mock.Anything, eventID, organizerID).Return(event.ErrCannotDeleteWithTickets)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: organizerID,
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "CANNOT_DELETE_WITH_TICKETS",
		},
		{
			name:    "unauthorized deletion",
			eventID: eventID.String(),
			setupMocks: func(mockService *MockEventService) {
				mockService.On("DeleteEvent", mock.Anything, eventID, mock.AnythingOfType("uuid.UUID")).Return(event.ErrUnauthorizedAccess)
			},
			setupAuth: func(c *gin.Context) {
				c.Set("user", &auth.JWTClaims{
					UserID: uuid.New(), // Different user
					Roles:  []string{"ORGANIZER"},
				})
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "UNAUTHORIZED_ACCESS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEventService)
			tt.setupMocks(mockService)

			handler := NewEventHandler(mockService, auth.NewJWTService("test-secret", "test-issuer", time.Hour))

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/events/"+tt.eventID, nil)
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.eventID}}
			
			// Setup auth
			tt.setupAuth(c)
			
			// Call handler
			handler.DeleteEvent(c)
			
			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var errorResponse eventDto.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Error)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestEventHandler_NewEventHandler(t *testing.T) {
	mockService := new(MockEventService)
	jwtService := auth.NewJWTService("test-secret", "test-issuer", time.Hour)
	handler := NewEventHandler(mockService, jwtService)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.eventService)
	assert.Equal(t, jwtService, handler.jwtService)
}