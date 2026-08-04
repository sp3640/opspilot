package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	InvitationStatusPending  = "Pending"
	InvitationStatusAccepted = "Accepted"
	InvitationStatusExpired  = "Expired"
	InvitationStatusRevoked  = "Revoked"
)

type Invitation struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index:idx_invitations_organization_id"`
	Email          string     `json:"email" gorm:"size:255;not null;index:idx_invitations_org_email,priority:2"`
	Role           string     `json:"role" gorm:"size:50;not null"`
	Token          string     `json:"token" gorm:"size:255;not null;uniqueIndex:idx_invitations_token"`
	Status         string     `json:"status" gorm:"size:20;not null;index:idx_invitations_status;index:idx_invitations_org_email,priority:1"`
	InvitedBy      uint       `json:"invited_by" gorm:"not null;index:idx_invitations_invited_by"`
	ExpiresAt      time.Time  `json:"expires_at" gorm:"not null;index:idx_invitations_expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (i *Invitation) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}

	return nil
}
