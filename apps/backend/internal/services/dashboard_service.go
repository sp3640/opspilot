package services

import (
	"os"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepository
}

func NewDashboardService(repo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) ValidateProjectAccess(userID uint, projectID uuid.UUID) error {
	if projectID == uuid.Nil {
		return nil
	}

	belongs, err := s.repo.ProjectBelongsToUser(projectID, userID)
	if err != nil {
		return err
	}
	if !belongs {
		return apperrors.ErrInvalidProject
	}

	return nil
}

// GetSummary returns a dashboard overview with all summary information.
// Maps repository DashboardSummary to DashboardOverviewResponse DTO.
func (s *DashboardService) GetSummary(userID uint) (*dto.DashboardOverviewResponse, error) {
	summary, err := s.repo.GetSummaryByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &dto.DashboardOverviewResponse{
		Greeting:          mapper.GenerateGreeting(),
		Environment:       getEnvironment(),
		LastDeployment:    "Not Available", // TODO: Integrate with deployment tracking
		Uptime:            "99.9%",         // TODO: Integrate with uptime monitoring
		TotalProjects:     summary.TotalProjects,
		TotalIncidents:    summary.TotalIncidents,
		OpenIncidents:     summary.OpenIncidents,
		CriticalIncidents: summary.CriticalIncidents,
		ResolvedIncidents: summary.ResolvedIncidents,
	}, nil
}

// GetRecentIncidents returns a list of recent incidents as dashboard responses.
// Maps models.Incident to DashboardIncidentResponse DTO.
func (s *DashboardService) GetRecentIncidents(userID uint, limit int) ([]dto.DashboardIncidentResponse, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	incidents, err := s.repo.ListRecentIncidentsByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	return mapper.MapIncidentsForDashboard(incidents), nil
}

// GetRecentActivity returns recent audit log entries as activity feed.
// Maps models.AuditLog to DashboardActivityResponse DTO with human-readable titles.
func (s *DashboardService) GetRecentActivity(userID uint, limit int) ([]dto.DashboardActivityResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	logs, err := s.repo.ListRecentActivityByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	return mapper.MapAuditLogs(logs), nil
}

// GetStats returns aggregated incident and project statistics.
// Queries real data from repositories instead of returning fake metrics.
func (s *DashboardService) GetStats(userID uint) (*dto.DashboardMetricsResponse, error) {
	repoStats, err := s.repo.GetIncidentStatsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Count projects by user
	projectCount, err := s.repo.CountProjectsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Count audit logs by user
	auditLogCount, err := s.repo.CountAuditLogsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Count incidents by user
	incidentCount, err := s.repo.CountIncidentsByUserID(userID)
	if err != nil {
		return nil, err
	}

	metrics := &dto.DashboardMetricsResponse{
		Projects:  projectCount,
		AuditLogs: auditLogCount,
		Incidents: dto.IncidentMetrics{
			Total: incidentCount,
		},
	}

	// Map status-based incident counts
	if count, ok := repoStats.ByStatus["OPEN"]; ok {
		metrics.Incidents.Open = count
	}
	if count, ok := repoStats.ByStatus["RESOLVED"]; ok {
		metrics.Incidents.Resolved = count
	}

	// Map severity-based critical incidents
	if count, ok := repoStats.BySeverity["P1"]; ok {
		metrics.Incidents.Critical = count
	}

	return metrics, nil
}

// GetServiceHealth returns high-level synthetic health indicators for core services.
func (s *DashboardService) GetServiceHealth(userID uint) ([]dto.DashboardHealthResponse, error) {
	incidentCount, err := s.repo.CountIncidentsByUserID(userID)
	if err != nil {
		return nil, err
	}

	apiStatus := "healthy"
	apiHealth := 98
	apiResponseTime := 42
	if incidentCount >= 10 {
		apiStatus = "degraded"
		apiHealth = 88
		apiResponseTime = 95
	}

	services := []dto.DashboardHealthResponse{
		{Name: "API", Status: apiStatus, ResponseTime: apiResponseTime, Health: apiHealth},
		{Name: "Database", Status: "healthy", ResponseTime: 28, Health: 99},
		{Name: "Storage", Status: "healthy", ResponseTime: 34, Health: 97},
		{Name: "Queue", Status: "healthy", ResponseTime: 40, Health: 96},
		{Name: "Authentication", Status: "healthy", ResponseTime: 31, Health: 99},
	}

	return services, nil
}

// getEnvironment retrieves the APP_ENV environment variable.
// Defaults to "Development" if not set.
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "Development"
	}
	return env
}
