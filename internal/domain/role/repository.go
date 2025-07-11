package role

import "context"

// Repository defines the data access interface for role operations
// This allows us to get roles from the database
type Repository interface {
	GetByName(ctx context.Context, name string) (*Role, error) // Gets a role by its name (like "USER", "ADMIN")
}
