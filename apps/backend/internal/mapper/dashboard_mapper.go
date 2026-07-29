package mapper

import (
	"fmt"
	"strings"
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapIncidentsForDashboard converts a slice of Incident models to DashboardIncidentResponse DTOs.
func MapIncidentsForDashboard(incidents []models.Incident) []dto.DashboardIncidentResponse {
	responses := make([]dto.DashboardIncidentResponse, 0, len(incidents))

	for _, incident := range incidents {
		responses = append(responses, dto.DashboardIncidentResponse{
			ID:        incident.ID,
			Title:     incident.Title,
			Severity:  incident.Severity,
			Status:    incident.Status,
			CreatedAt: incident.CreatedAt,
		})
	}

	return responses
}

// MapAuditLogs converts AuditLog models to human-readable DashboardActivityResponse DTOs.
// Generates meaningful titles and descriptions based on audit action and entity type.
func MapAuditLogs(logs []models.AuditLog) []dto.DashboardActivityResponse {
	activities := make([]dto.DashboardActivityResponse, 0, len(logs))

	for _, log := range logs {
		activity := dto.DashboardActivityResponse{
			ID:        log.ID,
			Timestamp: log.CreatedAt,
			Type:      log.EntityType,
			Status:    MapAuditActionToStatus(log.Action),
		}

		// Generate human-readable title and description
		activity.Title, activity.Description = GenerateActivityTitleAndDescription(log)

		activities = append(activities, activity)
	}

	return activities
}

// GenerateActivityTitleAndDescription creates human-readable activity text from AuditLog.
// Examples:
//   - AuditLog{EntityType: "PROJECT", Action: "CREATE"} → "Project Created", "Project created."
//   - AuditLog{EntityType: "INCIDENT", Action: "UPDATE"} → "Incident Updated", "Incident updated."
//   - AuditLog{EntityType: "USER", Action: "DELETE"} → "User Deleted", "User deleted."
func GenerateActivityTitleAndDescription(log models.AuditLog) (title, description string) {
	humanizedEntity := HumanizeEntityType(log.EntityType)
	actionName := humanizeAction(log.Action)

	title = fmt.Sprintf("%s %s", humanizedEntity, actionName)

	switch log.Action {
	case models.AuditActionCreate:
		description = fmt.Sprintf("%s created.", humanizedEntity)

	case models.AuditActionUpdate:
		if log.FieldName != "" {
			description = fmt.Sprintf("%s %s was updated.", humanizedEntity, log.FieldName)
		} else {
			description = fmt.Sprintf("%s updated.", humanizedEntity)
		}

	case models.AuditActionDelete:
		description = fmt.Sprintf("%s deleted.", humanizedEntity)

	default:
		description = fmt.Sprintf("%s was modified.", humanizedEntity)
	}

	return title, description
}

// HumanizeEntityType converts database entity type to human-readable format.
// Handles both upper-case ("PROJECT") and lower-case ("project") variants stored by different services.
// Examples: "PROJECT" → "Project", "incident" → "Incident", "AUDIT_LOG" → "Audit Log"
func HumanizeEntityType(entityType string) string {
	switch strings.ToUpper(entityType) {
	case "PROJECT":
		return "Project"
	case "INCIDENT":
		return "Incident"
	case "USER":
		return "User"
	case "COMMENT":
		return "Comment"
	case "AUDIT_LOG":
		return "Audit Log"
	default:
		// Fallback: capitalize first letter, lowercase the rest
		if len(entityType) == 0 {
			return "Item"
		}
		lower := strings.ToLower(entityType)
		return strings.ToUpper(lower[:1]) + lower[1:]
	}
}

// humanizeAction converts AuditAction enum to human-readable verb.
// Examples: "CREATE" → "Created", "UPDATE" → "Updated", "DELETE" → "Deleted"
func humanizeAction(action models.AuditAction) string {
	switch action {
	case models.AuditActionCreate:
		return "Created"
	case models.AuditActionUpdate:
		return "Updated"
	case models.AuditActionDelete:
		return "Deleted"
	default:
		return "Modified"
	}
}

// MapAuditActionToStatus converts AuditAction enum to a status string for frontend UI display.
// Maps CREATE/UPDATE to "success"/"info", DELETE to "warning"
func MapAuditActionToStatus(action models.AuditAction) string {
	switch action {
	case models.AuditActionCreate:
		return "success"
	case models.AuditActionUpdate:
		return "info"
	case models.AuditActionDelete:
		return "warning"
	default:
		return "info"
	}
}

// GenerateGreeting generates a context-aware greeting based on current server time.
// Returns "Good morning", "Good afternoon", "Good evening", or "Good night" based on hour of day.
func GenerateGreeting() string {
	hour := time.Now().Hour()

	switch {
	case hour >= 5 && hour < 12:
		return "Good morning"
	case hour >= 12 && hour < 17:
		return "Good afternoon"
	case hour >= 17 && hour < 21:
		return "Good evening"
	default:
		return "Good night"
	}
}
