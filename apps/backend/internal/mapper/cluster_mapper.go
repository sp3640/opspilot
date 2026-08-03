package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapCluster converts a Cluster model to a ClusterResponse DTO.
// This is the single authoritative mapping function for cluster responses.
func MapCluster(c models.Cluster) dto.ClusterResponse {
	return dto.ClusterResponse{
		ID:                  c.ID.String(),
		ProjectID:           c.ProjectID.String(),
		Name:                c.Name,
		Provider:            c.Provider,
		Status:              c.Status,
		IsDefault:           c.IsDefault,
		ConnectionType:      c.ConnectionType,
		KubeconfigEncrypted: c.KubeconfigEncrypted,
		APIEndpoint:         c.APIEndpoint,
		Region:              c.Region,
		Version:             c.Version,
		LastValidatedAt:     c.LastValidatedAt,
		LastDiscoveryAt:     c.LastDiscoveryAt,
		CreatedBy:           c.CreatedBy,
		ValidationError:     c.ValidationError,
		Metadata:            c.Metadata,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
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
