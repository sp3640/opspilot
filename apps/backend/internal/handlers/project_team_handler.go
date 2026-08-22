package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type ProjectTeamHandler struct {
	service *services.ProjectTeamService
}

func NewProjectTeamHandler(service *services.ProjectTeamService) *ProjectTeamHandler {
	return &ProjectTeamHandler{service: service}
}

func (h *ProjectTeamHandler) AssignTeam(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionProjectTeamManage) {
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

	var req dto.AssignTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	teamID, err := uuid.Parse(strings.TrimSpace(req.TeamID))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	assigned, err := h.service.AssignTeam(c.GetString("role"), organizationID, projectID, teamID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Team assigned to project successfully", assigned)
}

func (h *ProjectTeamHandler) RemoveTeam(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionProjectTeamManage) {
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

	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	if err := h.service.RemoveTeam(c.GetString("role"), organizationID, projectID, teamID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team removed from project successfully", nil)
}

func (h *ProjectTeamHandler) ListProjectTeams(c *gin.Context) {
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

	items, err := h.service.ListProjectTeams(c.GetString("role"), organizationID, projectID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Project teams fetched successfully", items)
}

func (h *ProjectTeamHandler) ListTeamProjects(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	items, err := h.service.ListTeamProjects(c.GetString("role"), organizationID, teamID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team projects fetched successfully", items)
}

func (h *ProjectTeamHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrProjectTeamAlreadyAssigned:
		response.Conflict(c, err.Error())
	case apperrors.ErrProjectNotFound, apperrors.ErrTeamNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrProjectForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
