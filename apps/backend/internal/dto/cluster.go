package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateClusterRequest is the validated input for cluster creation.
type CreateClusterRequest struct {
	ProjectID           uuid.UUID       `json:"project_id"`
	Name                string          `json:"name"`
	Provider            string          `json:"provider"`
	ConnectionType      string          `json:"connection_type"`
	KubernetesVersion   string          `json:"kubernetes_version"`
	Version             string          `json:"version"`
	APIEndpoint         string          `json:"api_endpoint"`
	Region              string          `json:"region"`
	CredentialType      string          `json:"credential_type"`
	Credential          string          `json:"credential"`
	KubeconfigEncrypted string          `json:"kubeconfig_encrypted"`
	Status              string          `json:"status"`
	LastValidatedAt     *time.Time      `json:"last_validated_at,omitempty"`
	LastDiscoveryAt     *time.Time      `json:"last_discovery_at,omitempty"`
	ValidationError     string          `json:"validation_error"`
	Metadata            json.RawMessage `json:"metadata"`
}

// UpdateClusterRequest is the validated input for a full cluster update.
type UpdateClusterRequest struct {
	ProjectID           uuid.UUID       `json:"project_id"`
	Name                string          `json:"name"`
	Provider            string          `json:"provider"`
	ConnectionType      string          `json:"connection_type"`
	KubernetesVersion   string          `json:"kubernetes_version"`
	Version             string          `json:"version"`
	APIEndpoint         string          `json:"api_endpoint"`
	Region              string          `json:"region"`
	CredentialType      string          `json:"credential_type"`
	Credential          string          `json:"credential"`
	KubeconfigEncrypted string          `json:"kubeconfig_encrypted"`
	Status              string          `json:"status"`
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
	ID                string          `json:"id"`
	OrganizationID    string          `json:"organizationId"`
	ProjectID         string          `json:"projectId,omitempty"`
	Name              string          `json:"name"`
	Provider          string          `json:"provider"`
	ConnectionType    string          `json:"connectionType,omitempty"`
	CredentialType    string          `json:"credentialType"`
	KubernetesVersion string          `json:"kubernetesVersion"`
	Version           string          `json:"version,omitempty"`
	Status            string          `json:"status"`
	APIEndpoint       string          `json:"apiEndpoint"`
	Region            string          `json:"region,omitempty"`
	IsDefault         bool            `json:"isDefault"`
	LastValidatedAt   *time.Time      `json:"lastValidatedAt,omitempty"`
	CreatedBy         uint            `json:"createdBy"`
	ValidationError   string          `json:"validationError"`
	Metadata          json.RawMessage `json:"metadata"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

type ValidationResponse struct {
	Status      string           `json:"status"`
	Healthy     bool             `json:"healthy"`
	ValidatedAt time.Time        `json:"validatedAt"`
	Error       string           `json:"error,omitempty"`
	Cluster     *ClusterResponse `json:"cluster,omitempty"`
}

// ClusterListResponse is the paginated cluster list returned by GET /clusters.
type ClusterListResponse struct {
	Items      []ClusterResponse `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"totalPages"`
}
