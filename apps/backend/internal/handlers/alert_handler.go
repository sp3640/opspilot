package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type AlertHandler struct {
	service *services.AlertService
}

type AttachIncidentRequest struct {
	IncidentID uint `json:"incidentId" binding:"required"`
}

func NewAlertHandler(service *services.AlertService) *AlertHandler {
	return &AlertHandler{service: service}
}

func (h *AlertHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionAlertManage) {
		return
	}

	var req dto.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	alert, err := h.service.CreateAlert(
		c.Request.Context(),
		req.ProjectID,
		req.IncidentID,
		req.Title,
		req.Description,
		req.Severity,
		req.Status,
		req.Source,
		req.ResourceType,
		req.ResourceID,
		req.Labels,
		req.Metadata,
		&req.FirstSeenAt,
		&req.LastSeenAt,
		userID,
		organizationID,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Alert created successfully", alert)
}

func (h *AlertHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "created_at", "updated_at", "severity", "status", "last_seen_at")
	if !ok {
		return
	}

	projectID, ok := parseOptionalUUID(c, "projectId")
	if !ok {
		return
	}
	req.ProjectID = projectID
	req.Status = strings.TrimSpace(strings.ToUpper(c.Query("status")))
	req.Severity = strings.TrimSpace(strings.ToUpper(c.Query("severity")))
	req.Source = strings.TrimSpace(strings.ToUpper(c.Query("source")))

	if req.Status != "" && !constants.IsValidAlertStatus(req.Status) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidAlertStatus.Error())
		return
	}
	if req.Severity != "" && !constants.IsValidAlertSeverity(req.Severity) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidAlertSeverity.Error())
		return
	}
	if req.Source != "" && !constants.IsValidAlertSource(req.Source) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidAlertSource.Error())
		return
	}

	result, err := h.service.ListAlerts(organizationID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Alerts fetched successfully", result)
}

func (h *AlertHandler) GetByID(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}

	alert, err := h.service.GetAlert(uint(alertID), organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Alert fetched successfully", alert)
}

func (h *AlertHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionAlertManage) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}

	var req dto.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	alert, err := h.service.UpdateAlert(
		c.Request.Context(),
		uint(alertID),
		userID,
		organizationID,
		req.ProjectID,
		req.IncidentID,
		req.Title,
		req.Description,
		req.Severity,
		req.Status,
		req.Source,
		req.ResourceType,
		req.ResourceID,
		req.Labels,
		req.Metadata,
		&req.FirstSeenAt,
		&req.LastSeenAt,
		req.AcknowledgedAt,
		req.ResolvedAt,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Alert updated successfully", alert)
}

func (h *AlertHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionAlertManage) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}

	err = h.service.DeleteAlert(c.Request.Context(), uint(alertID), userID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Alert deleted successfully", nil)
}

func (h *AlertHandler) Acknowledge(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionAlertAcknowledge) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}

	alert, err := h.service.AcknowledgeAlert(c.Request.Context(), uint(alertID), userID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Alert acknowledged successfully", alert)
}

func (h *AlertHandler) Resolve(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionAlertResolve) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}

	alert, err := h.service.ResolveAlert(c.Request.Context(), uint(alertID), userID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Alert resolved successfully", alert)
}

func (h *AlertHandler) Reopen(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionAlertReopen) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}

	alert, err := h.service.ReopenAlert(c.Request.Context(), uint(alertID), userID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Alert reopened successfully", alert)
}

func (h *AlertHandler) AttachIncident(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionAlertManage) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	alertID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid alert id")
		return
	}

	var req AttachIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	alert, err := h.service.AttachIncident(c.Request.Context(), uint(alertID), req.IncidentID, userID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Incident attached to alert successfully", alert)
}

func (h *AlertHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrAlertNotFound, apperrors.ErrIncidentNotFound:
		response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
	case apperrors.ErrProjectForbidden, apperrors.ErrInvalidProject:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidAlertSeverity, apperrors.ErrInvalidAlertStatus, apperrors.ErrInvalidAlertSource, apperrors.ErrInvalidAlertResourceType:
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
