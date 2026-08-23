package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationSLO is the SLO target an organization has configured for one
// application - the "Current SLO" every other SRE metric is measured
// against. One row per application (see the uniqueIndex on ApplicationID).
type ApplicationSLO struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID   uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`
	ApplicationID    uuid.UUID `json:"application_id" gorm:"type:uuid;not null;uniqueIndex"`
	TargetPercentage float64   `json:"target_percentage" gorm:"not null"`
	WindowDays       int       `json:"window_days" gorm:"not null"`
	CreatedBy        uint      `json:"created_by" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (a *ApplicationSLO) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	return nil
}
