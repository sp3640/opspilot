package dto

import "time"

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateProjectRequest is the validated input for project creation.
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=300"`
}

// UpdateProjectRequest is the validated input for a full project update.
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=300"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// ProjectOwnerResponse is the safe owner representation embedded in project responses.
type ProjectOwnerResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ProjectResponse is the canonical API representation of a project.
// It never exposes internal database fields such as DeletedAt or raw foreign keys.
type ProjectResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	Description string               `json:"description"`
	Environment string               `json:"environment"`
	Health      string               `json:"health"`
	Members     int                  `json:"members"`
	Services    int                  `json:"services"`
	Owner       ProjectOwnerResponse `json:"owner"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

// ProjectSummaryResponse is a lightweight representation used in list contexts
// where the full detail is not required.
type ProjectSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Environment string `json:"environment"`
	Health      string `json:"health"`
}

// ProjectListResponse is the paginated project list returned by GET /projects.
type ProjectListResponse struct {
	Items      []ProjectResponse `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"totalPages"`
}
