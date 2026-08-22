package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetricJSON is a json.RawMessage-compatible type that additionally
// implements sql.Scanner/driver.Valuer. Postgres' native jsonb driver
// returns []byte, which json.RawMessage (a bare []byte alias) already
// round-trips through GORM's generic scan path - but SQLite's driver
// reports "jsonb" columns as TEXT and hands back a plain string, which
// json.RawMessage cannot Scan on its own. This type makes both drivers work
// without changing the wire-level JSON shape callers see.
type MetricJSON json.RawMessage

func (j *MetricJSON) Scan(value any) error {
	if value == nil {
		*j = MetricJSON("null")
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*j = append((*j)[:0], v...)
		return nil
	case string:
		*j = MetricJSON(v)
		return nil
	default:
		return fmt.Errorf("unsupported Scan type %T for MetricJSON", value)
	}
}

func (j MetricJSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}

	return string(j), nil
}

func (j MetricJSON) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

func (j *MetricJSON) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(j).UnmarshalJSON(data)
}

type Metric struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index:idx_metrics_organization_id"`
	ProjectID      uuid.UUID  `json:"project_id" gorm:"type:uuid;not null;index:idx_metrics_project_id;index:idx_metrics_project_timestamp,priority:1;index:idx_metrics_project_cluster_resource,priority:1"`
	Project        *Project   `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ClusterID      uuid.UUID  `json:"cluster_id" gorm:"type:uuid;not null;index:idx_metrics_cluster_id;index:idx_metrics_project_cluster_resource,priority:2"`
	Cluster        *Cluster   `json:"-" gorm:"foreignKey:ClusterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ResourceID     uuid.UUID  `json:"resource_id" gorm:"type:uuid;not null;index:idx_metrics_resource_id;index:idx_metrics_project_cluster_resource,priority:3"`
	Resource       *Resource  `json:"-" gorm:"foreignKey:ResourceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ResourceKind   string     `json:"resource_kind" gorm:"size:64;not null;index:idx_metrics_resource_kind"`
	MetricType     string     `json:"metric_type" gorm:"size:32;not null;index:idx_metrics_metric_type;index:idx_metrics_metric_scope,priority:1"`
	MetricName     string     `json:"metric_name" gorm:"size:128;not null;index:idx_metrics_metric_name;index:idx_metrics_metric_scope,priority:2"`
	Value          float64    `json:"value" gorm:"not null"`
	Unit           string     `json:"unit" gorm:"size:32;not null;default:''"`
	Timestamp      time.Time  `json:"timestamp" gorm:"not null;index:idx_metrics_timestamp;index:idx_metrics_project_timestamp,priority:2;index:idx_metrics_metric_scope,priority:3"`
	Labels         MetricJSON `json:"labels" gorm:"type:jsonb;not null;default:'{}'"`
	Metadata       MetricJSON `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`

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
