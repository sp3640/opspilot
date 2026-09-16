package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/connector"
	githubconnector "github.com/sp3640/opspilot/backend/internal/connector/github"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

const (
	githubTokenCredentialKey = "access_token"
	// oauthStateTTL bounds how long a signed OAuth state token (and
	// therefore the browser redirect round trip to GitHub and back) stays
	// valid, limiting the replay window if a state value were ever leaked.
	oauthStateTTL      = 10 * time.Minute
	defaultCommitLimit = 20
	maxCommitLimit     = 50
	defaultPRLimit     = 20
	maxPRLimit         = 50
)

// GitHubService is the GitHub-specific business logic layer described by
// Sprint 28's architecture (HTTP Handler -> GitHub Service -> GitHub
// Connector -> GitHub API). It never talks to GitHub's HTTP API directly -
// that is entirely inside internal/connector/github.Client - and it never
// duplicates IntegrationService's encryption/status/audit bookkeeping: it
// composes IntegrationService for every one of those concerns.
type GitHubService struct {
	integrationService *IntegrationService
	client             *githubconnector.Client
	oauth              *githubconnector.OAuthConfig
	repoRepo           *repository.GitHubRepositoryRepository
	mappingRepo        *repository.ApplicationGitHubRepositoryRepository
	applicationRepo    *repository.ApplicationRepository
	deploymentRepo     *repository.DeploymentRepository
	auditService       *AuditService
	// stateSigningKey signs/verifies OAuth state tokens via HMAC-SHA256.
	// Reuses the existing JWT signing secret rather than introducing a new
	// key - the state token is not a credential itself (it carries no
	// secret, only organizationID/userID/expiry), it just needs to be
	// tamper-evident.
	stateSigningKey []byte
}

func NewGitHubService(
	integrationService *IntegrationService,
	client *githubconnector.Client,
	oauth *githubconnector.OAuthConfig,
	repoRepo *repository.GitHubRepositoryRepository,
	mappingRepo *repository.ApplicationGitHubRepositoryRepository,
	applicationRepo *repository.ApplicationRepository,
	deploymentRepo *repository.DeploymentRepository,
	stateSigningKey string,
) *GitHubService {
	return &GitHubService{
		integrationService: integrationService,
		client:             client,
		oauth:              oauth,
		repoRepo:           repoRepo,
		mappingRepo:        mappingRepo,
		applicationRepo:    applicationRepo,
		deploymentRepo:     deploymentRepo,
		stateSigningKey:    []byte(stateSigningKey),
	}
}

func (s *GitHubService) WithAuditService(auditService *AuditService) *GitHubService {
	s.auditService = auditService
	return s
}

// WithClient overrides the GitHub API client - used only by tests to point
// at an httptest.Server instead of the real GitHub API.
func (s *GitHubService) WithClient(client *githubconnector.Client) *GitHubService {
	s.client = client
	return s
}

// ─── OAuth ──────────────────────────────────────────────────────────────────

type oauthStateClaims struct {
	OrganizationID uuid.UUID `json:"org"`
	UserID         uint      `json:"user"`
	ExpiresAt      int64     `json:"exp"`
}

// StartOAuth returns the URL to redirect the browser to, embedding a
// signed, short-lived state token identifying which organization/user
// initiated the flow - GitHubOAuthCallback verifies this on the way back so
// the public callback endpoint can never be used to connect GitHub to an
// organization the caller didn't already have organization:manage on.
func (s *GitHubService) StartOAuth(organizationID uuid.UUID, userID uint) (*dto.GitHubOAuthStartResponse, error) {
	if !s.oauth.IsConfigured() {
		return nil, apperrors.ErrGitHubOAuthNotConfigured
	}

	state, err := s.signState(organizationID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.GitHubOAuthStartResponse{AuthorizeURL: s.oauth.BuildAuthorizeURL(state)}, nil
}

// CompleteOAuth verifies state, exchanges code for an access token, looks up
// the resulting GitHub identity, and creates or updates the organization's
// GitHub Integration. Returns the organizationID (for the caller to build a
// redirect) and never returns the access token itself.
func (s *GitHubService) CompleteOAuth(ctx context.Context, code, state string) (uuid.UUID, error) {
	if !s.oauth.IsConfigured() {
		return uuid.Nil, apperrors.ErrGitHubOAuthNotConfigured
	}

	organizationID, userID, err := s.verifyState(state)
	if err != nil {
		return uuid.Nil, err
	}

	token, err := s.oauth.ExchangeCode(ctx, code)
	if err != nil {
		return organizationID, mapGitHubError(err)
	}

	user, err := s.client.AuthenticatedUser(ctx, token)
	if err != nil {
		return organizationID, mapGitHubError(err)
	}

	name := fmt.Sprintf("GitHub (%s)", user.Login)
	if _, err := s.integrationService.UpsertConnectedIntegration(userID, organizationID, constants.IntegrationTypeGitHub, name, map[string]string{
		githubTokenCredentialKey: token,
	}); err != nil {
		return organizationID, err
	}

	return organizationID, nil
}

func (s *GitHubService) signState(organizationID uuid.UUID, userID uint) (string, error) {
	claims := oauthStateClaims{OrganizationID: organizationID, UserID: userID, ExpiresAt: time.Now().Add(oauthStateTTL).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	return encodedPayload + "." + s.sign(encodedPayload), nil
}

func (s *GitHubService) verifyState(state string) (uuid.UUID, uint, error) {
	dot := -1
	for i := len(state) - 1; i >= 0; i-- {
		if state[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return uuid.Nil, 0, apperrors.ErrGitHubOAuthStateInvalid
	}
	encodedPayload, signature := state[:dot], state[dot+1:]

	if !hmac.Equal([]byte(signature), []byte(s.sign(encodedPayload))) {
		return uuid.Nil, 0, apperrors.ErrGitHubOAuthStateInvalid
	}

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return uuid.Nil, 0, apperrors.ErrGitHubOAuthStateInvalid
	}
	var claims oauthStateClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil, 0, apperrors.ErrGitHubOAuthStateInvalid
	}
	if time.Now().Unix() > claims.ExpiresAt {
		return uuid.Nil, 0, apperrors.ErrGitHubOAuthStateInvalid
	}

	return claims.OrganizationID, claims.UserID, nil
}

func (s *GitHubService) sign(payload string) string {
	mac := hmac.New(sha256.New, s.stateSigningKey)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// GetIdentity returns the public identity associated with a connected GitHub
// integration. Credential decryption remains inside IntegrationService; the
// token is used only for this one API call and is never returned or audited.
func (s *GitHubService) GetIdentity(ctx context.Context, organizationID, integrationID uuid.UUID) (*dto.GitHubIdentityResponse, error) {
	_, cfg, err := s.getOwnedGitHubIntegration(organizationID, integrationID)
	if err != nil {
		return nil, err
	}

	user, err := s.client.AuthenticatedUser(ctx, cfg.Credentials[githubTokenCredentialKey])
	if err != nil {
		return nil, mapGitHubError(err)
	}

	return &dto.GitHubIdentityResponse{
		ID: user.ID, Login: user.Login, AvatarURL: user.AvatarURL, ProfileURL: user.HTMLURL,
	}, nil
}

// ─── Repository discovery ───────────────────────────────────────────────────

// DiscoverRepositories calls GitHub (bounded to 2 pages / 200 repositories)
// and upserts the results into GitHubRepository, so subsequent reads (and
// the frontend's default view) never need to re-hit GitHub. Also records
// the outcome as a connection check, since successfully listing
// repositories is as strong a "this credential works" signal as
// TestConnection's identity lookup.
func (s *GitHubService) DiscoverRepositories(ctx context.Context, userID uint, organizationID, integrationID uuid.UUID) (*dto.GitHubRepositoryListResponse, error) {
	integration, cfg, err := s.getOwnedGitHubIntegration(organizationID, integrationID)
	if err != nil {
		return nil, err
	}

	token := cfg.Credentials[githubTokenCredentialKey]
	var all []githubconnector.Repository
	const pages = 2
	for page := 1; page <= pages; page++ {
		repos, err := s.client.ListRepositories(ctx, token, page, 100)
		if err != nil {
			_, _ = s.integrationService.RecordConnectionOutcome(userID, organizationID, integration, "github_repository_sync", false, safeGitHubMessage(err))
			return nil, mapGitHubError(err)
		}
		all = append(all, repos...)
		if len(repos) < 100 {
			break
		}
	}

	for i := range all {
		repo := &models.GitHubRepository{
			OrganizationID: organizationID,
			IntegrationID:  integrationID,
			GitHubID:       all[i].ID,
			Owner:          all[i].Owner.Login,
			Name:           all[i].Name,
			FullName:       all[i].FullName,
			URL:            all[i].HTMLURL,
			DefaultBranch:  all[i].DefaultBranch,
			Private:        all[i].Private,
		}
		if err := s.repoRepo.Upsert(repo); err != nil {
			return nil, err
		}
	}

	if _, err := s.integrationService.RecordConnectionOutcome(userID, organizationID, integration, "github_repository_sync", true, fmt.Sprintf("Discovered %d repositories", len(all))); err != nil {
		return nil, err
	}

	return s.ListStoredRepositories(organizationID, integrationID, models.DefaultPage, models.DefaultLimit)
}

// ListStoredRepositories is a pure database read - no GitHub call - so the
// frontend can render this cheaply and often without ever hitting GitHub's
// rate limit; DiscoverRepositories is the explicit, user-initiated action
// that actually calls GitHub.
func (s *GitHubService) ListStoredRepositories(organizationID, integrationID uuid.UUID, page, limit int) (*dto.GitHubRepositoryListResponse, error) {
	repos, total, err := s.repoRepo.ListByIntegrationPaginated(integrationID, organizationID, page, limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.GitHubRepositoryResponse, 0, len(repos))
	for _, repo := range repos {
		items = append(items, mapGitHubRepository(repo))
	}
	return &dto.GitHubRepositoryListResponse{Items: items, Page: page, Limit: limit, Total: total, TotalPages: totalPages(total, limit)}, nil
}

func (s *GitHubService) SelectRepository(userID uint, organizationID, repositoryID uuid.UUID, selected bool) (*dto.GitHubRepositoryResponse, error) {
	repo, err := s.getOwnedRepository(repositoryID, organizationID)
	if err != nil {
		return nil, err
	}

	if err := s.repoRepo.UpdateSelected(repo.ID, organizationID, selected); err != nil {
		return nil, err
	}
	previousSelected := repo.Selected
	repo.Selected = selected

	if s.auditService != nil && previousSelected != selected {
		action := "selected"
		if !selected {
			action = "unselected"
		}
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID: userID, OrganizationID: organizationID, EntityType: "github_repository", EntityID: repo.ID.String(),
			Action: models.AuditActionUpdate, FieldName: "selected", OldValue: fmt.Sprintf("%t", previousSelected), NewValue: fmt.Sprintf("%t", selected),
			AfterState: marshalAuditState(map[string]any{"repository": repo.FullName, "action": action}),
		})
	}

	response := mapGitHubRepository(*repo)
	return &response, nil
}

// ─── Application <-> repository mapping ─────────────────────────────────────

func (s *GitHubService) GetApplicationMapping(organizationID, applicationID uuid.UUID) (*dto.ApplicationRepositoryMappingResponse, error) {
	mapping, err := s.mappingRepo.GetByApplicationID(applicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrApplicationRepositoryNotMapped
		}
		return nil, err
	}

	repo, err := s.repoRepo.GetByID(mapping.GitHubRepositoryID, organizationID)
	if err != nil {
		return nil, err
	}

	return mapApplicationMapping(*mapping, *repo), nil
}

func (s *GitHubService) MapApplicationRepository(userID uint, organizationID, applicationID, repositoryID uuid.UUID) (*dto.ApplicationRepositoryMappingResponse, error) {
	if _, err := s.applicationRepo.GetApplication(applicationID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrApplicationNotFound
		}
		return nil, err
	}

	repo, err := s.getOwnedRepository(repositoryID, organizationID)
	if err != nil {
		return nil, err
	}

	mapping := &models.ApplicationGitHubRepository{
		OrganizationID:     organizationID,
		ApplicationID:      applicationID,
		GitHubRepositoryID: repo.ID,
		CreatedBy:          userID,
	}
	if err := s.mappingRepo.Upsert(mapping); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID: userID, OrganizationID: organizationID, EntityType: "application_github_repository", EntityID: applicationID.String(),
			Action: models.AuditActionUpdate, FieldName: "repository",
			AfterState: marshalAuditState(map[string]any{"applicationId": applicationID.String(), "repository": repo.FullName}),
		})
	}

	return mapApplicationMapping(*mapping, *repo), nil
}

func (s *GitHubService) UnmapApplicationRepository(userID uint, organizationID, applicationID uuid.UUID) error {
	if _, err := s.applicationRepo.GetApplication(applicationID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrApplicationNotFound
		}
		return err
	}

	if err := s.mappingRepo.DeleteByApplicationID(applicationID, organizationID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID: userID, OrganizationID: organizationID, EntityType: "application_github_repository", EntityID: applicationID.String(),
			Action: models.AuditActionDelete,
		})
	}

	return nil
}

// ─── Commits / pull requests ────────────────────────────────────────────────

func (s *GitHubService) ListCommits(ctx context.Context, organizationID, repositoryID uuid.UUID, branch string, page, limit int) (*dto.GitHubCommitListResponse, error) {
	repo, token, err := s.resolveRepositoryToken(organizationID, repositoryID)
	if err != nil {
		return nil, err
	}
	if branch == "" {
		branch = repo.DefaultBranch
	}

	limit = boundLimit(limit, defaultCommitLimit, maxCommitLimit)
	commits, err := s.client.ListCommits(ctx, token, repo.Owner, repo.Name, branch, page, limit)
	if err != nil {
		return nil, mapGitHubError(err)
	}

	items := make([]dto.GitHubCommitResponse, 0, len(commits))
	for _, commit := range commits {
		items = append(items, mapGitHubCommit(commit))
	}
	return &dto.GitHubCommitListResponse{Items: items, Page: page, Limit: limit, HasMore: len(items) == limit}, nil
}

func (s *GitHubService) ListPullRequests(ctx context.Context, organizationID, repositoryID uuid.UUID, state string, page, limit int) (*dto.GitHubPullRequestListResponse, error) {
	repo, token, err := s.resolveRepositoryToken(organizationID, repositoryID)
	if err != nil {
		return nil, err
	}

	limit = boundLimit(limit, defaultPRLimit, maxPRLimit)
	pulls, err := s.client.ListPullRequests(ctx, token, repo.Owner, repo.Name, state, page, limit)
	if err != nil {
		return nil, mapGitHubError(err)
	}

	items := make([]dto.GitHubPullRequestResponse, 0, len(pulls))
	for _, pr := range pulls {
		items = append(items, mapGitHubPullRequest(pr))
	}
	return &dto.GitHubPullRequestListResponse{Items: items, Page: page, Limit: limit, State: state, HasMore: len(items) == limit}, nil
}

// ─── Deployment correlation ──────────────────────────────────────────────────

// GetDeploymentCorrelation implements exactly the "smallest read-only
// correlation path" Sprint 28 calls for: Deployment.CommitSHA (a real,
// caller-supplied Git SHA - never inferred from timestamps) resolved
// against the GitHub repository mapped to the deployment's application.
// Every case where that chain is missing a link reports Available=false
// with a specific Reason rather than guessing.
func (s *GitHubService) GetDeploymentCorrelation(ctx context.Context, organizationID, deploymentID uuid.UUID) (*dto.DeploymentGitHubCorrelationResponse, error) {
	deployment, err := s.deploymentRepo.GetByID(deploymentID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentNotFound
		}
		return nil, err
	}

	response := &dto.DeploymentGitHubCorrelationResponse{DeploymentID: deployment.ID.String(), CommitSHA: deployment.CommitSHA}

	if deployment.CommitSHA == nil || *deployment.CommitSHA == "" {
		response.Reason = "This deployment has no recorded commit SHA - correlation requires a caller-supplied commit_sha at deployment time."
		return response, nil
	}

	mapping, err := s.mappingRepo.GetByApplicationID(deployment.ApplicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Reason = "No GitHub repository is mapped to this application yet."
			return response, nil
		}
		return nil, err
	}

	repo, token, err := s.resolveRepositoryToken(organizationID, mapping.GitHubRepositoryID)
	if err != nil {
		return nil, err
	}

	commit, err := s.client.GetCommit(ctx, token, repo.Owner, repo.Name, *deployment.CommitSHA)
	if err != nil {
		if githubconnector.IsNotFound(err) {
			response.Reason = "Commit not found in the mapped repository - it may have been force-pushed away, or the mapping may be outdated."
			return response, nil
		}
		return nil, mapGitHubError(err)
	}
	commitResponse := mapGitHubCommit(*commit)
	response.Commit = &commitResponse

	// Pull-request association is best-effort: the commit itself is the
	// more important, positively-resolved fact, so a PR-lookup failure
	// never hides it.
	if pulls, err := s.client.ListPullRequestsForCommit(ctx, token, repo.Owner, repo.Name, *deployment.CommitSHA); err == nil {
		items := make([]dto.GitHubPullRequestResponse, 0, len(pulls))
		for _, pr := range pulls {
			items = append(items, mapGitHubPullRequest(pr))
		}
		response.PullRequests = items
	}

	response.Available = true
	return response, nil
}

// ─── helpers ────────────────────────────────────────────────────────────────

func (s *GitHubService) getOwnedGitHubIntegration(organizationID, integrationID uuid.UUID) (*models.Integration, connector.Config, error) {
	integration, cfg, err := s.integrationService.ResolveConnectorConfig(organizationID, integrationID)
	if err != nil {
		return nil, connector.Config{}, err
	}
	if integration.Type != constants.IntegrationTypeGitHub {
		return nil, connector.Config{}, apperrors.ErrIntegrationTypeMismatch
	}
	return integration, cfg, nil
}

func (s *GitHubService) getOwnedRepository(id, organizationID uuid.UUID) (*models.GitHubRepository, error) {
	repo, err := s.repoRepo.GetByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrGitHubRepositoryNotFound
		}
		return nil, err
	}
	return repo, nil
}

// resolveRepositoryToken loads a repository (organization-scoped) and
// decrypts the credentials of the integration that discovered it, in one
// step - the shared path ListCommits/ListPullRequests/GetDeploymentCorrelation
// all use to go from "a repository row" to "a usable token".
func (s *GitHubService) resolveRepositoryToken(organizationID, repositoryID uuid.UUID) (*models.GitHubRepository, string, error) {
	repo, err := s.getOwnedRepository(repositoryID, organizationID)
	if err != nil {
		return nil, "", err
	}

	_, cfg, err := s.getOwnedGitHubIntegration(organizationID, repo.IntegrationID)
	if err != nil {
		return nil, "", err
	}

	return repo, cfg.Credentials[githubTokenCredentialKey], nil
}

func boundLimit(requested, fallback, max int) int {
	if requested <= 0 {
		return fallback
	}
	if requested > max {
		return max
	}
	return requested
}

func totalPages(total int64, limit int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

// mapGitHubError translates this package's connector-level errors into the
// apperrors the HTTP layer already knows how to render as safe status
// codes/messages - never the raw connector error text, which could
// otherwise carry request/URL detail callers should not see.
func mapGitHubError(err error) error {
	switch {
	case errors.Is(err, githubconnector.ErrUnauthorized):
		return apperrors.ErrGitHubUnauthorized
	case errors.Is(err, githubconnector.ErrForbidden):
		return apperrors.ErrGitHubForbidden
	case errors.Is(err, githubconnector.ErrRateLimited):
		return apperrors.ErrGitHubRateLimited
	case errors.Is(err, githubconnector.ErrNotFound):
		return apperrors.ErrGitHubRepositoryNotFound
	case errors.Is(err, githubconnector.ErrTimeout):
		return apperrors.ErrGitHubTimeout
	case errors.Is(err, githubconnector.ErrUnavailable):
		return apperrors.ErrGitHubUnavailable
	case errors.Is(err, githubconnector.ErrTokenRequired):
		return apperrors.ErrGitHubNotConnected
	default:
		return apperrors.ErrGitHubUnavailable
	}
}

// safeGitHubMessage renders an error as the kind of short, safe string that
// is allowed into Integration.LastError / an audit NewValue - it never
// includes a token (mapGitHubError's inputs never carry one to begin with;
// this just keeps the message itself description-only).
func safeGitHubMessage(err error) string {
	return mapGitHubError(err).Error()
}

func mapGitHubRepository(repo models.GitHubRepository) dto.GitHubRepositoryResponse {
	return dto.GitHubRepositoryResponse{
		ID:            repo.ID.String(),
		IntegrationID: repo.IntegrationID.String(),
		GitHubID:      repo.GitHubID,
		Owner:         repo.Owner,
		Name:          repo.Name,
		FullName:      repo.FullName,
		URL:           repo.URL,
		DefaultBranch: repo.DefaultBranch,
		Private:       repo.Private,
		Selected:      repo.Selected,
		CreatedAt:     repo.CreatedAt,
		UpdatedAt:     repo.UpdatedAt,
	}
}

func mapApplicationMapping(mapping models.ApplicationGitHubRepository, repo models.GitHubRepository) *dto.ApplicationRepositoryMappingResponse {
	return &dto.ApplicationRepositoryMappingResponse{
		ApplicationID: mapping.ApplicationID.String(),
		Repository:    mapGitHubRepository(repo),
		CreatedAt:     mapping.CreatedAt,
		UpdatedAt:     mapping.UpdatedAt,
	}
}

func mapGitHubCommit(commit githubconnector.Commit) dto.GitHubCommitResponse {
	response := dto.GitHubCommitResponse{
		SHA:         commit.SHA,
		Message:     commit.Commit.Message,
		AuthorName:  commit.Commit.Author.Name,
		AuthorEmail: commit.Commit.Author.Email,
		Timestamp:   commit.Commit.Author.Date,
		URL:         commit.HTMLURL,
	}
	if commit.Author != nil {
		response.AuthorLogin = commit.Author.Login
	}
	return response
}

func mapGitHubPullRequest(pr githubconnector.PullRequest) dto.GitHubPullRequestResponse {
	return dto.GitHubPullRequestResponse{
		Number:      pr.Number,
		Title:       pr.Title,
		State:       pr.State,
		AuthorLogin: pr.User.Login,
		CreatedAt:   pr.CreatedAt,
		UpdatedAt:   pr.UpdatedAt,
		MergedAt:    pr.MergedAt,
		URL:         pr.HTMLURL,
		HeadSHA:     pr.Head.SHA,
		BaseBranch:  pr.Base.Ref,
	}
}
