package constants

import "strings"

const (
	DeploymentStatusPending    = "Pending"
	DeploymentStatusQueued     = "Queued"
	DeploymentStatusRunning    = "Running"
	DeploymentStatusSucceeded  = "Succeeded"
	DeploymentStatusFailed     = "Failed"
	DeploymentStatusCancelled  = "Cancelled"
	DeploymentStatusRolledBack = "RolledBack"

	DeploymentStrategyRollingUpdate = "RollingUpdate"
	DeploymentStrategyRecreate      = "Recreate"

	DeploymentEnvironmentDevelopment = "Development"
	DeploymentEnvironmentStaging     = "Staging"
	DeploymentEnvironmentProduction  = "Production"
)

var ValidDeploymentStatuses = []string{
	DeploymentStatusPending,
	DeploymentStatusQueued,
	DeploymentStatusRunning,
	DeploymentStatusSucceeded,
	DeploymentStatusFailed,
	DeploymentStatusCancelled,
	DeploymentStatusRolledBack,
}

var ValidDeploymentStrategies = []string{
	DeploymentStrategyRollingUpdate,
	DeploymentStrategyRecreate,
}

var ValidDeploymentEnvironments = []string{
	DeploymentEnvironmentDevelopment,
	DeploymentEnvironmentStaging,
	DeploymentEnvironmentProduction,
}

func NormalizeDeploymentStatus(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending":
		return DeploymentStatusPending, true
	case "queued":
		return DeploymentStatusQueued, true
	case "running":
		return DeploymentStatusRunning, true
	case "succeeded":
		return DeploymentStatusSucceeded, true
	case "failed":
		return DeploymentStatusFailed, true
	case "cancelled", "canceled":
		return DeploymentStatusCancelled, true
	case "rolledback", "rolled-back", "rolled_back":
		return DeploymentStatusRolledBack, true
	default:
		return "", false
	}
}

func NormalizeDeploymentStrategy(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "rollingupdate", "rolling-update", "rolling_update":
		return DeploymentStrategyRollingUpdate, true
	case "recreate":
		return DeploymentStrategyRecreate, true
	default:
		return "", false
	}
}

func NormalizeDeploymentEnvironment(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "development", "dev":
		return DeploymentEnvironmentDevelopment, true
	case "staging", "stage":
		return DeploymentEnvironmentStaging, true
	case "production", "prod":
		return DeploymentEnvironmentProduction, true
	default:
		return "", false
	}
}
