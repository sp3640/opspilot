package models

import "time"

type Comment struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Content    string `gorm:"not null" json:"content"`
	IncidentID uint   `gorm:"not null;index" json:"incident_id"`
	UserID     uint   `gorm:"not null;index" json:"user_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
