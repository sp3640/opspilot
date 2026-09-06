package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/connector"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
)

// redactedValue replaces every credential value in an audit snapshot. Keys
// are kept (so an audit reader can see *which* secret fields changed) but a
// value is never written to an audit log under any circumstance.
const redactedValue = "[REDACTED]"

// IntegrationService owns the full lifecycle of an organization's external
// integrations: CRUD, credential encryption, connector resolution, and
// status/last-check bookkeeping. It never returns a decrypted credential
// value to a caller outside this file, and every audit snapshot it builds
// is redacted before it ever reaches AuditService.
type IntegrationService struct {
	repo              *repository.IntegrationRepository
	cipher            security.ClusterCredentialCipher
	connectorRegistry *connector.Registry
	auditService      *AuditService
}

func NewIntegrationService(
	repo *repository.IntegrationRepository,
	cipher security.ClusterCredentialCipher,
	connectorRegistry *connector.Registry,
) *IntegrationService {
	return &IntegrationService{repo: repo, cipher: cipher, connectorRegistry: connectorRegistry}
}

func (s *IntegrationService) WithAuditService(auditService *AuditService) *IntegrationService {
	s.auditService = auditService
	return s
}

// ─── CRUD ──────────────────────────────────────────────────────────────────

func (s *IntegrationService) CreateIntegration(userID uint, organizationID uuid.UUID, req dto.CreateIntegrationRequest) (*dto.IntegrationResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrInvalidIntegrationName
	}
	if !constants.IsValidIntegrationType(req.Type) {
		return nil, apperrors.ErrInvalidIntegrationType
	}

	exists, err := s.repo.ExistsByTypeAndName(organizationID, req.Type, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrIntegrationAlreadyExists
	}

	encryptedCredentials, err := s.encryptCredentials(req.Credentials)
	if err != nil {
		return nil, err
	}

	integration := &models.Integration{
		OrganizationID:       organizationID,
		Type:                 req.Type,
		Name:                 name,
		Status:               constants.IntegrationStatusPending,
		ConfigMetadata:       marshalMetadata(req.Metadata),
		EncryptedCredentials: encryptedCredentials,
		CreatedBy:            userID,
	}
	if err := s.repo.Create(integration); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         userID,
			OrganizationID: organizationID,
			EntityType:     "integration",
			EntityID:       integration.ID.String(),
			Action:         models.AuditActionCreate,
			AfterState: marshalAuditState(map[string]any{
				"type":        req.Type,
				"name":        name,
				"status":      integration.Status,
				"metadata":    req.Metadata,
				"credentials": redactCredentials(req.Credentials),
			}),
		})
	}

	response := s.toResponse(*integration)
	return &response, nil
}

// UpsertConnectedIntegration creates (or, if one of this type already
// exists for the organization, updates the credentials of) an Integration,
// and immediately marks it CONNECTED - the path an OAuth-based connect flow
// uses (see GitHubService.CompleteOAuth) so reconnecting/re-authorizing
// never creates a duplicate integration row the way a second CreateIntegration
// call would. Requires the caller to have already positively verified the
// credentials work (e.g. via a successful identity lookup) before calling this.
func (s *IntegrationService) UpsertConnectedIntegration(userID uint, organizationID uuid.UUID, integrationType, name string, credentials map[string]string) (*dto.IntegrationResponse, error) {
	encryptedCredentials, err := s.encryptCredentials(credentials)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	existing, err := s.repo.GetFirstByType(organizationID, integrationType)
	isCreate := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !isCreate {
		return nil, err
	}

	integration := existing
	action := models.AuditActionUpdate
	if isCreate {
		integration = &models.Integration{
			OrganizationID: organizationID,
			Type:           integrationType,
			ConfigMetadata: "{}",
			CreatedBy:      userID,
		}
		action = models.AuditActionCreate
	}
	integration.Name = name
	integration.Status = constants.IntegrationStatusConnected
	integration.EncryptedCredentials = encryptedCredentials
	integration.LastCheckedAt = &now
	integration.LastError = ""

	if isCreate {
		if err := s.repo.Create(integration); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(integration); err != nil {
			return nil, err
		}
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID: userID, OrganizationID: organizationID, EntityType: "integration", EntityID: integration.ID.String(),
			Action: action, Result: models.AuditResultSuccess,
			FieldName: "credentials", NewValue: "connected via oauth",
			AfterState: marshalAuditState(map[string]any{"type": integrationType, "name": name, "status": integration.Status}),
		})
	}

	response := s.toResponse(*integration)
	return &response, nil
}

func (s *IntegrationService) ListIntegrations(organizationID uuid.UUID, req *models.PaginationRequest, integrationType, status string) (*dto.IntegrationListResponse, error) {
	items, total, err := s.repo.List(req, organizationID, integrationType, status)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.IntegrationResponse, 0, len(items))
	for _, integration := range items {
		responses = append(responses, s.toResponse(integration))
	}

	return &dto.IntegrationListResponse{
		Items:      responses,
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
	}, nil
}

func (s *IntegrationService) GetIntegration(organizationID, integrationID uuid.UUID) (*dto.IntegrationResponse, error) {
	integration, err := s.getOwnedIntegration(integrationID, organizationID)
	if err != nil {
		return nil, err
	}
	response := s.toResponse(*integration)
	return &response, nil
}

func (s *IntegrationService) UpdateIntegration(userID uint, organizationID, integrationID uuid.UUID, req dto.UpdateIntegrationRequest) (*dto.IntegrationResponse, error) {
	integration, err := s.getOwnedIntegration(integrationID, organizationID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrInvalidIntegrationName
	}

	previousName := integration.Name
	previousMetadata := integration.ConfigMetadata
	credentialsRotated := false

	integration.Name = name
	if req.Metadata != nil {
		integration.ConfigMetadata = marshalMetadata(req.Metadata)
	}

	if len(req.Credentials) > 0 {
		encryptedCredentials, err := s.encryptCredentials(req.Credentials)
		if err != nil {
			return nil, err
		}
		integration.EncryptedCredentials = encryptedCredentials
		credentialsRotated = true
	}

	if err := s.repo.Update(integration); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		entityID := integration.ID.String()
		if previousName != name {
			_ = s.auditService.LogUpdate(userID, organizationID, "integration", entityID, nil, nil, "name", previousName, name)
		}
		if req.Metadata != nil && previousMetadata != integration.ConfigMetadata {
			_ = s.auditService.LogEvent(AuditEventInput{
				UserID: userID, OrganizationID: organizationID, EntityType: "integration", EntityID: entityID,
				Action: models.AuditActionUpdate, FieldName: "metadata",
				BeforeState: previousMetadata, AfterState: integration.ConfigMetadata,
			})
		}
		if credentialsRotated {
			// Only the fact of rotation is audited, keyed by which fields
			// changed - values are never written, mirroring how
			// NotificationChannel target rotation is audited.
			_ = s.auditService.LogEvent(AuditEventInput{
				UserID: userID, OrganizationID: organizationID, EntityType: "integration", EntityID: entityID,
				Action: models.AuditActionUpdate, FieldName: "credentials",
				AfterState: marshalAuditState(map[string]any{"credentials": redactCredentials(req.Credentials)}),
			})
		}
	}

	response := s.toResponse(*integration)
	return &response, nil
}

func (s *IntegrationService) DeleteIntegration(userID uint, organizationID, integrationID uuid.UUID) error {
	integration, err := s.getOwnedIntegration(integrationID, organizationID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(integration.ID, organizationID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogDelete(userID, organizationID, "integration", integration.ID.String(), nil, nil)
	}

	return nil
}

// ─── Connection lifecycle ───────────────────────────────────────────────────

// TestConnection validates the integration's stored configuration/
// credentials against its connector's TestConnection operation, updates
// status/LastCheckedAt/LastError from the outcome, and audits both the
// test result and any resulting status change.
func (s *IntegrationService) TestConnection(ctx context.Context, userID uint, organizationID, integrationID uuid.UUID) (*dto.IntegrationCheckResponse, error) {
	return s.runCheck(ctx, userID, organizationID, integrationID, "test", func(c connector.Connector, cfg connector.Config) (*connector.Result, error) {
		return c.TestConnection(ctx, cfg)
	})
}

// CheckStatus runs the connector's lighter-weight HealthCheck operation -
// the "refresh status" action, distinct from a full configuration test.
func (s *IntegrationService) CheckStatus(ctx context.Context, userID uint, organizationID, integrationID uuid.UUID) (*dto.IntegrationCheckResponse, error) {
	return s.runCheck(ctx, userID, organizationID, integrationID, "check", func(c connector.Connector, cfg connector.Config) (*connector.Result, error) {
		return c.HealthCheck(ctx, cfg)
	})
}

func (s *IntegrationService) runCheck(
	ctx context.Context,
	userID uint,
	organizationID, integrationID uuid.UUID,
	auditVerb string,
	operation func(connector.Connector, connector.Config) (*connector.Result, error),
) (*dto.IntegrationCheckResponse, error) {
	integration, err := s.getOwnedIntegration(integrationID, organizationID)
	if err != nil {
		return nil, err
	}

	cfg, err := s.buildConnectorConfig(integration)
	if err != nil {
		return nil, err
	}

	previousStatus := integration.Status
	checkedAt := time.Now().UTC()

	c := s.connectorRegistry.Resolve(integration.Type)
	result, opErr := operation(c, cfg)

	var success bool
	var message string
	switch {
	case opErr != nil:
		message = opErr.Error()
	case result != nil && result.Success:
		success = true
		message = result.Message
	case result != nil:
		message = result.Message
	default:
		message = "connection check did not succeed"
	}

	return s.applyCheckOutcome(integration, userID, organizationID, auditVerb, success, message, &checkedAt, &previousStatus)
}

// applyCheckOutcome persists a connection status/LastCheckedAt/LastError
// update and audits it - the single write path shared by
// TestConnection/CheckStatus (via runCheck) and by RecordConnectionOutcome,
// which type-specific services (e.g. GitHubService) use when one of their
// own operations (repository discovery) also proves whether the stored
// credential currently works.
func (s *IntegrationService) applyCheckOutcome(
	integration *models.Integration,
	userID uint,
	organizationID uuid.UUID,
	auditVerb string,
	success bool,
	message string,
	checkedAtOverride *time.Time,
	previousStatusOverride *string,
) (*dto.IntegrationCheckResponse, error) {
	checkedAt := time.Now().UTC()
	if checkedAtOverride != nil {
		checkedAt = *checkedAtOverride
	}
	previousStatus := integration.Status
	if previousStatusOverride != nil {
		previousStatus = *previousStatusOverride
	}

	newStatus := constants.IntegrationStatusError
	lastError := message
	if success {
		newStatus = constants.IntegrationStatusConnected
		lastError = ""
	}

	if err := s.repo.UpdateCheckResult(integration.ID, organizationID, newStatus, checkedAt, lastError); err != nil {
		return nil, err
	}
	integration.Status = newStatus
	integration.LastCheckedAt = &checkedAt
	integration.LastError = lastError

	if s.auditService != nil {
		auditResult := models.AuditResultSuccess
		if !success {
			auditResult = models.AuditResultFailure
		}
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID: userID, OrganizationID: organizationID, EntityType: "integration", EntityID: integration.ID.String(),
			Action: models.AuditActionUpdate, Result: auditResult,
			FieldName: "connection_" + auditVerb, NewValue: message,
		})
		if previousStatus != newStatus {
			_ = s.auditService.LogUpdate(userID, organizationID, "integration", integration.ID.String(), nil, nil, "status", previousStatus, newStatus)
		}
	}

	response := s.toResponse(*integration)
	return &dto.IntegrationCheckResponse{
		Success:     success,
		Message:     message,
		CheckedAt:   checkedAt,
		Integration: response,
	}, nil
}

// ResolveConnectorConfig loads integration (organization-scoped, any type)
// and decrypts its credentials into a connector.Config - the one path any
// type-specific service (e.g. GitHubService) uses to get provider-ready
// credentials without duplicating this service's encryption logic.
func (s *IntegrationService) ResolveConnectorConfig(organizationID, integrationID uuid.UUID) (*models.Integration, connector.Config, error) {
	integration, err := s.getOwnedIntegration(integrationID, organizationID)
	if err != nil {
		return nil, connector.Config{}, err
	}
	cfg, err := s.buildConnectorConfig(integration)
	if err != nil {
		return nil, connector.Config{}, err
	}
	return integration, cfg, nil
}

// RecordConnectionOutcome persists a connection status/lastError update and
// audits it, exactly like TestConnection/CheckStatus do - used by
// type-specific services whose own operations (e.g. GitHubService's
// repository discovery) also prove whether the stored credential currently
// works, so that path updates the same status/audit trail rather than a
// second, parallel one.
func (s *IntegrationService) RecordConnectionOutcome(userID uint, organizationID uuid.UUID, integration *models.Integration, auditVerb string, success bool, message string) (*dto.IntegrationResponse, error) {
	result, err := s.applyCheckOutcome(integration, userID, organizationID, auditVerb, success, message, nil, nil)
	if err != nil {
		return nil, err
	}
	return &result.Integration, nil
}

// ─── helpers ────────────────────────────────────────────────────────────────

func (s *IntegrationService) getOwnedIntegration(id, organizationID uuid.UUID) (*models.Integration, error) {
	integration, err := s.repo.GetByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIntegrationNotFound
		}
		return nil, err
	}
	return integration, nil
}

// encryptCredentials fails safely: if no credentials were supplied, it
// returns an empty ciphertext (nothing to encrypt); if credentials were
// supplied but no cipher is configured, it returns
// ErrIntegrationEncryptionUnavailable rather than ever storing a secret in
// plaintext.
func (s *IntegrationService) encryptCredentials(credentials map[string]string) (string, error) {
	if len(credentials) == 0 {
		return "", nil
	}
	if s.cipher == nil {
		return "", apperrors.ErrIntegrationEncryptionUnavailable
	}

	encoded, err := json.Marshal(credentials)
	if err != nil {
		return "", err
	}
	return s.cipher.Encrypt(string(encoded))
}

// decryptCredentials is the only place EncryptedCredentials is ever turned
// back into a plaintext map - the result must never be attached to a
// response DTO, an audit snapshot, or a log statement.
func (s *IntegrationService) decryptCredentials(integration *models.Integration) (map[string]string, error) {
	if integration.EncryptedCredentials == "" {
		return map[string]string{}, nil
	}
	if s.cipher == nil {
		return nil, apperrors.ErrIntegrationEncryptionUnavailable
	}

	decrypted, err := s.cipher.Decrypt(integration.EncryptedCredentials)
	if err != nil {
		return nil, err
	}

	var credentials map[string]string
	if err := json.Unmarshal([]byte(decrypted), &credentials); err != nil {
		return nil, err
	}
	return credentials, nil
}

func (s *IntegrationService) buildConnectorConfig(integration *models.Integration) (connector.Config, error) {
	credentials, err := s.decryptCredentials(integration)
	if err != nil {
		return connector.Config{}, err
	}
	return connector.Config{
		Metadata:    unmarshalMetadata(integration.ConfigMetadata),
		Credentials: credentials,
	}, nil
}

func (s *IntegrationService) toResponse(integration models.Integration) dto.IntegrationResponse {
	capabilities := s.connectorRegistry.Resolve(integration.Type).Capabilities()

	return dto.IntegrationResponse{
		ID:                   integration.ID.String(),
		OrganizationID:       integration.OrganizationID.String(),
		Type:                 integration.Type,
		Name:                 integration.Name,
		Status:               integration.Status,
		Metadata:             unmarshalMetadata(integration.ConfigMetadata),
		HasCredentials:       integration.EncryptedCredentials != "",
		ConnectorImplemented: capabilities.Implemented,
		ConnectorDescription: capabilities.Description,
		LastCheckedAt:        integration.LastCheckedAt,
		LastError:            integration.LastError,
		CreatedAt:            integration.CreatedAt,
		UpdatedAt:            integration.UpdatedAt,
	}
}

func marshalMetadata(metadata map[string]string) string {
	if len(metadata) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func unmarshalMetadata(raw string) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	var metadata map[string]string
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return map[string]string{}
	}
	return metadata
}

// redactCredentials preserves keys (so an audit reader can see which secret
// fields were set/changed) but replaces every value with a fixed
// placeholder - the only form a credential is ever allowed to take inside
// an audit snapshot.
func redactCredentials(credentials map[string]string) map[string]string {
	if len(credentials) == 0 {
		return nil
	}
	redacted := make(map[string]string, len(credentials))
	for key := range credentials {
		redacted[key] = redactedValue
	}
	return redacted
}
