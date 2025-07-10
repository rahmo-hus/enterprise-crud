package user

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the business logic interface for user operations
// Handles user-related business rules and validations
type Service interface {
	CreateUser(ctx context.Context, email, username, password string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

// userService implements the Service interface
// Encapsulates business logic for user operations
type userService struct {
	repo Repository // Repository for data persistence
}

// NewUserService creates a new instance of userService
// Returns a service implementation for user business logic
func NewUserService(repo Repository) Service {
	return &userService{repo: repo}
}

// CreateUser creates a new user with the provided information
// Validates input, hashes password, and persists user data
func (s *userService) CreateUser(ctx context.Context, email, username, password string) (*User, error) {
	// Check if user already exists with this email
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Hash the password for secure storage
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user entity
	user := &User{
		ID:       uuid.New(),          // Generate unique identifier
		Email:    email,               // Set email address
		Username: username,            // Set username
		Password: string(hashedPassword), // Store hashed password
	}

	// Persist user to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email address
// Returns user if found, error if not found or database error occurs
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}