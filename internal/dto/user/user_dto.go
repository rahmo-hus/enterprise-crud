package user

import "github.com/google/uuid"

// CreateUserRequest represents the request payload for creating a new user
// Contains required fields for user registration
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"` // User's email address - must be unique
	Username string `json:"username" binding:"required,min=3" example:"john_doe"`      // Username - must be at least 3 characters
	Password string `json:"password" binding:"required,min=8" example:"password123"`   // Password - must be at least 8 characters
}

// UserResponse represents the response payload for user operations
// Excludes sensitive information like passwords
type UserResponse struct {
	ID       uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"` // Unique identifier for the user
	Email    string    `json:"email" example:"user@example.com"`                  // User's email address
	Username string    `json:"username" example:"john_doe"`                       // User's chosen username
	Roles    []string  `json:"roles" example:"USER,ADMIN"`                        // User's roles in the system
}

// LoginRequest represents the request payload for user login
// Contains credentials for authentication
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"` // User's email address
	Password string `json:"password" binding:"required" example:"password123"`         // User's password
}

// LoginResponse represents the response payload for successful login
// Contains user information and JWT token
type LoginResponse struct {
	User      UserResponse `json:"user"`                                                    // User information
	Token     string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT access token
	ExpiresAt int64        `json:"expires_at" example:"1735689600"`                         // Token expiration timestamp
}

// ErrorResponse represents error response structure
// Provides consistent error messaging across the API
type ErrorResponse struct {
	Error   string `json:"error" example:"Error message"`                        // Error message
	Message string `json:"message,omitempty" example:"Additional error details"` // Additional error details
}
