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

type kubernetesReplicaSetReader interface {
	ListReplicaSetsByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.ReplicaSetListResponse, error)
	GetReplicaSet(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.ReplicaSet, error)
}

type KubernetesReplicaSetHandler struct {
	service kubernetesReplicaSetReader
}

func NewKubernetesReplicaSetHandler(service kubernetesReplicaSetReader) *KubernetesReplicaSetHandler {
	return &KubernetesReplicaSetHandler{service: service}
}

func (h *KubernetesReplicaSetHandler) ListByApplication(c *gin.Context) {
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
	result, err := h.service.ListReplicaSetsByApplication(c.Request.Context(), applicationID, organizationID, namespace)
	if err != nil {
		handleKubernetesReplicaSetError(c, err)
		return
	}

	response.OK(c, "ReplicaSets fetched successfully", result)
}

func (h *KubernetesReplicaSetHandler) GetByNamespaceAndName(c *gin.Context) {
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
	result, err := h.service.GetReplicaSet(c.Request.Context(), applicationID, organizationID, namespace, name)
	if err != nil {
		handleKubernetesReplicaSetError(c, err)
		return
	}

	response.OK(c, "ReplicaSet fetched successfully", result)
}

func handleKubernetesReplicaSetError(c *gin.Context, err error) {
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
