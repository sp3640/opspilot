package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Application struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_applications_organization_id"`
	ProjectID      uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index:idx_applications_project_id;uniqueIndex:idx_applications_project_slug,priority:1"`
	Project        *Project  `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Name           string    `json:"name" gorm:"size:255;not null;index:idx_applications_name"`
	Slug           string    `json:"slug" gorm:"size:120;not null;uniqueIndex:idx_applications_project_slug,priority:2;index:idx_applications_slug"`
	Description    string    `json:"description" gorm:"type:text;not null;default:''"`
	RepositoryURL  string    `json:"repository_url" gorm:"size:500;not null;default:'';index:idx_applications_repository_url"`
	DefaultBranch  string    `json:"default_branch" gorm:"size:128;not null;default:'main'"`
	Runtime        string    `json:"runtime" gorm:"size:32;not null;index:idx_applications_runtime"`
	BuildCommand   string    `json:"build_command" gorm:"type:text;not null;default:''"`
	StartCommand   string    `json:"start_command" gorm:"type:text;not null;default:''"`
	Port           int       `json:"port" gorm:"not null;check:chk_applications_port,port >= 1 AND port <= 65535"`
	Environment    string    `json:"environment" gorm:"size:255;not null;default:''"`
	Status         string    `json:"status" gorm:"size:32;not null;default:'Draft';index:idx_applications_status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Application) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	return nil
}
