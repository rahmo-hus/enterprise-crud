package ticket

import (
	"github.com/google/uuid"
)

// Ticket represents a physical or digital ticket for an event
type Ticket struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	OrderID  uuid.UUID `gorm:"not null;type:uuid" json:"order_id"`
	EventID  uuid.UUID `gorm:"not null;type:uuid" json:"event_id"`
	UserID   uuid.UUID `gorm:"not null;type:uuid" json:"user_id"`
	SeatInfo *string   `gorm:"size:255" json:"seat_info,omitempty"` // Nullable
	QRCode   string    `gorm:"size:255;not null;unique" json:"qr_code"`
}

// TableName tells GORM what table to use for this model
func (Ticket) TableName() string {
	return "tickets"
}
