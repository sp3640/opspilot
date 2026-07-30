package repository

import (
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) GetByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) GetByIncidentID(incidentID uint) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.db.Where("incident_id = ?", incidentID).Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *CommentRepository) ListByIncidentID(req *models.PaginationRequest, incidentID uint) ([]models.Comment, int64, error) {
	if err := req.Validate("created_at", "updated_at"); err != nil {
		return nil, 0, err
	}
	query := r.db.Model(&models.Comment{}).Where("incident_id = ?", incidentID)

	if req.Search != "" {
		query = query.Where("content ILIKE ?", "%"+req.Search+"%")
	}
	if req.IncidentID != 0 {
		query = query.Where("incident_id = ?", req.IncidentID)
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
		case "updated_at":
			sortField = "updated_at"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var comments []models.Comment
	err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *CommentRepository) Update(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *CommentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}

func (r *CommentRepository) DeleteByIncidentID(incidentID uint) error {
	return r.db.Where("incident_id = ?", incidentID).Delete(&models.Comment{}).Error
}
