package repository

import (
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

func (r *IncidentRepository) Update(incident *models.Incident) error {
	return r.db.Save(incident).Error
}

func (r *IncidentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Incident{}, id).Error
}

func (r *IncidentRepository) ProjectBelongsToUser(projectID, userID uint) (bool, error) {
	var project models.Project
	if err := r.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return project.UserID == userID, nil
}
