package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_teams_organization_id;uniqueIndex:idx_teams_org_name,priority:1"`
	Name           string    `json:"name" gorm:"size:100;not null;uniqueIndex:idx_teams_org_name,priority:2"`
	Description    string    `json:"description" gorm:"type:text;not null;default:''"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate assigns a UUID when a team is created without an explicit ID.
func (t *Team) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	return nil
}
