package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type TeamHandler struct {
	service *services.TeamService
}

func NewTeamHandler(service *services.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionTeamManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	team, err := h.service.CreateTeam(organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Team created successfully", team)
}

func (h *TeamHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "name", "created_at", "updated_at")
	if !ok {
		return
	}

	result, err := h.service.ListTeams(organizationID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Teams fetched successfully", result)
}

func (h *TeamHandler) GetByID(c *gin.Context) {
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

	team, err := h.service.GetTeamByID(teamID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team fetched successfully", team)
}

func (h *TeamHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionTeamManage) {
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

	var req dto.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	team, err := h.service.UpdateTeam(teamID, organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team updated successfully", team)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionTeamManage) {
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

	if err := h.service.DeleteTeam(teamID, organizationID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team deleted successfully", nil)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionTeamManage) {
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

	var req dto.TeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	member, err := h.service.AddMember(teamID, organizationID, req.UserID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Team member added successfully", member)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionTeamManage) {
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

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.service.RemoveMember(teamID, organizationID, uint(userID)); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team member removed successfully", nil)
}

func (h *TeamHandler) ListMembers(c *gin.Context) {
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

	members, err := h.service.ListMembers(teamID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Team members fetched successfully", members)
}

func (h *TeamHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrTeamAlreadyExists, apperrors.ErrTeamMemberAlreadyExists:
		response.Conflict(c, err.Error())
	case apperrors.ErrTeamNotFound, apperrors.ErrTeamMemberNotFound, apperrors.ErrUserNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrTeamForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidTeamName, apperrors.ErrInvalidTeamDescription:
		response.BadRequest(c, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
