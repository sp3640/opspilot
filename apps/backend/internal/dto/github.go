package dto

import "time"

// ─── Repositories ──────────────────────────────────────────────────────────

// GitHubRepositoryResponse is the safe, stored-in-our-database shape for a
// discovered repository - never GitHub's raw JSON, and never anything that
// could contain a credential.
type GitHubRepositoryResponse struct {
	ID            string    `json:"id"`
	IntegrationID string    `json:"integrationId"`
	GitHubID      int64     `json:"githubId"`
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	FullName      string    `json:"fullName"`
	URL           string    `json:"url"`
	DefaultBranch string    `json:"defaultBranch"`
	Private       bool      `json:"private"`
	Selected      bool      `json:"selected"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type GitHubRepositoryListResponse struct {
	Items []GitHubRepositoryResponse `json:"items"`
	Total int                        `json:"total"`
}

type SelectGitHubRepositoryRequest struct {
	Selected bool `json:"selected"`
}

// ─── Application <-> repository mapping ────────────────────────────────────

type MapApplicationRepositoryRequest struct {
	RepositoryID string `json:"repository_id" binding:"required"`
}

type ApplicationRepositoryMappingResponse struct {
	ApplicationID string                   `json:"applicationId"`
	Repository    GitHubRepositoryResponse `json:"repository"`
	CreatedAt     time.Time                `json:"createdAt"`
	UpdatedAt     time.Time                `json:"updatedAt"`
}

// ─── Commits / pull requests ────────────────────────────────────────────────

type GitHubCommitResponse struct {
	SHA         string    `json:"sha"`
	Message     string    `json:"message"`
	AuthorName  string    `json:"authorName"`
	AuthorEmail string    `json:"authorEmail,omitempty"`
	AuthorLogin string    `json:"authorLogin,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	URL         string    `json:"url"`
}

type GitHubCommitListResponse struct {
	Items []GitHubCommitResponse `json:"items"`
}

type GitHubPullRequestResponse struct {
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	State       string     `json:"state"`
	AuthorLogin string     `json:"authorLogin"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	MergedAt    *time.Time `json:"mergedAt,omitempty"`
	URL         string     `json:"url"`
	HeadSHA     string     `json:"headSha"`
	BaseBranch  string     `json:"baseBranch"`
}

type GitHubPullRequestListResponse struct {
	Items []GitHubPullRequestResponse `json:"items"`
}

// ─── Deployment correlation ─────────────────────────────────────────────────

// DeploymentGitHubCorrelationResponse is the read-only "what does GitHub
// know about this deployment's commit" view. Available=false with a Reason
// covers every case where OpsPilot does not have enough trustworthy
// information to answer - it never guesses or infers from timestamps alone.
type DeploymentGitHubCorrelationResponse struct {
	DeploymentID string                      `json:"deploymentId"`
	CommitSHA    *string                     `json:"commitSha,omitempty"`
	Available    bool                        `json:"available"`
	Reason       string                      `json:"reason,omitempty"`
	Commit       *GitHubCommitResponse       `json:"commit,omitempty"`
	PullRequests []GitHubPullRequestResponse `json:"pullRequests,omitempty"`
}

// ─── OAuth ──────────────────────────────────────────────────────────────────

type GitHubOAuthStartResponse struct {
	AuthorizeURL string `json:"authorizeUrl"`
}
