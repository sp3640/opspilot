package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectTeam struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_project_teams_organization_id;uniqueIndex:idx_project_teams_org_project_team,priority:1"`
	ProjectID      uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index:idx_project_teams_project_id;uniqueIndex:idx_project_teams_org_project_team,priority:2"`
	Project        *Project  `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TeamID         uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index:idx_project_teams_team_id;uniqueIndex:idx_project_teams_org_project_team,priority:3"`
	Team           *Team     `json:"-" gorm:"foreignKey:TeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt      time.Time `json:"created_at"`
}

func (pt *ProjectTeam) BeforeCreate(_ *gorm.DB) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}

	return nil
}
