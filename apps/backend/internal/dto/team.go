package dto

import "time"

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type TeamMemberRequest struct {
	UserID uint `json:"userId" binding:"required"`
}

type TeamResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type TeamListResponse struct {
	Items      []TeamResponse `json:"items"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"totalPages"`
}

type TeamMemberResponse struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"teamId"`
	UserID    uint      `json:"userId"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type TeamMemberListResponse struct {
	Items []TeamMemberResponse `json:"items"`
	Total int64                `json:"total"`
}
