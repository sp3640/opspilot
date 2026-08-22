package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationTeam records that a team owns/is associated with an
// application, mirroring ProjectTeam's junction-table shape for the
// Application resource. A team can own multiple applications, and an
// application can be owned by multiple teams.
type ApplicationTeam struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID    `json:"organization_id" gorm:"type:uuid;not null;index:idx_application_teams_organization_id;uniqueIndex:idx_application_teams_org_app_team,priority:1"`
	ApplicationID  uuid.UUID    `json:"application_id" gorm:"type:uuid;not null;index:idx_application_teams_application_id;uniqueIndex:idx_application_teams_org_app_team,priority:2"`
	Application    *Application `json:"-" gorm:"foreignKey:ApplicationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TeamID         uuid.UUID    `json:"team_id" gorm:"type:uuid;not null;index:idx_application_teams_team_id;uniqueIndex:idx_application_teams_org_app_team,priority:3"`
	Team           *Team        `json:"-" gorm:"foreignKey:TeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt      time.Time    `json:"created_at"`
}

func (at *ApplicationTeam) BeforeCreate(_ *gorm.DB) error {
	if at.ID == uuid.Nil {
		at.ID = uuid.New()
	}

	return nil
}
