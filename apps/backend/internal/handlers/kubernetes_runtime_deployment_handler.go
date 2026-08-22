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

type kubernetesRuntimeDeploymentReader interface {
	ListDeploymentsByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.DeploymentRuntimeListResponse, error)
	GetDeployment(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.DeploymentRuntimeDetailResponse, error)
	ListDeploymentsForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.DeploymentRuntimeListResponse, error)
}

type KubernetesRuntimeDeploymentHandler struct {
	service kubernetesRuntimeDeploymentReader
}

func NewKubernetesRuntimeDeploymentHandler(service kubernetesRuntimeDeploymentReader) *KubernetesRuntimeDeploymentHandler {
	return &KubernetesRuntimeDeploymentHandler{service: service}
}

func (h *KubernetesRuntimeDeploymentHandler) ListByApplication(c *gin.Context) {
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
	result, err := h.service.ListDeploymentsByApplication(c.Request.Context(), applicationID, organizationID, namespace)
	if err != nil {
		handleKubernetesRuntimeDeploymentError(c, err)
		return
	}

	response.OK(c, "Runtime deployments fetched successfully", result)
}

func (h *KubernetesRuntimeDeploymentHandler) ListByCluster(c *gin.Context) {
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
	result, err := h.service.ListDeploymentsForCluster(c.Request.Context(), clusterID, organizationID, namespace)
	if err != nil {
		if err == apperrors.ErrClusterNotFound {
			response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
			return
		}
		handleKubernetesRuntimeDeploymentError(c, err)
		return
	}

	response.OK(c, "Runtime deployments fetched successfully", result)
}

func (h *KubernetesRuntimeDeploymentHandler) GetByNamespaceAndName(c *gin.Context) {
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
	result, err := h.service.GetDeployment(c.Request.Context(), applicationID, organizationID, namespace, name)
	if err != nil {
		handleKubernetesRuntimeDeploymentError(c, err)
		return
	}

	response.OK(c, "Runtime deployment fetched successfully", result)
}

func handleKubernetesRuntimeDeploymentError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationNotFound, apperrors.ErrRuntimeDeploymentNotFound, apperrors.ErrRuntimeDeploymentNamespaceNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationForbidden, apperrors.ErrProjectForbidden, apperrors.ErrRuntimeDeploymentForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidDeploymentNamespace:
		response.Error(c, http.StatusBadRequest, err.Error())
	case apperrors.ErrRuntimeDeploymentInvalidKubeconfig, apperrors.ErrRuntimeDeploymentClusterUnavailable:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrRuntimeDeploymentTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
