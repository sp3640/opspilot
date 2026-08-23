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

type SREHandler struct {
	service *services.SREMetricsService
}

func NewSREHandler(service *services.SREMetricsService) *SREHandler {
	return &SREHandler{service: service}
}

func (h *SREHandler) ConfigureSLO(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionSLOManage) {
		return
	}

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	var req dto.ApplicationSLOConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	result, err := h.service.ConfigureSLO(userID, applicationID, organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "SLO configuration saved successfully", result)
}

func (h *SREHandler) GetSLO(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionSLORead) {
		return
	}

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	result, err := h.service.GetSLOConfig(applicationID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "SLO configuration fetched successfully", result)
}

func (h *SREHandler) GetMetrics(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionSLORead) {
		return
	}

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	result, err := h.service.GetMetrics(applicationID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "SRE metrics fetched successfully", result)
}

func (h *SREHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationNotFound, apperrors.ErrSLONotConfigured:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidSLOTarget, apperrors.ErrInvalidSLOWindow:
		response.BadRequest(c, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
