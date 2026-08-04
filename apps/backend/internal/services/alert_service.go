package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
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
