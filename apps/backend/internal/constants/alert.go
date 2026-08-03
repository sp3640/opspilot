package constants

// Alert Severity levels
const (
	AlertSeverityCritical = "CRITICAL"
	AlertSeverityHigh     = "HIGH"
	AlertSeverityMedium   = "MEDIUM"
	AlertSeverityLow      = "LOW"
	AlertSeverityInfo     = "INFO"
)

// Alert Status values
const (
	AlertStatusOpen          = "OPEN"
	AlertStatusAcknowledged  = "ACKNOWLEDGED"
	AlertStatusInvestigating = "INVESTIGATING"
	AlertStatusResolved      = "RESOLVED"
)

// Alert Source values
const (
	AlertSourceManual     = "MANUAL"
	AlertSourceAI         = "AI"
	AlertSourceKubernetes = "KUBERNETES"
	AlertSourcePrometheus = "PROMETHEUS"
	AlertSourceGrafana    = "GRAFANA"
	AlertSourceAzure      = "AZURE"
	AlertSourceDocker     = "DOCKER"
	AlertSourceVM         = "VM"
)

// Alert Resource Type values
const (
	AlertResourceTypePod         = "POD"
	AlertResourceTypeDeployment  = "DEPLOYMENT"
	AlertResourceTypeNode        = "NODE"
	AlertResourceTypeService     = "SERVICE"
	AlertResourceTypeContainer   = "CONTAINER"
	AlertResourceTypeDatabase    = "DATABASE"
	AlertResourceTypeVM          = "VM"
	AlertResourceTypeApplication = "APPLICATION"
	AlertResourceTypeNamespace   = "NAMESPACE"
)

// ValidAlertSeverities is the slice of all valid alert severity values.
var ValidAlertSeverities = []string{
	AlertSeverityCritical,
	AlertSeverityHigh,
	AlertSeverityMedium,
	AlertSeverityLow,
	AlertSeverityInfo,
}

// ValidAlertStatuses is the slice of all valid alert status values.
var ValidAlertStatuses = []string{
	AlertStatusOpen,
	AlertStatusAcknowledged,
	AlertStatusInvestigating,
	AlertStatusResolved,
}

// ValidAlertSources is the slice of all valid alert source values.
var ValidAlertSources = []string{
	AlertSourceManual,
	AlertSourceAI,
	AlertSourceKubernetes,
	AlertSourcePrometheus,
	AlertSourceGrafana,
	AlertSourceAzure,
	AlertSourceDocker,
	AlertSourceVM,
}

// ValidAlertResourceTypes is the slice of all valid alert resource type values.
var ValidAlertResourceTypes = []string{
	AlertResourceTypePod,
	AlertResourceTypeDeployment,
	AlertResourceTypeNode,
	AlertResourceTypeService,
	AlertResourceTypeContainer,
	AlertResourceTypeDatabase,
	AlertResourceTypeVM,
	AlertResourceTypeApplication,
	AlertResourceTypeNamespace,
}

func IsValidAlertSeverity(severity string) bool {
	switch severity {
	case AlertSeverityCritical, AlertSeverityHigh, AlertSeverityMedium, AlertSeverityLow, AlertSeverityInfo:
		return true
	default:
		return false
	}
}

func IsValidAlertStatus(status string) bool {
	switch status {
	case AlertStatusOpen, AlertStatusAcknowledged, AlertStatusInvestigating, AlertStatusResolved:
		return true
	default:
		return false
	}
}

func IsValidAlertSource(source string) bool {
	switch source {
	case AlertSourceManual, AlertSourceAI, AlertSourceKubernetes, AlertSourcePrometheus, AlertSourceGrafana, AlertSourceAzure, AlertSourceDocker, AlertSourceVM:
		return true
	default:
		return false
	}
}

func IsValidAlertResourceType(resourceType string) bool {
	switch resourceType {
	case AlertResourceTypePod, AlertResourceTypeDeployment, AlertResourceTypeNode, AlertResourceTypeService, AlertResourceTypeContainer, AlertResourceTypeDatabase, AlertResourceTypeVM, AlertResourceTypeApplication, AlertResourceTypeNamespace:
		return true
	default:
		return false
	}
}
