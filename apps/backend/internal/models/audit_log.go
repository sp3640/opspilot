package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
	// AuditActionLogin covers both successful and failed authentication
	// attempts - distinguished by Result, not by a separate action value,
	// so filtering "every login-related event" never has to enumerate both.
	AuditActionLogin AuditAction = "LOGIN"
)

// AuditResult records whether the audited action actually succeeded. Every
// audit row has one; writers that don't pass a value default to Success in
// AuditService.LogEvent, since the vast majority of pre-Phase-23 call sites
// only ever recorded actions that had already succeeded.
type AuditResult string

const (
	AuditResultSuccess AuditResult = "SUCCESS"
	AuditResultFailure AuditResult = "FAILURE"
)

type AuditLog struct {
	ID             uint        `json:"id" gorm:"primaryKey"`
	OrganizationID uuid.UUID   `json:"organization_id" gorm:"type:uuid;not null;index"`
	UserID         uint        `json:"user_id" gorm:"not null;index"`
	ProjectID      *uuid.UUID  `json:"project_id,omitempty" gorm:"type:uuid;index"`
	ApplicationID  *uuid.UUID  `json:"application_id,omitempty" gorm:"type:uuid;index"`
	IncidentID     *uint       `json:"incident_id,omitempty" gorm:"index"`
	EntityType     string      `json:"entity_type" gorm:"size:50;not null;index"`
	EntityID       string      `json:"entity_id" gorm:"size:36;not null;index"`
	Action         AuditAction `json:"action" gorm:"size:20;not null;index"`
	Result         AuditResult `json:"result" gorm:"size:20;not null;default:SUCCESS;index"`
	FieldName      string      `json:"field_name" gorm:"size:100"`
	OldValue       string      `json:"old_value" gorm:"type:text"`
	NewValue       string      `json:"new_value" gorm:"type:text"`
	// BeforeState/AfterState hold an optional JSON snapshot for actions that
	// don't reduce to a single changed field (e.g. a resource's full state
	// at creation/deletion). Callers are responsible for excluding secrets
	// (passwords, tokens, kubeconfigs) before building these - see
	// services.marshalAuditState.
	BeforeState string `json:"before_state,omitempty" gorm:"type:text"`
	AfterState  string `json:"after_state,omitempty" gorm:"type:text"`
	// IPAddress/UserAgent are captured on a best-effort basis ("if safely
	// available") from the originating request; both are blank when not
	// captured for a given call site rather than fabricated.
	IPAddress string `json:"ip_address,omitempty" gorm:"size:64"`
	UserAgent string `json:"user_agent,omitempty" gorm:"size:255"`

	CreatedAt time.Time `json:"created_at"`
}
