package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func MapApplication(application models.Application) dto.ApplicationResponse {
	return dto.ApplicationResponse{
		ID:             application.ID.String(),
		OrganizationID: application.OrganizationID.String(),
		ProjectID:      application.ProjectID.String(),
		Name:           application.Name,
		Slug:           application.Slug,
		Description:    application.Description,
		RepositoryURL:  application.RepositoryURL,
		DefaultBranch:  application.DefaultBranch,
		Runtime:        application.Runtime,
		BuildCommand:   application.BuildCommand,
		StartCommand:   application.StartCommand,
		Port:           application.Port,
		Environment:    application.Environment,
		Status:         application.Status,
		CreatedAt:      application.CreatedAt,
		UpdatedAt:      application.UpdatedAt,
	}
}

func MapApplications(applications []models.Application) []dto.ApplicationResponse {
	responses := make([]dto.ApplicationResponse, 0, len(applications))
	for _, application := range applications {
		responses = append(responses, MapApplication(application))
	}

	return responses
}
