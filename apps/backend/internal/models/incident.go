package models

import (
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Title          string    `gorm:"size:255;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	Severity       string    `gorm:"size:20;not null" json:"severity"`
	Status         string    `gorm:"size:20;not null" json:"status"`

	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`

	// ApplicationID is the affected application, when known - optional,
	// since not every incident (e.g. a cluster-wide outage) is scoped to a
	// single application.
	ApplicationID *uuid.UUID   `gorm:"type:uuid;index" json:"application_id,omitempty"`
	Application   *Application `gorm:"foreignKey:ApplicationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`

	// OwnerTeamID is the team responsible for driving this incident to
	// resolution - optional, set explicitly by whoever creates/triages it.
	OwnerTeamID *uuid.UUID `gorm:"type:uuid;index" json:"owner_team_id,omitempty"`
	OwnerTeam   *Team      `gorm:"foreignKey:OwnerTeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`

	UserID uint `gorm:"not null;index" json:"user_id"`

	// ResolvedAt is set when Status transitions to RESOLVED and cleared if
	// the incident is reopened to any other status.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
