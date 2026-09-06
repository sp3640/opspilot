package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationGitHubRepository is the mapping from one OpsPilot Application
// to the single GitHub repository it corresponds to (Sprint 28, "Application
// <-> GitHub Repository"). One Application maps to at most one repository
// (uniqueIndex on ApplicationID) - the reverse is intentionally not unique,
// since a monorepo can back multiple Applications with the same repository.
//
// This is a small, dedicated mapping table rather than repurposing
// Application.RepositoryURL: that field is free-text with no FK to any
// specific integration/repository record, so it cannot safely stand in for
// a structured, organization-isolated relationship.
type ApplicationGitHubRepository struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_app_github_repo_organization_id"`
	ApplicationID  uuid.UUID `json:"application_id" gorm:"type:uuid;not null;uniqueIndex:idx_app_github_repo_application_id"`
	// Column pinned explicitly for the same reason as GitHubRepository.GitHubID.
	GitHubRepositoryID uuid.UUID `json:"github_repository_id" gorm:"column:github_repository_id;type:uuid;not null;index:idx_app_github_repo_repository_id"`
	CreatedBy          uint      `json:"created_by" gorm:"not null"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (m *ApplicationGitHubRepository) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
