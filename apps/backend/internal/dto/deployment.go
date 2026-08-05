package dto

import "time"

type CreateDeploymentRequest struct {
	ApplicationID      string `json:"applicationId" binding:"required,uuid"`
	ProjectID          string `json:"projectId" binding:"required,uuid"`
	TargetClusterID    string `json:"targetClusterId" binding:"required,uuid"`
	Image              string `json:"image" binding:"required,max=500"`
	ImageTag           string `json:"imageTag" binding:"omitempty,max=128"`
	Environment        string `json:"environment" binding:"required,max=32"`
	Namespace          string `json:"namespace" binding:"required,max=63"`
	ReplicaCount       int    `json:"replicaCount" binding:"required"`
	DeploymentStrategy string `json:"deploymentStrategy" binding:"required,max=64"`
}

type UpdateDeploymentRequest struct {
	TargetClusterID    *string `json:"targetClusterId" binding:"omitempty,uuid"`
	Image              *string `json:"image" binding:"omitempty,max=500"`
	ImageTag           *string `json:"imageTag" binding:"omitempty,max=128"`
	Environment        *string `json:"environment" binding:"omitempty,max=32"`
	Namespace          *string `json:"namespace" binding:"omitempty,max=63"`
	ReplicaCount       *int    `json:"replicaCount" binding:"omitempty"`
	DeploymentStrategy *string `json:"deploymentStrategy" binding:"omitempty,max=64"`
}

type RollbackDeploymentRequest struct {
	Revision int `json:"revision" binding:"required"`
}

type DeploymentResponse struct {
	ID                 string     `json:"id"`
	ApplicationID      string     `json:"applicationId"`
	ProjectID          string     `json:"projectId"`
	OrganizationID     string     `json:"organizationId"`
	Image              string     `json:"image"`
	ImageTag           string     `json:"imageTag"`
	Environment        string     `json:"environment"`
	Namespace          string     `json:"namespace"`
	ReplicaCount       int        `json:"replicaCount"`
	Status             string     `json:"status"`
	DeploymentStrategy string     `json:"deploymentStrategy"`
	TargetClusterID    string     `json:"targetClusterId"`
	CreatedBy          uint       `json:"createdBy"`
	UpdatedBy          uint       `json:"updatedBy"`
	StartedAt          *time.Time `json:"startedAt,omitempty"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type DeploymentListResponse struct {
	Items      []DeploymentResponse `json:"items"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	Total      int64                `json:"total"`
	TotalPages int                  `json:"totalPages"`
}

type RollbackDeploymentResponse struct {
	Deployment             DeploymentResponse `json:"deployment"`
	CurrentRevision        int                `json:"currentRevision"`
	RollbackSourceRevision int                `json:"rollbackSourceRevision"`
}
