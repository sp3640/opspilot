package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateClusterRequest is the validated input for cluster creation.
type CreateClusterRequest struct {
	ProjectID           uuid.UUID       `json:"project_id" binding:"required"`
	Name                string          `json:"name" binding:"required,min=1,max=255"`
	Provider            string          `json:"provider" binding:"required"`
	Status              string          `json:"status" binding:"required"`
	ConnectionType      string          `json:"connection_type" binding:"max=64"`
	KubeconfigEncrypted string          `json:"kubeconfig_encrypted"`
	APIEndpoint         string          `json:"api_endpoint" binding:"max=255"`
	Region              string          `json:"region" binding:"max=128"`
	Version             string          `json:"version" binding:"max=64"`
	LastValidatedAt     *time.Time      `json:"last_validated_at,omitempty"`
	LastDiscoveryAt     *time.Time      `json:"last_discovery_at,omitempty"`
	ValidationError     string          `json:"validation_error"`
	Metadata            json.RawMessage `json:"metadata"`
}

// UpdateClusterRequest is the validated input for a full cluster update.
type UpdateClusterRequest struct {
	ProjectID           uuid.UUID       `json:"project_id" binding:"required"`
	Name                string          `json:"name" binding:"required,min=1,max=255"`
	Provider            string          `json:"provider" binding:"required"`
	Status              string          `json:"status" binding:"required"`
	ConnectionType      string          `json:"connection_type" binding:"max=64"`
	KubeconfigEncrypted string          `json:"kubeconfig_encrypted"`
	APIEndpoint         string          `json:"api_endpoint" binding:"max=255"`
	Region              string          `json:"region" binding:"max=128"`
	Version             string          `json:"version" binding:"max=64"`
	LastValidatedAt     *time.Time      `json:"last_validated_at,omitempty"`
	LastDiscoveryAt     *time.Time      `json:"last_discovery_at,omitempty"`
	ValidationError     string          `json:"validation_error"`
	Metadata            json.RawMessage `json:"metadata"`
}

// ClusterFilterRequest is the validated input for cluster list filtering.
type ClusterFilterRequest struct {
	Page           int       `json:"page"`
	Limit          int       `json:"limit"`
	Search         string    `json:"search"`
	Sort           string    `json:"sort"`
	Order          string    `json:"order"`
	ProjectID      uuid.UUID `json:"project_id"`
	Provider       string    `json:"provider"`
	Status         string    `json:"status"`
	ConnectionType string    `json:"connection_type"`
	Region         string    `json:"region"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// ClusterResponse is the canonical API representation of a cluster.
// It never exposes internal database fields such as DeletedAt or raw relations.
type ClusterResponse struct {
	ID                  string          `json:"id"`
	ProjectID           string          `json:"projectId"`
	Name                string          `json:"name"`
	Provider            string          `json:"provider"`
	Status              string          `json:"status"`
	IsDefault           bool            `json:"isDefault"`
	ConnectionType      string          `json:"connectionType"`
	KubeconfigEncrypted string          `json:"kubeconfigEncrypted"`
	APIEndpoint         string          `json:"apiEndpoint"`
	Region              string          `json:"region"`
	Version             string          `json:"version"`
	LastValidatedAt     *time.Time      `json:"lastValidatedAt,omitempty"`
	LastDiscoveryAt     *time.Time      `json:"lastDiscoveryAt,omitempty"`
	CreatedBy           uint            `json:"createdBy"`
	ValidationError     string          `json:"validationError"`
	Metadata            json.RawMessage `json:"metadata"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

// ClusterListResponse is the paginated cluster list returned by GET /clusters.
type ClusterListResponse struct {
	Items      []ClusterResponse `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"totalPages"`
}
