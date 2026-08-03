package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapResource converts a Resource model to a ResourceResponse DTO.
// This is the single authoritative mapping function for resource responses.
func MapResource(r models.Resource) dto.ResourceResponse {
	var parentResourceID *string
	if r.ParentResourceID != nil {
		value := r.ParentResourceID.String()
		parentResourceID = &value
	}

	return dto.ResourceResponse{
		ID:               r.ID.String(),
		ProjectID:        r.ProjectID.String(),
		ParentResourceID: parentResourceID,
		Kind:             r.Kind,
		Name:             r.Name,
		DisplayName:      r.DisplayName,
		ExternalID:       r.ExternalID,
		Provider:         r.Provider,
		Region:           r.Region,
		Namespace:        r.Namespace,
		Cluster:          r.Cluster,
		Status:           r.Status,
		Health:           r.Health,
		Labels:           r.Labels,
		Annotations:      r.Annotations,
		Metadata:         r.Metadata,
		CreatedBy:        r.CreatedBy,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

// MapResources converts a slice of Resource models to a slice of ResourceResponse DTOs.
func MapResources(resources []models.Resource) []dto.ResourceResponse {
	responses := make([]dto.ResourceResponse, 0, len(resources))
	for _, r := range resources {
		responses = append(responses, MapResource(r))
	}
	return responses
}
