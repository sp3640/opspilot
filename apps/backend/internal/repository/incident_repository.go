package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type IncidentRepository struct {
	db *gorm.DB
}

func NewIncidentRepository(db *gorm.DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

func (r *IncidentRepository) Create(incident *models.Incident) error {
	return r.db.Create(incident).Error
}

func (r *IncidentRepository) GetByID(id uint, organizationID uuid.UUID) (*models.Incident, error) {
	var incident models.Incident
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&incident).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

// GetByIDAndOrganizationID returns an incident only when it belongs to organizationID.
func (r *IncidentRepository) GetByIDAndOrganizationID(id uint, organizationID uuid.UUID) (*models.Incident, error) {
	var incident models.Incident

	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&incident).Error; err != nil {
		return nil, err
	}

	return &incident, nil
}

func (r *IncidentRepository) ListByOrganizationID(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.Incident, int64, error) {
	if err := req.Validate("title", "severity", "status", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Incident{}).Where("organization_id = ?", organizationID)

	if req.Search != "" {
		query = query.Where(
			"title ILIKE ? OR description ILIKE ?",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
		)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Severity != "" {
		query = query.Where("severity = ?", req.Severity)
	}

	if req.ProjectID != uuid.Nil {
		query = query.Where("project_id = ?", req.ProjectID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	if req.Sort != "" {
		switch req.Sort {
		case "title":
			sortField = "title"
		case "severity":
			sortField = "severity"
		case "status":
			sortField = "status"
		case "created_at":
			sortField = "created_at"
		case "updated_at":
			sortField = "updated_at"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var incidents []models.Incident

	err := query.
		Order(sortField + " " + order).
		Limit(req.Limit).
		Offset((req.Page - 1) * req.Limit).
		Find(&incidents).Error

	if err != nil {
		return nil, 0, err
	}

	return incidents, total, nil
}

// Update writes every field of incident, including zero values - callers
// pass a fully-loaded-then-mutated incident (a full replace), and fields
// like ApplicationID/OwnerTeamID/ResolvedAt must be clearable to nil, which
// GORM's default Updates(struct) skips (it treats a nil pointer as "no
// change" unless every field is explicitly selected).
func (r *IncidentRepository) Update(incident *models.Incident) error {
	return r.db.Model(incident).Select("*").Updates(incident).Error
}

func (r *IncidentRepository) Delete(id uint, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Incident{}).Error
}

func (r *IncidentRepository) ProjectBelongsToOrganization(projectID, organizationID uuid.UUID) (bool, error) {
	var project models.Project

	if err := r.db.Where("id = ? AND organization_id = ?", projectID, organizationID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
