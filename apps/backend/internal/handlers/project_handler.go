package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type ProjectHandler struct {
	service *services.ProjectService
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func NewProjectHandler(service *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		service: service,
	}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)

	err := h.service.Create(
		c.Request.Context(),
		req.Name,
		req.Description,
		userID,
	)

	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Created(c, "Project created successfully", nil)
}

func (h *ProjectHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	req, ok := parsePagination(c, "name", "created_at", "updated_at")
	if !ok {
		return
	}
	result, err := h.service.ListMyProjects(userID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Projects fetched successfully", result)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	project, err := h.service.GetProjectByID(projectID, userID)
	if err != nil {
		switch err {
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
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
	userID := c.MustGet("userID").(uint)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project, err := h.service.UpdateProject(c.Request.Context(), projectID, userID, req.Name, req.Description)
	if err != nil {
		switch err {
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
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
	userID := c.MustGet("userID").(uint)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	err = h.service.DeleteProject(c.Request.Context(), projectID, userID)
	if err != nil {
		switch err {
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Project deleted successfully", nil)
}
