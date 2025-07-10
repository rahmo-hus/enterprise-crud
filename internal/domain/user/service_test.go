package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MockRepository is a mock implementation of Repository interface
// Used for testing service layer without database dependencies
type MockRepository struct {
	mock.Mock
}

// Create mocks the Create method of Repository interface
// Returns error based on test scenario configuration
func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// GetByEmail mocks the GetByEmail method of Repository interface  
// Returns user and error based on test scenario configuration
func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*User), args.Error(1)
}

// TestUserService_CreateUser tests the CreateUser method
// Covers successful creation, existing user, and error scenarios
func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name     string              // Test case name
		email    string              // Input email
		username string              // Input username
		password string              // Input password
		mockFunc func(*MockRepository) // Mock repository setup function
		wantErr  bool                // Expected error occurrence
		errMsg   string              // Expected error message
	}{
		{
			name:     "successful user creation",
			email:    "test@example.com",
			username: "testuser",
			password: "password123",
			mockFunc: func(m *MockRepository) {
				// Mock GetByEmail to return "not found" error (user doesn't exist)
				m.On("GetByEmail", mock.Anything, "test@example.com").Return((*User)(nil), gorm.ErrRecordNotFound)
				// Mock Create to succeed
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "user already exists",
			email:    "existing@example.com",
			username: "existinguser",
			password: "password123",
			mockFunc: func(m *MockRepository) {
				// Mock GetByEmail to return existing user
				existingUser := &User{
					ID:       uuid.New(),
					Email:    "existing@example.com",
					Username: "existinguser",
				}
				m.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			wantErr: true,
			errMsg:  "user with email existing@example.com already exists",
		},
		{
			name:     "repository create error",
			email:    "test@example.com",
			username: "testuser",
			password: "password123",
			mockFunc: func(m *MockRepository) {
				// Mock GetByEmail to return "not found" error
				m.On("GetByEmail", mock.Anything, "test@example.com").Return((*User)(nil), gorm.ErrRecordNotFound)
				// Mock Create to fail
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "failed to create user: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := new(MockRepository)
			tt.mockFunc(mockRepo)

			// Create service with mock repository
			service := NewUserService(mockRepo)

			// Execute test
			result, err := service.CreateUser(context.Background(), tt.email, tt.username, tt.password)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
				assert.Equal(t, tt.username, result.Username)
				assert.NotEmpty(t, result.ID)
				
				// Verify password is hashed
				err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(tt.password))
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUserService_GetUserByEmail tests the GetUserByEmail method
// Covers successful retrieval and error scenarios
func TestUserService_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name     string              // Test case name
		email    string              // Input email
		mockFunc func(*MockRepository) // Mock repository setup function
		wantErr  bool                // Expected error occurrence
		errMsg   string              // Expected error message
	}{
		{
			name:  "successful user retrieval",
			email: "test@example.com",
			mockFunc: func(m *MockRepository) {
				// Mock GetByEmail to return user
				user := &User{
					ID:       uuid.New(),
					Email:    "test@example.com",
					Username: "testuser",
				}
				m.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			mockFunc: func(m *MockRepository) {
				// Mock GetByEmail to return not found error
				m.On("GetByEmail", mock.Anything, "notfound@example.com").Return((*User)(nil), gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errMsg:  "failed to get user by email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := new(MockRepository)
			tt.mockFunc(mockRepo)

			// Create service with mock repository
			service := NewUserService(mockRepo)

			// Execute test
			result, err := service.GetUserByEmail(context.Background(), tt.email)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}