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

type ProjectHandler struct {
	service *services.ProjectService
}

func NewProjectHandler(service *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		service: service,
	}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	project, err := h.service.Create(c.Request.Context(), req.Name, req.Description, userID, organizationID)
	if err != nil {
		switch err {
		case apperrors.ErrProjectAlreadyExists:
			response.Conflict(c, err.Error())
		case apperrors.ErrInvalidProjectName, apperrors.ErrInvalidProjectDescription:
			response.BadRequest(c, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.Created(c, "Project created successfully", project)
}

func (h *ProjectHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "name", "created_at", "updated_at")
	if !ok {
		return
	}

	result, err := h.service.ListMyProjects(userID, organizationID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Projects fetched successfully", result)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	_, hasUser := c.Get("userID")
	if !hasUser {
		response.Unauthorized(c, "missing user context")
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

	project, err := h.service.GetProjectByID(projectID, organizationID)
	if err != nil {
		switch err {
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Project fetched successfully", project)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project, err := h.service.UpdateProject(c.Request.Context(), projectID, userID, organizationID, req.Name, req.Description)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidProjectName, apperrors.ErrInvalidProjectDescription:
			response.BadRequest(c, err.Error())
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Project updated successfully", project)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	err = h.service.DeleteProject(c.Request.Context(), projectID, userID, organizationID)
	if err != nil {
		switch err {
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Project deleted successfully", nil)
}
