package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type DeploymentHistoryHandler struct {
	service *services.DeploymentHistoryService
}

func NewDeploymentHistoryHandler(service *services.DeploymentHistoryService) *DeploymentHistoryHandler {
	return &DeploymentHistoryHandler{service: service}
}

func (h *DeploymentHistoryHandler) ListByDeployment(c *gin.Context) {
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

	req, ok := parsePagination(c, "revision", "created_at", "status")
	if !ok {
		return
	}

	history, err := h.service.ListDeploymentHistory(deploymentID, organizationID, req)
	if err != nil {
		handleDeploymentHistoryServiceError(c, err)
		return
	}

	response.OK(c, "Deployment history fetched successfully", history)
}

func (h *DeploymentHistoryHandler) GetRevision(c *gin.Context) {
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

	revision, err := strconv.Atoi(c.Param("revision"))
	if err != nil {
		response.BadRequest(c, "invalid deployment revision")
		return
	}

	history, err := h.service.GetDeploymentHistoryRevision(deploymentID, organizationID, revision)
	if err != nil {
		handleDeploymentHistoryServiceError(c, err)
		return
	}

	response.OK(c, "Deployment history revision fetched successfully", history)
}

func handleDeploymentHistoryServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrDeploymentForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrDeploymentNotFound, apperrors.ErrDeploymentHistoryNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrInvalidDeploymentRevision:
		response.BadRequest(c, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
