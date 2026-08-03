package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cluster represents a connection to an external infrastructure provider.
type Cluster struct {
	ID                  uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;not null"`
	ProjectID           uuid.UUID       `json:"project_id" gorm:"type:uuid;not null;index:idx_clusters_project_id;uniqueIndex:idx_clusters_project_default,where:is_default = true,priority:1"`
	Project             *Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Name                string          `json:"name" gorm:"size:255;not null;index:idx_clusters_name"`
	Provider            string          `json:"provider" gorm:"size:64;not null;index:idx_clusters_provider"`
	Status              string          `json:"status" gorm:"size:32;not null;index:idx_clusters_status"`
	IsDefault           bool            `json:"is_default" gorm:"not null;default:false;uniqueIndex:idx_clusters_project_default,where:is_default = true,priority:2"`
	ConnectionType      string          `json:"connection_type" gorm:"size:64;not null;default:''"`
	KubeconfigEncrypted string          `json:"kubeconfig_encrypted" gorm:"type:text;not null;default:''"`
	APIEndpoint         string          `json:"api_endpoint" gorm:"size:255;not null;default:''"`
	Region              string          `json:"region" gorm:"size:128;not null;default:''"`
	Version             string          `json:"version" gorm:"size:64;not null;default:''"`
	LastValidatedAt     *time.Time      `json:"last_validated_at,omitempty" gorm:"index:idx_clusters_last_validated_at"`
	LastDiscoveryAt     *time.Time      `json:"last_discovery_at,omitempty" gorm:"index:idx_clusters_last_discovery_at"`
	CreatedBy           uint            `json:"created_by" gorm:"not null;index:idx_clusters_created_by"`
	CreatedByUser       *User           `json:"-" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ValidationError     string          `json:"validation_error" gorm:"type:text;not null;default:''"`
	Metadata            json.RawMessage `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a UUID when a cluster is created without an explicit ID.
func (c *Cluster) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return nil
}
