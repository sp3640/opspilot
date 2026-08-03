package mapper

import (
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapAlert converts an Alert model to an AlertResponse DTO.
// This is the single authoritative mapping function for alert responses.
func MapAlert(a models.Alert) dto.AlertResponse {
	return dto.AlertResponse{
		ID:              a.ID,
		ProjectID:       a.ProjectID.String(),
		IncidentID:      a.IncidentID,
		Title:           a.Title,
		Description:     a.Description,
		Severity:        a.Severity,
		Status:          a.Status,
		Source:          a.Source,
		ResourceType:    a.ResourceType,
		ResourceID:      a.ResourceID,
		Fingerprint:     a.Fingerprint,
		OccurrenceCount: a.OccurrenceCount,
		Labels:          a.Labels,
		Metadata:        a.Metadata,
		FirstSeenAt:     a.FirstSeenAt,
		LastSeenAt:      a.LastSeenAt,
		AcknowledgedAt:  a.AcknowledgedAt,
		ResolvedAt:      a.ResolvedAt,
		CreatedBy:       a.CreatedBy,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}

// MapAlerts converts a slice of Alert models to a slice of AlertResponse DTOs.
func MapAlerts(alerts []models.Alert) []dto.AlertResponse {
	responses := make([]dto.AlertResponse, 0, len(alerts))
	for _, a := range alerts {
		responses = append(responses, MapAlert(a))
	}
	return responses
}
