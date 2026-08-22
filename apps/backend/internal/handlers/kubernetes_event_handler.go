package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/response"
)

type kubernetesEventReader interface {
	ListEventsByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.EventListResponse, error)
	GetEvent(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.EventDetailResponse, error)
	ListEventsForPod(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, podName string) (*dto.EventListResponse, error)
}

type KubernetesEventHandler struct {
	service kubernetesEventReader
}

func NewKubernetesEventHandler(service kubernetesEventReader) *KubernetesEventHandler {
	return &KubernetesEventHandler{service: service}
}

func (h *KubernetesEventHandler) ListByApplication(c *gin.Context) {
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

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	namespace := strings.TrimSpace(c.Query("namespace"))
	result, err := h.service.ListEventsByApplication(c.Request.Context(), applicationID, organizationID, namespace)
	if err != nil {
		handleKubernetesEventError(c, err)
		return
	}

	response.OK(c, "Events fetched successfully", result)
}

func (h *KubernetesEventHandler) GetByNamespaceAndName(c *gin.Context) {
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

	namespace := strings.TrimSpace(c.Param("namespace"))
	name := strings.TrimSpace(c.Param("name"))
	result, err := h.service.GetEvent(c.Request.Context(), applicationID, organizationID, namespace, name)
	if err != nil {
		handleKubernetesEventError(c, err)
		return
	}

	response.OK(c, "Event fetched successfully", result)
}

// GetPodEvents lists the real Kubernetes events involving a single pod,
// mirroring the applicationId-query-param convention used by GetPodLogs.
func (h *KubernetesEventHandler) GetPodEvents(c *gin.Context) {
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

	namespace := strings.TrimSpace(c.Param("namespace"))
	name := strings.TrimSpace(c.Param("name"))
	result, err := h.service.ListEventsForPod(c.Request.Context(), applicationID, organizationID, namespace, name)
	if err != nil {
		handleKubernetesEventError(c, err)
		return
	}

	response.OK(c, "Pod events fetched successfully", result)
}

func handleKubernetesEventError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationNotFound, apperrors.ErrEventNotFound, apperrors.ErrEventNamespaceNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationForbidden, apperrors.ErrProjectForbidden, apperrors.ErrEventForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidDeploymentNamespace:
		response.Error(c, http.StatusBadRequest, err.Error())
	case apperrors.ErrEventInvalidKubeconfig, apperrors.ErrEventClusterUnavailable:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrEventTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
