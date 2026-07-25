package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler(service *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	summary, err := h.service.GetSummary(userID)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Dashboard summary fetched successfully", summary)
}

func (h *DashboardHandler) RecentIncidents(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil {
		response.BadRequest(c, "invalid limit")
		return
	}

	incidents, err := h.service.GetRecentIncidents(userID, limit)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Recent incidents fetched successfully", incidents)
}

func (h *DashboardHandler) Activity(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		response.BadRequest(c, "invalid limit")
		return
	}

	activity, err := h.service.GetRecentActivity(userID, limit)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Recent activity fetched successfully", activity)
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	stats, err := h.service.GetStats(userID)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Dashboard stats fetched successfully", stats)
}
