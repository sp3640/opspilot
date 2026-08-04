package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a repository backed by the supplied GORM database.
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

// Create persists a project and translates duplicate slug violations to a domain error.
func (r *ProjectRepository) Create(project *models.Project) error {
	if err := r.db.Create(project).Error; err != nil {
		if isDuplicateSlugError(err) {
			return apperrors.ErrProjectAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID returns the project identified by id.
func (r *ProjectRepository) GetByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project

	err := r.db.First(&project, id).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// GetByIDAndOrganizationID returns a project only when it belongs to organizationID.
func (r *ProjectRepository) GetByIDAndOrganizationID(id uuid.UUID, organizationID uuid.UUID) (*models.Project, error) {
	var project models.Project

	err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&project).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// GetBySlug returns the active project identified by slug.
func (r *ProjectRepository) GetBySlug(slug string) (*models.Project, error) {
	var project models.Project

	if err := r.db.Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, err
	}

	return &project, nil
}

// GetAllByOrganizationID returns all projects in an organization.
func (r *ProjectRepository) GetAllByOrganizationID(organizationID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.Where("organization_id = ?", organizationID).Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// ListByOrganizationID returns a paginated project list and total for organizationID.
func (r *ProjectRepository) ListByOrganizationID(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.Project, int64, error) {
	if err := req.Validate("name", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}
	query := r.db.Model(&models.Project{}).Where("organization_id = ?", organizationID)

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

// CountByOrganizationID returns the number of active projects in organizationID.
func (r *ProjectRepository) CountByOrganizationID(organizationID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Project{}).Where("organization_id = ?", organizationID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// Update writes the project's non-zero fields and updates its UpdatedAt timestamp.
func (r *ProjectRepository) Update(project *models.Project) error {
	return r.db.Model(project).Updates(project).Error
}

// Delete soft-deletes the project identified by id and organization.
func (r *ProjectRepository) Delete(id uuid.UUID, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Project{}).Error
}

func isDuplicateSlugError(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == "idx_projects_slug"
}
