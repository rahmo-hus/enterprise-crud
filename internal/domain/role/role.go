package role

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user role in the system (like ADMIN or USER)
// This is a simple role model that defines what a user can do
type Role struct {
	// ID is the unique identifier for each role
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`

	// Name is the role name (ADMIN, USER, etc.)
	// Must be unique - no two roles can have the same name
	Name string `gorm:"unique;not null;size:50" json:"name" binding:"required"`

	// Description explains what this role can do
	Description string `gorm:"type:text" json:"description"`

	// Timestamps track when the role was created and last updated
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Common role names used in the application
const (
	RoleAdmin     = "ADMIN"     // Administrator with full access
	RoleUser      = "USER"      // Regular user with basic access
	RoleOrganizer = "ORGANIZER" // Event organizer with event management access
)

// TableName tells GORM what table to use for this model
func (Role) TableName() string {
	return "roles"
}
