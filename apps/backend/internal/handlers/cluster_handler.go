package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type clusterDefaultSetter interface {
	SetDefaultCluster(ctx *gin.Context, id uuid.UUID, userID uint) (*dto.ClusterResponse, error)
}

type ClusterHandler struct {
	service *services.ClusterService
}

func NewClusterHandler(service *services.ClusterService) *ClusterHandler {
	return &ClusterHandler{service: service}
}

func (h *ClusterHandler) Create(c *gin.Context) {
	var req dto.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterManage) {
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}
	cluster, err := h.service.CreateCluster(
		c.Request.Context(),
		req.ProjectID,
		req.Name,
		req.Provider,
		req.ConnectionType,
		req.KubeconfigEncrypted,
		req.APIEndpoint,
		req.Region,
		req.Metadata,
		userID,
		organizationID,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Cluster created successfully", cluster)
}

func (h *ClusterHandler) List(c *gin.Context) {
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterRead) {
		return
	}

	req, ok := parsePagination(c, "created_at", "updated_at", "name", "provider", "status", "last_validated_at", "last_discovery_at")
	if !ok {
		return
	}

	projectID, ok := parseOptionalUUID(c, "projectId")
	if !ok {
		return
	}
	req.ProjectID = projectID

	provider := strings.TrimSpace(strings.ToUpper(c.Query("provider")))
	status := strings.TrimSpace(strings.ToUpper(c.Query("status")))

	if provider != "" && !constants.IsValidClusterProvider(provider) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidClusterProvider.Error())
		return
	}
	if status != "" && !constants.IsValidClusterStatus(status) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidClusterStatus.Error())
		return
	}

	req.Provider = provider
	req.Status = status

	result, err := h.service.ListClusters(organizationID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Clusters fetched successfully", result)
}

func (h *ClusterHandler) GetByID(c *gin.Context) {
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterRead) {
		return
	}

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	cluster, err := h.service.GetCluster(clusterID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster fetched successfully", cluster)
}

func (h *ClusterHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterManage) {
		return
	}

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	var req dto.UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cluster, err := h.service.UpdateCluster(
		c.Request.Context(),
		clusterID,
		userID,
		organizationID,
		req.ProjectID,
		req.Name,
		req.Provider,
		req.ConnectionType,
		req.KubeconfigEncrypted,
		req.APIEndpoint,
		req.Region,
		req.Metadata,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster updated successfully", cluster)
}

func (h *ClusterHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterManage) {
		return
	}

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	err = h.service.DeleteCluster(c.Request.Context(), clusterID, userID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster deleted successfully", nil)
}

func (h *ClusterHandler) Validate(c *gin.Context) {
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterValidate) {
		return
	}

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	validationResult, err := h.service.ValidateClusterCredential(c.Request.Context(), clusterID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if validationResult == nil {
		response.InternalServerError(c, apperrors.ErrClusterNotFound)
		return
	}

	// A disconnected/unreachable/misconfigured cluster is an expected,
	// reportable outcome of validation, not a request error — the HTTP
	// envelope always succeeds; callers branch on the "connected" field.
	if validationResult.Connected {
		response.OK(c, "Cluster is reachable", validationResult)
		return
	}

	response.OK(c, "Cluster validation failed: "+validationResult.Error, validationResult)
}

func (h *ClusterHandler) SetDefault(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionClusterSetDefault) {
		return
	}

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	cluster, err := h.service.SetDefaultCluster(c.Request.Context(), clusterID, userID, organizationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster set as default successfully", cluster)
}

func (h *ClusterHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusForbidden, apperrors.ErrProjectForbidden.Error())
	case apperrors.ErrProjectForbidden, apperrors.ErrInvalidProject:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidClusterProvider, apperrors.ErrInvalidClusterStatus, apperrors.ErrInvalidClusterConnectionType, apperrors.ErrClusterCredentialRequired:
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
