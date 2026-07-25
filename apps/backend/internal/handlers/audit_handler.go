package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
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
	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	userID := c.MustGet("userID").(uint)
	req, ok := parsePagination(c, "created_at", "entity_type", "action")
	if !ok {
		return
	}
	result, err := h.service.ListIncidentAuditLogs(userID, uint(incidentID), req)
	if err != nil {
		switch err {
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
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
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	userID := c.MustGet("userID").(uint)
	req, ok := parsePagination(c, "created_at", "entity_type", "action")
	if !ok {
		return
	}
	result, err := h.service.ListProjectAuditLogs(userID, uint(projectID), req)
	if err != nil {
		switch err {
		case apperrors.ErrProjectNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Audit logs fetched successfully", result)
}
