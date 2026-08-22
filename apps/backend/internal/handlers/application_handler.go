package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type ApplicationHandler struct {
	service *services.ApplicationService
}

func NewApplicationHandler(service *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{service: service}
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionApplicationManage) {
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

	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	application, err := h.service.CreateApplication(c.Request.Context(), projectID, organizationID, req)
	if err != nil {
		handleApplicationServiceError(c, err)
		return
	}

	response.Created(c, "Application created successfully", application)
}

func (h *ApplicationHandler) ListByProject(c *gin.Context) {
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

	req, ok := parsePagination(c, "name", "slug", "runtime", "status", "created_at", "updated_at")
	if !ok {
		return
	}

	result, err := h.service.ListApplicationsByProject(projectID, organizationID, req)
	if err != nil {
		handleApplicationServiceError(c, err)
		return
	}

	response.OK(c, "Applications fetched successfully", result)
}

func (h *ApplicationHandler) GetByID(c *gin.Context) {
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

	application, err := h.service.GetApplication(applicationID, organizationID)
	if err != nil {
		handleApplicationServiceError(c, err)
		return
	}

	response.OK(c, "Application fetched successfully", application)
}

func (h *ApplicationHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionApplicationManage) {
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

	var req dto.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	application, err := h.service.UpdateApplication(c.Request.Context(), applicationID, organizationID, req)
	if err != nil {
		handleApplicationServiceError(c, err)
		return
	}

	response.OK(c, "Application updated successfully", application)
}

func (h *ApplicationHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionApplicationManage) {
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

	if err := h.service.DeleteApplication(c.Request.Context(), applicationID, organizationID); err != nil {
		handleApplicationServiceError(c, err)
		return
	}

	response.OK(c, "Application deleted successfully", nil)
}

func handleApplicationServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationForbidden, apperrors.ErrProjectForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrApplicationNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationAlreadyExists:
		response.Conflict(c, err.Error())
	case apperrors.ErrInvalidApplicationName, apperrors.ErrInvalidApplicationSlug, apperrors.ErrInvalidApplicationRuntime, apperrors.ErrInvalidApplicationStatus, apperrors.ErrInvalidApplicationPort, apperrors.ErrInvalidApplicationEnvironment:
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
