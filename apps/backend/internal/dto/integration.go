package dto

import "time"

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateIntegrationRequest is the validated input for connecting a new
// integration. Metadata is non-secret configuration (workspace name, base
// URL, ...); Credentials are secret fields (tokens, passwords) - encrypted
// immediately by the service layer and never echoed back by any response.
type CreateIntegrationRequest struct {
	Type        string            `json:"type" binding:"required"`
	Name        string            `json:"name" binding:"required,min=1,max=150"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Credentials map[string]string `json:"credentials,omitempty"`
}

// UpdateIntegrationRequest updates name/metadata and, only if provided,
// rotates the stored credentials. Omitting Credentials leaves the existing
// (encrypted) value untouched - the submitted secret is never required on
// every update, and never returned either way.
type UpdateIntegrationRequest struct {
	Name        string            `json:"name" binding:"required,min=1,max=150"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Credentials map[string]string `json:"credentials,omitempty"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// IntegrationResponse is the only shape ever returned for an Integration -
// it never serializes the model directly. It deliberately has no field for
// credentials (encrypted or decrypted) at all, and HasCredentials is a
// boolean, never a hint about their content.
type IntegrationResponse struct {
	ID                   string            `json:"id"`
	OrganizationID       string            `json:"organizationId"`
	Type                 string            `json:"type"`
	Name                 string            `json:"name"`
	Status               string            `json:"status"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	HasCredentials       bool              `json:"hasCredentials"`
	ConnectorImplemented bool              `json:"connectorImplemented"`
	ConnectorDescription string            `json:"connectorDescription,omitempty"`
	LastCheckedAt        *time.Time        `json:"lastCheckedAt,omitempty"`
	LastError            string            `json:"lastError,omitempty"`
	CreatedAt            time.Time         `json:"createdAt"`
	UpdatedAt            time.Time         `json:"updatedAt"`
}

type IntegrationListResponse struct {
	Items      []IntegrationResponse `json:"items"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"totalPages"`
}

// IntegrationCheckResponse is returned by both /test and /check - the
// immediate outcome of that one attempt, plus the integration's resulting
// state so the frontend need not make a second request to refresh it.
type IntegrationCheckResponse struct {
	Success     bool                `json:"success"`
	Message     string              `json:"message"`
	CheckedAt   time.Time           `json:"checkedAt"`
	Integration IntegrationResponse `json:"integration"`
}
