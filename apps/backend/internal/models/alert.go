package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Alert struct {
	ID              uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID       uuid.UUID       `json:"project_id" gorm:"type:uuid;not null;index:idx_alerts_project_id;index:idx_alerts_project_fingerprint,priority:1"`
	Project         *Project        `json:"-" gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	IncidentID      *uint           `json:"incident_id,omitempty" gorm:"index:idx_alerts_incident_id"`
	Incident        *Incident       `json:"-" gorm:"foreignKey:IncidentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Title           string          `json:"title" gorm:"size:255;not null;index:idx_alerts_title"`
	Description     string          `json:"description" gorm:"type:text;not null;default:''"`
	Severity        string          `json:"severity" gorm:"size:20;not null;index:idx_alerts_severity"`
	Status          string          `json:"status" gorm:"size:20;not null;index:idx_alerts_status"`
	Source          string          `json:"source" gorm:"size:20;not null;index:idx_alerts_source"`
	ResourceType    string          `json:"resource_type" gorm:"size:32;not null;index:idx_alerts_resource_type"`
	ResourceID      string          `json:"resource_id" gorm:"size:255;not null;index:idx_alerts_resource_id"`
	Fingerprint     string          `json:"fingerprint" gorm:"size:191;not null;index:idx_alerts_project_fingerprint,priority:2"`
	OccurrenceCount int             `json:"occurrence_count" gorm:"not null;default:1;check:chk_alerts_occurrence_count_positive,occurrence_count >= 1"`
	Labels          json.RawMessage `json:"labels" gorm:"type:jsonb;not null;default:'{}'"`
	Metadata        json.RawMessage `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`
	FirstSeenAt     time.Time       `json:"first_seen_at" gorm:"not null;index:idx_alerts_first_seen_at"`
	LastSeenAt      time.Time       `json:"last_seen_at" gorm:"not null;index:idx_alerts_last_seen_at"`
	AcknowledgedAt  *time.Time      `json:"acknowledged_at,omitempty" gorm:"index:idx_alerts_acknowledged_at"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty" gorm:"index:idx_alerts_resolved_at"`
	CreatedBy       uint            `json:"created_by" gorm:"not null;index:idx_alerts_created_by"`
	CreatedByUser   *User           `json:"-" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
