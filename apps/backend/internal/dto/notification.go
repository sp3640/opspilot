package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateNotificationChannelRequest is the validated input for creating a
// notification channel. Target's meaning depends on Type: an email address
// for EMAIL, a webhook URL for SLACK/TEAMS/WEBHOOK. It is never echoed back
// by the API after creation - see NotificationChannelResponse.MaskedTarget.
type CreateNotificationChannelRequest struct {
	Name   string     `json:"name" binding:"required,min=1,max=150"`
	Type   string     `json:"type" binding:"required"`
	Target string     `json:"target" binding:"required"`
	TeamID *uuid.UUID `json:"team_id,omitempty"`
	Events []string   `json:"events" binding:"required,min=1"`
}

// UpdateNotificationChannelRequest updates a channel's name/events/enabled
// state, and optionally rotates its target. Target is a pointer so omitting
// it leaves the existing (encrypted) target untouched.
type UpdateNotificationChannelRequest struct {
	Name    string   `json:"name" binding:"required,min=1,max=150"`
	Target  *string  `json:"target,omitempty"`
	Events  []string `json:"events" binding:"required,min=1"`
	Enabled bool     `json:"enabled"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// NotificationChannelResponse never includes the decrypted target - only a
// masked hint sufficient for a human to recognize which channel they
// configured (e.g. "j***@example.com" or "https://hooks.slack.com/***").
type NotificationChannelResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	TeamID         *string   `json:"teamId,omitempty"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	MaskedTarget   string    `json:"maskedTarget"`
	Events         []string  `json:"events"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type NotificationChannelListResponse struct {
	Items      []NotificationChannelResponse `json:"items"`
	Page       int                           `json:"page"`
	Limit      int                           `json:"limit"`
	Total      int64                         `json:"total"`
	TotalPages int                           `json:"totalPages"`
}
