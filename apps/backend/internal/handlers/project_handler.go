package handlers

import (
	"github.com/gin-gonic/gin"

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
		req.Name,
		req.Description,
		userID,
	)

	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Created(c, "Project created successfully", nil)
}

func (h *ProjectHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	projects, err := h.service.GetMyProjects(userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OK(c, "Projects fetched successfully", projects)
}