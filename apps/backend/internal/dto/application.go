package dto

import "time"

// CreateApplicationRequest is the validated input for application creation.
type CreateApplicationRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=255"`
	Slug          string `json:"slug" binding:"omitempty,max=120"`
	Description   string `json:"description" binding:"omitempty,max=500"`
	RepositoryURL string `json:"repository_url" binding:"omitempty,max=500"`
	DefaultBranch string `json:"default_branch" binding:"omitempty,max=128"`
	Runtime       string `json:"runtime" binding:"required"`
	BuildCommand  string `json:"build_command" binding:"omitempty"`
	StartCommand  string `json:"start_command" binding:"omitempty"`
	Port          int    `json:"port" binding:"required"`
	Environment   string `json:"environment" binding:"omitempty,max=255"`
	Status        string `json:"status" binding:"omitempty,max=32"`
}

// UpdateApplicationRequest is the validated input for application updates.
type UpdateApplicationRequest struct {
	Name          string `json:"name" binding:"omitempty,min=1,max=255"`
	Slug          string `json:"slug" binding:"omitempty,max=120"`
	Description   string `json:"description" binding:"omitempty,max=500"`
	RepositoryURL string `json:"repository_url" binding:"omitempty,max=500"`
	DefaultBranch string `json:"default_branch" binding:"omitempty,max=128"`
	Runtime       string `json:"runtime" binding:"omitempty"`
	BuildCommand  string `json:"build_command" binding:"omitempty"`
	StartCommand  string `json:"start_command" binding:"omitempty"`
	Port          int    `json:"port" binding:"omitempty"`
	Environment   string `json:"environment" binding:"omitempty,max=255"`
	Status        string `json:"status" binding:"omitempty,max=32"`
}

// ApplicationResponse is the canonical API representation of an application.
type ApplicationResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	RepositoryURL  string    `json:"repositoryUrl"`
	DefaultBranch  string    `json:"defaultBranch"`
	Runtime        string    `json:"runtime"`
	BuildCommand   string    `json:"buildCommand"`
	StartCommand   string    `json:"startCommand"`
	Port           int       `json:"port"`
	Environment    string    `json:"environment"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// ApplicationListResponse is the paginated application list returned by GET /projects/:projectId/applications.
type ApplicationListResponse struct {
	Items      []ApplicationResponse `json:"items"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"totalPages"`
}
