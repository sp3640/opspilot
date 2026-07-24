package services

import (
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepository
}

func NewDashboardService(repo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) GetSummary(userID uint) (*repository.DashboardSummary, error) {
	return s.repo.GetSummaryByUserID(userID)
}

func (s *DashboardService) GetRecentIncidents(userID uint, limit int) ([]models.Incident, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.ListRecentIncidentsByUserID(userID, limit)
}

func (s *DashboardService) GetRecentActivity(userID uint, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.ListRecentActivityByUserID(userID, limit)
}

func (s *DashboardService) GetStats(userID uint) (*repository.IncidentStats, error) {
	return s.repo.GetIncidentStatsByUserID(userID)
}
