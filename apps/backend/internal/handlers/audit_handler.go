package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
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
	req, ok := parsePagination(c, "created_at", "entity_type", "action")
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
	req, ok := parsePagination(c, "created_at", "entity_type", "action")
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
