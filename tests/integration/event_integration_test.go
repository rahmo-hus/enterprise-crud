//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"enterprise-crud/internal/app"
	"enterprise-crud/internal/dto/event"
	"enterprise-crud/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventIntegration(t *testing.T) {
	// Setup test database
	testDB := SetupTestDatabase(t)
	defer testDB.Close()
	defer testDB.Cleanup(t)

	// Create test fixtures
	fixtures := NewTestFixtures(testDB)
	_, organizerRole, _ := fixtures.StandardRoles(t)

	// Create database connection for the app
	dbConn := &database.Connection{DB: testDB.DB}

	// Create the application dependencies
	cfg := CreateTestConfig()

	// Create dependencies
	deps, err := CreateTestDependencies(cfg, dbConn)
	require.NoError(t, err, "Failed to create dependencies")

	// Create the application instance
	application := app.NewWireApp(cfg, dbConn, deps.UserHandler, deps.EventHandler, deps.OrderHandler)

	// Create HTTP handler
	router := application.SetupRouter()

	t.Run("POST /events - Create Event", func(t *testing.T) {
		// Create test data
		organizer := fixtures.CreateUser(t, "organizer@test.com", "organizer", "password123", organizerRole)
		_ = organizer // Used in request
		venue := fixtures.CreateVenue(t, "Test Venue", 100)

		// Prepare request payload
		createEventReq := event.CreateEventRequest{
			VenueID:      venue.ID,
			Title:        "Test Event",
			Description:  "Test event description",
			EventDate:    time.Now().Add(24 * time.Hour),
			TicketPrice:  50.0,
			TotalTickets: 100,
		}

		payload, err := json.Marshal(createEventReq)
		require.NoError(t, err)

		// Create HTTP request
		req, err := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusCreated, w.Code)

		var response event.EventResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, createEventReq.Title, response.Title)
		assert.Equal(t, createEventReq.Description, response.Description)
		assert.Equal(t, createEventReq.TicketPrice, response.TicketPrice)
		assert.Equal(t, createEventReq.TotalTickets, response.TotalTickets)
		assert.NotEmpty(t, response.ID)
	})

	t.Run("GET /events/{id} - Get Event", func(t *testing.T) {
		// Create test data
		organizer := fixtures.CreateUser(t, "organizer2@test.com", "organizer2", "password123", organizerRole)
		venue := fixtures.CreateVenue(t, "Test Venue 2", 200)
		testEvent := fixtures.CreateEvent(t, venue, organizer, "Test Event 2", 75.0, 150)

		// Create HTTP request
		req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/events/%s", testEvent.ID.String()), nil)
		require.NoError(t, err)

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response event.EventResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, testEvent.Title, response.Title)
		assert.Equal(t, testEvent.Description, response.Description)
		assert.Equal(t, testEvent.TicketPrice, response.TicketPrice)
		assert.Equal(t, testEvent.ID.String(), response.ID)
	})

	t.Run("PUT /events/{id} - Update Event", func(t *testing.T) {
		// Create test data
		organizer := fixtures.CreateUser(t, "organizer3@test.com", "organizer3", "password123", organizerRole)
		venue := fixtures.CreateVenue(t, "Test Venue 3", 300)
		testEvent := fixtures.CreateEvent(t, venue, organizer, "Original Event", 100.0, 200)

		// Prepare update payload
		updateEventReq := event.UpdateEventRequest{
			VenueID:      venue.ID,
			Title:        "Updated Event",
			Description:  "Updated description",
			EventDate:    time.Now().Add(24 * time.Hour),
			TicketPrice:  125.0,
			TotalTickets: 180,
		}

		payload, err := json.Marshal(updateEventReq)
		require.NoError(t, err)

		// Create HTTP request
		req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/events/%s", testEvent.ID.String()), bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response event.EventResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, updateEventReq.Title, response.Title)
		assert.Equal(t, updateEventReq.Description, response.Description)
		assert.Equal(t, updateEventReq.TicketPrice, response.TicketPrice)
		assert.Equal(t, updateEventReq.TotalTickets, response.TotalTickets)
	})

	t.Run("DELETE /events/{id} - Delete Event", func(t *testing.T) {
		// Create test data
		organizer := fixtures.CreateUser(t, "organizer4@test.com", "organizer4", "password123", organizerRole)
		venue := fixtures.CreateVenue(t, "Test Venue 4", 400)
		testEvent := fixtures.CreateEvent(t, venue, organizer, "Delete Event", 50.0, 100)

		// Create HTTP request
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/events/%s", testEvent.ID.String()), nil)
		require.NoError(t, err)

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify event is deleted - try to get it
		req, err = http.NewRequest("GET", fmt.Sprintf("/api/v1/events/%s", testEvent.ID.String()), nil)
		require.NoError(t, err)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GET /events - List Events", func(t *testing.T) {
		// Create test data
		organizer := fixtures.CreateUser(t, "organizer5@test.com", "organizer5", "password123", organizerRole)
		venue := fixtures.CreateVenue(t, "Test Venue 5", 500)
		event1 := fixtures.CreateEvent(t, venue, organizer, "Event 1", 50.0, 100)
		event2 := fixtures.CreateEvent(t, venue, organizer, "Event 2", 75.0, 150)

		// Create HTTP request
		req, err := http.NewRequest("GET", "/api/v1/events", nil)
		require.NoError(t, err)

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response []event.EventResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(response), 2)

		// Check if our test events are in the response
		eventTitles := make([]string, len(response))
		for i, e := range response {
			eventTitles[i] = e.Title
		}
		assert.Contains(t, eventTitles, event1.Title)
		assert.Contains(t, eventTitles, event2.Title)
	})

	t.Run("POST /events - Create Event with Invalid Data", func(t *testing.T) {
		// Test with invalid data
		createEventReq := event.CreateEventRequest{
			VenueID:      uuid.UUID{}, // Invalid UUID
			Title:        "",          // Empty title
			Description:  "Test description",
			EventDate:    time.Now().Add(-24 * time.Hour), // Past date
			TicketPrice:  -10.0,                           // Negative price
			TotalTickets: 0,                               // Zero tickets
		}

		payload, err := json.Marshal(createEventReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /events/{id} - Get Non-existent Event", func(t *testing.T) {
		// Create HTTP request with non-existent ID
		req, err := http.NewRequest("GET", "/api/v1/events/550e8400-e29b-41d4-a716-446655440000", nil)
		require.NoError(t, err)

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GET /events - List Events with Filters", func(t *testing.T) {
		// Create test data
		organizer := fixtures.CreateUser(t, "organizer6@test.com", "organizer6", "password123", organizerRole)
		venue := fixtures.CreateVenue(t, "Test Venue 6", 600)

		// Create events with different statuses
		activeEvent := fixtures.CreateEvent(t, venue, organizer, "Active Event", 50.0, 100)

		// Create HTTP request with status filter
		req, err := http.NewRequest("GET", "/api/v1/events?status=active", nil)
		require.NoError(t, err)

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response []event.EventResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// All events should be active
		for _, e := range response {
			assert.Equal(t, "active", e.Status)
		}

		// Our test event should be in the response
		eventTitles := make([]string, len(response))
		for i, e := range response {
			eventTitles[i] = e.Title
		}
		assert.Contains(t, eventTitles, activeEvent.Title)
	})
}
