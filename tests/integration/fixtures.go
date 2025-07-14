//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"enterprise-crud/internal/domain/event"
	"enterprise-crud/internal/domain/order"
	"enterprise-crud/internal/domain/role"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/domain/venue"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestFixtures provides test data creation helpers
type TestFixtures struct {
	db *TestDatabase
}

// NewTestFixtures creates a new test fixtures instance
func NewTestFixtures(db *TestDatabase) *TestFixtures {
	return &TestFixtures{db: db}
}

// CreateRole creates a test role
func (f *TestFixtures) CreateRole(t *testing.T, name, description string) *role.Role {
	roleEntity := &role.Role{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
	}

	err := f.db.DB.Create(roleEntity).Error
	require.NoError(t, err, "Failed to create test role")

	return roleEntity
}

// CreateUser creates a test user with roles
func (f *TestFixtures) CreateUser(t *testing.T, email, username, password string, roles ...*role.Role) *user.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password")

	userEntity := &user.User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
		Roles:    make([]role.Role, len(roles)),
	}

	// Convert role pointers to values
	for i, r := range roles {
		userEntity.Roles[i] = *r
	}

	err = f.db.DB.Create(userEntity).Error
	require.NoError(t, err, "Failed to create test user")

	return userEntity
}

// CreateVenue creates a test venue
func (f *TestFixtures) CreateVenue(t *testing.T, name string, capacity int) *venue.Venue {
	venueEntity := &venue.Venue{
		ID:       uuid.New(),
		Name:     name,
		Address:  "Test Location",
		Capacity: capacity,
	}

	err := f.db.DB.Create(venueEntity).Error
	require.NoError(t, err, "Failed to create test venue")

	return venueEntity
}

// CreateEvent creates a test event
func (f *TestFixtures) CreateEvent(t *testing.T, venue *venue.Venue, organizer *user.User, title string, ticketPrice float64, totalTickets int) *event.Event {
	eventEntity := &event.Event{
		ID:               uuid.New(),
		VenueID:          venue.ID,
		OrganizerID:      organizer.ID,
		Title:            title,
		Description:      "Test event description",
		EventDate:        time.Now().Add(24 * time.Hour), // Tomorrow
		TicketPrice:      ticketPrice,
		AvailableTickets: totalTickets,
		TotalTickets:     totalTickets,
		Status:           event.StatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := f.db.DB.Create(eventEntity).Error
	require.NoError(t, err, "Failed to create test event")

	return eventEntity
}

// CreateOrder creates a test order
func (f *TestFixtures) CreateOrder(t *testing.T, user *user.User, event *event.Event, quantity int) *order.Order {
	orderEntity := &order.Order{
		ID:          uuid.New(),
		UserID:      user.ID,
		EventID:     event.ID,
		Quantity:    quantity,
		TotalAmount: event.TicketPrice * float64(quantity),
		Status:      order.StatusPending,
		CreatedAt:   time.Now(),
	}

	err := f.db.DB.Create(orderEntity).Error
	require.NoError(t, err, "Failed to create test order")

	return orderEntity
}

// CreateCompleteTestData creates a complete set of test data (roles, user, venue, event, order)
func (f *TestFixtures) CreateCompleteTestData(t *testing.T) (*user.User, *event.Event, *order.Order) {
	// Create roles
	userRole := f.CreateRole(t, "USER", "Regular user")
	organizerRole := f.CreateRole(t, "ORGANIZER", "Event organizer")

	// Create users
	regularUser := f.CreateUser(t, "user@test.com", "testuser", "password123", userRole)
	organizer := f.CreateUser(t, "organizer@test.com", "testorganizer", "password123", organizerRole)

	// Create venue
	venue := f.CreateVenue(t, "Test Venue", 100)

	// Create event
	event := f.CreateEvent(t, venue, organizer, "Test Event", 50.0, 100)

	// Create order
	order := f.CreateOrder(t, regularUser, event, 2)

	return regularUser, event, order
}

// StandardRoles creates the standard application roles
func (f *TestFixtures) StandardRoles(t *testing.T) (*role.Role, *role.Role, *role.Role) {
	userRole := f.CreateRole(t, "USER", "Regular user role")
	organizerRole := f.CreateRole(t, "ORGANIZER", "Event organizer role")
	adminRole := f.CreateRole(t, "ADMIN", "Administrator role")

	return userRole, organizerRole, adminRole
}
