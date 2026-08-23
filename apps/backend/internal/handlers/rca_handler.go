package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

// RCAHandler exposes the deterministic Incident Intelligence / Root Cause
// Analysis endpoint. It is read-only: it never mutates the incident, and
// gates on the same PermissionIncidentRead every other incident-viewing
// endpoint uses - no new elevated permission is needed to read a
// correlation over data the caller could already see individually.
type RCAHandler struct {
	service *services.RCAService
}

func NewRCAHandler(service *services.RCAService) *RCAHandler {
	return &RCAHandler{service: service}
}

func (h *RCAHandler) GetRootCauseAnalysis(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionIncidentRead) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	analysis, err := h.service.Analyze(c.Request.Context(), uint(incidentID), organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Incident intelligence generated successfully", analysis)
}

func (h *RCAHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrIncidentNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
