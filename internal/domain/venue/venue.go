package venue

import (
	"time"

	"github.com/google/uuid"
)

// Venue represents a location where events can be held
type Venue struct {
	// ID is the unique identifier for each venue
	ID uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`

	// Name is the venue name
	Name string `gorm:"not null;size:255" json:"name" binding:"required"`

	// Address is the venue address
	Address string `gorm:"not null;type:text" json:"address" binding:"required"`

	// Capacity is the maximum number of people the venue can hold
	Capacity int `gorm:"not null;check:capacity > 0" json:"capacity" binding:"required,min=1"`

	// Description provides additional information about the venue
	Description string `gorm:"type:text" json:"description"`

	// Timestamps track when the venue was created and last updated
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName tells GORM what table to use for this model
func (Venue) TableName() string {
	return "venues"
}
