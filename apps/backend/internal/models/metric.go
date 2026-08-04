package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Metric struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index:idx_metrics_organization_id"`
	ProjectID      uuid.UUID       `json:"project_id" gorm:"type:uuid;not null;index:idx_metrics_project_id;index:idx_metrics_project_timestamp,priority:1;index:idx_metrics_project_cluster_resource,priority:1"`
	Project        *Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ClusterID      uuid.UUID       `json:"cluster_id" gorm:"type:uuid;not null;index:idx_metrics_cluster_id;index:idx_metrics_project_cluster_resource,priority:2"`
	Cluster        *Cluster        `json:"-" gorm:"foreignKey:ClusterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ResourceID     uuid.UUID       `json:"resource_id" gorm:"type:uuid;not null;index:idx_metrics_resource_id;index:idx_metrics_project_cluster_resource,priority:3"`
	Resource       *Resource       `json:"-" gorm:"foreignKey:ResourceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ResourceKind   string          `json:"resource_kind" gorm:"size:64;not null;index:idx_metrics_resource_kind"`
	MetricType     string          `json:"metric_type" gorm:"size:32;not null;index:idx_metrics_metric_type;index:idx_metrics_metric_scope,priority:1"`
	MetricName     string          `json:"metric_name" gorm:"size:128;not null;index:idx_metrics_metric_name;index:idx_metrics_metric_scope,priority:2"`
	Value          float64         `json:"value" gorm:"not null"`
	Unit           string          `json:"unit" gorm:"size:32;not null;default:''"`
	Timestamp      time.Time       `json:"timestamp" gorm:"not null;index:idx_metrics_timestamp;index:idx_metrics_project_timestamp,priority:2;index:idx_metrics_metric_scope,priority:3"`
	Labels         json.RawMessage `json:"labels" gorm:"type:jsonb;not null;default:'{}'"`
	Metadata       json.RawMessage `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`

	CreatedAt time.Time `json:"created_at"`
}

type MetricAggregatePoint struct {
	Bucket time.Time
	Value  float64
	Count  int64
}

// BeforeCreate assigns a UUID when a metric is created without an explicit ID.
func (m *Metric) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	return nil
}
