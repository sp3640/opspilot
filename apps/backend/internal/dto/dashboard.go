package dto

import "time"

// DashboardOverviewResponse represents the high-level dashboard summary.
// This DTO abstracts away database implementation details and presents
// a clean API contract for the frontend.
type DashboardOverviewResponse struct {
	Greeting       string `json:"greeting"`
	Environment    string `json:"environment"`
	LastDeployment string `json:"lastDeployment"`
	Uptime         string `json:"uptime"`

	// Incident Statistics
	TotalProjects     int64 `json:"totalProjects"`
	TotalIncidents    int64 `json:"totalIncidents"`
	OpenIncidents     int64 `json:"openIncidents"`
	CriticalIncidents int64 `json:"criticalIncidents"`
	ResolvedIncidents int64 `json:"resolvedIncidents"`
}

// IncidentMetrics represents aggregated incident statistics.
type IncidentMetrics struct {
	Total    int64 `json:"total"`
	Open     int64 `json:"open"`
	Critical int64 `json:"critical"`
	Resolved int64 `json:"resolved"`
}

// DashboardMetricsResponse provides aggregated metrics for dashboard display.
// All metrics are returned as explicit fields to prevent the frontend from
// understanding backend aggregation logic (e.g., maps, groups).
type DashboardMetricsResponse struct {
	Projects  int64           `json:"projects"`
	Incidents IncidentMetrics `json:"incidents"`
	AuditLogs int64           `json:"auditLogs"`
	Users     int64           `json:"users"`

	// Optional fields for extended metrics
	ProjectsDeploying           *int64 `json:"projectsDeploying,omitempty"`
	IncidentsRequiringAttention *int64 `json:"incidentsRequiringAttention,omitempty"`
	NewUsersThisWeek            *int64 `json:"newUsersThisWeek,omitempty"`
}

// DashboardActivityResponse represents a single activity event.
// Activities are derived from AuditLog entries with human-readable titles and descriptions.
type DashboardActivityResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`   // e.g., "incident", "project", "user", "audit"
	Status      string    `json:"status"` // e.g., "success", "warning", "info", "error"
	Timestamp   time.Time `json:"timestamp"`
}

// DashboardIncidentResponse represents an incident for dashboard display.
// Only exposes fields required by the frontend UI.
// Never exposes database implementation details.
type DashboardIncidentResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Severity  string    `json:"severity"` // e.g., "P1", "P2", "P3", "P4"
	Status    string    `json:"status"`   // e.g., "OPEN", "IN_PROGRESS", "RESOLVED"
	CreatedAt time.Time `json:"createdAt"`
}

// DashboardHealthResponse represents a single service health status.
// Used to display system health and dependency availability.
type DashboardHealthResponse struct {
	Name         string `json:"name"`         // e.g., "API", "Database", "Storage"
	Status       string `json:"status"`       // e.g., "healthy", "degraded", "unhealthy"
	ResponseTime int    `json:"responseTime"` // milliseconds
	Health       int    `json:"health"`       // 0-100 percentage
}
