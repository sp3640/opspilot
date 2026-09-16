package repository

import (
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"gorm.io/gorm"
)

// GitHubRepositoryRepository persists the GitHubRepository cache rows
// (Sprint 28). Organization isolation is enforced on every read the same
// way every other Sprint 27/28 repository does it: organizationID is
// always part of the WHERE clause, never trusted from the caller alone.
type GitHubRepositoryRepository struct {
	db *gorm.DB
}

func NewGitHubRepositoryRepository(db *gorm.DB) *GitHubRepositoryRepository {
	return &GitHubRepositoryRepository{db: db}
}

func (r *GitHubRepositoryRepository) GetByID(id, organizationID uuid.UUID) (*models.GitHubRepository, error) {
	var repo models.GitHubRepository
	if err := r.db.Where("id = ? AND organization_id = ?", id, organizationID).First(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *GitHubRepositoryRepository) ListByIntegration(integrationID, organizationID uuid.UUID) ([]models.GitHubRepository, error) {
	var repos []models.GitHubRepository
	err := r.db.
		Where("integration_id = ? AND organization_id = ?", integrationID, organizationID).
		Order("full_name ASC").
		Find(&repos).Error
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// ListByIntegrationPaginated returns a stable, organization-scoped page of
// cached repositories. Full name followed by ID makes paging deterministic
// even if a repository is renamed between requests.
func (r *GitHubRepositoryRepository) ListByIntegrationPaginated(integrationID, organizationID uuid.UUID, page, limit int) ([]models.GitHubRepository, int64, error) {
	query := r.db.Model(&models.GitHubRepository{}).
		Where("integration_id = ? AND organization_id = ?", integrationID, organizationID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var repos []models.GitHubRepository
	if err := query.Order("full_name ASC").Order("id ASC").Limit(limit).Offset((page - 1) * limit).Find(&repos).Error; err != nil {
		return nil, 0, err
	}
	return repos, total, nil
}

// Upsert inserts repo if no row exists yet for (IntegrationID, GitHubID),
// otherwise updates the mutable metadata fields in place - the mechanism
// that keeps repository discovery from ever duplicating a row on repeated
// syncs. Selected is deliberately never overwritten by a sync (it's a
// user choice, not GitHub metadata).
func (r *GitHubRepositoryRepository) Upsert(repo *models.GitHubRepository) error {
	var existing models.GitHubRepository
	err := r.db.
		Where("integration_id = ? AND github_id = ?", repo.IntegrationID, repo.GitHubID).
		First(&existing).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(repo).Error
		}
		return err
	}

	existing.Owner = repo.Owner
	existing.Name = repo.Name
	existing.FullName = repo.FullName
	existing.URL = repo.URL
	existing.DefaultBranch = repo.DefaultBranch
	existing.Private = repo.Private
	if err := r.db.Save(&existing).Error; err != nil {
		return err
	}
	*repo = existing
	return nil
}

func (r *GitHubRepositoryRepository) UpdateSelected(id, organizationID uuid.UUID, selected bool) error {
	return r.db.Model(&models.GitHubRepository{}).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Update("selected", selected).Error
}
