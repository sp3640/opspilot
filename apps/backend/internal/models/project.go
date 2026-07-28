package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Slug        string    `json:"slug" gorm:"size:120;not null;uniqueIndex:idx_projects_slug"`
	Description string    `json:"description" gorm:"type:text;not null;default:''"`
	Environment string    `json:"environment" gorm:"size:20;not null;default:development;index:idx_projects_environment"`
	Health      string    `json:"health" gorm:"size:20;not null;default:healthy;index:idx_projects_health"`
	OwnerID     uint      `json:"owner_id" gorm:"not null;index:idx_projects_owner_id"`
	Members     int       `json:"members" gorm:"not null;default:0;check:chk_projects_members_non_negative,members >= 0"`
	Services    int       `json:"services" gorm:"not null;default:0;check:chk_projects_services_non_negative,services >= 0"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a UUID when a project is created without an explicit ID.
func (p *Project) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	return nil
}
