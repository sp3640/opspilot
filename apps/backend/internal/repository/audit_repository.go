package repository

import (
	"github.com/google/uuid"
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

func (r *AuditRepository) GetByIncidentID(incidentID uint, organizationID uuid.UUID) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.Where("incident_id = ? AND organization_id = ?", incidentID, organizationID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AuditRepository) GetByProjectID(projectID, organizationID uuid.UUID) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.Where("project_id = ? AND organization_id = ?", projectID, organizationID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AuditRepository) ListByProjectID(req *models.PaginationRequest, projectID, organizationID uuid.UUID) ([]models.AuditLog, int64, error) {
	if err := req.Validate("created_at", "entity_type", "action"); err != nil {
		return nil, 0, err
	}
	query := r.db.Model(&models.AuditLog{}).Where("project_id = ? AND organization_id = ?", projectID, organizationID)

	if req.Search != "" {
		query = query.Where("entity_type ILIKE ?", "%"+req.Search+"%")
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.EntityType != "" {
		query = query.Where("entity_type = ?", req.EntityType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	if req.Sort != "" {
		switch req.Sort {
		case "created_at":
			sortField = "created_at"
		case "entity_type":
			sortField = "entity_type"
		case "action":
			sortField = "action"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var logs []models.AuditLog
	err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditRepository) ListByIncidentID(req *models.PaginationRequest, incidentID uint, organizationID uuid.UUID) ([]models.AuditLog, int64, error) {
	if err := req.Validate("created_at", "entity_type", "action"); err != nil {
		return nil, 0, err
	}
	query := r.db.Model(&models.AuditLog{}).Where("incident_id = ? AND organization_id = ?", incidentID, organizationID)

	if req.Search != "" {
		query = query.Where("entity_type ILIKE ?", "%"+req.Search+"%")
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.EntityType != "" {
		query = query.Where("entity_type = ?", req.EntityType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	if req.Sort != "" {
		switch req.Sort {
		case "created_at":
			sortField = "created_at"
		case "entity_type":
			sortField = "entity_type"
		case "action":
			sortField = "action"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var logs []models.AuditLog
	err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditRepository) ClearIncidentReference(incidentID uint, organizationID uuid.UUID) error {
	return r.db.Model(&models.AuditLog{}).
		Where("incident_id = ? AND organization_id = ?", incidentID, organizationID).
		Update("incident_id", nil).Error
}
