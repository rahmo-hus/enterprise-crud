package venue

import (
	"time"

	"github.com/google/uuid"
)

// CreateVenueRequest represents the request structure for creating a new venue
type CreateVenueRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Address     string `json:"address" binding:"required,min=1"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
	Description string `json:"description,omitempty"`
}

// UpdateVenueRequest represents the request structure for updating a venue
type UpdateVenueRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Address     string `json:"address" binding:"required,min=1"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
	Description string `json:"description,omitempty"`
}

// VenueResponse represents the response structure for venue operations
type VenueResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Capacity    int       `json:"capacity"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VenueListResponse represents the response structure for listing venues
type VenueListResponse struct {
	Venues []VenueResponse `json:"venues"`
	Count  int             `json:"count"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents success response structure
type SuccessResponse struct {
	Message string `json:"message"`
}
