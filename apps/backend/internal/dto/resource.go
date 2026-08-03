package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateResourceRequest is the validated input for resource creation.
type CreateResourceRequest struct {
	ProjectID        uuid.UUID       `json:"project_id" binding:"required"`
	ParentResourceID *uuid.UUID      `json:"parent_resource_id,omitempty"`
	Kind             string          `json:"kind" binding:"required"`
	Name             string          `json:"name" binding:"required,min=1,max=255"`
	DisplayName      string          `json:"display_name" binding:"max=255"`
	ExternalID       string          `json:"external_id" binding:"max=255"`
	Provider         string          `json:"provider" binding:"max=64"`
	Region           string          `json:"region" binding:"max=128"`
	Namespace        string          `json:"namespace" binding:"max=255"`
	Cluster          string          `json:"cluster" binding:"max=255"`
	Status           string          `json:"status" binding:"required"`
	Health           string          `json:"health" binding:"required"`
	Labels           json.RawMessage `json:"labels"`
	Annotations      json.RawMessage `json:"annotations"`
	Metadata         json.RawMessage `json:"metadata"`
}

// UpdateResourceRequest is the validated input for a full resource update.
type UpdateResourceRequest struct {
	ProjectID        uuid.UUID       `json:"project_id" binding:"required"`
	ParentResourceID *uuid.UUID      `json:"parent_resource_id,omitempty"`
	Kind             string          `json:"kind" binding:"required"`
	Name             string          `json:"name" binding:"required,min=1,max=255"`
	DisplayName      string          `json:"display_name" binding:"max=255"`
	ExternalID       string          `json:"external_id" binding:"max=255"`
	Provider         string          `json:"provider" binding:"max=64"`
	Region           string          `json:"region" binding:"max=128"`
	Namespace        string          `json:"namespace" binding:"max=255"`
	Cluster          string          `json:"cluster" binding:"max=255"`
	Status           string          `json:"status" binding:"required"`
	Health           string          `json:"health" binding:"required"`
	Labels           json.RawMessage `json:"labels"`
	Annotations      json.RawMessage `json:"annotations"`
	Metadata         json.RawMessage `json:"metadata"`
}

// ResourceFilterRequest is the validated input for resource list filtering.
type ResourceFilterRequest struct {
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
	Search    string    `json:"search"`
	Sort      string    `json:"sort"`
	Order     string    `json:"order"`
	ProjectID uuid.UUID `json:"project_id"`
	Kind      string    `json:"kind"`
	Status    string    `json:"status"`
	Health    string    `json:"health"`
	Provider  string    `json:"provider"`
	Region    string    `json:"region"`
	Namespace string    `json:"namespace"`
	Cluster   string    `json:"cluster"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// ResourceResponse is the canonical API representation of a resource.
// It never exposes internal database fields such as DeletedAt or raw relations.
type ResourceResponse struct {
	ID               string          `json:"id"`
	ProjectID        string          `json:"projectId"`
	ParentResourceID *string         `json:"parentResourceId,omitempty"`
	Kind             string          `json:"kind"`
	Name             string          `json:"name"`
	DisplayName      string          `json:"displayName"`
	ExternalID       string          `json:"externalId"`
	Provider         string          `json:"provider"`
	Region           string          `json:"region"`
	Namespace        string          `json:"namespace"`
	Cluster          string          `json:"cluster"`
	Status           string          `json:"status"`
	Health           string          `json:"health"`
	Labels           json.RawMessage `json:"labels"`
	Annotations      json.RawMessage `json:"annotations"`
	Metadata         json.RawMessage `json:"metadata"`
	CreatedBy        uint            `json:"createdBy"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

// ResourceListResponse is the paginated resource list returned by GET /resources.
type ResourceListResponse struct {
	Items      []ResourceResponse `json:"items"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"totalPages"`
}
