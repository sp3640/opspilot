package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/response"
)

type kubernetesLogReader interface {
	GetPodLogs(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, podName string, container string, tailLines *int64, sinceSeconds *int64, timestamps bool, previous bool) (*dto.PodLogResponse, error)
}

type KubernetesLogHandler struct {
	service kubernetesLogReader
}

func NewKubernetesLogHandler(service kubernetesLogReader) *KubernetesLogHandler {
	return &KubernetesLogHandler{service: service}
}

func (h *KubernetesLogHandler) GetPodLogs(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}
	if h.service == nil {
		response.InternalServerError(c)
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	applicationIDValue := strings.TrimSpace(c.Query("applicationId"))
	if applicationIDValue == "" {
		response.BadRequest(c, "applicationId is required")
		return
	}

	applicationID, err := uuid.Parse(applicationIDValue)
	if err != nil {
		response.BadRequest(c, "invalid applicationId")
		return
	}

	tailLines, ok := parseOptionalInt64Query(c, "tailLines")
	if !ok {
		return
	}
	sinceSeconds, ok := parseOptionalInt64Query(c, "sinceSeconds")
	if !ok {
		return
	}
	timestamps, ok := parseOptionalBoolQuery(c, "timestamps")
	if !ok {
		return
	}
	previous, ok := parseOptionalBoolQuery(c, "previous")
	if !ok {
		return
	}

	namespace := strings.TrimSpace(c.Param("namespace"))
	podName := strings.TrimSpace(c.Param("pod"))
	if podName == "" {
		podName = strings.TrimSpace(c.Param("name"))
	}
	container := strings.TrimSpace(c.Query("container"))
	result, err := h.service.GetPodLogs(c.Request.Context(), applicationID, organizationID, namespace, podName, container, tailLines, sinceSeconds, timestamps, previous)
	if err != nil {
		handleKubernetesLogError(c, err)
		return
	}

	response.OK(c, "Pod logs fetched successfully", result)
}

func parseOptionalInt64Query(c *gin.Context, key string) (*int64, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, true
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		response.BadRequest(c, "invalid "+key)
		return nil, false
	}

	return &parsed, true
}

func parseOptionalBoolQuery(c *gin.Context, key string) (bool, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return false, true
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		response.BadRequest(c, "invalid "+key)
		return false, false
	}

	return parsed, true
}

func handleKubernetesLogError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationNotFound, apperrors.ErrLogPodNotFound, apperrors.ErrLogContainerNotFound, apperrors.ErrLogNamespaceNotFound, apperrors.ErrLogPreviousNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationForbidden, apperrors.ErrProjectForbidden, apperrors.ErrLogForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidDeploymentNamespace:
		response.Error(c, http.StatusBadRequest, err.Error())
	case apperrors.ErrLogInvalidKubeconfig, apperrors.ErrLogClusterUnavailable:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrLogTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
