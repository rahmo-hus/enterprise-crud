package user

import "context"

// Repository defines the data access interface for user operations
// This is the repository pattern similar to Spring Data JPA repositories
// Abstracts database operations and provides a clean interface for data access
type Repository interface {
	Create(ctx context.Context, user *User) error                // Persists a new user to the database
	GetByEmail(ctx context.Context, email string) (*User, error) // Retrieves a user by their email address
}
