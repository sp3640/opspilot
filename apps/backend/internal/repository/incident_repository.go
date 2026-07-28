package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
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

func (r *IncidentRepository) GetByID(id uint) (*models.Incident, error) {
	var incident models.Incident
	if err := r.db.First(&incident, id).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

func (r *IncidentRepository) GetByIDAndUserID(id, userID uint) (*models.Incident, error) {
	var incident models.Incident

	if err := r.db.Where("id = ?", id).First(&incident).Error; err != nil {
		return nil, err
	}

	if incident.UserID != userID {
		return nil, apperrors.ErrProjectForbidden
	}

	return &incident, nil
}

func (r *IncidentRepository) GetAllByUserID(userID uint) ([]models.Incident, error) {
	var incidents []models.Incident

	if err := r.db.Where("user_id = ?", userID).Find(&incidents).Error; err != nil {
		return nil, err
	}

	return incidents, nil
}

func (r *IncidentRepository) ListByUserID(req *models.PaginationRequest, userID uint) ([]models.Incident, int64, error) {
	if err := req.Validate("title", "severity", "status", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Incident{}).Where("user_id = ?", userID)

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

func (r *IncidentRepository) Update(incident *models.Incident) error {
	return r.db.Save(incident).Error
}

func (r *IncidentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Incident{}, id).Error
}

func (r *IncidentRepository) ProjectBelongsToUser(projectID uuid.UUID, userID uint) (bool, error) {
	var project models.Project

	if err := r.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return project.OwnerID == userID, nil
}
