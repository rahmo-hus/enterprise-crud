package user

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id" gorm:"primaryKey, type:uuid"`
	Email    string    `json:"email" gorm:"unique; not null"`
	Username string    `json:"username" gorm:"unique; not null"`
	Password string    `json:"-" gorm:"not null"`
}
