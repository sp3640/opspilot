package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Deployment struct {
	ID                 uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;not null"`
	ApplicationID      uuid.UUID    `json:"application_id" gorm:"type:uuid;not null;index:idx_deployments_application_id;index:idx_deployments_org_app_created,priority:2"`
	Application        *Application `json:"-" gorm:"foreignKey:ApplicationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ProjectID          uuid.UUID    `json:"project_id" gorm:"type:uuid;not null;index:idx_deployments_project_id;index:idx_deployments_org_project_created,priority:2"`
	Project            *Project     `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	OrganizationID     uuid.UUID    `json:"organization_id" gorm:"type:uuid;not null;index:idx_deployments_organization_id;index:idx_deployments_org_created,priority:1;index:idx_deployments_org_app_created,priority:1;index:idx_deployments_org_project_created,priority:1"`
	Image              string       `json:"image" gorm:"size:500;not null"`
	ImageTag           string       `json:"image_tag" gorm:"size:128;not null;default:''"`
	Environment        string       `json:"environment" gorm:"size:32;not null;index:idx_deployments_environment"`
	Namespace          string       `json:"namespace" gorm:"size:63;not null;index:idx_deployments_namespace"`
	ReplicaCount       int          `json:"replica_count" gorm:"not null;check:chk_deployments_replica_count,replica_count > 0"`
	Status             string       `json:"status" gorm:"size:32;not null;index:idx_deployments_status"`
	DeploymentStrategy string       `json:"deployment_strategy" gorm:"size:64;not null"`
	TargetClusterID    uuid.UUID    `json:"target_cluster_id" gorm:"type:uuid;not null;index:idx_deployments_target_cluster_id"`
	TargetCluster      *Cluster     `json:"-" gorm:"foreignKey:TargetClusterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	// CommitSHA and Author are optional provenance a caller may supply (e.g.
	// manually, or by a future CI/CD integration) - there is no automatic
	// source for either today, so both stay nil unless explicitly provided.
	CommitSHA   *string        `json:"commit_sha,omitempty" gorm:"size:64"`
	Author      *string        `json:"author,omitempty" gorm:"size:255"`
	CreatedBy   uint           `json:"created_by" gorm:"not null;index:idx_deployments_created_by"`
	UpdatedBy   uint           `json:"updated_by" gorm:"not null;index:idx_deployments_updated_by"`
	StartedAt   *time.Time     `json:"started_at,omitempty" gorm:"index:idx_deployments_started_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty" gorm:"index:idx_deployments_completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (d *Deployment) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	return nil
}
