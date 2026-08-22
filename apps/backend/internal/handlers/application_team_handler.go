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

type ApplicationTeamHandler struct {
	service *services.ApplicationTeamService
}

func NewApplicationTeamHandler(service *services.ApplicationTeamService) *ApplicationTeamHandler {
	return &ApplicationTeamHandler{service: service}
}

func (h *ApplicationTeamHandler) AssignTeam(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionApplicationTeamManage) {
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

	var req dto.AssignApplicationTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	teamID, err := uuid.Parse(strings.TrimSpace(req.TeamID))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	assigned, err := h.service.AssignTeam(c.GetString("role"), organizationID, applicationID, teamID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Team assigned to application successfully", assigned)
}

func (h *ApplicationTeamHandler) RemoveTeam(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionApplicationTeamManage) {
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

	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	if err := h.service.RemoveTeam(c.GetString("role"), organizationID, applicationID, teamID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team removed from application successfully", nil)
}

func (h *ApplicationTeamHandler) ListApplicationTeams(c *gin.Context) {
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

	items, err := h.service.ListApplicationTeams(c.GetString("role"), organizationID, applicationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Application teams fetched successfully", items)
}

func (h *ApplicationTeamHandler) ListTeamApplications(c *gin.Context) {
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

	items, err := h.service.ListTeamApplications(c.GetString("role"), organizationID, teamID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team applications fetched successfully", items)
}

func (h *ApplicationTeamHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationTeamAlreadyAssigned:
		response.Conflict(c, err.Error())
	case apperrors.ErrApplicationNotFound, apperrors.ErrTeamNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
