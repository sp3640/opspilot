package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeploymentHistory struct {
	ID                 uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;not null"`
	DeploymentID       uuid.UUID   `json:"deployment_id" gorm:"type:uuid;not null;index:idx_deployment_history_deployment_id;index:idx_deployment_history_org_deployment_revision,priority:2"`
	ApplicationID      uuid.UUID   `json:"application_id" gorm:"type:uuid;not null;index:idx_deployment_history_application_id"`
	ProjectID          uuid.UUID   `json:"project_id" gorm:"type:uuid;not null;index:idx_deployment_history_project_id"`
	OrganizationID     uuid.UUID   `json:"organization_id" gorm:"type:uuid;not null;index:idx_deployment_history_organization_id;index:idx_deployment_history_org_deployment_revision,priority:1"`
	Revision           int         `json:"revision" gorm:"not null;check:chk_deployment_history_revision,revision > 0;index:idx_deployment_history_revision;uniqueIndex:idx_deployment_history_org_deployment_revision,priority:3"`
	Image              string      `json:"image" gorm:"size:500;not null"`
	ImageTag           string      `json:"image_tag" gorm:"size:128;not null;default:''"`
	Environment        string      `json:"environment" gorm:"size:32;not null"`
	Namespace          string      `json:"namespace" gorm:"size:63;not null"`
	ReplicaCount       int         `json:"replica_count" gorm:"not null;check:chk_deployment_history_replica_count,replica_count > 0"`
	DeploymentStrategy string      `json:"deployment_strategy" gorm:"size:64;not null"`
	Status             string      `json:"status" gorm:"size:32;not null;index:idx_deployment_history_status"`
	CommitSHA          *string     `json:"commit_sha,omitempty" gorm:"size:64"`
	Author             *string     `json:"author,omitempty" gorm:"size:255"`
	ChangeSummary      string      `json:"change_summary" gorm:"size:500;not null"`
	TriggeredBy        uint        `json:"triggered_by" gorm:"not null;index:idx_deployment_history_triggered_by"`
	CreatedAt          time.Time   `json:"created_at" gorm:"index:idx_deployment_history_created_at"`
	Deployment         *Deployment `json:"-" gorm:"foreignKey:DeploymentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (d *DeploymentHistory) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	return nil
}
