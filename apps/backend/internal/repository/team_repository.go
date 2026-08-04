package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *models.Team) error {
	if err := r.db.Create(team).Error; err != nil {
		if isDuplicateTeamNameError(err) {
			return apperrors.ErrTeamAlreadyExists
		}
		return err
	}

	return nil
}

func (r *TeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	if err := r.db.First(&team, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &team, nil
}

func (r *TeamRepository) List(req *models.PaginationRequest, organizationID uuid.UUID) ([]models.Team, int64, error) {
	if err := req.Validate("name", "created_at", "updated_at"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Team{}).Where("organization_id = ?", organizationID)
	if req.Search != "" {
		query = query.Where("name ILIKE ?", "%"+req.Search+"%")
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

	items := make([]models.Team, 0)
	if err := query.Order(sortField + " " + order).Limit(req.Limit).Offset((req.Page - 1) * req.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *TeamRepository) Update(team *models.Team) error {
	if err := r.db.Model(team).Updates(team).Error; err != nil {
		if isDuplicateTeamNameError(err) {
			return apperrors.ErrTeamAlreadyExists
		}
		return err
	}

	return nil
}

func (r *TeamRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Team{}, "id = ?", id).Error
}

func (r *TeamRepository) ExistsByName(name string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Team{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *TeamRepository) ExistsByNameInOrganization(name string, organizationID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Team{}).Where("organization_id = ? AND name = ?", organizationID, name).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func isDuplicateTeamNameError(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == "idx_teams_org_name"
}
