package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type IncidentHandler struct {
	service *services.IncidentService
}

func NewIncidentHandler(service *services.IncidentService) *IncidentHandler {
	return &IncidentHandler{service: service}
}

func (h *IncidentHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	var req dto.CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	incident, err := h.service.CreateIncident(
		c.Request.Context(),
		req.Title,
		req.Description,
		req.Severity,
		req.Status,
		req.ProjectID,
		userID,
		organizationID,
	)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidSeverity:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrInvalidStatus:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrInvalidProject:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.Created(c, "Incident created successfully", incident)
}

func (h *IncidentHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "title", "severity", "status", "created_at", "updated_at")
	if !ok {
		return
	}
	projectID, ok := parseOptionalUUID(c, "projectId")
	if !ok {
		return
	}

	req.ProjectID = projectID
	result, err := h.service.ListMyIncidents(organizationID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Incidents fetched successfully", result)
}

func (h *IncidentHandler) GetByID(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
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

	incident, err := h.service.GetIncidentByID(uint(incidentID), organizationID)
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

	response.OK(c, "Incident fetched successfully", incident)
}

func (h *IncidentHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	var req dto.UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	incident, err := h.service.UpdateIncident(
		c.Request.Context(),
		uint(incidentID),
		userID,
		organizationID,
		req.Title,
		req.Description,
		req.Severity,
		req.Status,
		req.ProjectID,
	)
	if err != nil {
		switch err {
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		case apperrors.ErrInvalidSeverity:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrInvalidStatus:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrInvalidProject:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Incident updated successfully", incident)
}

func (h *IncidentHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePlatformAdmin(c) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	err = h.service.DeleteIncident(c.Request.Context(), uint(incidentID), userID, organizationID)
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

	response.OK(c, "Incident deleted successfully", nil)
}
