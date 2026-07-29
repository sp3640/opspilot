package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapProject converts a Project model and its owner to a ProjectResponse DTO.
// This is the single authoritative mapping function for project responses.
func MapProject(p models.Project, owner models.User) dto.ProjectResponse {
	return dto.ProjectResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Environment: p.Environment,
		Health:      p.Health,
		Members:     p.Members,
		Services:    p.Services,
		Owner: dto.ProjectOwnerResponse{
			ID:   owner.ID,
			Name: owner.Name,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// MapProjects converts a slice of Project models with a shared owner to ProjectResponse DTOs.
func MapProjects(projects []models.Project, owner models.User) []dto.ProjectResponse {
	responses := make([]dto.ProjectResponse, 0, len(projects))
	for _, p := range projects {
		responses = append(responses, MapProject(p, owner))
	}
	return responses
}

// MapProjectSummary converts a Project model to a lightweight ProjectSummaryResponse DTO.
// Used in contexts where only identity and status fields are required.
func MapProjectSummary(p models.Project) dto.ProjectSummaryResponse {
	return dto.ProjectSummaryResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Slug:        p.Slug,
		Environment: p.Environment,
		Health:      p.Health,
	}
}
