package database

import (
	"context"
	"enterprise-crud/internal/domain/user"
	"gorm.io/gorm"
)

// userRepository implements the user.Repository interface
// Handles database operations for user entities
type userRepository struct {
	db *gorm.DB // Database connection instance
}

// NewUserRepository creates a new instance of userRepository
// Returns a repository implementation for user operations
func NewUserRepository(db *gorm.DB) user.Repository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database
// Returns error if user creation fails or constraints are violated
func (r *userRepository) Create(ctx context.Context, user *user.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByEmail retrieves a user by their email address
// Returns user if found, nil and error if not found or database error occurs
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
