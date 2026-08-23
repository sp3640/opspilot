package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ApplicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) CreateApplication(application *models.Application) error {
	if err := r.db.Create(application).Error; err != nil {
		if isDuplicateApplicationSlugError(err) {
			return apperrors.ErrApplicationAlreadyExists
		}
		return err
	}

	return nil
}

func (r *ApplicationRepository) UpdateApplication(application *models.Application) error {
	return r.db.Model(application).Select("*").Omit("ID", "OrganizationID", "ProjectID", "CreatedAt").Updates(application).Error
}

func (r *ApplicationRepository) DeleteApplication(id uuid.UUID, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Application{}).Error
}

func (r *ApplicationRepository) GetApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
	var application models.Application
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&application).Error; err != nil {
		return nil, err
	}

	return &application, nil
}

func (r *ApplicationRepository) FindByIDAnyOrganization(id uuid.UUID) (*models.Application, error) {
	var application models.Application
	if err := r.db.Where("id = ?", id).First(&application).Error; err != nil {
		return nil, err
	}

	return &application, nil
}

func (r *ApplicationRepository) ListApplicationsByProject(projectID uuid.UUID, organizationID uuid.UUID, req *models.PaginationRequest) ([]models.Application, int64, error) {
	if err := req.Validate("name", "slug", "runtime", "status", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Application{}).Where("organization_id = ? AND project_id = ?", organizationID, projectID)

	if req.Search != "" {
		query = query.Where(
			"name ILIKE ? OR slug ILIKE ? OR description ILIKE ? OR repository_url ILIKE ? OR default_branch ILIKE ? OR runtime ILIKE ? OR status ILIKE ? OR environment ILIKE ?",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
		)
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
		case "slug":
			sortField = "slug"
		case "runtime":
			sortField = "runtime"
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

	var applications []models.Application
	if err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&applications).Error; err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}

// ListApplicationsByOrganization returns every application across every
// project in the organization - used by the operations dashboard, which
// needs an org-wide application count/list rather than one project at a
// time.
func (r *ApplicationRepository) ListApplicationsByOrganization(organizationID uuid.UUID, req *models.PaginationRequest) ([]models.Application, int64, error) {
	if err := req.Validate("name", "slug", "runtime", "status", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Application{}).Where("organization_id = ?", organizationID)

	if req.Search != "" {
		query = query.Where(
			"name ILIKE ? OR slug ILIKE ? OR description ILIKE ? OR repository_url ILIKE ? OR default_branch ILIKE ? OR runtime ILIKE ? OR status ILIKE ? OR environment ILIKE ?",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
			"%"+req.Search+"%",
		)
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
		case "slug":
			sortField = "slug"
		case "runtime":
			sortField = "runtime"
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

	var applications []models.Application
	if err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&applications).Error; err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}

func (r *ApplicationRepository) GetBySlug(projectID uuid.UUID, organizationID uuid.UUID, slug string) (*models.Application, error) {
	var application models.Application
	if err := r.db.Where("project_id = ? AND organization_id = ? AND slug = ?", projectID, organizationID, slug).First(&application).Error; err != nil {
		return nil, err
	}

	return &application, nil
}

func isDuplicateApplicationSlugError(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == "idx_applications_project_slug"
}
