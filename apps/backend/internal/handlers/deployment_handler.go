package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type DeploymentHandler struct {
	service *services.DeploymentService
}

func NewDeploymentHandler(service *services.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{service: service}
}

func (h *DeploymentHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	userID := c.MustGet("userID").(uint)

	var req dto.CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	deployment, err := h.service.CreateDeployment(c.Request.Context(), organizationID, userID, req)
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.Created(c, "Deployment created successfully", deployment)
}

func (h *DeploymentHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "created_at", "updated_at", "started_at", "completed_at", "status")
	if !ok {
		return
	}

	applicationID, appOK := parseOptionalUUID(c, "applicationId")
	if !appOK {
		return
	}
	projectID, projectOK := parseOptionalUUID(c, "projectId")
	if !projectOK {
		return
	}
	if applicationID != uuid.Nil && projectID != uuid.Nil {
		response.BadRequest(c, "applicationId and projectId cannot both be set")
		return
	}

	var (
		result *dto.DeploymentListResponse
		err    error
	)
	if applicationID != uuid.Nil {
		result, err = h.service.ListApplicationDeployments(applicationID, organizationID, req)
	} else if projectID != uuid.Nil {
		result, err = h.service.ListProjectDeployments(projectID, organizationID, req)
	} else {
		result, err = h.service.ListDeployments(organizationID, req)
	}
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployments fetched successfully", result)
}

func (h *DeploymentHandler) GetByID(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid deployment id")
		return
	}

	deployment, err := h.service.GetDeployment(id, organizationID)
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployment fetched successfully", deployment)
}

func (h *DeploymentHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	userID := c.MustGet("userID").(uint)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid deployment id")
		return
	}

	var req dto.UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	deployment, err := h.service.UpdateDeployment(c.Request.Context(), id, organizationID, userID, req)
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployment updated successfully", deployment)
}

func (h *DeploymentHandler) Cancel(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	userID := c.MustGet("userID").(uint)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid deployment id")
		return
	}

	deployment, err := h.service.CancelDeployment(c.Request.Context(), id, organizationID, userID)
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployment cancelled successfully", deployment)
}

func (h *DeploymentHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid deployment id")
		return
	}

	if err := h.service.DeleteDeployment(c.Request.Context(), id, organizationID); err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployment deleted successfully", nil)
}

func (h *DeploymentHandler) Rollback(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	userID := c.MustGet("userID").(uint)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid deployment id")
		return
	}

	var req dto.RollbackDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.RollbackDeployment(c.Request.Context(), id, organizationID, userID, req.Revision)
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployment rolled back successfully", result)
}

func (h *DeploymentHandler) ListByApplication(c *gin.Context) {
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

	req, ok := parsePagination(c, "created_at", "updated_at", "started_at", "completed_at", "status")
	if !ok {
		return
	}

	result, err := h.service.ListApplicationDeployments(applicationID, organizationID, req)
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployments fetched successfully", result)
}

func (h *DeploymentHandler) ListByProject(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	req, ok := parsePagination(c, "created_at", "updated_at", "started_at", "completed_at", "status")
	if !ok {
		return
	}

	result, err := h.service.ListProjectDeployments(projectID, organizationID, req)
	if err != nil {
		handleDeploymentServiceError(c, err)
		return
	}

	response.OK(c, "Deployments fetched successfully", result)
}

func handleDeploymentServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrDeploymentForbidden, apperrors.ErrProjectForbidden, apperrors.ErrApplicationForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrDeploymentNotFound, apperrors.ErrApplicationNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusBadRequest, err.Error())
	case apperrors.ErrDeploymentInvalidKubeconfig, apperrors.ErrDeploymentNamespaceNotFound, apperrors.ErrDeploymentImageInvalid:
		response.Error(c, http.StatusBadRequest, err.Error())
	case apperrors.ErrDeploymentClusterUnreachable, apperrors.ErrDeploymentExecutionPermissionDenied, apperrors.ErrDeploymentExecutionTimeout, apperrors.ErrDeploymentExecutionFailed:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrInvalidApplication, apperrors.ErrInvalidProject, apperrors.ErrInvalidDeploymentImage, apperrors.ErrInvalidDeploymentReplica, apperrors.ErrInvalidDeploymentStrategy, apperrors.ErrInvalidDeploymentEnvironment, apperrors.ErrInvalidDeploymentNamespace, apperrors.ErrInvalidDeploymentStatus, apperrors.ErrInvalidDeploymentRevision, apperrors.ErrDeploymentRollbackLatest:
		response.BadRequest(c, err.Error())
	case apperrors.ErrDeploymentHistoryNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
