package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationChannel is an organization- (or, optionally, team-) scoped
// destination for operational notifications. TeamID nil means the channel
// receives every subscribed event for the whole organization; a non-nil
// TeamID additionally scopes it to events tied to that team (currently only
// incidents carry a team, via OwnerTeamID - see NotificationService.Dispatch).
type NotificationChannel struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index"`
	TeamID         *uuid.UUID `json:"team_id,omitempty" gorm:"type:uuid;index"`
	Team           *Team      `json:"-" gorm:"foreignKey:TeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name           string     `json:"name" gorm:"size:150;not null"`
	// Type is one of constants.NotificationChannel* (EMAIL/SLACK/TEAMS/WEBHOOK).
	Type string `json:"type" gorm:"size:20;not null;index"`
	// EncryptedConfig is an AES-256-GCM ciphertext (security.ClusterCredentialCipher)
	// of a small JSON blob whose shape depends on Type - {"email":"..."} for
	// EMAIL, {"url":"..."} for SLACK/TEAMS/WEBHOOK. Never decrypted for
	// display: the API only ever returns a masked target derived from it.
	EncryptedConfig string `json:"-" gorm:"type:text;not null"`
	// Events is a JSON array of constants.NotificationEvent* strings this
	// channel is subscribed to.
	Events    string         `json:"-" gorm:"type:text;not null;default:'[]'"`
	Enabled   bool           `json:"enabled" gorm:"not null;default:true;index"`
	CreatedBy uint           `json:"created_by" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (n *NotificationChannel) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}

	return nil
}
