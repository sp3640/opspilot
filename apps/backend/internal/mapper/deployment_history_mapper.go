package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func MapDeploymentHistory(history models.DeploymentHistory) dto.DeploymentHistoryResponse {
	return dto.DeploymentHistoryResponse{
		ID:                 history.ID.String(),
		DeploymentID:       history.DeploymentID.String(),
		ApplicationID:      history.ApplicationID.String(),
		ProjectID:          history.ProjectID.String(),
		OrganizationID:     history.OrganizationID.String(),
		Revision:           history.Revision,
		Image:              history.Image,
		ImageTag:           history.ImageTag,
		Environment:        history.Environment,
		Namespace:          history.Namespace,
		ReplicaCount:       history.ReplicaCount,
		DeploymentStrategy: history.DeploymentStrategy,
		Status:             history.Status,
		ChangeSummary:      history.ChangeSummary,
		TriggeredBy:        history.TriggeredBy,
		CreatedAt:          history.CreatedAt,
	}
}

func MapDeploymentHistoryItems(items []models.DeploymentHistory) []dto.DeploymentHistoryResponse {
	responses := make([]dto.DeploymentHistoryResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, MapDeploymentHistory(item))
	}

	return responses
}
