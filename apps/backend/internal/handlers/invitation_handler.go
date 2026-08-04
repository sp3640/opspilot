package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type InvitationHandler struct {
	service *services.InvitationService
}

func NewInvitationHandler(service *services.InvitationService) *InvitationHandler {
	return &InvitationHandler{service: service}
}

func (h *InvitationHandler) Invite(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	userID := c.MustGet("userID").(uint)
	role := strings.TrimSpace(c.GetString("role"))

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	var req dto.InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	invitation, err := h.service.InviteUser(userID, role, organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Invitation created successfully", invitation)
}

func (h *InvitationHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	role := strings.TrimSpace(c.GetString("role"))

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "created_at", "updated_at", "email", "expires_at", "status")
	if !ok {
		return
	}

	invitations, err := h.service.ListInvitations(role, organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Invitations fetched successfully", invitations)
}

func (h *InvitationHandler) Accept(c *gin.Context) {
	if c.GetString("email") == "" {
		response.Unauthorized(c, "missing authenticated user")
		return
	}

	userID := c.MustGet("userID").(uint)
	userEmail := strings.TrimSpace(c.GetString("email"))

	var req dto.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	invitation, err := h.service.AcceptInvitation(userID, userEmail, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Invitation accepted successfully", invitation)
}

func (h *InvitationHandler) Revoke(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	role := strings.TrimSpace(c.GetString("role"))

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	invitationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid invitation id")
		return
	}

	if err := h.service.RevokeInvitation(invitationID, role, organizationID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Invitation revoked successfully", nil)
}

func (h *InvitationHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrInvitationAlreadyExists:
		response.Conflict(c, err.Error())
	case apperrors.ErrInvitationNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrInvitationForbidden, apperrors.ErrInvitationEmailMismatch:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvitationExpired, apperrors.ErrInvitationNotPending, apperrors.ErrInvalidInvitationToken, apperrors.ErrInvalidInvitationEmail, apperrors.ErrInvalidInvitationRole:
		response.BadRequest(c, err.Error())
	case apperrors.ErrUserNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
