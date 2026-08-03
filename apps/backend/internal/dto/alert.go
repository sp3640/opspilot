package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateAlertRequest is the validated input for alert creation.
type CreateAlertRequest struct {
	ProjectID       uuid.UUID       `json:"project_id" binding:"required"`
	IncidentID      *uint           `json:"incident_id,omitempty"`
	Title           string          `json:"title" binding:"required,min=1,max=255"`
	Description     string          `json:"description" binding:"max=2000"`
	Severity        string          `json:"severity" binding:"required"`
	Status          string          `json:"status" binding:"required"`
	Source          string          `json:"source" binding:"required"`
	ResourceType    string          `json:"resource_type" binding:"required"`
	ResourceID      string          `json:"resource_id" binding:"required,min=1,max=255"`
	Fingerprint     string          `json:"fingerprint" binding:"required,min=1,max=191"`
	OccurrenceCount int             `json:"occurrence_count" binding:"required,min=1"`
	Labels          json.RawMessage `json:"labels"`
	Metadata        json.RawMessage `json:"metadata"`
	FirstSeenAt     time.Time       `json:"first_seen_at" binding:"required"`
	LastSeenAt      time.Time       `json:"last_seen_at" binding:"required"`
	AcknowledgedAt  *time.Time      `json:"acknowledged_at,omitempty"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty"`
}

// UpdateAlertRequest is the validated input for a full alert update.
type UpdateAlertRequest struct {
	ProjectID       uuid.UUID       `json:"project_id" binding:"required"`
	IncidentID      *uint           `json:"incident_id,omitempty"`
	Title           string          `json:"title" binding:"required,min=1,max=255"`
	Description     string          `json:"description" binding:"max=2000"`
	Severity        string          `json:"severity" binding:"required"`
	Status          string          `json:"status" binding:"required"`
	Source          string          `json:"source" binding:"required"`
	ResourceType    string          `json:"resource_type" binding:"required"`
	ResourceID      string          `json:"resource_id" binding:"required,min=1,max=255"`
	Fingerprint     string          `json:"fingerprint" binding:"required,min=1,max=191"`
	OccurrenceCount int             `json:"occurrence_count" binding:"required,min=1"`
	Labels          json.RawMessage `json:"labels"`
	Metadata        json.RawMessage `json:"metadata"`
	FirstSeenAt     time.Time       `json:"first_seen_at" binding:"required"`
	LastSeenAt      time.Time       `json:"last_seen_at" binding:"required"`
	AcknowledgedAt  *time.Time      `json:"acknowledged_at,omitempty"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty"`
}

// AlertFilterRequest is the validated input for alert list filtering.
type AlertFilterRequest struct {
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
	Search    string    `json:"search"`
	Sort      string    `json:"sort"`
	Order     string    `json:"order"`
	Status    string    `json:"status"`
	Severity  string    `json:"severity"`
	Source    string    `json:"source"`
	ProjectID uuid.UUID `json:"project_id"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// AlertResponse is the canonical API representation of an alert.
// It never exposes internal database fields such as DeletedAt or raw relations.
type AlertResponse struct {
	ID              uint            `json:"id"`
	ProjectID       string          `json:"projectId"`
	IncidentID      *uint           `json:"incidentId,omitempty"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Severity        string          `json:"severity"`
	Status          string          `json:"status"`
	Source          string          `json:"source"`
	ResourceType    string          `json:"resourceType"`
	ResourceID      string          `json:"resourceId"`
	Fingerprint     string          `json:"fingerprint"`
	OccurrenceCount int             `json:"occurrenceCount"`
	Labels          json.RawMessage `json:"labels"`
	Metadata        json.RawMessage `json:"metadata"`
	FirstSeenAt     time.Time       `json:"firstSeenAt"`
	LastSeenAt      time.Time       `json:"lastSeenAt"`
	AcknowledgedAt  *time.Time      `json:"acknowledgedAt,omitempty"`
	ResolvedAt      *time.Time      `json:"resolvedAt,omitempty"`
	CreatedBy       uint            `json:"createdBy"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// AlertPaginationResponse is reusable pagination metadata for alert list responses.
type AlertPaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// AlertListResponse is the paginated alert list response.
type AlertListResponse struct {
	Items      []AlertResponse `json:"items"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"totalPages"`
}
