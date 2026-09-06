package constants

// Integration type identifiers (Sprint 27 - Integration Foundation). These
// name the external systems OpsPilot will eventually connect to; no
// connector is implemented for any of them yet (see internal/connector) -
// this sprint only builds the persistence, security, and lifecycle
// scaffolding they will plug into. Centralized here so the type identifier
// is never scattered as a raw string literal across handlers/services.
//
// Note: this is a distinct identifier space from
// constants.NotificationChannel* (EMAIL/SLACK/TEAMS/WEBHOOK, uppercase) -
// Integration and NotificationChannel are different domain models (one
// represents an external-system connection, the other a place to send an
// operational alert) that happen to share a couple of provider names.
const (
	IntegrationTypeGitHub        = "github"
	IntegrationTypeSlack         = "slack"
	IntegrationTypeEmail         = "email"
	IntegrationTypePrometheus    = "prometheus"
	IntegrationTypeLoki          = "loki"
	IntegrationTypeOpenTelemetry = "opentelemetry"
	IntegrationTypeAzure         = "azure"
	IntegrationTypeTeams         = "teams"
	IntegrationTypeWebhook       = "webhook"
)

var ValidIntegrationTypes = []string{
	IntegrationTypeGitHub,
	IntegrationTypeSlack,
	IntegrationTypeEmail,
	IntegrationTypePrometheus,
	IntegrationTypeLoki,
	IntegrationTypeOpenTelemetry,
	IntegrationTypeAzure,
	IntegrationTypeTeams,
	IntegrationTypeWebhook,
}

func IsValidIntegrationType(value string) bool {
	switch value {
	case IntegrationTypeGitHub, IntegrationTypeSlack, IntegrationTypeEmail, IntegrationTypePrometheus,
		IntegrationTypeLoki, IntegrationTypeOpenTelemetry, IntegrationTypeAzure, IntegrationTypeTeams, IntegrationTypeWebhook:
		return true
	default:
		return false
	}
}

// Integration lifecycle status. This represents where the integration
// record itself is in its connect/disconnect lifecycle - never the general
// health of the external service (that distinction matters: an
// "external service outage" is not the same fact as "we have not connected
// this integration yet", and the two must never be conflated in one field).
const (
	IntegrationStatusPending      = "PENDING"
	IntegrationStatusConnected    = "CONNECTED"
	IntegrationStatusDisconnected = "DISCONNECTED"
	IntegrationStatusError        = "ERROR"
)

var ValidIntegrationStatuses = []string{
	IntegrationStatusPending,
	IntegrationStatusConnected,
	IntegrationStatusDisconnected,
	IntegrationStatusError,
}

func IsValidIntegrationStatus(value string) bool {
	switch value {
	case IntegrationStatusPending, IntegrationStatusConnected, IntegrationStatusDisconnected, IntegrationStatusError:
		return true
	default:
		return false
	}
}
