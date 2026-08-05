package dto

import "time"

type DeploymentHistoryResponse struct {
	ID                 string    `json:"id"`
	DeploymentID       string    `json:"deploymentId"`
	ApplicationID      string    `json:"applicationId"`
	ProjectID          string    `json:"projectId"`
	OrganizationID     string    `json:"organizationId"`
	Revision           int       `json:"revision"`
	Image              string    `json:"image"`
	ImageTag           string    `json:"imageTag"`
	Environment        string    `json:"environment"`
	Namespace          string    `json:"namespace"`
	ReplicaCount       int       `json:"replicaCount"`
	DeploymentStrategy string    `json:"deploymentStrategy"`
	Status             string    `json:"status"`
	ChangeSummary      string    `json:"changeSummary"`
	TriggeredBy        uint      `json:"triggeredBy"`
	CreatedAt          time.Time `json:"createdAt"`
}

type DeploymentHistoryListResponse struct {
	Items      []DeploymentHistoryResponse `json:"items"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	Total      int64                       `json:"total"`
	TotalPages int                         `json:"totalPages"`
}
