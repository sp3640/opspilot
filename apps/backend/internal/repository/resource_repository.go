package repository

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ResourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) Create(resource *models.Resource) error {
	return r.db.Create(resource).Error
}

func (r *ResourceRepository) BulkCreate(resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	return r.db.Create(&resources).Error
}

func (r *ResourceRepository) Update(resource *models.Resource) error {
	return r.db.Model(resource).Select("*").Omit("ID", "CreatedAt", "DeletedAt").Updates(resource).Error
}

func (r *ResourceRepository) BulkUpdate(resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for index := range resources {
			resource := resources[index]
			if err := tx.Model(&resource).Select("*").Omit("ID", "CreatedAt", "DeletedAt").Updates(&resource).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ResourceRepository) FindByID(id uuid.UUID, organizationID uuid.UUID) (*models.Resource, error) {
	var resource models.Resource
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&resource).Error; err != nil {
		return nil, err
	}

	return &resource, nil
}

func (r *ResourceRepository) FindByExternalID(projectID, organizationID uuid.UUID, kind, externalID string) (*models.Resource, error) {
	var resource models.Resource
	if err := r.db.Unscoped().Where("project_id = ? AND organization_id = ? AND kind = ? AND external_id = ?", projectID, organizationID, kind, externalID).First(&resource).Error; err != nil {
		return nil, err
	}

	return &resource, nil
}

func (r *ResourceRepository) List(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.Resource, int64, error) {
	if err := req.Validate("created_at", "updated_at", "name", "kind", "status", "health"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Resource{}).
		Where("resources.organization_id = ?", organizationID)

	if req.Search != "" {
		query = query.Where(
			"resources.name ILIKE ? OR resources.display_name ILIKE ? OR resources.external_id ILIKE ? OR resources.namespace ILIKE ? OR resources.cluster ILIKE ?",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
		)
	}

	if req.Status != "" {
		query = query.Where("resources.status = ?", req.Status)
	}

	if req.Kind != "" {
		query = query.Where("resources.kind = ?", req.Kind)
	}

	if req.Health != "" {
		query = query.Where("resources.health = ?", req.Health)
	}

	if req.ProjectID != uuid.Nil {
		query = query.Where("resources.project_id = ?", req.ProjectID)
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
		case "kind":
			sortField = "kind"
		case "status":
			sortField = "status"
		case "health":
			sortField = "health"
		}
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var resources []models.Resource
	err := query.
		Order("resources." + sortField + " " + order).
		Limit(req.Limit).
		Offset((req.Page - 1) * req.Limit).
		Find(&resources).Error
	if err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

func (r *ResourceRepository) ListByProject(projectID, organizationID uuid.UUID) ([]models.Resource, error) {
	var resources []models.Resource
	if err := r.db.Where("project_id = ? AND organization_id = ?", projectID, organizationID).Order("kind ASC").Order("name ASC").Find(&resources).Error; err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *ResourceRepository) SoftDelete(id, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Resource{}).Error
}

func (r *ResourceRepository) BulkSoftDelete(resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(resources))
	for _, resource := range resources {
		ids = append(ids, resource.ID)
	}

	return r.db.Where("id IN ?", ids).Delete(&models.Resource{}).Error
}

func (r *ResourceRepository) Restore(id uuid.UUID) error {
	return r.db.Unscoped().Model(&models.Resource{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *ResourceRepository) BulkRestore(resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(resources))
	for _, resource := range resources {
		ids = append(ids, resource.ID)
	}

	return r.db.Unscoped().Model(&models.Resource{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
}

func (r *ResourceRepository) Delete(id, organizationID uuid.UUID) error {
	return r.db.Unscoped().Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Resource{}).Error
}

func (r *ResourceRepository) Exists(projectID, organizationID uuid.UUID, kind, externalID string) (bool, error) {
	var count int64
	if err := r.db.Unscoped().Model(&models.Resource{}).Where("project_id = ? AND organization_id = ? AND kind = ? AND external_id = ?", projectID, organizationID, kind, externalID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ResourceRepository) ProjectBelongsToOrganization(projectID, organizationID uuid.UUID) (bool, error) {
	var project models.Project

	if err := r.db.Where("id = ? AND organization_id = ?", projectID, organizationID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *ResourceRepository) ListResources(ctx context.Context, projectID string) ([]models.Resource, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	err = r.db.WithContext(ctx).Unscoped().Where("project_id = ?", projectUUID).Order("kind ASC").Order("name ASC").Find(&resources).Error
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *ResourceRepository) CreateResources(ctx context.Context, resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(&resources).Error
}

func (r *ResourceRepository) UpdateResources(ctx context.Context, resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for index := range resources {
			resource := resources[index]
			if err := tx.Model(&resource).Select("*").Omit("ID", "CreatedAt", "DeletedAt").Updates(&resource).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ResourceRepository) SoftDeleteResources(ctx context.Context, resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(resources))
	for _, resource := range resources {
		ids = append(ids, resource.ID)
	}

	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&models.Resource{}).Error
}

func (r *ResourceRepository) RestoreResources(ctx context.Context, resources []models.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(resources))
	for _, resource := range resources {
		ids = append(ids, resource.ID)
	}

	return r.db.WithContext(ctx).Unscoped().Model(&models.Resource{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
}

func resourceEntityID(resource models.Resource) string {
	if resource.ID != uuid.Nil {
		return resource.ID.String()
	}

	return resource.ProjectID.String() + ":" + resource.Kind + ":" + resource.ExternalID
}

func resourceProjectIDString(projectID uuid.UUID) string {
	if projectID == uuid.Nil {
		return ""
	}

	return projectID.String()
}

func resourceCountString(value int) string {
	return strconv.Itoa(value)
}
