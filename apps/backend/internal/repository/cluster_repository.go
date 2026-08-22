package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ClusterRepository struct {
	db *gorm.DB
}

func NewClusterRepository(db *gorm.DB) *ClusterRepository {
	return &ClusterRepository{db: db}
}

func (r *ClusterRepository) Create(cluster *models.Cluster) error {
	return r.db.Create(cluster).Error
}

func (r *ClusterRepository) Update(cluster *models.Cluster) error {
	return r.db.Model(cluster).Select("*").Omit("ID", "CreatedAt", "DeletedAt").Updates(cluster).Error
}

func (r *ClusterRepository) Delete(id uuid.UUID, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Cluster{}).Error
}

func (r *ClusterRepository) FindByID(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	var cluster models.Cluster
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&cluster).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepository) FindByProject(projectID uuid.UUID, organizationID uuid.UUID) ([]models.Cluster, error) {
	var clusters []models.Cluster
	if err := r.db.
		Where("project_id = ? AND organization_id = ?", projectID, organizationID).
		Order("is_default DESC").
		Order("created_at ASC").
		Find(&clusters).Error; err != nil {
		return nil, err
	}

	return clusters, nil
}

func (r *ClusterRepository) List(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.Cluster, int64, error) {
	if err := req.Validate("created_at", "updated_at", "name", "provider", "status", "last_validated_at", "last_discovery_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Cluster{}).
		Where("clusters.organization_id = ?", organizationID)

	if req.Search != "" {
		query = query.Where(
			"clusters.name ILIKE ? OR clusters.provider ILIKE ? OR clusters.api_endpoint ILIKE ? OR clusters.region ILIKE ?",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
		)
	}

	if req.Provider != "" {
		query = query.Where("clusters.provider = ?", req.Provider)
	}

	if req.Status != "" {
		query = query.Where("clusters.status = ?", req.Status)
	}

	if req.ProjectID != uuid.Nil {
		query = query.Where("clusters.project_id = ?", req.ProjectID)
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
		case "name":
			sortField = "name"
		case "provider":
			sortField = "provider"
		case "status":
			sortField = "status"
		case "last_validated_at":
			sortField = "last_validated_at"
		case "last_discovery_at":
			sortField = "last_discovery_at"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var clusters []models.Cluster
	err := query.
		Order("clusters.is_default DESC").
		Order("clusters." + sortField + " " + order).
		Limit(req.Limit).
		Offset((req.Page - 1) * req.Limit).
		Find(&clusters).Error
	if err != nil {
		return nil, 0, err
	}

	return clusters, total, nil
}

func (r *ClusterRepository) GetDefaultCluster(projectID uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	var cluster models.Cluster
	if err := r.db.Where("project_id = ? AND organization_id = ? AND is_default = ?", projectID, organizationID, true).First(&cluster).Error; err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (r *ClusterRepository) SetDefaultCluster(projectID, clusterID, organizationID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Cluster{}).Where("project_id = ? AND organization_id = ?", projectID, organizationID).Update("is_default", false).Error; err != nil {
			return err
		}

		result := tx.Model(&models.Cluster{}).
			Where("project_id = ? AND id = ? AND organization_id = ?", projectID, clusterID, organizationID).
			Update("is_default", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func (r *ClusterRepository) UpdateValidation(id uuid.UUID, organizationID uuid.UUID, status string, lastValidatedAt *time.Time, validationError, kubernetesVersion string) error {
	return r.db.Model(&models.Cluster{}).Where("id = ? AND organization_id = ?", id, organizationID).Updates(map[string]interface{}{
		"status":             status,
		"last_validated_at":  lastValidatedAt,
		"validation_error":   validationError,
		"kubernetes_version": kubernetesVersion,
	}).Error
}

func (r *ClusterRepository) UpdateDiscovery(id uuid.UUID, organizationID uuid.UUID, lastDiscoveryAt time.Time) error {
	return r.db.Model(&models.Cluster{}).Where("id = ? AND organization_id = ?", id, organizationID).Update("last_discovery_at", lastDiscoveryAt).Error
}

func (r *ClusterRepository) ProjectBelongsToOrganization(projectID uuid.UUID, organizationID uuid.UUID) (bool, error) {
	var project models.Project

	if err := r.db.Where("id = ? AND organization_id = ?", projectID, organizationID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
