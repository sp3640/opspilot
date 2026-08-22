package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateClusterRequest is the validated input for cluster creation.
//
// Status, KubernetesVersion, ValidationError and the last-validated/discovery
// timestamps are intentionally absent: they reflect real connectivity state
// and are only ever set by the server (CreateCluster / ValidateClusterCredential),
// never trusted from client input.
type CreateClusterRequest struct {
	ProjectID           uuid.UUID       `json:"project_id"`
	Name                string          `json:"name"`
	Provider            string          `json:"provider"`
	ConnectionType      string          `json:"connection_type"`
	APIEndpoint         string          `json:"api_endpoint"`
	Region              string          `json:"region"`
	CredentialType      string          `json:"credential_type"`
	KubeconfigEncrypted string          `json:"kubeconfig_encrypted"`
	Metadata            json.RawMessage `json:"metadata"`
}

// UpdateClusterRequest is the validated input for a cluster update.
//
// KubeconfigEncrypted is a pointer so omitting the field from the JSON body
// (nil) preserves the stored credential unchanged, while sending it (even as
// an explicit value) replaces it. This is what prevents an edit that only
// changes e.g. the name from silently wiping stored credentials: the frontend
// never has the plaintext/ciphertext to round-trip, so it must be able to
// leave the field out entirely.
type UpdateClusterRequest struct {
	ProjectID           uuid.UUID       `json:"project_id"`
	Name                string          `json:"name"`
	Provider            string          `json:"provider"`
	ConnectionType      string          `json:"connection_type"`
	APIEndpoint         string          `json:"api_endpoint"`
	Region              string          `json:"region"`
	CredentialType      string          `json:"credential_type"`
	KubeconfigEncrypted *string         `json:"kubeconfig_encrypted,omitempty"`
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
// It never exposes internal database fields such as DeletedAt or raw
// relations, and — critically — never exposes the stored credential
// (EncryptedCredential/KubeconfigEncrypted are not mapped here at all).
type ClusterResponse struct {
	ID                string          `json:"id"`
	OrganizationID    string          `json:"organizationId"`
	ProjectID         string          `json:"projectId,omitempty"`
	Name              string          `json:"name"`
	Provider          string          `json:"provider"`
	ConnectionType    string          `json:"connectionType,omitempty"`
	CredentialType    string          `json:"credentialType"`
	KubernetesVersion string          `json:"kubernetesVersion,omitempty"`
	Status            string          `json:"status"`
	APIEndpoint       string          `json:"apiEndpoint"`
	Region            string          `json:"region,omitempty"`
	IsDefault         bool            `json:"isDefault"`
	LastValidatedAt   *time.Time      `json:"lastValidatedAt,omitempty"`
	LastDiscoveryAt   *time.Time      `json:"lastDiscoveryAt,omitempty"`
	CreatedBy         uint            `json:"createdBy"`
	ValidationError   string          `json:"validationError"`
	Metadata          json.RawMessage `json:"metadata"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// ValidationResponse is the result of a real connectivity check against the
// cluster's Kubernetes API server (see ClusterService.ValidateClusterCredential).
// It always reports the outcome — including an unreachable/invalid cluster —
// with success:true at the HTTP envelope level, since "the cluster is
// unreachable" is an expected, reportable outcome rather than a request error.
type ValidationResponse struct {
	Connected         bool             `json:"connected"`
	Status            string           `json:"status"`
	KubernetesVersion string           `json:"kubernetesVersion,omitempty"`
	APIServerURL      string           `json:"apiServerUrl,omitempty"`
	LatencyMs         int64            `json:"latencyMs"`
	ValidatedAt       time.Time        `json:"validatedAt"`
	Error             string           `json:"error,omitempty"`
	Cluster           *ClusterResponse `json:"cluster,omitempty"`
}

// ClusterListResponse is the paginated cluster list returned by GET /clusters.
type ClusterListResponse struct {
	Items      []ClusterResponse `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"totalPages"`
}
