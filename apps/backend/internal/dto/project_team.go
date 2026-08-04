package dto

import "time"

type AssignTeamRequest struct {
	TeamID string `json:"teamId" binding:"required"`
}

type ProjectTeamResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	TeamID         string    `json:"teamId"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ProjectTeamListResponse struct {
	Items []ProjectTeamResponse `json:"items"`
	Total int64                 `json:"total"`
}
