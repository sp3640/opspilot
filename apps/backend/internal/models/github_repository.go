package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GitHubRepository caches what OpsPilot knows about one GitHub repository
// discovered through a connected GitHub Integration - the "GitHub Data
// Model" from Sprint 28. It exists independently of any Application
// mapping (see ApplicationGitHubRepository): an organization can discover
// and Select a repository before deciding which application, if any, it
// corresponds to.
//
// GitHubID (GitHub's own numeric repository ID) is the natural external
// key; the uniqueIndex on (IntegrationID, GitHubID) is what prevents
// discovery from ever duplicating a repository row on repeated syncs.
type GitHubRepository struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_github_repositories_organization_id"`
	IntegrationID  uuid.UUID `json:"integration_id" gorm:"type:uuid;not null;index:idx_github_repositories_integration_id;uniqueIndex:idx_github_repositories_integration_github_id,priority:1"`
	// GitHubID's column is pinned explicitly to github_id - GORM's default
	// snake_case namer does not know "GitHub" is one word and would
	// otherwise produce git_hub_id, silently diverging from every repo/
	// service query in this package that spells the column github_id.
	GitHubID      int64  `json:"github_id" gorm:"column:github_id;not null;uniqueIndex:idx_github_repositories_integration_github_id,priority:2"`
	Owner         string `json:"owner" gorm:"size:255;not null"`
	Name          string `json:"name" gorm:"size:255;not null"`
	FullName      string `json:"full_name" gorm:"size:511;not null"`
	URL           string `json:"url" gorm:"size:500;not null"`
	DefaultBranch string `json:"default_branch" gorm:"size:255;not null;default:'main'"`
	Private       bool   `json:"private" gorm:"not null;default:false"`
	// Selected marks that the organization has explicitly chosen to track
	// this repository - distinct from (and a prerequisite to, in the UI)
	// mapping it to a specific Application.
	Selected  bool      `json:"selected" gorm:"not null;default:false;index:idx_github_repositories_selected"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *GitHubRepository) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
