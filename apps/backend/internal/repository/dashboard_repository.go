package repository

import (
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

type DashboardSummary struct {
	TotalProjects     int64 `json:"totalProjects"`
	TotalIncidents    int64 `json:"totalIncidents"`
	OpenIncidents     int64 `json:"openIncidents"`
	CriticalIncidents int64 `json:"criticalIncidents"`
	ResolvedIncidents int64 `json:"resolvedIncidents"`
}

type IncidentStats struct {
	ByStatus   map[string]int64 `json:"byStatus"`
	BySeverity map[string]int64 `json:"bySeverity"`
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetSummaryByUserID(userID uint) (*DashboardSummary, error) {
	summary := &DashboardSummary{}

	// Total Projects
	if err := r.db.Model(&models.Project{}).
		Where("owner_id = ?", userID).
		Count(&summary.TotalProjects).Error; err != nil {
		return nil, err
	}

	// Total Incidents
	if err := r.db.Model(&models.Incident{}).
		Where("user_id = ?", userID).
		Count(&summary.TotalIncidents).Error; err != nil {
		return nil, err
	}

	// Open Incidents
	if err := r.db.Model(&models.Incident{}).
		Where("user_id = ? AND status = ?", userID, "OPEN").
		Count(&summary.OpenIncidents).Error; err != nil {
		return nil, err
	}

	// Critical Incidents
	if err := r.db.Model(&models.Incident{}).
		Where("user_id = ? AND severity = ?", userID, "P1").
		Count(&summary.CriticalIncidents).Error; err != nil {
		return nil, err
	}

	// Resolved Incidents
	if err := r.db.Model(&models.Incident{}).
		Where("user_id = ? AND status = ?", userID, "RESOLVED").
		Count(&summary.ResolvedIncidents).Error; err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *DashboardRepository) ListRecentIncidentsByUserID(userID uint, limit int) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

func (r *DashboardRepository) ListRecentActivityByUserID(userID uint, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *DashboardRepository) GetIncidentStatsByUserID(userID uint) (*IncidentStats, error) {
	stats := &IncidentStats{
		ByStatus:   make(map[string]int64),
		BySeverity: make(map[string]int64),
	}

	var statusGroups []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&models.Incident{}).
		Where("user_id = ?", userID).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusGroups).Error; err != nil {
		return nil, err
	}
	for _, item := range statusGroups {
		stats.ByStatus[item.Status] = item.Count
	}

	var severityGroups []struct {
		Severity string
		Count    int64
	}
	if err := r.db.Model(&models.Incident{}).
		Where("user_id = ?", userID).
		Select("severity, COUNT(*) as count").
		Group("severity").
		Scan(&severityGroups).Error; err != nil {
		return nil, err
	}
	for _, item := range severityGroups {
		stats.BySeverity[item.Severity] = item.Count
	}

	return stats, nil
}
