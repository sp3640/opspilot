package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func MapDeployment(deployment models.Deployment) dto.DeploymentResponse {
	return dto.DeploymentResponse{
		ID:                 deployment.ID.String(),
		ApplicationID:      deployment.ApplicationID.String(),
		ProjectID:          deployment.ProjectID.String(),
		OrganizationID:     deployment.OrganizationID.String(),
		Image:              deployment.Image,
		ImageTag:           deployment.ImageTag,
		Environment:        deployment.Environment,
		Namespace:          deployment.Namespace,
		ReplicaCount:       deployment.ReplicaCount,
		Status:             deployment.Status,
		DeploymentStrategy: deployment.DeploymentStrategy,
		TargetClusterID:    deployment.TargetClusterID.String(),
		CreatedBy:          deployment.CreatedBy,
		UpdatedBy:          deployment.UpdatedBy,
		StartedAt:          deployment.StartedAt,
		CompletedAt:        deployment.CompletedAt,
		CreatedAt:          deployment.CreatedAt,
		UpdatedAt:          deployment.UpdatedAt,
	}
}

func MapDeployments(items []models.Deployment) []dto.DeploymentResponse {
	responses := make([]dto.DeploymentResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, MapDeployment(item))
	}

	return responses
}
