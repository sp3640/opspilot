package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	kubeintegration "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type clusterDefaultSetter interface {
	SetDefaultCluster(ctx *gin.Context, id uuid.UUID, userID uint) (*dto.ClusterResponse, error)
}

type ClusterHandler struct {
	service *services.ClusterService
}

type ClusterValidationResponse struct {
	Connected      bool                 `json:"connected"`
	ClusterVersion string               `json:"clusterVersion,omitempty"`
	APIServerURL   string               `json:"apiServerUrl,omitempty"`
	LatencyMs      int64                `json:"latencyMs"`
	ValidatedAt    time.Time            `json:"validatedAt"`
	Error          string               `json:"error,omitempty"`
	Cluster        *dto.ClusterResponse `json:"cluster,omitempty"`
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

	userID := c.MustGet("userID").(uint)
	cluster, err := h.service.CreateCluster(
		c.Request.Context(),
		req.ProjectID,
		req.Name,
		req.Provider,
		req.Status,
		req.ConnectionType,
		req.KubeconfigEncrypted,
		req.APIEndpoint,
		req.Region,
		req.Version,
		req.ValidationError,
		req.Metadata,
		req.LastValidatedAt,
		req.LastDiscoveryAt,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "Cluster created successfully", cluster)
}

func (h *ClusterHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

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

	req.Status = status
	if provider != "" {
		if req.Search == "" {
			req.Search = provider
		} else {
			req.Search = req.Search + " " + provider
		}
	}

	result, err := h.service.ListClusters(userID, req)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	if provider != "" {
		filtered := make([]dto.ClusterResponse, 0, len(result.Items))
		for _, item := range result.Items {
			if strings.EqualFold(strings.TrimSpace(item.Provider), provider) {
				filtered = append(filtered, item)
			}
		}
		result.Items = filtered
		result.Total = int64(len(filtered))
		result.TotalPages = 1
	}

	response.OK(c, "Clusters fetched successfully", result)
}

func (h *ClusterHandler) GetByID(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	cluster, err := h.service.GetCluster(clusterID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster fetched successfully", cluster)
}

func (h *ClusterHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

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
		req.ProjectID,
		req.Name,
		req.Provider,
		req.Status,
		req.ConnectionType,
		req.KubeconfigEncrypted,
		req.APIEndpoint,
		req.Region,
		req.Version,
		req.ValidationError,
		req.Metadata,
		req.LastValidatedAt,
		req.LastDiscoveryAt,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster updated successfully", cluster)
}

func (h *ClusterHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	err = h.service.DeleteCluster(c.Request.Context(), clusterID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster deleted successfully", nil)
}

func (h *ClusterHandler) Validate(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	cluster, err := h.service.GetCluster(clusterID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if strings.TrimSpace(strings.ToUpper(cluster.Provider)) != constants.ClusterProviderKubernetes {
		response.Error(c, http.StatusBadRequest, apperrors.ErrInvalidClusterProvider.Error())
		return
	}

	client := kubeintegration.NewClient([]byte(cluster.KubeconfigEncrypted))
	validator := kubeintegration.NewValidator(client)

	validatedAt := time.Now().UTC()
	updatedStatus := constants.ClusterStatusDisconnected
	updatedValidationError := ""
	updatedVersion := cluster.Version
	updatedAPIEndpoint := cluster.APIEndpoint

	validationResult, validationErr := validator.ValidateConnection(c.Request.Context())
	payload := ClusterValidationResponse{
		Connected:   false,
		ValidatedAt: validatedAt,
	}

	if validationErr == nil && validationResult != nil {
		payload.Connected = true
		payload.LatencyMs = validationResult.Latency.Milliseconds()
		payload.ValidatedAt = validationResult.ValidatedAt
		payload.APIServerURL = validationResult.APIServerURL
		if validationResult.ClusterVersion != nil {
			payload.ClusterVersion = validationResult.ClusterVersion.GitVersion
		}

		validatedAt = validationResult.ValidatedAt
		updatedStatus = constants.ClusterStatusConnected
		if payload.ClusterVersion != "" {
			updatedVersion = payload.ClusterVersion
		}
		if payload.APIServerURL != "" {
			updatedAPIEndpoint = payload.APIServerURL
		}
	} else if validationErr != nil {
		payload.Error = validationErr.Error()
		updatedValidationError = validationErr.Error()
	}

	updatedCluster, err := h.service.UpdateCluster(
		c.Request.Context(),
		clusterID,
		userID,
		uuid.MustParse(cluster.ProjectID),
		cluster.Name,
		cluster.Provider,
		updatedStatus,
		cluster.ConnectionType,
		cluster.KubeconfigEncrypted,
		updatedAPIEndpoint,
		cluster.Region,
		updatedVersion,
		updatedValidationError,
		cluster.Metadata,
		&validatedAt,
		cluster.LastDiscoveryAt,
	)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	payload.Cluster = updatedCluster

	if payload.Connected {
		response.OK(c, "Cluster validation successful", payload)
		return
	}

	response.OK(c, "Cluster validation failed", payload)
}

func (h *ClusterHandler) SetDefault(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	clusterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid cluster id")
		return
	}

	cluster, err := h.service.SetDefaultCluster(c.Request.Context(), clusterID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "Cluster set as default successfully", cluster)
}

func (h *ClusterHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case apperrors.ErrClusterNotFound:
		response.Error(c, http.StatusNotFound, err.Error())
	case apperrors.ErrProjectForbidden, apperrors.ErrInvalidProject:
		response.Error(c, http.StatusForbidden, err.Error())
	case apperrors.ErrInvalidClusterProvider, apperrors.ErrInvalidClusterStatus, apperrors.ErrInvalidClusterConnectionType:
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}
