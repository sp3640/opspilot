package models

import "time"

type Project struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	UserID      uint      `gorm:"not null;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}