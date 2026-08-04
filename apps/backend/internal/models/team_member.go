package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamMember struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	TeamID    uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index:idx_team_members_team_id;uniqueIndex:idx_team_members_team_user,priority:1"`
	Team      *Team     `json:"-" gorm:"foreignKey:TeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID    uint      `json:"user_id" gorm:"not null;index:idx_team_members_user_id;uniqueIndex:idx_team_members_team_user,priority:2"`
	User      *User     `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate assigns a UUID when a team membership is created without an explicit ID.
func (tm *TeamMember) BeforeCreate(_ *gorm.DB) error {
	if tm.ID == uuid.Nil {
		tm.ID = uuid.New()
	}

	return nil
}
