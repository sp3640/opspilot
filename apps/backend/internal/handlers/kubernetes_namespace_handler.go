package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
)

type kubernetesNamespaceReader interface {
	ListNamespacesForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID) (*dto.NamespaceListResponse, error)
}

type KubernetesNamespaceHandler struct {
	service kubernetesNamespaceReader
}

func NewKubernetesNamespaceHandler(service kubernetesNamespaceReader) *KubernetesNamespaceHandler {
	return &KubernetesNamespaceHandler{service: service}
}

func (h *KubernetesNamespaceHandler) ListByCluster(c *gin.Context) {
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

	result, err := h.service.ListNamespacesForCluster(c.Request.Context(), clusterID, organizationID)
	if err != nil {
		handleKubernetesNamespaceError(c, err)
		return
	}

	response.OK(c, "Namespaces fetched successfully", result)
}

func handleKubernetesNamespaceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrProjectForbidden, apperrors.ErrNamespaceForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
	case apperrors.ErrNamespaceInvalidKubeconfig, apperrors.ErrNamespaceClusterUnavailable:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrNamespaceTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
