package database

import (
	"context"
	"enterprise-crud/internal/domain/role"
	"gorm.io/gorm"
)

// roleRepository implements the role.Repository interface
// Handles database operations for role entities
type roleRepository struct {
	db *gorm.DB // Database connection instance
}

// NewRoleRepository creates a new instance of roleRepository
// Returns a repository implementation for role operations
func NewRoleRepository(db *gorm.DB) role.Repository {
	return &roleRepository{db: db}
}

// GetByName retrieves a role by its name (like "USER" or "ADMIN")
// This is used when assigning roles to users during registration
func (r *roleRepository) GetByName(ctx context.Context, name string) (*role.Role, error) {
	var roleEntity role.Role

	// Find the role by name in the database
	// This is a simple SELECT WHERE name = ? query
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&roleEntity).Error
	if err != nil {
		return nil, err // Return error if role not found
	}

	return &roleEntity, nil // Return the role if found
}