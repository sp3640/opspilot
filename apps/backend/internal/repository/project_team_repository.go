package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

type ProjectTeamRepository struct {
	db *gorm.DB
}

func NewProjectTeamRepository(db *gorm.DB) *ProjectTeamRepository {
	return &ProjectTeamRepository{db: db}
}

func (r *ProjectTeamRepository) AssignTeam(mapping *models.ProjectTeam) error {
	return r.db.Create(mapping).Error
}

func (r *ProjectTeamRepository) RemoveTeam(projectID, teamID, organizationID uuid.UUID) error {
	return r.db.Where("project_id = ? AND team_id = ? AND organization_id = ?", projectID, teamID, organizationID).Delete(&models.ProjectTeam{}).Error
}

func (r *ProjectTeamRepository) ListProjectTeams(projectID, organizationID uuid.UUID) ([]models.ProjectTeam, error) {
	items := make([]models.ProjectTeam, 0)
	if err := r.db.
		Where("project_id = ? AND organization_id = ?", projectID, organizationID).
		Order("created_at ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ProjectTeamRepository) ListTeamProjects(teamID, organizationID uuid.UUID) ([]models.ProjectTeam, error) {
	items := make([]models.ProjectTeam, 0)
	if err := r.db.
		Where("team_id = ? AND organization_id = ?", teamID, organizationID).
		Order("created_at ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ProjectTeamRepository) IsAssigned(projectID, teamID, organizationID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ProjectTeam{}).
		Where("project_id = ? AND team_id = ? AND organization_id = ?", projectID, teamID, organizationID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
