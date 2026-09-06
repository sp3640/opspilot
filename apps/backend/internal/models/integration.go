package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Integration represents one organization's connection to an external
// system (GitHub, Slack, Prometheus, ...). This sprint (27) only builds the
// persistence/lifecycle/security scaffolding - no real connector exists yet
// for any Type (see internal/connector), so every integration created today
// necessarily starts and stays in a not-yet-connected state until a future
// sprint implements its connector.
type Integration struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_integrations_organization_id"`
	// Type is one of constants.IntegrationType* - validated, never a
	// free-form string chosen by a caller.
	Type string `json:"type" gorm:"size:32;not null;index:idx_integrations_type"`
	Name string `json:"name" gorm:"size:150;not null"`
	// Status is one of constants.IntegrationStatus* - the integration
	// record's own lifecycle state, distinct from external-service health.
	Status string `json:"status" gorm:"size:20;not null;default:'PENDING';index:idx_integrations_status"`
	// ConfigMetadata is a JSON object of non-secret configuration (e.g. a
	// workspace name, base URL, repository slug) - safe to return via the
	// API and to write into audit snapshots as-is.
	ConfigMetadata string `json:"-" gorm:"type:text;not null;default:'{}'"`
	// EncryptedCredentials is an AES-256-GCM ciphertext (see
	// security.ClusterCredentialCipher) of a JSON object of secret fields
	// (tokens, passwords, client secrets). Never decrypted for display and
	// never included in any API response, audit snapshot, or log line -
	// see IntegrationService for the only code paths allowed to decrypt it.
	EncryptedCredentials string `json:"-" gorm:"type:text;not null;default:''"`
	// LastCheckedAt/LastError record the most recent TestConnection or
	// HealthCheck attempt - both are best-effort observability, never
	// authoritative proof the external service is currently reachable.
	LastCheckedAt *time.Time     `json:"last_checked_at,omitempty"`
	LastError     string         `json:"last_error" gorm:"type:text;not null;default:''"`
	CreatedBy     uint           `json:"created_by" gorm:"not null"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (i *Integration) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}

	return nil
}
