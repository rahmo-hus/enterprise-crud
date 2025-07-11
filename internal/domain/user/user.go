package user

import (
	"time"

	"enterprise-crud/internal/domain/role"
	"github.com/google/uuid"
)

// User represents the core user entity in the domain layer
// This is the central business object similar to a JPA Entity in Spring Boot
// Contains all user-related data and business rules
type User struct {
	ID       uuid.UUID `json:"id" gorm:"primaryKey, type:uuid"`  // Unique identifier, auto-generated UUID primary key
	Email    string    `json:"email" gorm:"unique; not null"`    // User email address, must be unique across system
	Username string    `json:"username" gorm:"unique; not null"` // User chosen username, must be unique across system
	Password string    `json:"-" gorm:"not null"`                // Encrypted password, excluded from JSON serialization for security

	// Roles defines what this user can do in the system
	// Many-to-many relationship: one user can have multiple roles, one role can belong to multiple users
	// GORM will automatically handle the user_roles junction table
	Roles []role.Role `json:"roles" gorm:"many2many:user_roles;"`

	// Timestamps track when the user was created and last updated
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
