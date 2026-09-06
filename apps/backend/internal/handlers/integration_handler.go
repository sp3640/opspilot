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

// IntegrationHandler exposes organization-scoped integration management.
// Read endpoints require only organization membership (mirroring every
// other org-wide read, e.g. AuditHandler/NotificationHandler); every
// mutation additionally requires organization:manage - reused as-is rather
// than introducing a new permission family, since Platform-Admin-only
// management is exactly what this endpoint set needs and organization:manage
// already means precisely that in the existing RBAC matrix.
type IntegrationHandler struct {
	service *services.IntegrationService
}

func NewIntegrationHandler(service *services.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{service: service}
}

func (h *IntegrationHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	var req dto.CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	integration, err := h.service.CreateIntegration(userID, organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Integration created successfully", integration)
}

func (h *IntegrationHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "created_at", "name", "type", "status")
	if !ok {
		return
	}
	integrationType := strings.ToLower(strings.TrimSpace(c.Query("type")))
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))

	result, err := h.service.ListIntegrations(organizationID, req, integrationType, status)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Integrations fetched successfully", result)
}

func (h *IntegrationHandler) GetByID(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	integration, err := h.service.GetIntegration(organizationID, integrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Integration fetched successfully", integration)
}

func (h *IntegrationHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	var req dto.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	integration, err := h.service.UpdateIntegration(userID, organizationID, integrationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Integration updated successfully", integration)
}

func (h *IntegrationHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	userID := c.MustGet("userID").(uint)
	if err := h.service.DeleteIntegration(userID, organizationID, integrationID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Integration deleted successfully", nil)
}

func (h *IntegrationHandler) Test(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	userID := c.MustGet("userID").(uint)
	result, err := h.service.TestConnection(c.Request.Context(), userID, organizationID, integrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Connection test completed", result)
}

func (h *IntegrationHandler) Check(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid integration id")
		return
	}

	userID := c.MustGet("userID").(uint)
	result, err := h.service.CheckStatus(c.Request.Context(), userID, organizationID, integrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Status check completed", result)
}

func (h *IntegrationHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrIntegrationNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrInvalidIntegrationType, apperrors.ErrInvalidIntegrationName,
		apperrors.ErrIntegrationCredentialsRequired:
		response.BadRequest(c, err.Error())
	case apperrors.ErrIntegrationAlreadyExists:
		response.Conflict(c, err.Error())
	case apperrors.ErrIntegrationEncryptionUnavailable:
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
