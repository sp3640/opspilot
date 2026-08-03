package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/discovery"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type resourceClusterReader interface {
	GetCluster(id uuid.UUID, userID uint) (*dto.ClusterResponse, error)
}

type resourceSyncRunner interface {
	Run(ctx context.Context, clusterID uuid.UUID) (*discovery.WorkerResult, error)
}

type ResourceHandler struct {
	service       *services.ResourceService
	clusterReader resourceClusterReader
	syncRunner    resourceSyncRunner
}

type SyncResourcesRequest struct {
	ProjectID uuid.UUID `json:"projectId" binding:"required"`
	ClusterID uuid.UUID `json:"clusterId" binding:"required"`
}

func NewResourceHandler(service *services.ResourceService, clusterReader resourceClusterReader, syncRunner resourceSyncRunner) *ResourceHandler {
	return &ResourceHandler{
		service:       service,
		clusterReader: clusterReader,
		syncRunner:    syncRunner,
	}
}

func (h *ResourceHandler) Create(c *gin.Context) {
	var req dto.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	resource, err := h.service.CreateResource(
		c.Request.Context(),
		req.ProjectID,
		req.ParentResourceID,
		req.Kind,
		req.Name,
		req.DisplayName,
		req.ExternalID,
		req.Provider,
		req.Region,
		req.Namespace,
		req.Cluster,
		req.Status,
		req.Health,
		req.Labels,
		req.Annotations,
		req.Metadata,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Resource created successfully", resource)
}

func (h *ResourceHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	req, ok := parsePagination(c, "created_at", "updated_at", "name", "kind", "status", "health")
	if !ok {
		return
	}

	projectID, ok := parseOptionalUUID(c, "projectId")
	if !ok {
		return
	}
	req.ProjectID = projectID
	req.Kind = strings.TrimSpace(c.Query("kind"))
	req.Status = strings.TrimSpace(strings.ToUpper(c.Query("status")))
	req.Health = strings.TrimSpace(strings.ToUpper(c.Query("health")))

	if req.Kind != "" && !constants.IsValidResourceKind(req.Kind) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidResourceKind.Error())
		return
	}
	if req.Status != "" && !constants.IsValidResourceStatus(req.Status) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidResourceStatus.Error())
		return
	}
	if req.Health != "" && !constants.IsValidResourceHealth(req.Health) {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidResourceHealth.Error())
		return
	}

	result, err := h.service.ListResources(userID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "Resources fetched successfully", result)
}

func (h *ResourceHandler) GetByID(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	resourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	resource, err := h.service.GetResource(resourceID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Resource fetched successfully", resource)
}

func (h *ResourceHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	resourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	var req dto.UpdateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resource, err := h.service.UpdateResource(
		c.Request.Context(),
		resourceID,
		userID,
		req.ProjectID,
		req.ParentResourceID,
		req.Kind,
		req.Name,
		req.DisplayName,
		req.ExternalID,
		req.Provider,
		req.Region,
		req.Namespace,
		req.Cluster,
		req.Status,
		req.Health,
		req.Labels,
		req.Annotations,
		req.Metadata,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Resource updated successfully", resource)
}

func (h *ResourceHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	resourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid resource id")
		return
	}

	err = h.service.DeleteResource(c.Request.Context(), resourceID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Resource deleted successfully", nil)
}

func (h *ResourceHandler) Sync(c *gin.Context) {
	if h.syncRunner == nil || h.clusterReader == nil {
		response.InternalServerError(c)
		return
	}

	var req SyncResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	cluster, err := h.clusterReader.GetCluster(req.ClusterID, userID)
	if err != nil {
		switch err {
		case apperrors.ErrClusterNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	if cluster.ProjectID != req.ProjectID.String() {
		response.Error(c, http.StatusForbidden, apperrors.ErrInvalidProject.Error())
		return
	}

	result, err := h.syncRunner.Run(c.Request.Context(), req.ClusterID)
	if err != nil {
		switch err {
		case discovery.ErrClusterDisconnected:
			response.Error(c, http.StatusConflict, err.Error())
		case discovery.ErrDiscoveryNotSupported:
			response.Error(c, http.StatusBadRequest, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}
	if result == nil {
		response.InternalServerError(c)
		return
	}
	if result.Skipped {
		response.Error(c, http.StatusConflict, result.SkipReason)
		return
	}
	if result.SyncResult == nil {
		response.InternalServerError(c)
		return
	}

	response.OK(c, "Resources synchronized successfully", result.SyncResult)
}

func (h *ResourceHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrResourceNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrProjectForbidden, apperrors.ErrInvalidProject:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidResourceKind, apperrors.ErrInvalidResourceStatus, apperrors.ErrInvalidResourceHealth:
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
