// Package github implements OpsPilot's first real integration connector
// (Sprint 28), satisfying the generic internal/connector.Connector
// abstraction from Sprint 27 without changing it. It is read-only by
// design: repository discovery, commit history, and pull request history
// only - no GitHub write operation (create/merge/delete/branch/comment) is
// implemented anywhere in this package.
package github

import "time"

// User is the authenticated GitHub identity - the "who did we connect as"
// fact shown to the organization after connecting.
type User struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// Repository is GitHub's own repository shape, trimmed to the fields this
// package actually uses. Never returned directly to an HTTP caller -
// GitHubService maps this into its own DTO.
type Repository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Owner         struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// Commit mirrors GitHub's nested commit shape: commit.author (the raw Git
// author identity, always present) is distinct from author (the linked
// GitHub account, which may be null for commits authored outside GitHub).
type Commit struct {
	SHA     string `json:"sha"`
	HTMLURL string `json:"html_url"`
	Commit  struct {
		Message string `json:"message"`
		Author  struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
}

// PullRequest mirrors the fields this package uses from GitHub's pull
// request shape.
type PullRequest struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	MergedAt  *time.Time `json:"merged_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		SHA string `json:"sha"`
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}
