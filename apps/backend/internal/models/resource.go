package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Resource struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID   uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index:idx_resources_organization_id"`
	ProjectID        uuid.UUID       `json:"project_id" gorm:"type:uuid;not null;index:idx_resources_project_id;index:idx_resources_project_external,priority:1;index:idx_resources_project_kind_name,priority:1"`
	Project          *Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ParentResourceID *uuid.UUID      `json:"parent_resource_id,omitempty" gorm:"type:uuid;index:idx_resources_parent_resource_id"`
	Parent           *Resource       `json:"-" gorm:"foreignKey:ParentResourceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Kind             string          `json:"kind" gorm:"size:64;not null;index:idx_resources_kind;index:idx_resources_project_kind_name,priority:2"`
	Name             string          `json:"name" gorm:"size:255;not null;index:idx_resources_name;index:idx_resources_project_kind_name,priority:3"`
	DisplayName      string          `json:"display_name" gorm:"size:255;not null;default:''"`
	ExternalID       string          `json:"external_id" gorm:"size:255;not null;default:'';index:idx_resources_project_external,priority:2"`
	Provider         string          `json:"provider" gorm:"size:64;not null;default:'';index:idx_resources_provider"`
	Region           string          `json:"region" gorm:"size:128;not null;default:'';index:idx_resources_region"`
	Namespace        string          `json:"namespace" gorm:"size:255;not null;default:'';index:idx_resources_namespace"`
	Cluster          string          `json:"cluster" gorm:"size:255;not null;default:'';index:idx_resources_cluster"`
	Status           string          `json:"status" gorm:"size:32;not null;default:UNKNOWN;index:idx_resources_status"`
	Health           string          `json:"health" gorm:"size:32;not null;default:UNKNOWN;index:idx_resources_health"`
	Labels           json.RawMessage `json:"labels" gorm:"type:jsonb;not null;default:'{}'"`
	Annotations      json.RawMessage `json:"annotations" gorm:"type:jsonb;not null;default:'{}'"`
	Metadata         json.RawMessage `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`
	CreatedBy        uint            `json:"created_by" gorm:"not null;index:idx_resources_created_by"`
	CreatedByUser    *User           `json:"-" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a UUID when a resource is created without an explicit ID.
func (r *Resource) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	return nil
}
