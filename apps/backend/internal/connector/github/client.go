package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.github.com"
	// requestTimeout bounds every single GitHub API call - an unreachable
	// or hanging GitHub must never block an OpsPilot request indefinitely.
	requestTimeout = 15 * time.Second
	// maxPerPage bounds every list call - GitHub's own maximum, and never
	// exceeded so no caller can accidentally request an unbounded page size.
	maxPerPage = 100
)

// Client is a small, dependency-free GitHub REST API v3 client scoped to
// exactly the read operations this sprint needs. BaseURL is overridable so
// tests can point it at an httptest.Server instead of the real GitHub API -
// no test in this codebase depends on live GitHub.
type Client struct {
	BaseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: requestTimeout}
	}
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

// AuthenticatedUser returns the identity a token resolves to - GitHub's
// cheapest possible "does this credential work" call, used both for
// TestConnection/HealthCheck and to show the connected account.
func (c *Client) AuthenticatedUser(ctx context.Context, token string) (*User, error) {
	var user User
	if err := c.get(ctx, token, "/user", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ListRepositories returns repositories the token's owner can access,
// newest-activity first. page/perPage are both bounded by the caller via
// GitHubService; perPage is clamped to GitHub's own maximum here as a
// second line of defense.
func (c *Client) ListRepositories(ctx context.Context, token string, page, perPage int) ([]Repository, error) {
	query := url.Values{
		"per_page":    {strconv.Itoa(clampPerPage(perPage))},
		"page":        {strconv.Itoa(maxInt(page, 1))},
		"sort":        {"updated"},
		"affiliation": {"owner,collaborator,organization_member"},
	}
	var repos []Repository
	if err := c.get(ctx, token, "/user/repos", query, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

// ListCommits returns recent commits for owner/repo, optionally scoped to a
// branch/ref. Bounded by perPage - never an unbounded history walk.
func (c *Client) ListCommits(ctx context.Context, token, owner, repo, branch string, perPage int) ([]Commit, error) {
	query := url.Values{"per_page": {strconv.Itoa(clampPerPage(perPage))}}
	if strings.TrimSpace(branch) != "" {
		query.Set("sha", branch)
	}
	var commits []Commit
	path := fmt.Sprintf("/repos/%s/%s/commits", url.PathEscape(owner), url.PathEscape(repo))
	if err := c.get(ctx, token, path, query, &commits); err != nil {
		return nil, err
	}
	return commits, nil
}

// GetCommit fetches exactly one commit by SHA - the deployment-correlation
// lookup, given a caller-supplied (never inferred) commit SHA.
func (c *Client) GetCommit(ctx context.Context, token, owner, repo, sha string) (*Commit, error) {
	var commit Commit
	path := fmt.Sprintf("/repos/%s/%s/commits/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(sha))
	if err := c.get(ctx, token, path, nil, &commit); err != nil {
		return nil, err
	}
	return &commit, nil
}

// ListPullRequests returns recent pull requests (any state) for owner/repo,
// newest-updated first, bounded by perPage.
func (c *Client) ListPullRequests(ctx context.Context, token, owner, repo string, perPage int) ([]PullRequest, error) {
	query := url.Values{
		"per_page":  {strconv.Itoa(clampPerPage(perPage))},
		"state":     {"all"},
		"sort":      {"updated"},
		"direction": {"desc"},
	}
	var pulls []PullRequest
	path := fmt.Sprintf("/repos/%s/%s/pulls", url.PathEscape(owner), url.PathEscape(repo))
	if err := c.get(ctx, token, path, query, &pulls); err != nil {
		return nil, err
	}
	return pulls, nil
}

// ListPullRequestsForCommit returns pull requests associated with a
// specific commit SHA - the second half of deployment correlation, given a
// commit that has already been positively resolved via GetCommit.
func (c *Client) ListPullRequestsForCommit(ctx context.Context, token, owner, repo, sha string) ([]PullRequest, error) {
	var pulls []PullRequest
	path := fmt.Sprintf("/repos/%s/%s/commits/%s/pulls", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(sha))
	if err := c.get(ctx, token, path, nil, &pulls); err != nil {
		return nil, err
	}
	return pulls, nil
}

func (c *Client) get(ctx context.Context, token, path string, query url.Values, out any) error {
	if strings.TrimSpace(token) == "" {
		return ErrTokenRequired
	}

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	target := c.BaseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("build github request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if requestCtx.Err() != nil {
			return ErrTimeout
		}
		return ErrUnavailable
	}
	defer resp.Body.Close()

	if err := mapStatus(resp); err != nil {
		return err
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}
	return nil
}

// mapStatus turns a GitHub HTTP response into one of this package's typed
// errors, distinguishing a genuine rate limit (403/429 with
// X-RateLimit-Remaining: 0) from a plain permissions-forbidden response -
// the two require very different user-facing guidance.
func mapStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusForbidden:
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return ErrRateLimited
		}
		return ErrForbidden
	default:
		if resp.StatusCode >= 500 {
			return ErrUnavailable
		}
		return fmt.Errorf("github returned unexpected status %d", resp.StatusCode)
	}
}

func clampPerPage(perPage int) int {
	if perPage <= 0 {
		return 30
	}
	if perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsNotFound reports whether err (possibly wrapped) is ErrNotFound - used by
// callers that treat "commit not found" as a distinct, expected outcome
// rather than a generic failure.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
