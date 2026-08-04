package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser          = "User"
	RolePlatformAdmin = "Platform Admin"
)

type User struct {
	ID             uint          `gorm:"primaryKey"`
	Name           string        `gorm:"size:100;not null"`
	Email          string        `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash   string        `gorm:"not null"`
	Role           string        `gorm:"size:50;not null;default:'User'"`
	OrganizationID *uuid.UUID    `gorm:"type:uuid;index:idx_users_organization_id"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
