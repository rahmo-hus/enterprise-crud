package database

import (
	"context"
	"testing"

	"enterprise-crud/internal/domain/venue"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// VenueRepositoryTestSuite is a test suite for venue repository
type VenueRepositoryTestSuite struct {
	suite.Suite
	db        *gorm.DB
	venueRepo venue.Repository
}

// SetupSuite runs before all tests in the suite
func (suite *VenueRepositoryTestSuite) SetupSuite() {
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

	suite.db = db
	suite.venueRepo = NewVenueRepository(db)
}

// TearDownSuite runs after all tests in the suite
func (suite *VenueRepositoryTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	suite.Require().NoError(err)
	sqlDB.Close()
}

// SetupTest runs before each test
func (suite *VenueRepositoryTestSuite) SetupTest() {
	// Clear all tables before each test
	suite.db.Exec("DELETE FROM venues")
}

func (suite *VenueRepositoryTestSuite) TestCreate() {
	testVenue := &venue.Venue{
		ID:          uuid.New(),
		Name:        "Test Venue",
		Address:     "123 Test Street",
		Capacity:    100,
		Description: "A test venue",
	}

	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, testVenue.ID)
}

func (suite *VenueRepositoryTestSuite) TestGetByID() {
	// Create and save a venue
	testVenue := &venue.Venue{
		ID:          uuid.New(),
		Name:        "Test Venue",
		Address:     "123 Test Street",
		Capacity:    100,
		Description: "A test venue",
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Test successful retrieval
	foundVenue, err := suite.venueRepo.GetByID(context.Background(), testVenue.ID)
	suite.NoError(err)
	suite.Equal(testVenue.ID, foundVenue.ID)
	suite.Equal(testVenue.Name, foundVenue.Name)
	suite.Equal(testVenue.Address, foundVenue.Address)
	suite.Equal(testVenue.Capacity, foundVenue.Capacity)

	// Test venue not found
	nonExistentID := uuid.New()
	foundVenue, err = suite.venueRepo.GetByID(context.Background(), nonExistentID)
	suite.Error(err)
	suite.Nil(foundVenue)
	suite.True(venue.IsVenueNotFoundError(err))
}

func (suite *VenueRepositoryTestSuite) TestGetAll() {
	// Create multiple venues
	venues := []*venue.Venue{
		{
			ID:          uuid.New(),
			Name:        "Venue 1",
			Address:     "123 Test Street",
			Capacity:    100,
			Description: "First venue",
		},
		{
			ID:          uuid.New(),
			Name:        "Venue 2",
			Address:     "456 Test Avenue",
			Capacity:    200,
			Description: "Second venue",
		},
	}

	// Save all venues
	for _, v := range venues {
		err := suite.venueRepo.Create(context.Background(), v)
		suite.Require().NoError(err)
	}

	// Test GetAll
	allVenues, err := suite.venueRepo.GetAll(context.Background())
	suite.NoError(err)
	suite.Len(allVenues, 2)

	// Verify venues are returned
	venueNames := make([]string, len(allVenues))
	for i, v := range allVenues {
		venueNames[i] = v.Name
	}
	suite.Contains(venueNames, "Venue 1")
	suite.Contains(venueNames, "Venue 2")
}

func (suite *VenueRepositoryTestSuite) TestUpdate() {
	// Create and save a venue
	testVenue := &venue.Venue{
		ID:          uuid.New(),
		Name:        "Original Name",
		Address:     "Original Address",
		Capacity:    100,
		Description: "Original Description",
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Update the venue
	testVenue.Name = "Updated Name"
	testVenue.Address = "Updated Address"
	testVenue.Capacity = 200
	testVenue.Description = "Updated Description"

	err = suite.venueRepo.Update(context.Background(), testVenue)
	suite.NoError(err)

	// Verify update
	updatedVenue, err := suite.venueRepo.GetByID(context.Background(), testVenue.ID)
	suite.NoError(err)
	suite.Equal("Updated Name", updatedVenue.Name)
	suite.Equal("Updated Address", updatedVenue.Address)
	suite.Equal(200, updatedVenue.Capacity)
	suite.Equal("Updated Description", updatedVenue.Description)
}

func (suite *VenueRepositoryTestSuite) TestDelete() {
	// Create and save a venue
	testVenue := &venue.Venue{
		ID:          uuid.New(),
		Name:        "Test Venue",
		Address:     "123 Test Street",
		Capacity:    100,
		Description: "A test venue",
	}
	err := suite.venueRepo.Create(context.Background(), testVenue)
	suite.Require().NoError(err)

	// Delete the venue
	err = suite.venueRepo.Delete(context.Background(), testVenue.ID)
	suite.NoError(err)

	// Verify deletion
	deletedVenue, err := suite.venueRepo.GetByID(context.Background(), testVenue.ID)
	suite.Error(err)
	suite.Nil(deletedVenue)
	suite.True(venue.IsVenueNotFoundError(err))
}

func (suite *VenueRepositoryTestSuite) TestNewVenueRepository() {
	repo := NewVenueRepository(suite.db)
	suite.NotNil(repo)
}

func TestVenueRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(VenueRepositoryTestSuite))
}