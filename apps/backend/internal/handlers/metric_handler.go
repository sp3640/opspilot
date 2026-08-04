package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type MetricHandler struct {
	service *services.MetricService
}

func NewMetricHandler(service *services.MetricService) *MetricHandler {
	return &MetricHandler{service: service}
}

func (h *MetricHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "created_at", "timestamp", "metric_type", "metric_name", "resource_kind", "value")
	if !ok {
		return
	}

	projectID, ok := parseOptionalUUID(c, "projectId")
	if !ok {
		return
	}
	req.ProjectID = projectID

	result, err := h.service.GetMetrics(organizationID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Metrics fetched successfully", result)
}

func (h *MetricHandler) GetLatest(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	projectID, ok := parseRequiredUUID(c, "projectId")
	if !ok {
		return
	}

	clusterID, ok := parseOptionalUUIDPointer(c, "clusterId")
	if !ok {
		return
	}
	resourceID, ok := parseOptionalUUIDPointer(c, "resourceId")
	if !ok {
		return
	}

	metricType := strings.TrimSpace(c.Query("metricType"))
	metricName := strings.TrimSpace(c.Query("metricName"))
	if metricType == "" {
		response.BadRequest(c, "metricType is required")
		return
	}

	result, err := h.service.GetLatest(organizationID, projectID, clusterID, resourceID, metricType, metricName)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Latest metric fetched successfully", result)
}

func (h *MetricHandler) GetHistory(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	projectID, ok := parseRequiredUUID(c, "projectId")
	if !ok {
		return
	}

	clusterID, ok := parseOptionalUUIDPointer(c, "clusterId")
	if !ok {
		return
	}
	resourceID, ok := parseOptionalUUIDPointer(c, "resourceId")
	if !ok {
		return
	}

	metricType := strings.TrimSpace(c.Query("metricType"))
	metricName := strings.TrimSpace(c.Query("metricName"))
	if metricType == "" {
		response.BadRequest(c, "metricType is required")
		return
	}

	start, ok := parseOptionalTime(c, "start")
	if !ok {
		return
	}
	end, ok := parseOptionalTime(c, "end")
	if !ok {
		return
	}

	limit := models.DefaultLimit
	if req, ok := parsePagination(c, "timestamp", "created_at"); ok {
		limit = req.Limit
	} else {
		return
	}

	result, err := h.service.GetHistory(organizationID, projectID, clusterID, resourceID, metricType, metricName, start, end, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Metric history fetched successfully", result)
}

func (h *MetricHandler) Aggregate(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	projectID, ok := parseRequiredUUID(c, "projectId")
	if !ok {
		return
	}

	metricType := strings.TrimSpace(c.Query("metricType"))
	metricName := strings.TrimSpace(c.Query("metricName"))
	if metricType == "" {
		response.BadRequest(c, "metricType is required")
		return
	}

	interval := strings.TrimSpace(c.Query("interval"))
	if interval == "" {
		interval = "hour"
	}

	startTime, ok := parseRequiredTime(c, "start")
	if !ok {
		return
	}
	endTime, ok := parseRequiredTime(c, "end")
	if !ok {
		return
	}

	result, err := h.service.Aggregate(organizationID, projectID, metricType, metricName, interval, startTime, endTime)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Metric aggregate fetched successfully", result)
}

func (h *MetricHandler) GetResourceMetrics(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	resourceIDValue, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	projectID, ok := parseRequiredUUID(c, "projectId")
	if !ok {
		return
	}

	metricType := strings.TrimSpace(c.Query("metricType"))
	metricName := strings.TrimSpace(c.Query("metricName"))
	if metricType == "" {
		response.BadRequest(c, "metricType is required")
		return
	}

	start, ok := parseOptionalTime(c, "start")
	if !ok {
		return
	}
	end, ok := parseOptionalTime(c, "end")
	if !ok {
		return
	}

	limit := models.DefaultLimit
	if req, ok := parsePagination(c, "timestamp", "created_at"); ok {
		limit = req.Limit
	} else {
		return
	}

	result, err := h.service.GetHistory(organizationID, projectID, nil, &resourceIDValue, metricType, metricName, start, end, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Resource metrics fetched successfully", result)
}

func (h *MetricHandler) GetClusterMetrics(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	clusterIDValue, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	projectID, ok := parseRequiredUUID(c, "projectId")
	if !ok {
		return
	}

	metricType := strings.TrimSpace(c.Query("metricType"))
	metricName := strings.TrimSpace(c.Query("metricName"))
	if metricType == "" {
		response.BadRequest(c, "metricType is required")
		return
	}

	start, ok := parseOptionalTime(c, "start")
	if !ok {
		return
	}
	end, ok := parseOptionalTime(c, "end")
	if !ok {
		return
	}

	limit := models.DefaultLimit
	if req, ok := parsePagination(c, "timestamp", "created_at"); ok {
		limit = req.Limit
	} else {
		return
	}

	result, err := h.service.GetHistory(organizationID, projectID, &clusterIDValue, nil, metricType, metricName, start, end, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster metrics fetched successfully", result)
}

func (h *MetricHandler) handleServiceError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "metric not found")
		return
	}

	switch err {
	case apperrors.ErrInvalidProject, apperrors.ErrProjectForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidMetricType, apperrors.ErrInvalidTimeRange:
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}

func parseRequiredUUID(c *gin.Context, key string) (uuid.UUID, bool) {
	value, ok := parseOptionalUUID(c, key)
	if !ok {
		return uuid.Nil, false
	}
	if value == uuid.Nil {
		response.BadRequest(c, key+" is required")
		return uuid.Nil, false
	}

	return value, true
}

func parseOptionalUUIDPointer(c *gin.Context, key string) (*uuid.UUID, bool) {
	value, ok := parseOptionalUUID(c, key)
	if !ok {
		return nil, false
	}
	if value == uuid.Nil {
		return nil, true
	}

	return &value, true
}

func parseRequiredTime(c *gin.Context, key string) (time.Time, bool) {
	parsed, ok := parseOptionalTime(c, key)
	if !ok {
		return time.Time{}, false
	}
	if parsed == nil {
		response.BadRequest(c, key+" is required")
		return time.Time{}, false
	}

	return parsed.UTC(), true
}

func parseOptionalTime(c *gin.Context, key string) (*time.Time, bool) {
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
