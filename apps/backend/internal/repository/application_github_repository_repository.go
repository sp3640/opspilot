package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

// ApplicationGitHubRepositoryRepository persists the Application <-> GitHub
// repository mapping (Sprint 28).
type ApplicationGitHubRepositoryRepository struct {
	db *gorm.DB
}

func NewApplicationGitHubRepositoryRepository(db *gorm.DB) *ApplicationGitHubRepositoryRepository {
	return &ApplicationGitHubRepositoryRepository{db: db}
}

func (r *ApplicationGitHubRepositoryRepository) GetByApplicationID(applicationID, organizationID uuid.UUID) (*models.ApplicationGitHubRepository, error) {
	var mapping models.ApplicationGitHubRepository
	err := r.db.
		Where("application_id = ? AND organization_id = ?", applicationID, organizationID).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// Upsert replaces any existing mapping for mapping.ApplicationID with the
// new GitHubRepositoryID (one repository per application) - a re-mapping is
// a normal, expected operation, not an error.
func (r *ApplicationGitHubRepositoryRepository) Upsert(mapping *models.ApplicationGitHubRepository) error {
	var existing models.ApplicationGitHubRepository
	err := r.db.Where("application_id = ?", mapping.ApplicationID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(mapping).Error
		}
		return err
	}

	existing.GitHubRepositoryID = mapping.GitHubRepositoryID
	if err := r.db.Save(&existing).Error; err != nil {
		return err
	}
	*mapping = existing
	return nil
}

func (r *ApplicationGitHubRepositoryRepository) DeleteByApplicationID(applicationID, organizationID uuid.UUID) error {
	return r.db.
		Where("application_id = ? AND organization_id = ?", applicationID, organizationID).
		Delete(&models.ApplicationGitHubRepository{}).Error
}
