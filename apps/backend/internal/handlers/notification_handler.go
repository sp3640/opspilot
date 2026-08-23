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

type NotificationHandler struct {
	service *services.NotificationService
}

func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionNotificationManage) {
		return
	}

	var req dto.CreateNotificationChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	channel, err := h.service.CreateChannel(userID, organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Notification channel created successfully", channel)
}

func (h *NotificationHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionNotificationRead) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "created_at", "name", "type")
	if !ok {
		return
	}
	teamID, ok := parseOptionalUUID(c, "teamId")
	if !ok {
		return
	}
	channelType := strings.ToUpper(strings.TrimSpace(c.Query("type")))

	result, err := h.service.ListChannels(organizationID, req, teamID, channelType)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Notification channels fetched successfully", result)
}

func (h *NotificationHandler) GetByID(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionNotificationRead) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid channel id")
		return
	}

	channel, err := h.service.GetChannel(organizationID, channelID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Notification channel fetched successfully", channel)
}

func (h *NotificationHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionNotificationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid channel id")
		return
	}

	var req dto.UpdateNotificationChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	channel, err := h.service.UpdateChannel(userID, organizationID, channelID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Notification channel updated successfully", channel)
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionNotificationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid channel id")
		return
	}

	userID := c.MustGet("userID").(uint)
	if err := h.service.DeleteChannel(userID, organizationID, channelID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Notification channel deleted successfully", nil)
}

func (h *NotificationHandler) Test(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionNotificationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid channel id")
		return
	}

	if err := h.service.SendTestNotification(c.Request.Context(), organizationID, channelID); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	response.OK(c, "Test notification sent successfully", nil)
}

func (h *NotificationHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrNotificationChannelNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrInvalidNotificationChannelType, apperrors.ErrInvalidNotificationChannelName,
		apperrors.ErrInvalidNotificationEvent, apperrors.ErrInvalidNotificationTarget, apperrors.ErrNotificationChannelTeamMismatch:
		response.BadRequest(c, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
