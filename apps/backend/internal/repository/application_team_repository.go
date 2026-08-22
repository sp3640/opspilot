package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ApplicationTeamRepository struct {
	db *gorm.DB
}

func NewApplicationTeamRepository(db *gorm.DB) *ApplicationTeamRepository {
	return &ApplicationTeamRepository{db: db}
}

func (r *ApplicationTeamRepository) AssignTeam(mapping *models.ApplicationTeam) error {
	return r.db.Create(mapping).Error
}

func (r *ApplicationTeamRepository) RemoveTeam(applicationID, teamID, organizationID uuid.UUID) error {
	return r.db.Where("application_id = ? AND team_id = ? AND organization_id = ?", applicationID, teamID, organizationID).Delete(&models.ApplicationTeam{}).Error
}

func (r *ApplicationTeamRepository) ListApplicationTeams(applicationID, organizationID uuid.UUID) ([]models.ApplicationTeam, error) {
	items := make([]models.ApplicationTeam, 0)
	if err := r.db.
		Where("application_id = ? AND organization_id = ?", applicationID, organizationID).
		Order("created_at ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ApplicationTeamRepository) ListTeamApplications(teamID, organizationID uuid.UUID) ([]models.ApplicationTeam, error) {
	items := make([]models.ApplicationTeam, 0)
	if err := r.db.
		Where("team_id = ? AND organization_id = ?", teamID, organizationID).
		Order("created_at ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ApplicationTeamRepository) IsAssigned(applicationID, teamID, organizationID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ApplicationTeam{}).
		Where("application_id = ? AND team_id = ? AND organization_id = ?", applicationID, teamID, organizationID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
