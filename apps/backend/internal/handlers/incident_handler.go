package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type IncidentHandler struct {
	service *services.IncidentService
}

type CreateIncidentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Severity    string `json:"severity" binding:"required"`
	Status      string `json:"status" binding:"required"`
	ProjectID   uint   `json:"project_id" binding:"required"`
}

type UpdateIncidentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Severity    string `json:"severity" binding:"required"`
	Status      string `json:"status" binding:"required"`
	ProjectID   uint   `json:"project_id" binding:"required"`
}

func NewIncidentHandler(service *services.IncidentService) *IncidentHandler {
	return &IncidentHandler{service: service}
}

func (h *IncidentHandler) Create(c *gin.Context) {
	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)

	incident, err := h.service.CreateIncident(
		c.Request.Context(),
		req.Title,
		req.Description,
		req.Severity,
		req.Status,
		req.ProjectID,
		userID,
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
	userID := c.MustGet("userID").(uint)

	req, ok := parsePagination(c, "title", "severity", "status", "created_at", "updated_at")
	if !ok {
		return
	}
	projectID, ok := parseOptionalPositiveUint(c, "projectID")
	if !ok {
		return
	}

	req.ProjectID = projectID
	result, err := h.service.ListMyIncidents(userID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Incidents fetched successfully", result)
}

func (h *IncidentHandler) GetByID(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	incident, err := h.service.GetIncidentByID(uint(incidentID), userID)
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

	response.OK(c, "Incident fetched successfully", incident)
}

func (h *IncidentHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	var req UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	incident, err := h.service.UpdateIncident(
		c.Request.Context(),
		uint(incidentID),
		userID,
		req.Title,
		req.Description,
		req.Severity,
		req.Status,
		req.ProjectID,
	)
	if err != nil {
		switch err {
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
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
	userID := c.MustGet("userID").(uint)

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	err = h.service.DeleteIncident(c.Request.Context(), uint(incidentID), userID)
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

	response.OK(c, "Incident deleted successfully", nil)
}
