package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapIncident converts an Incident model to an IncidentResponse DTO.
// This is the single authoritative mapping function for incident responses.
func MapIncident(i models.Incident) dto.IncidentResponse {
	var applicationID *string
	if i.ApplicationID != nil {
		value := i.ApplicationID.String()
		applicationID = &value
	}

	var ownerTeamID *string
	if i.OwnerTeamID != nil {
		value := i.OwnerTeamID.String()
		ownerTeamID = &value
	}

	return dto.IncidentResponse{
		ID:            i.ID,
		Title:         i.Title,
		Description:   i.Description,
		Severity:      i.Severity,
		Status:        i.Status,
		ProjectID:     i.ProjectID.String(),
		ApplicationID: applicationID,
		OwnerTeamID:   ownerTeamID,
		ResolvedAt:    i.ResolvedAt,
		CreatedAt:     i.CreatedAt,
		UpdatedAt:     i.UpdatedAt,
	}
}

// MapIncidents converts a slice of Incident models to a slice of IncidentResponse DTOs.
func MapIncidents(incidents []models.Incident) []dto.IncidentResponse {
	responses := make([]dto.IncidentResponse, 0, len(incidents))
	for _, i := range incidents {
		responses = append(responses, MapIncident(i))
	}
	return responses
}

// MapIncidentSummary converts an Incident model to a lightweight IncidentSummaryResponse DTO.
// Used in contexts where only identity and status fields are required.
func MapIncidentSummary(i models.Incident) dto.IncidentSummaryResponse {
	return dto.IncidentSummaryResponse{
		ID:        i.ID,
		Title:     i.Title,
		Severity:  i.Severity,
		Status:    i.Status,
		ProjectID: i.ProjectID.String(),
		UpdatedAt: i.UpdatedAt,
	}
}
