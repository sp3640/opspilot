package models

import "time"

type Incident struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"size:255;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Severity    string `gorm:"size:20;not null" json:"severity"`
	Status      string `gorm:"size:20;not null" json:"status"`

	ProjectID uint `gorm:"not null;index" json:"project_id"`
	UserID    uint `gorm:"not null;index" json:"user_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
