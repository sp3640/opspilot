package constants

// Notification channel types - the provider abstraction implemented for
// Phase 24. Each is implementable cleanly with the Go standard library
// against this codebase's existing architecture (net/smtp for email,
// net/http for the three webhook-style integrations), so no external SDK
// or vendored dependency was introduced for any of them.
const (
	NotificationChannelEmail   = "EMAIL"
	NotificationChannelSlack   = "SLACK"
	NotificationChannelTeams   = "TEAMS"
	NotificationChannelWebhook = "WEBHOOK"
)

var ValidNotificationChannelTypes = []string{
	NotificationChannelEmail,
	NotificationChannelSlack,
	NotificationChannelTeams,
	NotificationChannelWebhook,
}

func IsValidNotificationChannelType(value string) bool {
	switch value {
	case NotificationChannelEmail, NotificationChannelSlack, NotificationChannelTeams, NotificationChannelWebhook:
		return true
	default:
		return false
	}
}

// Notification event types - the triggers a channel can subscribe to.
const (
	NotificationEventCriticalAlert       = "CRITICAL_ALERT"
	NotificationEventSev1Incident        = "SEV1_INCIDENT"
	NotificationEventIncidentAssigned    = "INCIDENT_ASSIGNED"
	NotificationEventDeploymentFailed    = "DEPLOYMENT_FAILED"
	NotificationEventDeploymentRecovered = "DEPLOYMENT_RECOVERED"
)

var ValidNotificationEventTypes = []string{
	NotificationEventCriticalAlert,
	NotificationEventSev1Incident,
	NotificationEventIncidentAssigned,
	NotificationEventDeploymentFailed,
	NotificationEventDeploymentRecovered,
}

func IsValidNotificationEventType(value string) bool {
	switch value {
	case NotificationEventCriticalAlert, NotificationEventSev1Incident, NotificationEventIncidentAssigned,
		NotificationEventDeploymentFailed, NotificationEventDeploymentRecovered:
		return true
	default:
		return false
	}
}
