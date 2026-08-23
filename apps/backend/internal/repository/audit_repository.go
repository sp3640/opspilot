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
	query := r.db.Model(&models.AuditLog{}).Where("project_id = ? AND organization_id = ?", projectID, organizationID)
	return r.list(query, req)
}

func (r *AuditRepository) ListByIncidentID(req *models.PaginationRequest, incidentID uint, organizationID uuid.UUID) ([]models.AuditLog, int64, error) {
	query := r.db.Model(&models.AuditLog{}).Where("incident_id = ? AND organization_id = ?", incidentID, organizationID)
	return r.list(query, req)
}

// ListByEntity returns the audit trail for any single entity (e.g. an
// alert), generalizing GetByIncidentID/GetByProjectID to the EntityType +
// EntityID columns every audit log row already carries.
func (r *AuditRepository) ListByEntity(req *models.PaginationRequest, entityType, entityID string, organizationID uuid.UUID) ([]models.AuditLog, int64, error) {
	query := r.db.Model(&models.AuditLog{}).Where("entity_type = ? AND entity_id = ? AND organization_id = ?", entityType, entityID, organizationID)
	return r.list(query, req)
}

// ListByOrganization returns every audit log in the organization, unscoped
// by project/incident/entity - backing the Organization Audit view.
func (r *AuditRepository) ListByOrganization(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.AuditLog, int64, error) {
	query := r.db.Model(&models.AuditLog{}).Where("organization_id = ?", organizationID)
	return r.list(query, req)
}

// list applies every filter Phase 23's audit views need (user, action,
// resource/entity type, application, result, free-text entity-type search,
// and a created_at date range) on top of a caller-scoped query, then
// paginates. Centralizing this once means every list method above gains
// new filters/sort fields identically rather than drifting across four
// near-duplicate query builders.
func (r *AuditRepository) list(query *gorm.DB, req *models.PaginationRequest) ([]models.AuditLog, int64, error) {
	if err := req.Validate("created_at", "entity_type", "action"); err != nil {
		return nil, 0, err
	}

	if req.Search != "" {
		query = query.Where("entity_type ILIKE ?", "%"+req.Search+"%")
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.EntityType != "" {
		query = query.Where("entity_type = ?", req.EntityType)
	}
	if req.Result != "" {
		query = query.Where("result = ?", req.Result)
	}
	if req.UserID != 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.ApplicationID != uuid.Nil {
		query = query.Where("application_id = ?", req.ApplicationID)
	}
	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", *req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("created_at <= ?", *req.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	switch req.Sort {
	case "entity_type":
		sortField = "entity_type"
	case "action":
		sortField = "action"
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
