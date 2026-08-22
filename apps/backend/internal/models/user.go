package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/rbac"
)

const (
	// RoleUser is the legacy pre-RBAC role value. It is no longer assigned to
	// new users (see UserService.Register) and holds no permissions under
	// the rbac matrix; it is kept only so existing code/tests can recognize
	// not-yet-migrated rows. The startup role-normalization migration
	// rewrites it to RoleViewer.
	RoleUser = "User"

	RolePlatformAdmin  = string(rbac.RolePlatformAdmin)
	RoleDevOpsEngineer = string(rbac.RoleDevOpsEngineer)
	RoleDeveloper      = string(rbac.RoleDeveloper)
	RoleViewer         = string(rbac.RoleViewer)
)

type User struct {
	ID             uint          `gorm:"primaryKey"`
	Name           string        `gorm:"size:100;not null"`
	Email          string        `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash   string        `gorm:"not null"`
	Role           string        `gorm:"size:50;not null;default:'Viewer'"`
	OrganizationID *uuid.UUID    `gorm:"type:uuid;index:idx_users_organization_id"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
