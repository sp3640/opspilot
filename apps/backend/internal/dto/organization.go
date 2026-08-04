package dto

import "time"

// CreateOrganizationRequest is the validated input for organization creation.
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Slug        string `json:"slug" binding:"omitempty,max=120"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateOrganizationRequest is the validated input for organization update.
type UpdateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Slug        string `json:"slug" binding:"omitempty,max=120"`
	Description string `json:"description" binding:"max=500"`
}

// OrganizationResponse is the canonical API representation of an organization.
type OrganizationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	OwnerID     uint      `json:"ownerId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// OrganizationListResponse is the paginated organization list response.
type OrganizationListResponse struct {
	Items      []OrganizationResponse `json:"items"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	Total      int64                  `json:"total"`
	TotalPages int                    `json:"totalPages"`
}
