package repository

import (
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepository) GetByIncidentID(incidentID uint) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.Where("incident_id = ?", incidentID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AuditRepository) GetByProjectID(projectID uint) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
