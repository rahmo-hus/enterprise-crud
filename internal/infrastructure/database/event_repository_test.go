package database

import (
	"context"
	"testing"
	"time"

	"enterprise-crud/internal/domain/event"
	"enterprise-crud/internal/domain/venue"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// EventRepositoryTestSuite is a test suite for event repository
type EventRepositoryTestSuite struct {
	suite.Suite
	db        *gorm.DB
	eventRepo event.Repository
	venueRepo venue.Repository
}

// SetupSuite runs before all tests in the suite
func (suite *EventRepositoryTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	suite.Require().NoError(err)

	// Create tables manually for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS venues (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			address TEXT NOT NULL,
			capacity INTEGER NOT NULL,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	suite.Require().NoError(err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			venue_id TEXT NOT NULL,
			organizer_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			event_date DATETIME NOT NULL,
			ticket_price DECIMAL(10,2) NOT NULL,
			available_tickets INTEGER NOT NULL,
			total_tickets INTEGER NOT NULL,
			status TEXT DEFAULT 'ACTIVE',
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY (venue_id) REFERENCES venues(id)
		)
	`).Error
	suite.Require().NoError(err)

	suite.db = db
	suite.eventRepo = NewEventRepository(db)
	suite.venueRepo = NewVenueRepository(db)
}

// TearDownSuite runs after all tests in the suite
func (suite *EventRepositoryTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	suite.Require().NoError(err)
	sqlDB.Close()
}

// SetupTest runs before each test
func (suite *EventRepositoryTestSuite) SetupTest() {
	// Clear all tables before each test
	suite.db.Exec("DELETE FROM events")
	suite.db.Exec("DELETE FROM venues")
}

func (suite *EventRepositoryTestSuite) TestCreate() {
	// Create a venue first
	testVenue := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Test Venue",
		Address:  "123 Test St",
		Capacity: 100,
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Create an event
	testEvent := &event.Event{
		ID:               uuid.New(),
		VenueID:          testVenue.ID,
		OrganizerID:      uuid.New(),
		Title:            "Test Event",
		Description:      "Test Description",
		EventDate:        time.Now().Add(24 * time.Hour),
		TicketPrice:      50.0,
		AvailableTickets: 50,
		TotalTickets:     50,
		Status:           event.StatusActive,
	}

	err = suite.eventRepo.Create(context.Background(), testEvent)
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, testEvent.ID)
}

func (suite *EventRepositoryTestSuite) TestGetByID() {
	// Create a venue first
	testVenue := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Test Venue",
		Address:  "123 Test St",
		Capacity: 100,
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Create and save an event
	testEvent := &event.Event{
		ID:               uuid.New(),
		VenueID:          testVenue.ID,
		OrganizerID:      uuid.New(),
		Title:            "Test Event",
		Description:      "Test Description",
		EventDate:        time.Now().Add(24 * time.Hour),
		TicketPrice:      50.0,
		AvailableTickets: 50,
		TotalTickets:     50,
		Status:           event.StatusActive,
	}
	err = suite.eventRepo.Create(context.Background(), testEvent)
	suite.Require().NoError(err)

	// Test successful retrieval
	foundEvent, err := suite.eventRepo.GetByID(context.Background(), testEvent.ID)
	suite.NoError(err)
	suite.Equal(testEvent.ID, foundEvent.ID)
	suite.Equal(testEvent.Title, foundEvent.Title)
	suite.Equal(testEvent.VenueID, foundEvent.VenueID)

	// Test event not found
	nonExistentID := uuid.New()
	foundEvent, err = suite.eventRepo.GetByID(context.Background(), nonExistentID)
	suite.Error(err)
	suite.Nil(foundEvent)
	suite.True(event.IsEventNotFoundError(err))
}

func (suite *EventRepositoryTestSuite) TestGetAll() {
	// Create a venue first
	testVenue := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Test Venue",
		Address:  "123 Test St",
		Capacity: 100,
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Create multiple events
	events := []*event.Event{
		{
			ID:               uuid.New(),
			VenueID:          testVenue.ID,
			OrganizerID:      uuid.New(),
			Title:            "Event 1",
			EventDate:        time.Now().Add(24 * time.Hour),
			TicketPrice:      50.0,
			AvailableTickets: 50,
			TotalTickets:     50,
			Status:           event.StatusActive,
		},
		{
			ID:               uuid.New(),
			VenueID:          testVenue.ID,
			OrganizerID:      uuid.New(),
			Title:            "Event 2",
			EventDate:        time.Now().Add(48 * time.Hour),
			TicketPrice:      75.0,
			AvailableTickets: 25,
			TotalTickets:     25,
			Status:           event.StatusActive,
		},
	}

	// Save all events
	for _, e := range events {
		err := suite.eventRepo.Create(context.Background(), e)
		suite.Require().NoError(err)
	}

	// Test GetAll
	allEvents, err := suite.eventRepo.GetAll(context.Background())
	suite.NoError(err)
	suite.Len(allEvents, 2)
	
	// Events should be ordered by event_date
	suite.Equal("Event 1", allEvents[0].Title)
	suite.Equal("Event 2", allEvents[1].Title)
}

func (suite *EventRepositoryTestSuite) TestGetByOrganizer() {
	// Create a venue first
	testVenue := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Test Venue",
		Address:  "123 Test St",
		Capacity: 100,
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	organizerID := uuid.New()
	
	// Create events for different organizers
	events := []*event.Event{
		{
			ID:               uuid.New(),
			VenueID:          testVenue.ID,
			OrganizerID:      organizerID,
			Title:            "Organizer Event 1",
			EventDate:        time.Now().Add(24 * time.Hour),
			TicketPrice:      50.0,
			AvailableTickets: 50,
			TotalTickets:     50,
			Status:           event.StatusActive,
		},
		{
			ID:               uuid.New(),
			VenueID:          testVenue.ID,
			OrganizerID:      organizerID,
			Title:            "Organizer Event 2",
			EventDate:        time.Now().Add(48 * time.Hour),
			TicketPrice:      75.0,
			AvailableTickets: 25,
			TotalTickets:     25,
			Status:           event.StatusActive,
		},
		{
			ID:               uuid.New(),
			VenueID:          testVenue.ID,
			OrganizerID:      uuid.New(), // Different organizer
			Title:            "Other Event",
			EventDate:        time.Now().Add(72 * time.Hour),
			TicketPrice:      100.0,
			AvailableTickets: 10,
			TotalTickets:     10,
			Status:           event.StatusActive,
		},
	}

	// Save all events
	for _, e := range events {
		err := suite.eventRepo.Create(context.Background(), e)
		suite.Require().NoError(err)
	}

	// Test GetByOrganizer
	organizerEvents, err := suite.eventRepo.GetByOrganizer(context.Background(), organizerID)
	suite.NoError(err)
	suite.Len(organizerEvents, 2)
	
	// Verify events belong to correct organizer
	for _, e := range organizerEvents {
		suite.Equal(organizerID, e.OrganizerID)
	}
}

func (suite *EventRepositoryTestSuite) TestGetByVenue() {
	// Create venues
	venue1 := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Venue 1",
		Address:  "123 Test St",
		Capacity: 100,
	}
	venue2 := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Venue 2",
		Address:  "456 Test Ave",
		Capacity: 200,
	}
	
	err := suite.venueRepo.Create(context.Background(), venue1)
	suite.Require().NoError(err)
	err = suite.venueRepo.Create(context.Background(), venue2)
	suite.Require().NoError(err)

	// Create events for different venues
	events := []*event.Event{
		{
			ID:               uuid.New(),
			VenueID:          venue1.ID,
			OrganizerID:      uuid.New(),
			Title:            "Venue 1 Event 1",
			EventDate:        time.Now().Add(24 * time.Hour),
			TicketPrice:      50.0,
			AvailableTickets: 50,
			TotalTickets:     50,
			Status:           event.StatusActive,
		},
		{
			ID:               uuid.New(),
			VenueID:          venue1.ID,
			OrganizerID:      uuid.New(),
			Title:            "Venue 1 Event 2",
			EventDate:        time.Now().Add(48 * time.Hour),
			TicketPrice:      75.0,
			AvailableTickets: 25,
			TotalTickets:     25,
			Status:           event.StatusActive,
		},
		{
			ID:               uuid.New(),
			VenueID:          venue2.ID,
			OrganizerID:      uuid.New(),
			Title:            "Venue 2 Event",
			EventDate:        time.Now().Add(72 * time.Hour),
			TicketPrice:      100.0,
			AvailableTickets: 10,
			TotalTickets:     10,
			Status:           event.StatusActive,
		},
	}

	// Save all events
	for _, e := range events {
		err := suite.eventRepo.Create(context.Background(), e)
		suite.Require().NoError(err)
	}

	// Test GetByVenue
	venue1Events, err := suite.eventRepo.GetByVenue(context.Background(), venue1.ID)
	suite.NoError(err)
	suite.Len(venue1Events, 2)
	
	// Verify events belong to correct venue
	for _, e := range venue1Events {
		suite.Equal(venue1.ID, e.VenueID)
	}
}

func (suite *EventRepositoryTestSuite) TestUpdate() {
	// Create a venue first
	testVenue := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Test Venue",
		Address:  "123 Test St",
		Capacity: 100,
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Create and save an event
	testEvent := &event.Event{
		ID:               uuid.New(),
		VenueID:          testVenue.ID,
		OrganizerID:      uuid.New(),
		Title:            "Original Title",
		Description:      "Original Description",
		EventDate:        time.Now().Add(24 * time.Hour),
		TicketPrice:      50.0,
		AvailableTickets: 50,
		TotalTickets:     50,
		Status:           event.StatusActive,
	}
	err = suite.eventRepo.Create(context.Background(), testEvent)
	suite.Require().NoError(err)

	// Update the event
	testEvent.Title = "Updated Title"
	testEvent.Description = "Updated Description"
	testEvent.TicketPrice = 75.0

	err = suite.eventRepo.Update(context.Background(), testEvent)
	suite.NoError(err)

	// Verify update
	updatedEvent, err := suite.eventRepo.GetByID(context.Background(), testEvent.ID)
	suite.NoError(err)
	suite.Equal("Updated Title", updatedEvent.Title)
	suite.Equal("Updated Description", updatedEvent.Description)
	suite.Equal(75.0, updatedEvent.TicketPrice)
}

func (suite *EventRepositoryTestSuite) TestDelete() {
	// Create a venue first
	testVenue := &venue.Venue{
		ID:       uuid.New(),
		Name:     "Test Venue",
		Address:  "123 Test St",
		Capacity: 100,
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Create and save an event
	testEvent := &event.Event{
		ID:               uuid.New(),
		VenueID:          testVenue.ID,
		OrganizerID:      uuid.New(),
		Title:            "Test Event",
		Description:      "Test Description",
		EventDate:        time.Now().Add(24 * time.Hour),
		TicketPrice:      50.0,
		AvailableTickets: 50,
		TotalTickets:     50,
		Status:           event.StatusActive,
	}
	err = suite.eventRepo.Create(context.Background(), testEvent)
	suite.Require().NoError(err)

	// Delete the event
	err = suite.eventRepo.Delete(context.Background(), testEvent.ID)
	suite.NoError(err)

	// Verify deletion
	deletedEvent, err := suite.eventRepo.GetByID(context.Background(), testEvent.ID)
	suite.Error(err)
	suite.Nil(deletedEvent)
	suite.True(event.IsEventNotFoundError(err))
}

func (suite *EventRepositoryTestSuite) TestNewEventRepository() {
	repo := NewEventRepository(suite.db)
	suite.NotNil(repo)
}

func TestEventRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(EventRepositoryTestSuite))
}