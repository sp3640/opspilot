package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type DeploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) Create(deployment *models.Deployment) error {
	return r.db.Create(deployment).Error
}

func (r *DeploymentRepository) Update(deployment *models.Deployment) error {
	return r.db.Model(deployment).
		Select("Image", "ImageTag", "Environment", "Namespace", "ReplicaCount", "DeploymentStrategy", "TargetClusterID", "CommitSHA", "Author", "UpdatedBy", "UpdatedAt").
		Updates(deployment).Error
}

func (r *DeploymentRepository) Delete(id uuid.UUID, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Deployment{}).Error
}

func (r *DeploymentRepository) GetByID(id uuid.UUID, organizationID uuid.UUID) (*models.Deployment, error) {
	var deployment models.Deployment
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&deployment).Error; err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (r *DeploymentRepository) GetByIDAnyOrganization(id uuid.UUID) (*models.Deployment, error) {
	var deployment models.Deployment
	if err := r.db.Unscoped().Where("id = ?", id).First(&deployment).Error; err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (r *DeploymentRepository) ListByApplication(applicationID, organizationID uuid.UUID, req *models.PaginationRequest) ([]models.Deployment, int64, error) {
	query := r.db.Model(&models.Deployment{}).Where("organization_id = ? AND application_id = ?", organizationID, applicationID)
	return r.list(query, req)
}

func (r *DeploymentRepository) ListByProject(projectID, organizationID uuid.UUID, req *models.PaginationRequest) ([]models.Deployment, int64, error) {
	query := r.db.Model(&models.Deployment{}).Where("organization_id = ? AND project_id = ?", organizationID, projectID)
	return r.list(query, req)
}

func (r *DeploymentRepository) ListByOrganization(organizationID uuid.UUID, req *models.PaginationRequest) ([]models.Deployment, int64, error) {
	query := r.db.Model(&models.Deployment{}).Where("organization_id = ?", organizationID)
	return r.list(query, req)
}

func (r *DeploymentRepository) GetLatestDeployment(applicationID, organizationID uuid.UUID) (*models.Deployment, error) {
	var deployment models.Deployment
	if err := r.db.
		Where("organization_id = ? AND application_id = ?", organizationID, applicationID).
		Order("created_at DESC").
		First(&deployment).Error; err != nil {
		return nil, err
	}

	return &deployment, nil
}

// GetPreviousDeployment returns the most recent deployment for
// applicationID that was created strictly before before, excluding
// excludeID itself - used to detect a "recovered" deployment (this one
// succeeded and the previous attempt for the same application had failed).
func (r *DeploymentRepository) GetPreviousDeployment(applicationID, organizationID, excludeID uuid.UUID, before time.Time) (*models.Deployment, error) {
	var deployment models.Deployment
	err := r.db.
		Where("organization_id = ? AND application_id = ? AND id <> ? AND created_at < ?", organizationID, applicationID, excludeID, before).
		Order("created_at DESC").
		First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *DeploymentRepository) UpdateStatus(id, organizationID uuid.UUID, status string, startedAt, completedAt *time.Time, updatedBy uint) error {
	updates := map[string]any{
		"status":     status,
		"updated_by": updatedBy,
	}
	if startedAt != nil {
		updates["started_at"] = startedAt
	}
	if completedAt != nil {
		updates["completed_at"] = completedAt
	}

	return r.db.Model(&models.Deployment{}).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Updates(updates).Error
}

func (r *DeploymentRepository) ApplyRollback(deployment *models.Deployment) error {
	return r.db.Model(&models.Deployment{}).
		Where("id = ? AND organization_id = ?", deployment.ID, deployment.OrganizationID).
		Updates(map[string]any{
			"image":               deployment.Image,
			"image_tag":           deployment.ImageTag,
			"environment":         deployment.Environment,
			"namespace":           deployment.Namespace,
			"replica_count":       deployment.ReplicaCount,
			"deployment_strategy": deployment.DeploymentStrategy,
			"status":              deployment.Status,
			"commit_sha":          deployment.CommitSHA,
			"author":              deployment.Author,
			"started_at":          nil,
			"completed_at":        nil,
			"updated_by":          deployment.UpdatedBy,
			"updated_at":          time.Now().UTC(),
		}).Error
}

func (r *DeploymentRepository) list(query *gorm.DB, req *models.PaginationRequest) ([]models.Deployment, int64, error) {
	if err := req.Validate("created_at", "updated_at", "started_at", "completed_at", "status"); err != nil {
		return nil, 0, err
	}

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("image ILIKE ? OR image_tag ILIKE ? OR environment ILIKE ? OR namespace ILIKE ? OR status ILIKE ?", search, search, search, search, search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	if req.Sort != "" {
		sortField = req.Sort
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	items := make([]models.Deployment, 0)
	if err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
