package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(organization *models.Organization) error {
	if err := r.db.Create(organization).Error; err != nil {
		if isDuplicateOrganizationSlugError(err) {
			return apperrors.ErrOrganizationAlreadyExists
		}
		return err
	}

	return nil
}

func (r *OrganizationRepository) GetByID(id uuid.UUID) (*models.Organization, error) {
	var organization models.Organization

	if err := r.db.First(&organization, id).Error; err != nil {
		return nil, err
	}

	return &organization, nil
}

func (r *OrganizationRepository) GetBySlug(slug string) (*models.Organization, error) {
	var organization models.Organization

	if err := r.db.Where("slug = ?", slug).First(&organization).Error; err != nil {
		return nil, err
	}

	return &organization, nil
}

func (r *OrganizationRepository) GetFirst() (*models.Organization, error) {
	var organization models.Organization

	if err := r.db.Order("created_at ASC").First(&organization).Error; err != nil {
		return nil, err
	}

	return &organization, nil
}

func (r *OrganizationRepository) Update(organization *models.Organization) error {
	return r.db.Model(organization).Updates(organization).Error
}

func (r *OrganizationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Organization{}, id).Error
}

func (r *OrganizationRepository) List(req *models.PaginationRequest, ownerID uint) ([]models.Organization, int64, error) {
	if err := req.Validate("name", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Organization{}).Where("owner_id = ?", ownerID)

	if req.Search != "" {
		query = query.Where("name ILIKE ? OR slug ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	switch req.Sort {
	case "name":
		sortField = "name"
	case "updated_at":
		sortField = "updated_at"
	case "created_at":
		sortField = "created_at"
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	organizations := make([]models.Organization, 0)
	if err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&organizations).Error; err != nil {
		return nil, 0, err
	}

	return organizations, total, nil
}

func (r *OrganizationRepository) CountOrganizations() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Organization{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func isDuplicateOrganizationSlugError(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == "idx_organizations_slug"
}
