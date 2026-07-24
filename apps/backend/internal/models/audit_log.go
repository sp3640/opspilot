package models

import "time"

type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
)

type AuditLog struct {
	ID         uint        `json:"id" gorm:"primaryKey"`
	UserID     uint        `json:"user_id" gorm:"not null;index"`
	ProjectID  *uint       `json:"project_id,omitempty" gorm:"index"`
	IncidentID *uint       `json:"incident_id,omitempty" gorm:"index"`
	EntityType string      `json:"entity_type" gorm:"size:50;not null;index"`
	EntityID   uint        `json:"entity_id" gorm:"not null;index"`
	Action     AuditAction `json:"action" gorm:"size:20;not null;index"`
	FieldName  string      `json:"field_name" gorm:"size:100"`
	OldValue   string      `json:"old_value" gorm:"type:text"`
	NewValue   string      `json:"new_value" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
}
