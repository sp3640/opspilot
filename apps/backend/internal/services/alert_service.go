package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/alerting"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type AlertService struct {
	repo         *repository.AlertRepository
	incidentRepo *repository.IncidentRepository
	auditRepo    *AuditService
}

func NewAlertService(repo *repository.AlertRepository, incidentRepo *repository.IncidentRepository, auditService *AuditService) *AlertService {
	return &AlertService{
		repo:         repo,
		incidentRepo: incidentRepo,
		auditRepo:    auditService,
	}
}

func GenerateFingerprint(projectID uuid.UUID, resourceType, resourceID, severity, title string) string {
	canonical := strings.Join([]string{
		projectID.String(),
		resourceType,
		resourceID,
		severity,
		title,
	}, "|")

	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])
}

func (s *AlertService) CreateAlert(
	ctx context.Context,
	projectID uuid.UUID,
	incidentID *uint,
	title,
	description,
	severity,
	status,
	source,
	resourceType,
	resourceID string,
	labels,
	metadata json.RawMessage,
	firstSeenAt,
	lastSeenAt *time.Time,
	userID uint,
	organizationID uuid.UUID,
) (*dto.AlertResponse, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	resourceID = strings.TrimSpace(resourceID)

	if err := validateAlertInput(severity, status, source, resourceType); err != nil {
		return nil, err
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(projectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	fingerprint := GenerateFingerprint(projectID, resourceType, resourceID, severity, title)

	existing, err := s.repo.FindByFingerprint(projectID, organizationID, fingerprint)
	if err == nil {
		now := time.Now()
		updatedMetadata := existing.Metadata
		if len(metadata) > 0 {
			updatedMetadata = metadata
		}

		previousCount := existing.OccurrenceCount
		if err := s.repo.IncrementOccurrence(existing.ID, now, updatedMetadata); err != nil {
			return nil, err
		}

		existing.OccurrenceCount = previousCount + 1
		existing.LastSeenAt = now
		if len(metadata) > 0 {
			existing.Metadata = metadata
		}

		if s.auditRepo != nil {
			entityID := strconv.FormatUint(uint64(existing.ID), 10)
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &existing.ProjectID, existing.IncidentID, "occurrence_count", strconv.Itoa(previousCount), strconv.Itoa(existing.OccurrenceCount)); err != nil {
				logAuditFailure(ctx, "update", "alert", existing.ID, err)
			}
		}

		response := mapper.MapAlert(*existing)
		return &response, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	resolvedLabels := labels
	if len(resolvedLabels) == 0 {
		resolvedLabels = json.RawMessage(`{}`)
	}

	resolvedMetadata := metadata
	if len(resolvedMetadata) == 0 {
		resolvedMetadata = json.RawMessage(`{}`)
	}

	firstSeen := time.Now()
	if firstSeenAt != nil && !firstSeenAt.IsZero() {
		firstSeen = *firstSeenAt
	}

	lastSeen := firstSeen
	if lastSeenAt != nil && !lastSeenAt.IsZero() {
		lastSeen = *lastSeenAt
	}

	alert := &models.Alert{
		OrganizationID:  organizationID,
		ProjectID:       projectID,
		IncidentID:      incidentID,
		Title:           title,
		Description:     description,
		Severity:        severity,
		Status:          status,
		Source:          source,
		ResourceType:    resourceType,
		ResourceID:      resourceID,
		Fingerprint:     fingerprint,
		OccurrenceCount: 1,
		Labels:          resolvedLabels,
		Metadata:        resolvedMetadata,
		FirstSeenAt:     firstSeen,
		LastSeenAt:      lastSeen,
		CreatedBy:       userID,
	}

	if err := s.repo.Create(alert); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogCreate(userID, organizationID, "alert", strconv.FormatUint(uint64(alert.ID), 10), &alert.ProjectID, alert.IncidentID); err != nil {
			logAuditFailure(ctx, "create", "alert", alert.ID, err)
		}
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

func (s *AlertService) UpdateAlert(
	ctx context.Context,
	id,
	userID uint,
	organizationID uuid.UUID,
	projectID uuid.UUID,
	incidentID *uint,
	title,
	description,
	severity,
	status,
	source,
	resourceType,
	resourceID string,
	labels,
	metadata json.RawMessage,
	firstSeenAt,
	lastSeenAt *time.Time,
	acknowledgedAt,
	resolvedAt *time.Time,
) (*dto.AlertResponse, error) {
	if err := validateAlertInput(severity, status, source, resourceType); err != nil {
		return nil, err
	}

	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return nil, err
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(projectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	if incidentID != nil {
		if err := s.ensureIncidentBelongsToUser(*incidentID, organizationID, projectID); err != nil {
			return nil, err
		}
	}

	previousTitle := alert.Title
	previousDescription := alert.Description
	previousSeverity := alert.Severity
	previousStatus := alert.Status
	previousSource := alert.Source
	previousResourceType := alert.ResourceType
	previousResourceID := alert.ResourceID

	alert.ProjectID = projectID
	alert.IncidentID = incidentID
	alert.Title = strings.TrimSpace(title)
	alert.Description = strings.TrimSpace(description)
	alert.Severity = severity
	alert.Status = status
	alert.Source = source
	alert.ResourceType = resourceType
	alert.ResourceID = strings.TrimSpace(resourceID)
	if len(labels) > 0 {
		alert.Labels = labels
	}
	if len(metadata) > 0 {
		alert.Metadata = metadata
	}
	if firstSeenAt != nil && !firstSeenAt.IsZero() {
		alert.FirstSeenAt = *firstSeenAt
	}
	if lastSeenAt != nil && !lastSeenAt.IsZero() {
		alert.LastSeenAt = *lastSeenAt
	}
	alert.AcknowledgedAt = acknowledgedAt
	alert.ResolvedAt = resolvedAt
	alert.Fingerprint = GenerateFingerprint(alert.ProjectID, alert.ResourceType, alert.ResourceID, alert.Severity, alert.Title)

	if err := s.repo.Update(alert); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		entityID := strconv.FormatUint(uint64(alert.ID), 10)
		if previousTitle != alert.Title {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "title", previousTitle, alert.Title); err != nil {
				logAuditFailure(ctx, "update", "alert", alert.ID, err)
			}
		}
		if previousDescription != alert.Description {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "description", previousDescription, alert.Description); err != nil {
				logAuditFailure(ctx, "update", "alert", alert.ID, err)
			}
		}
		if previousSeverity != alert.Severity {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "severity", previousSeverity, alert.Severity); err != nil {
				logAuditFailure(ctx, "update", "alert", alert.ID, err)
			}
		}
		if previousStatus != alert.Status {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "status", previousStatus, alert.Status); err != nil {
				logAuditFailure(ctx, "update", "alert", alert.ID, err)
			}
		}
		if previousSource != alert.Source {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "source", previousSource, alert.Source); err != nil {
				logAuditFailure(ctx, "update", "alert", alert.ID, err)
			}
		}
		if previousResourceType != alert.ResourceType {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "resource_type", previousResourceType, alert.ResourceType); err != nil {
				logAuditFailure(ctx, "update", "alert", alert.ID, err)
			}
		}
		if previousResourceID != alert.ResourceID {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "resource_id", previousResourceID, alert.ResourceID); err != nil {
				logAuditFailure(ctx, "update", "alert", alert.ID, err)
			}
		}
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

func (s *AlertService) DeleteAlert(ctx context.Context, id, userID uint, organizationID uuid.UUID) error {
	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(alert.ID, organizationID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogDelete(userID, organizationID, "alert", strconv.FormatUint(uint64(alert.ID), 10), &alert.ProjectID, alert.IncidentID); err != nil {
			logAuditFailure(ctx, "delete", "alert", alert.ID, err)
		}
	}

	return nil
}

func (s *AlertService) GetAlert(id uint, organizationID uuid.UUID) (*dto.AlertResponse, error) {
	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return nil, err
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

func (s *AlertService) ListAlerts(organizationID uuid.UUID, req *models.PaginationRequest) (*dto.AlertListResponse, error) {
	items, total, err := s.repo.List(req, organizationID)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.AlertListResponse{
		Items:      mapper.MapAlerts(items),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *AlertService) ResolveAlert(ctx context.Context, id, userID uint, organizationID uuid.UUID) (*dto.AlertResponse, error) {
	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return nil, err
	}

	resolvedAt := time.Now()
	if err := s.repo.Resolve(alert.ID, resolvedAt); err != nil {
		return nil, err
	}

	previousStatus := alert.Status
	alert.Status = constants.AlertStatusResolved
	alert.ResolvedAt = &resolvedAt

	if s.auditRepo != nil {
		if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", strconv.FormatUint(uint64(alert.ID), 10), &alert.ProjectID, alert.IncidentID, "status", previousStatus, constants.AlertStatusResolved); err != nil {
			logAuditFailure(ctx, "resolve", "alert", alert.ID, err)
		}
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

func (s *AlertService) AcknowledgeAlert(ctx context.Context, id, userID uint, organizationID uuid.UUID) (*dto.AlertResponse, error) {
	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return nil, err
	}

	acknowledgedAt := time.Now()
	if err := s.repo.Acknowledge(alert.ID, acknowledgedAt); err != nil {
		return nil, err
	}

	previousStatus := alert.Status
	alert.Status = constants.AlertStatusAcknowledged
	alert.AcknowledgedAt = &acknowledgedAt

	if s.auditRepo != nil {
		if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", strconv.FormatUint(uint64(alert.ID), 10), &alert.ProjectID, alert.IncidentID, "status", previousStatus, constants.AlertStatusAcknowledged); err != nil {
			logAuditFailure(ctx, "acknowledge", "alert", alert.ID, err)
		}
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

func (s *AlertService) AttachIncident(ctx context.Context, id, incidentID, userID uint, organizationID uuid.UUID) (*dto.AlertResponse, error) {
	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return nil, err
	}

	incident, err := s.incidentRepo.GetByIDAndOrganizationID(incidentID, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}
	if incident.ProjectID != alert.ProjectID {
		return nil, apperrors.ErrInvalidProject
	}

	if err := s.repo.AttachIncident(alert.ID, incidentID); err != nil {
		return nil, err
	}

	previousIncidentID := alert.IncidentID
	statusBefore := alert.Status
	alert.IncidentID = &incidentID
	alert.Status = constants.AlertStatusInvestigating

	if s.auditRepo != nil {
		entityID := strconv.FormatUint(uint64(alert.ID), 10)
		oldIncident := ""
		if previousIncidentID != nil {
			oldIncident = strconv.FormatUint(uint64(*previousIncidentID), 10)
		}
		newIncident := strconv.FormatUint(uint64(incidentID), 10)
		if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "incident_id", oldIncident, newIncident); err != nil {
			logAuditFailure(ctx, "attach_incident", "alert", alert.ID, err)
		}
		if statusBefore != alert.Status {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", entityID, &alert.ProjectID, alert.IncidentID, "status", statusBefore, alert.Status); err != nil {
				logAuditFailure(ctx, "attach_incident", "alert", alert.ID, err)
			}
		}
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

// RefreshLastSeen is retained for scheduled alert freshness workflows and dedup pipelines.
func (s *AlertService) RefreshLastSeen(ctx context.Context, id, userID uint, organizationID uuid.UUID) (*dto.AlertResponse, error) {
	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	previousCount := alert.OccurrenceCount
	if err := s.repo.IncrementOccurrence(alert.ID, now, nil); err != nil {
		return nil, err
	}

	alert.OccurrenceCount = previousCount + 1
	alert.LastSeenAt = now

	if s.auditRepo != nil {
		if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", strconv.FormatUint(uint64(alert.ID), 10), &alert.ProjectID, alert.IncidentID, "occurrence_count", strconv.Itoa(previousCount), strconv.Itoa(alert.OccurrenceCount)); err != nil {
			logAuditFailure(ctx, "refresh_last_seen", "alert", alert.ID, err)
		}
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

func (s *AlertService) ReopenAlert(ctx context.Context, id, userID uint, organizationID uuid.UUID) (*dto.AlertResponse, error) {
	alert, err := s.getOwnedAlert(id, organizationID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.repo.Reopen(alert.ID, now); err != nil {
		return nil, err
	}

	previousStatus := alert.Status
	alert.Status = constants.AlertStatusOpen
	alert.ResolvedAt = nil
	alert.LastSeenAt = now

	if s.auditRepo != nil {
		if err := s.auditRepo.LogUpdate(userID, organizationID, "alert", strconv.FormatUint(uint64(alert.ID), 10), &alert.ProjectID, alert.IncidentID, "status", previousStatus, constants.AlertStatusOpen); err != nil {
			logAuditFailure(ctx, "reopen", "alert", alert.ID, err)
		}
	}

	response := mapper.MapAlert(*alert)
	return &response, nil
}

// conditionKey identifies "the same underlying problem" across evaluation
// passes and severity changes, so the reconciler can update an existing
// alert in place instead of creating a parallel duplicate whenever severity
// shifts (which, because severity is part of Alert's fingerprint, would
// otherwise produce a different fingerprint and a brand-new row).
type conditionKey struct {
	resourceType string
	resourceID   string
	condition    string
}

// ReconcileConditions is the alert-evaluation engine's only entry point: it
// turns a snapshot of currently-firing conditions into Alert rows, per
// organization/project/source, using rules keyed on (resourceType,
// resourceID, condition) rather than Alert's fingerprint - fingerprints
// include severity, so a plain fingerprint lookup would miss a recurring
// condition whose severity differs from when it was last open, and create a
// duplicate instead of reopening the existing alert. Rules:
//  1. No existing alert for this resource+condition -> create one (OPEN).
//  2. An existing active (non-RESOLVED) alert, same severity -> occurrence
//     count bumped, nothing else changes.
//  3. An existing active alert, different severity -> updated in place
//     (title/description/severity/metadata), never duplicated.
//  4. An existing RESOLVED alert for the same resource+condition recurs ->
//     updated in place and reopened, never duplicated.
//  5. An alert this engine previously opened for a condition that is no
//     longer firing is auto-resolved: an alert that no longer reflects
//     reality is noise, not signal.
//
// organizationID is resolved from projectID (mirroring MetricService's
// StoreSnapshot), since the caller - a background evaluation loop - only
// knows the project a cluster belongs to.
func (s *AlertService) ReconcileConditions(ctx context.Context, projectID uuid.UUID, userID uint, source string, conditions []alerting.EvaluatedCondition) error {
	organizationID, err := s.repo.GetProjectOrganizationID(projectID)
	if err != nil {
		return err
	}

	all, err := s.repo.ListBySource(projectID, organizationID, source)
	if err != nil {
		return err
	}

	byKey := make(map[conditionKey]*models.Alert, len(all))
	for i := range all {
		alert := &all[i]
		var meta alerting.ConditionMetadata
		if unmarshalErr := json.Unmarshal(alert.Metadata, &meta); unmarshalErr != nil || meta.Condition == "" {
			continue // not an engine-managed alert (or predates this metadata shape) - leave it untouched
		}
		byKey[conditionKey{alert.ResourceType, alert.ResourceID, meta.Condition}] = alert
	}

	seen := make(map[conditionKey]bool, len(conditions))
	for _, condition := range conditions {
		key := conditionKey{condition.ResourceType, condition.ResourceID, condition.ConditionKey}
		seen[key] = true

		metadata, err := json.Marshal(alerting.ConditionMetadata{
			Condition:     condition.ConditionKey,
			CurrentValue:  condition.CurrentValue,
			Threshold:     condition.Threshold,
			Unit:          condition.Unit,
			ClusterID:     condition.ClusterID,
			ClusterName:   condition.ClusterName,
			ApplicationID: condition.ApplicationID,
		})
		if err != nil {
			logger.Error(ctx, "alert reconciliation: failed to marshal condition metadata", slog.Any("error", err))
			continue
		}

		existing, found := byKey[key]
		switch {
		case found && existing.Status != constants.AlertStatusResolved && existing.Severity == condition.Severity:
			if incErr := s.repo.IncrementOccurrence(existing.ID, time.Now().UTC(), metadata); incErr != nil {
				logger.Error(ctx, "alert reconciliation: failed to refresh alert", slog.Uint64("alert_id", uint64(existing.ID)), slog.Any("error", incErr))
			}

		case found && existing.Status != constants.AlertStatusResolved:
			now := time.Now().UTC()
			if _, updateErr := s.UpdateAlert(
				ctx, existing.ID, userID, organizationID, existing.ProjectID, existing.IncidentID,
				condition.Title, condition.Description, condition.Severity, existing.Status, source,
				condition.ResourceType, condition.ResourceID, existing.Labels, metadata,
				&existing.FirstSeenAt, &now, existing.AcknowledgedAt, existing.ResolvedAt,
			); updateErr != nil {
				logger.Error(ctx, "alert reconciliation: failed to escalate/update alert", slog.Uint64("alert_id", uint64(existing.ID)), slog.Any("error", updateErr))
			}

		case found: // previously RESOLVED, condition has recurred
			now := time.Now().UTC()
			if _, updateErr := s.UpdateAlert(
				ctx, existing.ID, userID, organizationID, existing.ProjectID, existing.IncidentID,
				condition.Title, condition.Description, condition.Severity, existing.Status, source,
				condition.ResourceType, condition.ResourceID, existing.Labels, metadata,
				&existing.FirstSeenAt, &now, existing.AcknowledgedAt, nil,
			); updateErr != nil {
				logger.Error(ctx, "alert reconciliation: failed to update recurring alert", slog.Uint64("alert_id", uint64(existing.ID)), slog.Any("error", updateErr))
				continue
			}
			if _, reopenErr := s.ReopenAlert(ctx, existing.ID, userID, organizationID); reopenErr != nil {
				logger.Error(ctx, "alert reconciliation: failed to reopen recurring alert", slog.Uint64("alert_id", uint64(existing.ID)), slog.Any("error", reopenErr))
			}

		default:
			if _, createErr := s.CreateAlert(
				ctx, projectID, nil, condition.Title, condition.Description, condition.Severity,
				constants.AlertStatusOpen, source, condition.ResourceType, condition.ResourceID,
				json.RawMessage(`{}`), metadata, nil, nil, userID, organizationID,
			); createErr != nil {
				logger.Error(ctx, "alert reconciliation: failed to create alert", slog.String("condition", condition.ConditionKey), slog.Any("error", createErr))
			}
		}
	}

	for key, alert := range byKey {
		if seen[key] || alert.Status == constants.AlertStatusResolved {
			continue
		}
		if _, resolveErr := s.ResolveAlert(ctx, alert.ID, userID, organizationID); resolveErr != nil {
			logger.Error(ctx, "alert reconciliation: failed to auto-resolve cleared condition", slog.Uint64("alert_id", uint64(alert.ID)), slog.Any("error", resolveErr))
		}
	}

	return nil
}

// ListAuditLogs returns the alert's own audit trail (every acknowledge/
// resolve/reopen/escalation/occurrence-count change already recorded by
// this service) - the real, timestamped basis for an alert timeline, rather
// than reconstructing one from guesses.
func (s *AlertService) ListAuditLogs(organizationID uuid.UUID, id uint, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.getOwnedAlert(id, organizationID); err != nil {
		return nil, err
	}
	if s.auditRepo == nil {
		return &models.PaginationResponse{Page: req.Page, Limit: req.Limit, Items: []models.AuditLog{}}, nil
	}

	return s.auditRepo.ListEntityAuditLogs(organizationID, "alert", strconv.FormatUint(uint64(id), 10), req)
}

func (s *AlertService) getOwnedAlert(id uint, organizationID uuid.UUID) (*models.Alert, error) {
	alert, err := s.repo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrAlertNotFound
		}
		return nil, err
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(alert.ProjectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrProjectForbidden
	}

	return alert, nil
}

func (s *AlertService) ensureIncidentBelongsToUser(incidentID uint, organizationID uuid.UUID, projectID uuid.UUID) error {
	if s.incidentRepo == nil {
		return apperrors.ErrIncidentNotFound
	}

	incident, err := s.incidentRepo.GetByIDAndOrganizationID(incidentID, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrIncidentNotFound
		}
		return err
	}

	if incident.ProjectID != projectID {
		return apperrors.ErrInvalidProject
	}

	return nil
}

func validateAlertInput(severity, status, source, resourceType string) error {
	if !constants.IsValidAlertSeverity(severity) {
		return apperrors.ErrInvalidAlertSeverity
	}
	if !constants.IsValidAlertStatus(status) {
		return apperrors.ErrInvalidAlertStatus
	}
	if !constants.IsValidAlertSource(source) {
		return apperrors.ErrInvalidAlertSource
	}
	if !constants.IsValidAlertResourceType(resourceType) {
		return apperrors.ErrInvalidAlertResourceType
	}

	return nil
}
