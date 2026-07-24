package repository

import (
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) GetByID(id uint) (*models.Project, error) {
	var project models.Project

	err := r.db.First(&project, id).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) GetAllByUserID(userID uint) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.Where("user_id = ?", userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}