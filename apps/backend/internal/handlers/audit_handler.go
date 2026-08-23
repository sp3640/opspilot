package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type AuditHandler struct {
	service *services.AuditService
}

func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) GetIncidentAuditLogs(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	req, ok := parseAuditPagination(c)
	if !ok {
		return
	}
	result, err := h.service.ListIncidentAuditLogs(userID, organizationID, uint(incidentID), req)
	if err != nil {
		switch err {
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Audit logs fetched successfully", result)
}

func (h *AuditHandler) GetProjectAuditLogs(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	req, ok := parseAuditPagination(c)
	if !ok {
		return
	}
	result, err := h.service.ListProjectAuditLogs(userID, organizationID, projectID, req)
	if err != nil {
		switch err {
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Audit logs fetched successfully", result)
}

// GetOrganizationAuditLogs returns every audit log in the caller's own
// organization, unscoped by project/incident/entity - the Organization
// Audit view (Phase 23). Read-only and gated the same as every other
// org-wide list endpoint (org membership only).
func (h *AuditHandler) GetOrganizationAuditLogs(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	req, ok := parseAuditPagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListOrganizationAuditLogs(organizationID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Audit logs fetched successfully", result)
}

// parseAuditPagination layers the audit-specific filters (user, application,
// date range - all three set directly rather than via the generic
// ParsePagination parser, mirroring how ProjectID/IncidentID already work)
// on top of the shared pagination/action/entityType/result parsing.
func parseAuditPagination(c *gin.Context) (*models.PaginationRequest, bool) {
	req, ok := parsePagination(c, "created_at", "entity_type", "action")
	if !ok {
		return nil, false
	}

	userID, ok := parseOptionalUint(c, "userId")
	if !ok {
		return nil, false
	}
	req.UserID = userID

	applicationID, ok := parseOptionalUUID(c, "applicationId")
	if !ok {
		return nil, false
	}
	req.ApplicationID = applicationID

	dateFrom, ok := parseOptionalTime(c, "dateFrom")
	if !ok {
		return nil, false
	}
	req.DateFrom = dateFrom

	dateTo, ok := parseOptionalTime(c, "dateTo")
	if !ok {
		return nil, false
	}
	req.DateTo = dateTo

	return req, true
}
