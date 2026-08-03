package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler(service *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if _, _, ok := h.validateQuery(c, userID); !ok {
		return
	}

	summary, err := h.service.GetSummary(userID)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Dashboard overview fetched successfully", summary)
}

func (h *DashboardHandler) Resources(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if _, _, ok := h.validateQuery(c, userID); !ok {
		return
	}

	resources, err := h.service.GetRecentActivity(userID, 10)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Dashboard resources fetched successfully", resources)
}

func (h *DashboardHandler) Alerts(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if _, _, ok := h.validateQuery(c, userID); !ok {
		return
	}

	alerts, err := h.service.GetRecentIncidents(userID, 10)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Dashboard alerts fetched successfully", alerts)
}

func (h *DashboardHandler) Clusters(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if _, _, ok := h.validateQuery(c, userID); !ok {
		return
	}

	clusters, err := h.service.GetServiceHealth(userID)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Dashboard clusters fetched successfully", clusters)
}

func (h *DashboardHandler) Metrics(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if _, _, ok := h.validateQuery(c, userID); !ok {
		return
	}

	metrics, err := h.service.GetStats(userID)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Dashboard metrics fetched successfully", metrics)
}

func (h *DashboardHandler) validateQuery(c *gin.Context, userID uint) (uuid.UUID, *time.Time, bool) {
	projectID, ok := parseOptionalUUID(c, "projectId")
	if !ok {
		return uuid.Nil, nil, false
	}

	start, ok := parseOptionalRFC3339(c, "start")
	if !ok {
		return uuid.Nil, nil, false
	}
	end, ok := parseOptionalRFC3339(c, "end")
	if !ok {
		return uuid.Nil, nil, false
	}
	if start != nil && end != nil && end.Before(*start) {
		response.BadRequest(c, "end must be greater than or equal to start")
		return uuid.Nil, nil, false
	}

	if err := h.service.ValidateProjectAccess(userID, projectID); err != nil {
		switch err {
		case apperrors.ErrInvalidProject:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return uuid.Nil, nil, false
	}

	return projectID, start, true
}

func parseOptionalRFC3339(c *gin.Context, key string) (*time.Time, bool) {
	rawValue, provided := c.GetQuery(key)
	if !provided || strings.TrimSpace(rawValue) == "" {
		return nil, true
	}

	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(rawValue))
	if err != nil {
		response.BadRequest(c, "invalid "+key)
		return nil, false
	}

	value := parsed.UTC()
	return &value, true
}
