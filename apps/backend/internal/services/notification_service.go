package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/notification"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
)

// channelConfig is the JSON shape encrypted into
// NotificationChannel.EncryptedConfig. Which field is populated depends on
// the channel's Type.
type channelConfig struct {
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// NotificationService owns every notification channel's lifecycle (CRUD,
// always encrypting the target before it touches the database) and is the
// single place that decides which channels an event fans out to
// (Dispatch). Providers themselves (internal/notification) know nothing
// about channels, encryption, or events - they only know how to send one
// Message to one already-resolved target.
type NotificationService struct {
	repo           *repository.NotificationChannelRepository
	cipher         security.ClusterCredentialCipher
	providers      map[string]notification.Provider
	deploymentRepo *repository.DeploymentRepository
	teamRepo       *repository.TeamRepository
	auditService   *AuditService
}

func NewNotificationService(
	repo *repository.NotificationChannelRepository,
	cipher security.ClusterCredentialCipher,
	providers map[string]notification.Provider,
) *NotificationService {
	return &NotificationService{repo: repo, cipher: cipher, providers: providers}
}

// WithDeploymentRepo enables the "previous deployment" lookup that backs
// DEPLOYMENT_RECOVERED detection. Without it, recovery notifications are
// simply never sent (no fabricated "recovered" event without real evidence
// of a prior failure).
func (s *NotificationService) WithDeploymentRepo(deploymentRepo *repository.DeploymentRepository) *NotificationService {
	s.deploymentRepo = deploymentRepo
	return s
}

// WithTeamRepo enables validating that a channel's TeamID belongs to the
// creating organization. Without it, team-scoped channels cannot be created.
func (s *NotificationService) WithTeamRepo(teamRepo *repository.TeamRepository) *NotificationService {
	s.teamRepo = teamRepo
	return s
}

func (s *NotificationService) WithAuditService(auditService *AuditService) *NotificationService {
	s.auditService = auditService
	return s
}

// ─── CRUD ──────────────────────────────────────────────────────────────────

func (s *NotificationService) CreateChannel(userID uint, organizationID uuid.UUID, req dto.CreateNotificationChannelRequest) (*dto.NotificationChannelResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrInvalidNotificationChannelName
	}
	if !constants.IsValidNotificationChannelType(req.Type) {
		return nil, apperrors.ErrInvalidNotificationChannelType
	}

	events, err := normalizeAndValidateEvents(req.Events)
	if err != nil {
		return nil, err
	}

	cfg, err := buildChannelConfig(req.Type, req.Target)
	if err != nil {
		return nil, err
	}

	var teamID *uuid.UUID
	if req.TeamID != nil && *req.TeamID != uuid.Nil {
		if err := s.validateTeam(*req.TeamID, organizationID); err != nil {
			return nil, err
		}
		teamID = req.TeamID
	}

	encryptedConfig, err := s.encryptConfig(cfg)
	if err != nil {
		return nil, err
	}

	channel := &models.NotificationChannel{
		OrganizationID:  organizationID,
		TeamID:          teamID,
		Name:            name,
		Type:            req.Type,
		EncryptedConfig: encryptedConfig,
		Events:          marshalNotificationEvents(events),
		Enabled:         true,
		CreatedBy:       userID,
	}
	if err := s.repo.Create(channel); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         userID,
			OrganizationID: organizationID,
			EntityType:     "notification_channel",
			EntityID:       channel.ID.String(),
			Action:         models.AuditActionCreate,
			AfterState:     marshalAuditState(map[string]any{"name": name, "type": req.Type, "events": events, "enabled": true}),
		})
	}

	response := s.toResponse(*channel, cfg)
	return &response, nil
}

func (s *NotificationService) ListChannels(organizationID uuid.UUID, req *models.PaginationRequest, teamID uuid.UUID, channelType string) (*dto.NotificationChannelListResponse, error) {
	items, total, err := s.repo.List(req, organizationID, teamID, channelType)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.NotificationChannelResponse, 0, len(items))
	for _, channel := range items {
		responses = append(responses, s.toMaskedResponse(channel))
	}

	return &dto.NotificationChannelListResponse{
		Items:      responses,
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
	}, nil
}

func (s *NotificationService) GetChannel(organizationID, channelID uuid.UUID) (*dto.NotificationChannelResponse, error) {
	channel, err := s.getOwnedChannel(channelID, organizationID)
	if err != nil {
		return nil, err
	}
	response := s.toMaskedResponse(*channel)
	return &response, nil
}

func (s *NotificationService) UpdateChannel(userID uint, organizationID, channelID uuid.UUID, req dto.UpdateNotificationChannelRequest) (*dto.NotificationChannelResponse, error) {
	channel, err := s.getOwnedChannel(channelID, organizationID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrInvalidNotificationChannelName
	}
	events, err := normalizeAndValidateEvents(req.Events)
	if err != nil {
		return nil, err
	}

	previousName := channel.Name
	previousEnabled := channel.Enabled
	targetRotated := false

	channel.Name = name
	channel.Events = marshalNotificationEvents(events)
	channel.Enabled = req.Enabled

	if req.Target != nil && strings.TrimSpace(*req.Target) != "" {
		cfg, err := buildChannelConfig(channel.Type, *req.Target)
		if err != nil {
			return nil, err
		}
		encryptedConfig, err := s.encryptConfig(cfg)
		if err != nil {
			return nil, err
		}
		channel.EncryptedConfig = encryptedConfig
		targetRotated = true
	}

	if err := s.repo.Update(channel); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		entityID := channel.ID.String()
		if previousName != name {
			_ = s.auditService.LogUpdate(userID, organizationID, "notification_channel", entityID, nil, nil, "name", previousName, name)
		}
		if previousEnabled != req.Enabled {
			_ = s.auditService.LogUpdate(userID, organizationID, "notification_channel", entityID, nil, nil, "enabled", fmt.Sprintf("%t", previousEnabled), fmt.Sprintf("%t", req.Enabled))
		}
		if targetRotated {
			// The actual target is never written to an audit record - only
			// the fact that it changed, mirroring how cluster credential
			// rotation is audited elsewhere in this codebase.
			_ = s.auditService.LogUpdate(userID, organizationID, "notification_channel", entityID, nil, nil, "target", "rotated", "rotated")
		}
	}

	response := s.toMaskedResponse(*channel)
	return &response, nil
}

func (s *NotificationService) DeleteChannel(userID uint, organizationID, channelID uuid.UUID) error {
	channel, err := s.getOwnedChannel(channelID, organizationID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(channel.ID, organizationID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogDelete(userID, organizationID, "notification_channel", channel.ID.String(), nil, nil)
	}

	return nil
}

// SendTestNotification sends a real, immediate test message and, unlike
// Dispatch, surfaces any failure to the caller - a user explicitly testing
// their configuration needs the honest result, not a best-effort swallow.
func (s *NotificationService) SendTestNotification(ctx context.Context, organizationID, channelID uuid.UUID) error {
	channel, err := s.getOwnedChannel(channelID, organizationID)
	if err != nil {
		return err
	}

	msg := notification.Message{
		Event: "TEST",
		Title: "OpsPilot test notification",
		Body:  "This is a test notification confirming that this channel is configured correctly.",
		Fields: map[string]string{
			"Channel": channel.Name,
		},
	}
	return s.sendToChannel(ctx, *channel, msg)
}

// ─── Dispatch ──────────────────────────────────────────────────────────────

// Dispatch fans msg out, best-effort, to every enabled channel subscribed
// to eventType within organizationID (plus, when teamID is non-nil, every
// enabled channel scoped to that team). It never returns an error and never
// blocks the caller's own operation on a delivery failure - each attempt is
// logged (and, when auditService is configured, audited with its
// success/failure Result) independently.
func (s *NotificationService) Dispatch(ctx context.Context, organizationID uuid.UUID, teamID *uuid.UUID, eventType string, msg notification.Message, actorUserID uint) {
	channels, err := s.repo.ListEnabledForDispatch(organizationID, teamID)
	if err != nil {
		logger.Error(ctx, "notification dispatch: failed to list channels", slog.String("event", eventType), slog.Any("error", err))
		return
	}

	for _, channel := range channels {
		if !channelSubscribed(channel.Events, eventType) {
			continue
		}

		result := models.AuditResultSuccess
		if err := s.sendToChannel(ctx, channel, msg); err != nil {
			result = models.AuditResultFailure
			logger.Error(ctx, "notification dispatch failed",
				slog.String("channel_id", channel.ID.String()),
				slog.String("channel_type", channel.Type),
				slog.String("event", eventType),
				slog.Any("error", err),
			)
		}

		if s.auditService != nil {
			_ = s.auditService.LogEvent(AuditEventInput{
				UserID:         actorUserID,
				OrganizationID: organizationID,
				EntityType:     "notification_dispatch",
				EntityID:       channel.ID.String(),
				Action:         models.AuditActionCreate,
				Result:         result,
				FieldName:      "event",
				NewValue:       eventType,
			})
		}
	}
}

func (s *NotificationService) sendToChannel(ctx context.Context, channel models.NotificationChannel, msg notification.Message) error {
	if !channel.Enabled {
		return fmt.Errorf("channel is disabled")
	}

	provider := s.providers[channel.Type]
	if provider == nil {
		return fmt.Errorf("no provider registered for channel type %s", channel.Type)
	}

	decrypted, err := s.cipher.Decrypt(channel.EncryptedConfig)
	if err != nil {
		return fmt.Errorf("decrypt channel config: %w", err)
	}

	var cfg channelConfig
	if err := json.Unmarshal([]byte(decrypted), &cfg); err != nil {
		return fmt.Errorf("parse channel config: %w", err)
	}

	return provider.Send(ctx, extractTarget(channel.Type, cfg), msg)
}

// ─── Event-specific helpers (called by AlertService/IncidentService/DeploymentStatusUpdater) ──

func (s *NotificationService) NotifyCriticalAlert(ctx context.Context, alert *models.Alert, userID uint) {
	msg := notification.Message{
		Event:    constants.NotificationEventCriticalAlert,
		Title:    fmt.Sprintf("Critical alert: %s", alert.Title),
		Body:     alert.Description,
		Severity: alert.Severity,
		Fields: map[string]string{
			"Resource": alert.ResourceType + "/" + alert.ResourceID,
			"Source":   alert.Source,
			"Project":  alert.ProjectID.String(),
		},
	}
	s.Dispatch(ctx, alert.OrganizationID, nil, constants.NotificationEventCriticalAlert, msg, userID)
}

func (s *NotificationService) NotifySev1Incident(ctx context.Context, incident *models.Incident, userID uint) {
	msg := notification.Message{
		Event:    constants.NotificationEventSev1Incident,
		Title:    fmt.Sprintf("SEV-1 incident: %s", incident.Title),
		Body:     incident.Description,
		Severity: incident.Severity,
		Fields: map[string]string{
			"Project": incident.ProjectID.String(),
			"Status":  incident.Status,
		},
	}
	s.Dispatch(ctx, incident.OrganizationID, incident.OwnerTeamID, constants.NotificationEventSev1Incident, msg, userID)
}

func (s *NotificationService) NotifyIncidentAssigned(ctx context.Context, incident *models.Incident, assigneeUserID, actorUserID uint) {
	msg := notification.Message{
		Event: constants.NotificationEventIncidentAssigned,
		Title: fmt.Sprintf("Incident assigned: %s", incident.Title),
		Body:  fmt.Sprintf("Incident #%d was assigned to user %d.", incident.ID, assigneeUserID),
		Fields: map[string]string{
			"Project":  incident.ProjectID.String(),
			"Severity": incident.Severity,
		},
	}
	s.Dispatch(ctx, incident.OrganizationID, incident.OwnerTeamID, constants.NotificationEventIncidentAssigned, msg, actorUserID)
}

func (s *NotificationService) NotifyDeploymentFailed(ctx context.Context, deployment *models.Deployment, userID uint, reason string) {
	body := "Deployment failed."
	if reason != "" {
		body = "Deployment failed: " + reason
	}
	msg := notification.Message{
		Event: constants.NotificationEventDeploymentFailed,
		Title: fmt.Sprintf("Deployment failed: %s", deployment.Image+":"+deployment.ImageTag),
		Body:  body,
		Fields: map[string]string{
			"Environment": deployment.Environment,
			"Namespace":   deployment.Namespace,
			"Application": deployment.ApplicationID.String(),
		},
	}
	s.Dispatch(ctx, deployment.OrganizationID, nil, constants.NotificationEventDeploymentFailed, msg, userID)
}

// NotifyDeploymentOutcomeSucceeded checks whether this successful deployment
// represents a recovery (the immediately preceding deployment attempt for
// the same application had failed or was rolled back) and, only then, fires
// DEPLOYMENT_RECOVERED. A normal Succeeded-to-Succeeded deploy is not a
// recovery and never triggers a notification. Requires WithDeploymentRepo;
// without it, no recovery detection is possible and nothing is sent.
func (s *NotificationService) NotifyDeploymentOutcomeSucceeded(ctx context.Context, deployment *models.Deployment, userID uint) {
	if s.deploymentRepo == nil {
		return
	}

	previous, err := s.deploymentRepo.GetPreviousDeployment(deployment.ApplicationID, deployment.OrganizationID, deployment.ID, deployment.CreatedAt)
	if err != nil {
		return
	}
	if previous.Status != constants.DeploymentStatusFailed && previous.Status != constants.DeploymentStatusRolledBack {
		return
	}

	msg := notification.Message{
		Event: constants.NotificationEventDeploymentRecovered,
		Title: fmt.Sprintf("Deployment recovered: %s", deployment.Image+":"+deployment.ImageTag),
		Body:  "The previous deployment attempt for this application had failed; this deployment succeeded.",
		Fields: map[string]string{
			"Environment": deployment.Environment,
			"Namespace":   deployment.Namespace,
			"Application": deployment.ApplicationID.String(),
		},
	}
	s.Dispatch(ctx, deployment.OrganizationID, nil, constants.NotificationEventDeploymentRecovered, msg, userID)
}

// ─── helpers ────────────────────────────────────────────────────────────────

func (s *NotificationService) getOwnedChannel(id, organizationID uuid.UUID) (*models.NotificationChannel, error) {
	channel, err := s.repo.GetByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotificationChannelNotFound
		}
		return nil, err
	}
	return channel, nil
}

func (s *NotificationService) validateTeam(teamID, organizationID uuid.UUID) error {
	if s.teamRepo == nil {
		return apperrors.ErrNotificationChannelTeamMismatch
	}
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotificationChannelTeamMismatch
		}
		return err
	}
	if team.OrganizationID != organizationID {
		return apperrors.ErrNotificationChannelTeamMismatch
	}
	return nil
}

func (s *NotificationService) encryptConfig(cfg channelConfig) (string, error) {
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return s.cipher.Encrypt(string(encoded))
}

// toResponse builds a response from a config the caller already has in hand
// (avoids a redundant decrypt immediately after encrypting it on create).
func (s *NotificationService) toResponse(channel models.NotificationChannel, cfg channelConfig) dto.NotificationChannelResponse {
	var teamID *string
	if channel.TeamID != nil {
		v := channel.TeamID.String()
		teamID = &v
	}

	return dto.NotificationChannelResponse{
		ID:             channel.ID.String(),
		OrganizationID: channel.OrganizationID.String(),
		TeamID:         teamID,
		Name:           channel.Name,
		Type:           channel.Type,
		MaskedTarget:   maskTarget(channel.Type, extractTarget(channel.Type, cfg)),
		Events:         unmarshalNotificationEvents(channel.Events),
		Enabled:        channel.Enabled,
		CreatedAt:      channel.CreatedAt,
		UpdatedAt:      channel.UpdatedAt,
	}
}

// toMaskedResponse decrypts only enough to compute a masked target hint -
// the decrypted value itself never leaves this function.
func (s *NotificationService) toMaskedResponse(channel models.NotificationChannel) dto.NotificationChannelResponse {
	maskedTarget := "••••"
	if decrypted, err := s.cipher.Decrypt(channel.EncryptedConfig); err == nil {
		var cfg channelConfig
		if json.Unmarshal([]byte(decrypted), &cfg) == nil {
			maskedTarget = maskTarget(channel.Type, extractTarget(channel.Type, cfg))
		}
	}

	var teamID *string
	if channel.TeamID != nil {
		v := channel.TeamID.String()
		teamID = &v
	}

	return dto.NotificationChannelResponse{
		ID:             channel.ID.String(),
		OrganizationID: channel.OrganizationID.String(),
		TeamID:         teamID,
		Name:           channel.Name,
		Type:           channel.Type,
		MaskedTarget:   maskedTarget,
		Events:         unmarshalNotificationEvents(channel.Events),
		Enabled:        channel.Enabled,
		CreatedAt:      channel.CreatedAt,
		UpdatedAt:      channel.UpdatedAt,
	}
}

func buildChannelConfig(channelType, target string) (channelConfig, error) {
	target = strings.TrimSpace(target)

	switch channelType {
	case constants.NotificationChannelEmail:
		if !strings.Contains(target, "@") {
			return channelConfig{}, apperrors.ErrInvalidNotificationTarget
		}
		return channelConfig{Email: target}, nil

	case constants.NotificationChannelSlack, constants.NotificationChannelTeams, constants.NotificationChannelWebhook:
		parsed, err := url.ParseRequestURI(target)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return channelConfig{}, apperrors.ErrInvalidNotificationTarget
		}
		return channelConfig{URL: target}, nil

	default:
		return channelConfig{}, apperrors.ErrInvalidNotificationChannelType
	}
}

func extractTarget(channelType string, cfg channelConfig) string {
	if channelType == constants.NotificationChannelEmail {
		return cfg.Email
	}
	return cfg.URL
}

func maskTarget(channelType, target string) string {
	if channelType == constants.NotificationChannelEmail {
		return maskEmail(target)
	}
	return maskURL(target)
}

func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return "••••"
	}
	local := email[:at]
	domain := email[at:]
	if len(local) <= 1 {
		return "•" + domain
	}
	return local[:1] + strings.Repeat("•", len(local)-1) + domain
}

func maskURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "••••"
	}
	return parsed.Scheme + "://" + parsed.Host + "/••••"
}

func normalizeAndValidateEvents(events []string) ([]string, error) {
	seen := make(map[string]bool, len(events))
	normalized := make([]string, 0, len(events))
	for _, event := range events {
		event = strings.ToUpper(strings.TrimSpace(event))
		if !constants.IsValidNotificationEventType(event) {
			return nil, apperrors.ErrInvalidNotificationEvent
		}
		if seen[event] {
			continue
		}
		seen[event] = true
		normalized = append(normalized, event)
	}
	if len(normalized) == 0 {
		return nil, apperrors.ErrInvalidNotificationEvent
	}
	return normalized, nil
}

func marshalNotificationEvents(events []string) string {
	if len(events) == 0 {
		return "[]"
	}
	encoded, err := json.Marshal(events)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func unmarshalNotificationEvents(raw string) []string {
	var events []string
	if err := json.Unmarshal([]byte(raw), &events); err != nil {
		return []string{}
	}
	return events
}

func channelSubscribed(eventsJSON, eventType string) bool {
	for _, event := range unmarshalNotificationEvents(eventsJSON) {
		if event == eventType {
			return true
		}
	}
	return false
}
