package dto

import "time"

type AssignApplicationTeamRequest struct {
	TeamID string `json:"teamId" binding:"required"`
}

type ApplicationTeamResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ApplicationID  string    `json:"applicationId"`
	TeamID         string    `json:"teamId"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ApplicationTeamListResponse struct {
	Items []ApplicationTeamResponse `json:"items"`
	Total int64                     `json:"total"`
}
