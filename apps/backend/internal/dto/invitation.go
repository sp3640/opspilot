package dto

import "time"

type InviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}

type InvitationResponse struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Token          string     `json:"token"`
	Status         string     `json:"status"`
	InvitedBy      uint       `json:"invitedBy"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type InvitationListResponse struct {
	Items      []InvitationResponse `json:"items"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	Total      int64                `json:"total"`
	TotalPages int                  `json:"totalPages"`
}
