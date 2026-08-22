package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateIncidentRequest is the validated input for incident creation.
type CreateIncidentRequest struct {
	Title         string     `json:"title" binding:"required,min=1,max=255"`
	Description   string     `json:"description" binding:"max=1000"`
	Severity      string     `json:"severity" binding:"required"`
	Status        string     `json:"status" binding:"required"`
	ProjectID     uuid.UUID  `json:"project_id" binding:"required"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	OwnerTeamID   *uuid.UUID `json:"owner_team_id,omitempty"`
}

// UpdateIncidentRequest is the validated input for a full incident update.
type UpdateIncidentRequest struct {
	Title         string     `json:"title" binding:"required,min=1,max=255"`
	Description   string     `json:"description" binding:"max=1000"`
	Severity      string     `json:"severity" binding:"required"`
	Status        string     `json:"status" binding:"required"`
	ProjectID     uuid.UUID  `json:"project_id" binding:"required"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	OwnerTeamID   *uuid.UUID `json:"owner_team_id,omitempty"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// IncidentResponse is the canonical API representation of an incident.
// It never exposes internal database fields such as UserID or raw foreign keys.
type IncidentResponse struct {
	ID            uint       `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Severity      string     `json:"severity"`
	Status        string     `json:"status"`
	ProjectID     string     `json:"projectId"`
	ApplicationID *string    `json:"applicationId,omitempty"`
	OwnerTeamID   *string    `json:"ownerTeamId,omitempty"`
	ResolvedAt    *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// IncidentSummaryResponse is a lightweight representation used in list contexts
// where the full detail is not required.
type IncidentSummaryResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Severity  string    `json:"severity"`
	Status    string    `json:"status"`
	ProjectID string    `json:"projectId"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// IncidentListResponse is the paginated incident list returned by GET /incidents.
type IncidentListResponse struct {
	Items      []IncidentResponse `json:"items"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"totalPages"`
}
