package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapCluster converts a Cluster model to a ClusterResponse DTO.
// This is the single authoritative mapping function for cluster responses.
func MapCluster(c models.Cluster) dto.ClusterResponse {
	return dto.ClusterResponse{
		ID:                c.ID.String(),
		OrganizationID:    c.OrganizationID.String(),
		ProjectID:         c.ProjectID.String(),
		Name:              c.Name,
		Provider:          c.Provider,
		ConnectionType:    c.ConnectionType,
		CredentialType:    c.CredentialType,
		KubernetesVersion: c.KubernetesVersion,
		Version:           c.Version,
		Status:            c.Status,
		APIEndpoint:       c.APIEndpoint,
		Region:            c.Region,
		IsDefault:         c.IsDefault,
		LastValidatedAt:   c.LastValidatedAt,
		CreatedBy:         c.CreatedBy,
		ValidationError:   c.ValidationError,
		Metadata:          c.Metadata,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

// MapClusters converts a slice of Cluster models to a slice of ClusterResponse DTOs.
func MapClusters(clusters []models.Cluster) []dto.ClusterResponse {
	responses := make([]dto.ClusterResponse, 0, len(clusters))
	for _, c := range clusters {
		responses = append(responses, MapCluster(c))
	}
	return responses
}
