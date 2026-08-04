package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Slug        string    `json:"slug" gorm:"size:120;not null;uniqueIndex:idx_organizations_slug"`
	Description string    `json:"description" gorm:"type:text;not null;default:''"`
	OwnerID     uint      `json:"owner_id" gorm:"not null;index:idx_organizations_owner_id"`
	Owner       *User     `json:"-" gorm:"foreignKey:OwnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a UUID when an organization is created without an explicit ID.
func (o *Organization) BeforeCreate(_ *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}

	return nil
}
