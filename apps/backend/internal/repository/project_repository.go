package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
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

func (r *ProjectRepository) GetByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project

	err := r.db.First(&project, id).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) GetByIDAndUserID(id uuid.UUID, userID uint) (*models.Project, error) {
	var project models.Project

	err := r.db.Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, apperrors.ErrProjectForbidden
	}

	return &project, nil
}

func (r *ProjectRepository) GetAllByUserID(userID uint) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.Where("owner_id = ?", userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectRepository) ListByUserID(req *models.PaginationRequest, userID uint) ([]models.Project, int64, error) {
	if err := req.Validate("name", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}
	query := r.db.Model(&models.Project{}).Where("owner_id = ?", userID)

	if req.Search != "" {
		query = query.Where("name ILIKE ?", "%"+req.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	if req.Sort != "" {
		switch req.Sort {
		case "name":
			sortField = "name"
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

	var projects []models.Project
	err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *ProjectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Project{}, id).Error
}
