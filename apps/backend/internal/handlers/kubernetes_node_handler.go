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

type kubernetesNodeReader interface {
	ListNodesForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID) (*dto.NodeListResponse, error)
}

type KubernetesNodeHandler struct {
	service kubernetesNodeReader
}

func NewKubernetesNodeHandler(service kubernetesNodeReader) *KubernetesNodeHandler {
	return &KubernetesNodeHandler{service: service}
}

func (h *KubernetesNodeHandler) ListByCluster(c *gin.Context) {
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

	result, err := h.service.ListNodesForCluster(c.Request.Context(), clusterID, organizationID)
	if err != nil {
		handleKubernetesNodeError(c, err)
		return
	}

	response.OK(c, "Nodes fetched successfully", result)
}

func handleKubernetesNodeError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrProjectForbidden, apperrors.ErrNodeForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrClusterNotFound:
		// Cross-org or nonexistent cluster: respond identically to a
		// forbidden cluster rather than leaking existence via a distinct
		// status code.
		response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
	case apperrors.ErrNodeInvalidKubeconfig, apperrors.ErrNodeClusterUnavailable:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrNodeTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
