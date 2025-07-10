package user

import "github.com/google/uuid"

// User represents the core user entity in the domain layer
// This is the central business object similar to a JPA Entity in Spring Boot
// Contains all user-related data and business rules
type User struct {
	ID       uuid.UUID `json:"id" gorm:"primaryKey, type:uuid"`  // Unique identifier, auto-generated UUID primary key
	Email    string    `json:"email" gorm:"unique; not null"`    // User email address, must be unique across system
	Username string    `json:"username" gorm:"unique; not null"` // User chosen username, must be unique across system
	Password string    `json:"-" gorm:"not null"`                // Encrypted password, excluded from JSON serialization for security
}
