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

type kubernetesIngressReader interface {
	ListIngressesByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.IngressListResponse, error)
	GetIngress(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.IngressResponse, error)
}

type KubernetesIngressHandler struct {
	service kubernetesIngressReader
}

func NewKubernetesIngressHandler(service kubernetesIngressReader) *KubernetesIngressHandler {
	return &KubernetesIngressHandler{service: service}
}

func (h *KubernetesIngressHandler) ListByApplication(c *gin.Context) {
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
	result, err := h.service.ListIngressesByApplication(c.Request.Context(), applicationID, organizationID, namespace)
	if err != nil {
		handleKubernetesIngressError(c, err)
		return
	}

	response.OK(c, "Ingresses fetched successfully", result)
}

func (h *KubernetesIngressHandler) GetByNamespaceAndName(c *gin.Context) {
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
	result, err := h.service.GetIngress(c.Request.Context(), applicationID, organizationID, namespace, name)
	if err != nil {
		handleKubernetesIngressError(c, err)
		return
	}

	response.OK(c, "Ingress fetched successfully", result)
}

func handleKubernetesIngressError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrApplicationNotFound, apperrors.ErrIngressNotFound, apperrors.ErrIngressNamespaceNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrApplicationForbidden, apperrors.ErrProjectForbidden, apperrors.ErrIngressForbidden:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidDeploymentNamespace:
		response.Error(c, http.StatusBadRequest, err.Error())
	case apperrors.ErrIngressInvalidKubeconfig, apperrors.ErrIngressClusterUnavailable:
		response.Error(c, http.StatusBadGateway, err.Error())
	case apperrors.ErrIngressTimeout:
		response.Error(c, http.StatusGatewayTimeout, err.Error())
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
