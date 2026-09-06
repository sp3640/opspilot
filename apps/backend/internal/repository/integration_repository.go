package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type IntegrationRepository struct {
	db *gorm.DB
}

func NewIntegrationRepository(db *gorm.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

func (r *IntegrationRepository) Create(integration *models.Integration) error {
	return r.db.Create(integration).Error
}

func (r *IntegrationRepository) GetByID(id, organizationID uuid.UUID) (*models.Integration, error) {
	var integration models.Integration
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&integration).Error; err != nil {
		return nil, err
	}
	return &integration, nil
}

// ExistsByTypeAndName reports whether organizationID already has a
// (non-deleted) integration of the given type and name - the "duplicate
// integration" case the service layer rejects on create.
func (r *IntegrationRepository) ExistsByTypeAndName(organizationID uuid.UUID, integrationType, name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Integration{}).
		Where("organization_id = ? AND type = ? AND name = ?", organizationID, integrationType, name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetFirstByType returns the organization's integration of the given type,
// if one exists - used by OAuth-based connect flows so reconnecting/
// re-authorizing updates the same row instead of creating a duplicate.
func (r *IntegrationRepository) GetFirstByType(organizationID uuid.UUID, integrationType string) (*models.Integration, error) {
	var integration models.Integration
	err := r.db.
		Where("organization_id = ? AND type = ?", organizationID, integrationType).
		Order("created_at ASC").
		First(&integration).Error
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (r *IntegrationRepository) Update(integration *models.Integration) error {
	return r.db.Save(integration).Error
}

func (r *IntegrationRepository) Delete(id, organizationID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, organizationID).Delete(&models.Integration{}).Error
}

// List returns a paginated, optionally type/status-filtered set of
// integrations for organizationID. integrationType == "" / status == ""
// mean "no filter".
func (r *IntegrationRepository) List(req *models.PaginationRequest, organizationID uuid.UUID, integrationType, status string) ([]models.Integration, int64, error) {
	if err := req.Validate("created_at", "name", "type", "status"); err != nil {
		return nil, 0, err
	}

	query := r.db.Model(&models.Integration{}).Where("organization_id = ?", organizationID)
	if integrationType != "" {
		query = query.Where("type = ?", integrationType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	var integrations []models.Integration
	err := query.
		Order(req.Sort + " " + order).
		Limit(req.Limit).
		Offset((req.Page - 1) * req.Limit).
		Find(&integrations).Error
	if err != nil {
		return nil, 0, err
	}

	return integrations, total, nil
}

// UpdateCheckResult records the outcome of a TestConnection/HealthCheck
// attempt - status, LastCheckedAt, and LastError together, in one write, so
// a caller can never observe a status change without its accompanying
// check timestamp.
func (r *IntegrationRepository) UpdateCheckResult(id, organizationID uuid.UUID, status string, checkedAt time.Time, lastError string) error {
	return r.db.Model(&models.Integration{}).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Updates(map[string]any{
			"status":          status,
			"last_checked_at": checkedAt,
			"last_error":      lastError,
		}).Error
}
