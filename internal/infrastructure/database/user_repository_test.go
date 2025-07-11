package database

import (
	"context"
	"testing"

	"enterprise-crud/internal/domain/user"
	"github.com/google/uuid"
	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// UserRepositoryTestSuite is a test suite for user repository
// Uses in-memory SQLite database for testing
type UserRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB        // Test database instance
	repo user.Repository // Repository under test
}

// SetupSuite runs before all tests in the suite
// Initializes test database and repository
func (suite *UserRepositoryTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing with custom configuration
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	suite.Require().NoError(err)

	// Create tables manually for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS roles (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	suite.Require().NoError(err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	suite.Require().NoError(err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_roles (
			user_id TEXT,
			role_id TEXT,
			PRIMARY KEY (user_id, role_id)
		)
	`).Error
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewUserRepository(db)
}

// TearDownSuite runs after all tests in the suite
// Cleans up test database
func (suite *UserRepositoryTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// SetupTest runs before each test case
// Cleans up test data
func (suite *UserRepositoryTestSuite) SetupTest() {
	// Clean up test data
	suite.db.Exec("DELETE FROM users")
}

// TestCreate tests the Create method of user repository
// Covers successful creation and constraint violations
func (suite *UserRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	// Test successful user creation
	testUser := &user.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword",
	}

	err := suite.repo.Create(ctx, testUser)
	suite.NoError(err)

	// Verify user was created in database
	var createdUser user.User
	err = suite.db.Where("email = ?", testUser.Email).First(&createdUser).Error
	suite.NoError(err)
	suite.Equal(testUser.Email, createdUser.Email)
	suite.Equal(testUser.Username, createdUser.Username)

	// Test duplicate email constraint
	duplicateUser := &user.User{
		ID:       uuid.New(),
		Email:    "test@example.com", // Same email as previous user
		Username: "differentuser",
		Password: "hashedpassword",
	}

	err = suite.repo.Create(ctx, duplicateUser)
	suite.Error(err) // Should fail due to unique constraint
}

// TestGetByEmail tests the GetByEmail method of user repository
// Covers successful retrieval and not found scenarios
func (suite *UserRepositoryTestSuite) TestGetByEmail() {
	ctx := context.Background()

	// Create test user in database
	testUser := &user.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword",
	}

	err := suite.db.Create(testUser).Error
	suite.Require().NoError(err)

	// Test successful retrieval
	foundUser, err := suite.repo.GetByEmail(ctx, "test@example.com")
	suite.NoError(err)
	suite.NotNil(foundUser)
	suite.Equal(testUser.Email, foundUser.Email)
	suite.Equal(testUser.Username, foundUser.Username)
	suite.Equal(testUser.ID, foundUser.ID)

	// Test user not found
	notFoundUser, err := suite.repo.GetByEmail(ctx, "nonexistent@example.com")
	suite.Error(err)
	suite.Nil(notFoundUser)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestGetByEmail_WithMultipleUsers tests retrieval with multiple users
// Ensures correct user is returned when multiple users exist
func (suite *UserRepositoryTestSuite) TestGetByEmail_WithMultipleUsers() {
	ctx := context.Background()

	// Create multiple test users
	users := []*user.User{
		{
			ID:       uuid.New(),
			Email:    "user1@example.com",
			Username: "user1",
			Password: "hashedpassword1",
		},
		{
			ID:       uuid.New(),
			Email:    "user2@example.com",
			Username: "user2",
			Password: "hashedpassword2",
		},
		{
			ID:       uuid.New(),
			Email:    "user3@example.com",
			Username: "user3",
			Password: "hashedpassword3",
		},
	}

	// Insert all users
	for _, u := range users {
		err := suite.db.Create(u).Error
		suite.Require().NoError(err)
	}

	// Test retrieval of specific user
	foundUser, err := suite.repo.GetByEmail(ctx, "user2@example.com")
	suite.NoError(err)
	suite.NotNil(foundUser)
	suite.Equal("user2@example.com", foundUser.Email)
	suite.Equal("user2", foundUser.Username)
}

// TestNewUserRepository tests the constructor function
// Verifies repository instance creation
func (suite *UserRepositoryTestSuite) TestNewUserRepository() {
	repo := NewUserRepository(suite.db)
	suite.NotNil(repo)
	suite.Implements((*user.Repository)(nil), repo)
}

// TestUserRepositoryTestSuite runs the test suite
// Entry point for running all repository tests
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
