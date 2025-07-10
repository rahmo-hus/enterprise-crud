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
//
// CONTEXT USAGE:
// - ctx: Allows the database operation to be cancelled or timeout
// - WithContext(ctx): Enables request tracing, deadlines, and cancellation
// - If HTTP request is cancelled, this DB operation will also be cancelled
// - Prevents hanging database connections and resource leaks
//
// ERROR HANDLING:
// - .Error: GORM returns the last error that occurred
// - Returns nil if successful, error if failed
// - Handles constraint violations (unique email/username)
// - Database connection errors, validation errors, etc.
//
// Returns error if user creation fails or constraints are violated
func (r *userRepository) Create(ctx context.Context, user *user.User) error {
	// WithContext(ctx) ensures this DB operation:
	// 1. Can be cancelled if the HTTP request is cancelled
	// 2. Will timeout if the context has a deadline
	// 3. Allows distributed tracing across services
	// 4. Prevents resource leaks and hanging connections
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByEmail retrieves a user by their email address
//
// CONTEXT USAGE:
// - Same context benefits as Create method
// - Enables cancellation and timeout for SELECT queries
// - Essential for long-running queries or network delays
//
// SQL QUERY EXPLANATION:
// - Where("email = ?", email): Parameterized query prevents SQL injection
// - First(&u): Retrieves the first matching record
// - .Error: Returns error if no record found or database error
//
// GORM BEHAVIOR:
// - First() returns ErrRecordNotFound if no user exists
// - Where() adds WHERE clause to the SQL query
// - The ? placeholder is safely replaced with the email parameter
//
// Returns user if found, nil and error if not found or database error occurs
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User

	// WithContext(ctx) + Where() + First() sequence:
	// 1. WithContext(ctx): Enables cancellation and tracing
	// 2. Where("email = ?", email): Adds WHERE clause (SQL injection safe)
	// 3. First(&u): Executes SELECT query and scans result into u
	// 4. .Error: Returns the error (nil if successful, ErrRecordNotFound if no match)
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err // Return nil user and the error (could be ErrRecordNotFound)
	}
	return &u, nil // Return pointer to user and nil error
}
