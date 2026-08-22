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
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
)

type kubernetesServiceReader interface {
	ListServicesByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.ServiceListResponse, error)
	GetService(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.ServiceResponse, error)
	ListServicesForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.ServiceListResponse, error)
}

type KubernetesServiceHandler struct {
	service kubernetesServiceReader
}

func NewKubernetesServiceHandler(service kubernetesServiceReader) *KubernetesServiceHandler {
	return &KubernetesServiceHandler{service: service}
}

func (h *KubernetesServiceHandler) ListByApplication(c *gin.Context) {
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
	result, err := h.service.ListServicesByApplication(c.Request.Context(), applicationID, organizationID, namespace)
	if err != nil {
		handleKubernetesServiceError(c, err)
		return
	}

	response.OK(c, "Services fetched successfully", result)
}

func (h *KubernetesServiceHandler) ListByCluster(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterRead) {
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

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	namespace := strings.TrimSpace(c.Query("namespace"))
	result, err := h.service.ListServicesForCluster(c.Request.Context(), clusterID, organizationID, namespace)
	if err != nil {
		if err == apperrors.ErrClusterNotFound {
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
			return
		}
		handleKubernetesServiceError(c, err)
		return
	}

	response.OK(c, "Services fetched successfully", result)
}

func (h *KubernetesServiceHandler) GetByNamespaceAndName(c *gin.Context) {
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
	result, err := h.service.GetService(c.Request.Context(), applicationID, organizationID, namespace, name)
	if err != nil {
		handleKubernetesServiceError(c, err)
		return
	}

	response.OK(c, "Service fetched successfully", result)
}

func handleKubernetesServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationNotFound, apperrors.ErrServiceNotFound, apperrors.ErrServiceNamespaceNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationForbidden, apperrors.ErrProjectForbidden, apperrors.ErrServiceForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidDeploymentNamespace:
		response.Error(c, http.StatusBadRequest, err.Error())
	case apperrors.ErrServiceInvalidKubeconfig, apperrors.ErrServiceClusterUnavailable:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrServiceTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
