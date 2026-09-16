package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

// GitHubHandler is the HTTP layer for Sprint 28's GitHub connector. It
// never talks to GitHub itself and never sees a decrypted credential -
// every method here does nothing but parse/validate the request, delegate
// to GitHubService, and render the response through the existing envelope.
// Read endpoints require organization membership; every mutation
// (repository sync/select, application mapping, OAuth start) additionally
// requires organization:manage, reusing Sprint 27's permission rather than
// introducing a GitHub-specific one.
type GitHubHandler struct {
	service *services.GitHubService
	cfg     *config.Config
}

func NewGitHubHandler(service *services.GitHubService, cfg *config.Config) *GitHubHandler {
	return &GitHubHandler{service: service, cfg: cfg}
}

// ─── OAuth ──────────────────────────────────────────────────────────────────

func (h *GitHubHandler) OAuthStart(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	result, err := h.service.StartOAuth(organizationID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "GitHub OAuth authorization URL generated", result)
}

// OAuthCallback is intentionally NOT behind AuthMiddleware (GitHub's
// redirect carries no Authorization header) - the signed state token is
// the entire authorization boundary here, verified inside GitHubService.
func (h *GitHubHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	frontendBase := "http://localhost:3000"
	if h.cfg != nil && h.cfg.FrontendBaseURL != "" {
		frontendBase = h.cfg.FrontendBaseURL
	}

	if code == "" || state == "" {
		c.Redirect(http.StatusFound, frontendBase+"/organization?github=error&reason=missing_parameters")
		return
	}

	if _, err := h.service.CompleteOAuth(c.Request.Context(), code, state); err != nil {
		c.Redirect(http.StatusFound, frontendBase+"/organization?github=error")
		return
	}

	c.Redirect(http.StatusFound, frontendBase+"/organization?github=connected")
}

// ─── Repository discovery ───────────────────────────────────────────────────

func (h *GitHubHandler) GetIdentity(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	result, err := h.service.GetIdentity(c.Request.Context(), organizationID, integrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "GitHub identity fetched successfully", result)
}

func (h *GitHubHandler) ListRepositories(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	page, limit, ok := parseGitHubPagination(c, models.MaxLimit)
	if !ok {
		return
	}

	result, err := h.service.ListStoredRepositories(organizationID, integrationID, page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "GitHub repositories fetched successfully", result)
}

func (h *GitHubHandler) SyncRepositories(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	userID := c.MustGet("userID").(uint)
	result, err := h.service.DiscoverRepositories(c.Request.Context(), userID, organizationID, integrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "GitHub repositories synced successfully", result)
}

func (h *GitHubHandler) SelectRepository(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	repositoryID, err := uuid.Parse(c.Param("repositoryId"))
	if err != nil {
		response.BadRequest(c, "invalid repository id")
		return
	}

	var req dto.SelectGitHubRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	result, err := h.service.SelectRepository(userID, organizationID, repositoryID, req.Selected)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Repository updated successfully", result)
}

// ─── Commits / pull requests ────────────────────────────────────────────────

func (h *GitHubHandler) ListCommits(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	repositoryID, err := uuid.Parse(c.Param("repositoryId"))
	if err != nil {
		response.BadRequest(c, "invalid repository id")
		return
	}

	page, limit, ok := parseGitHubPagination(c, 50)
	if !ok {
		return
	}
	branch := c.Query("branch")

	result, err := h.service.ListCommits(c.Request.Context(), organizationID, repositoryID, branch, page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Commits fetched successfully", result)
}

func (h *GitHubHandler) ListPullRequests(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	repositoryID, err := uuid.Parse(c.Param("repositoryId"))
	if err != nil {
		response.BadRequest(c, "invalid repository id")
		return
	}

	page, limit, ok := parseGitHubPagination(c, 50)
	if !ok {
		return
	}

	state := strings.ToLower(strings.TrimSpace(c.DefaultQuery("state", "all")))
	if state != "open" && state != "closed" && state != "all" {
		response.BadRequest(c, "state must be one of: open, closed, all")
		return
	}

	result, err := h.service.ListPullRequests(c.Request.Context(), organizationID, repositoryID, state, page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Pull requests fetched successfully", result)
}

// ─── Application <-> repository mapping ─────────────────────────────────────

func (h *GitHubHandler) GetApplicationMapping(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	result, err := h.service.GetApplicationMapping(organizationID, applicationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Repository mapping fetched successfully", result)
}

func (h *GitHubHandler) MapApplicationRepository(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	var req dto.MapApplicationRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	repositoryID, err := uuid.Parse(req.RepositoryID)
	if err != nil {
		response.BadRequest(c, "invalid repository id")
		return
	}

	userID := c.MustGet("userID").(uint)
	result, err := h.service.MapApplicationRepository(userID, organizationID, applicationID, repositoryID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Application mapped to repository successfully", result)
}

func (h *GitHubHandler) UnmapApplicationRepository(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	userID := c.MustGet("userID").(uint)
	if err := h.service.UnmapApplicationRepository(userID, organizationID, applicationID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Application repository mapping removed successfully", nil)
}

// ─── Deployment correlation ──────────────────────────────────────────────────

func (h *GitHubHandler) GetDeploymentCorrelation(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	deploymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid deployment id")
		return
	}

	result, err := h.service.GetDeploymentCorrelation(c.Request.Context(), organizationID, deploymentID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Deployment GitHub correlation fetched successfully", result)
}

// ─── helpers ────────────────────────────────────────────────────────────────

func parseGitHubPagination(c *gin.Context, maxLimit int) (int, int, bool) {
	page, limit := models.DefaultPage, models.DefaultLimit
	if raw := c.Query("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < models.DefaultPage {
			response.BadRequest(c, "page must be at least 1")
			return 0, 0, false
		}
		page = parsed
	}
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxLimit {
			response.BadRequest(c, "limit must be between 1 and "+strconv.Itoa(maxLimit))
			return 0, 0, false
		}
		limit = parsed
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit, true
}

func (h *GitHubHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrIntegrationNotFound, apperrors.ErrGitHubRepositoryNotFound, apperrors.ErrApplicationNotFound,
		apperrors.ErrDeploymentNotFound, apperrors.ErrApplicationRepositoryNotMapped:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrIntegrationTypeMismatch, apperrors.ErrGitHubOAuthStateInvalid:
		response.BadRequest(c, err.Error())
	case apperrors.ErrGitHubOAuthNotConfigured:
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	case apperrors.ErrGitHubUnauthorized:
		response.Error(c, http.StatusUnauthorized, err.Error())
	case apperrors.ErrGitHubForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrGitHubRateLimited:
		response.Error(c, http.StatusTooManyRequests, err.Error())
	case apperrors.ErrGitHubTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	case apperrors.ErrGitHubUnavailable, apperrors.ErrGitHubNotConnected:
		response.Error(c, http.StatusBadGateway, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
