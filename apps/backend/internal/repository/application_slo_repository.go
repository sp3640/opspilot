package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ApplicationSLORepository struct {
	db *gorm.DB
}

func NewApplicationSLORepository(db *gorm.DB) *ApplicationSLORepository {
	return &ApplicationSLORepository{db: db}
}

func (r *ApplicationSLORepository) GetByApplicationID(applicationID, organizationID uuid.UUID) (*models.ApplicationSLO, error) {
	var slo models.ApplicationSLO
	if err := r.db.Where("application_id = ? AND organization_id = ?", applicationID, organizationID).First(&slo).Error; err != nil {
		return nil, err
	}
	return &slo, nil
}

// Upsert creates slo if no row exists for its ApplicationID yet, otherwise
// updates the existing row's target/window in place (preserving its ID and
// CreatedBy) - there is exactly one SLO configuration per application.
func (r *ApplicationSLORepository) Upsert(slo *models.ApplicationSLO) error {
	existing, err := r.GetByApplicationID(slo.ApplicationID, slo.OrganizationID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(slo).Error
		}
		return err
	}

	existing.TargetPercentage = slo.TargetPercentage
	existing.WindowDays = slo.WindowDays
	if err := r.db.Save(existing).Error; err != nil {
		return err
	}
	*slo = *existing
	return nil
}
