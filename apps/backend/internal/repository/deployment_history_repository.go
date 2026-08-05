package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type DeploymentHistoryRepository struct {
	db *gorm.DB
}

func NewDeploymentHistoryRepository(db *gorm.DB) *DeploymentHistoryRepository {
	return &DeploymentHistoryRepository{db: db}
}

func (r *DeploymentHistoryRepository) Create(history *models.DeploymentHistory) error {
	return r.db.Create(history).Error
}

func (r *DeploymentHistoryRepository) GetByID(id, organizationID uuid.UUID) (*models.DeploymentHistory, error) {
	var history models.DeploymentHistory
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&history).Error; err != nil {
		return nil, err
	}

	return &history, nil
}

func (r *DeploymentHistoryRepository) ListDeploymentHistory(deploymentID, organizationID uuid.UUID, req *models.PaginationRequest) ([]models.DeploymentHistory, int64, error) {
	if err := req.Validate("revision", "created_at", "status"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.DeploymentHistory{}).
		Where("deployment_id = ? AND organization_id = ?", deploymentID, organizationID)

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where(
			"image ILIKE ? OR image_tag ILIKE ? OR environment ILIKE ? OR namespace ILIKE ? OR status ILIKE ? OR change_summary ILIKE ?",
			search,
			search,
			search,
			search,
			search,
			search,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "revision"
	if req.Sort != "" {
		sortField = req.Sort
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	items := make([]models.DeploymentHistory, 0)
	if err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *DeploymentHistoryRepository) GetLatestRevision(deploymentID, organizationID uuid.UUID) (int, error) {
	var history models.DeploymentHistory
	if err := r.db.Where("deployment_id = ? AND organization_id = ?", deploymentID, organizationID).
		Order("revision DESC").
		First(&history).Error; err != nil {
		return 0, err
	}

	return history.Revision, nil
}

func (r *DeploymentHistoryRepository) GetRevision(deploymentID, organizationID uuid.UUID, revision int) (*models.DeploymentHistory, error) {
	var history models.DeploymentHistory
	if err := r.db.Where("deployment_id = ? AND organization_id = ? AND revision = ?", deploymentID, organizationID, revision).First(&history).Error; err != nil {
		return nil, err
	}

	return &history, nil
}

func (r *DeploymentHistoryRepository) DeleteOldHistory(deploymentID, organizationID uuid.UUID, keepLatest int) error {
	if keepLatest <= 0 {
		return r.db.Where("deployment_id = ? AND organization_id = ?", deploymentID, organizationID).Delete(&models.DeploymentHistory{}).Error
	}

	keepSubquery := r.db.Model(&models.DeploymentHistory{}).
		Select("id").
		Where("deployment_id = ? AND organization_id = ?", deploymentID, organizationID).
		Order("revision DESC").
		Limit(keepLatest)

	return r.db.Where("deployment_id = ? AND organization_id = ? AND id NOT IN (?)", deploymentID, organizationID, keepSubquery).
		Delete(&models.DeploymentHistory{}).Error
}
