package models

import (
	"time"
)

type User struct {
	UserID    string    `gorm:"primaryKey" json:"user_id"`
	Nama      string    `gorm:"size:255;not null" json:"nama"`
	Lab       string    `gorm:"size:255;not null" json:"lab"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
