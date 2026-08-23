package services

import (
	"context"
	"errors"
	"strconv"
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

type IncidentService struct {
	repo                *repository.IncidentRepository
	commentRepo         *repository.CommentRepository
	auditStorage        *repository.AuditRepository
	auditRepo           *AuditService
	applicationRepo     *repository.ApplicationRepository
	teamRepo            *repository.TeamRepository
	userRepo            *repository.UserRepository
	notificationService *NotificationService
}

func NewIncidentService(
	repo *repository.IncidentRepository,
	commentRepo *repository.CommentRepository,
	auditStorage *repository.AuditRepository,
	auditService *AuditService,
) *IncidentService {
	return &IncidentService{
		repo:         repo,
		commentRepo:  commentRepo,
		auditStorage: auditStorage,
		auditRepo:    auditService,
	}
}

// WithApplicationRepo enables validating/resolving Incident.ApplicationID
// (the affected application). Optional: without it, an incident simply
// cannot be linked to an application.
func (s *IncidentService) WithApplicationRepo(applicationRepo *repository.ApplicationRepository) *IncidentService {
	s.applicationRepo = applicationRepo
	return s
}

// WithTeamRepo enables validating/resolving Incident.OwnerTeamID.
func (s *IncidentService) WithTeamRepo(teamRepo *repository.TeamRepository) *IncidentService {
	s.teamRepo = teamRepo
	return s
}

// WithUserRepo enables validating that an AssignIncident target belongs to
// the incident's organization. Without it, assignment is unavailable.
func (s *IncidentService) WithUserRepo(userRepo *repository.UserRepository) *IncidentService {
	s.userRepo = userRepo
	return s
}

// WithNotificationService enables firing SEV1_INCIDENT and
// INCIDENT_ASSIGNED notifications. Optional: without it, incidents are
// still created/updated/assigned normally, just without notification fan-out.
func (s *IncidentService) WithNotificationService(notificationService *NotificationService) *IncidentService {
	s.notificationService = notificationService
	return s
}

// resolveApplication validates that applicationID (if provided) exists,
// belongs to organizationID, and belongs to the same project the incident
// is (or will be) scoped to - an incident cannot claim to affect an
// application from a different project.
func (s *IncidentService) resolveApplication(applicationID *uuid.UUID, projectID, organizationID uuid.UUID) error {
	if applicationID == nil {
		return nil
	}
	if s.applicationRepo == nil {
		return apperrors.ErrIncidentApplicationNotFound
	}

	application, err := s.applicationRepo.GetApplication(*applicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrIncidentApplicationNotFound
		}
		return err
	}
	if application.ProjectID != projectID {
		return apperrors.ErrIncidentApplicationMismatch
	}

	return nil
}

// resolveOwnerTeam validates that ownerTeamID (if provided) exists and
// belongs to organizationID.
func (s *IncidentService) resolveOwnerTeam(ownerTeamID *uuid.UUID, organizationID uuid.UUID) error {
	if ownerTeamID == nil {
		return nil
	}
	if s.teamRepo == nil {
		return apperrors.ErrIncidentOwnerTeamNotFound
	}

	team, err := s.teamRepo.GetByID(*ownerTeamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrIncidentOwnerTeamNotFound
		}
		return err
	}
	if team.OrganizationID != organizationID {
		return apperrors.ErrIncidentOwnerTeamNotFound
	}

	return nil
}

// CreateIncident persists a new incident and returns its DTO representation.
func (s *IncidentService) CreateIncident(ctx context.Context, title, description, severity, status string, projectID uuid.UUID, applicationID, ownerTeamID *uuid.UUID, userID uint, organizationID uuid.UUID) (*dto.IncidentResponse, error) {
	if !isValidSeverity(severity) {
		return nil, apperrors.ErrInvalidSeverity
	}
	if !isValidStatus(status) {
		return nil, apperrors.ErrInvalidStatus
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(projectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	if err := s.resolveApplication(applicationID, projectID, organizationID); err != nil {
		return nil, err
	}
	if err := s.resolveOwnerTeam(ownerTeamID, organizationID); err != nil {
		return nil, err
	}

	var resolvedAt *time.Time
	if status == constants.StatusResolved {
		now := time.Now()
		resolvedAt = &now
	}

	incident := &models.Incident{
		OrganizationID: organizationID,
		Title:          title,
		Description:    description,
		Severity:       severity,
		Status:         status,
		ProjectID:      projectID,
		ApplicationID:  applicationID,
		OwnerTeamID:    ownerTeamID,
		UserID:         userID,
		ResolvedAt:     resolvedAt,
	}

	if err := s.repo.Create(incident); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogCreate(userID, organizationID, "incident", strconv.FormatUint(uint64(incident.ID), 10), &projectID, nil); err != nil {
			logAuditFailure(ctx, "create", "incident", incident.ID, err)
		}
	}

	if s.notificationService != nil && incident.Severity == constants.SeverityP0 {
		s.notificationService.NotifySev1Incident(ctx, incident, userID)
	}

	response := mapper.MapIncident(*incident)
	return &response, nil
}

// ListMyIncidents returns a paginated list of incidents for the user.
func (s *IncidentService) ListMyIncidents(organizationID uuid.UUID, req *models.PaginationRequest) (*dto.IncidentListResponse, error) {
	items, total, err := s.repo.ListByOrganizationID(req, organizationID)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &dto.IncidentListResponse{
		Items:      mapper.MapIncidents(items),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetIncidentByID returns a single incident the user owns, as a DTO.
func (s *IncidentService) GetIncidentByID(id uint, organizationID uuid.UUID) (*dto.IncidentResponse, error) {
	incident, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	response := mapper.MapIncident(*incident)
	return &response, nil
}

// UpdateIncident applies changes and returns the updated DTO.
func (s *IncidentService) UpdateIncident(ctx context.Context, id, userID uint, organizationID uuid.UUID, title, description, severity, status string, projectID uuid.UUID, applicationID, ownerTeamID *uuid.UUID) (*dto.IncidentResponse, error) {
	incident, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	if !isValidSeverity(severity) {
		return nil, apperrors.ErrInvalidSeverity
	}
	if !isValidStatus(status) {
		return nil, apperrors.ErrInvalidStatus
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(projectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	if err := s.resolveApplication(applicationID, projectID, organizationID); err != nil {
		return nil, err
	}
	if err := s.resolveOwnerTeam(ownerTeamID, organizationID); err != nil {
		return nil, err
	}

	previousTitle := incident.Title
	previousDescription := incident.Description
	previousSeverity := incident.Severity
	previousStatus := incident.Status
	previousProjectID := incident.ProjectID
	previousApplicationID := incident.ApplicationID
	previousOwnerTeamID := incident.OwnerTeamID

	incident.Title = title
	incident.Description = description
	incident.Severity = severity
	incident.Status = status
	incident.ProjectID = projectID
	incident.ApplicationID = applicationID
	incident.OwnerTeamID = ownerTeamID

	// ResolvedAt tracks the current status: set the moment it becomes
	// RESOLVED, cleared the moment it moves to anything else (reopened).
	if status == constants.StatusResolved && previousStatus != constants.StatusResolved {
		now := time.Now()
		incident.ResolvedAt = &now
	} else if status != constants.StatusResolved {
		incident.ResolvedAt = nil
	}

	if err := s.repo.Update(incident); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		entityIDStr := strconv.FormatUint(uint64(incident.ID), 10)
		if previousTitle != title {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "incident", entityIDStr, &projectID, nil, "title", previousTitle, title); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousDescription != description {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "incident", entityIDStr, &projectID, nil, "description", previousDescription, description); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousSeverity != severity {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "incident", entityIDStr, &projectID, nil, "severity", previousSeverity, severity); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousStatus != status {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "incident", entityIDStr, &projectID, nil, "status", previousStatus, status); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousProjectID != projectID {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "incident", entityIDStr, &projectID, nil, "project_id", previousProjectID.String(), projectID.String()); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if !uuidPointerStringsEqual(previousApplicationID, applicationID) {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "incident", entityIDStr, &projectID, nil, "application_id", uuidPointerString(previousApplicationID), uuidPointerString(applicationID)); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if !uuidPointerStringsEqual(previousOwnerTeamID, ownerTeamID) {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "incident", entityIDStr, &projectID, nil, "owner_team_id", uuidPointerString(previousOwnerTeamID), uuidPointerString(ownerTeamID)); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
	}

	if s.notificationService != nil && severity == constants.SeverityP0 && previousSeverity != constants.SeverityP0 {
		s.notificationService.NotifySev1Incident(ctx, incident, userID)
	}

	response := mapper.MapIncident(*incident)
	return &response, nil
}

// AssignIncident sets the user responsible for driving this incident to
// resolution (distinct from OwnerTeamID, which team) and fires an
// INCIDENT_ASSIGNED notification.
func (s *IncidentService) AssignIncident(ctx context.Context, id, actorID uint, organizationID uuid.UUID, assigneeUserID uint) (*dto.IncidentResponse, error) {
	incident, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	if s.userRepo == nil {
		return nil, apperrors.ErrIncidentAssigneeNotFound
	}
	assignee, err := s.userRepo.GetByID(assigneeUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentAssigneeNotFound
		}
		return nil, err
	}
	if assignee.OrganizationID == nil || *assignee.OrganizationID != organizationID {
		return nil, apperrors.ErrIncidentAssigneeNotFound
	}

	previousAssigneeID := incident.AssigneeID
	incident.AssigneeID = &assigneeUserID
	if err := s.repo.Update(incident); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		entityIDStr := strconv.FormatUint(uint64(incident.ID), 10)
		if err := s.auditRepo.LogUpdate(actorID, organizationID, "incident", entityIDStr, &incident.ProjectID, nil, "assignee_id", uintPointerString(previousAssigneeID), strconv.FormatUint(uint64(assigneeUserID), 10)); err != nil {
			logAuditFailure(ctx, "assign", "incident", incident.ID, err)
		}
	}

	if s.notificationService != nil {
		s.notificationService.NotifyIncidentAssigned(ctx, incident, assigneeUserID, actorID)
	}

	response := mapper.MapIncident(*incident)
	return &response, nil
}

// AcknowledgeIncident records the first time an incident is acknowledged -
// the source data for the MTTA SRE metric. Idempotent: acknowledging an
// already-acknowledged incident is a no-op that returns its current state.
func (s *IncidentService) AcknowledgeIncident(ctx context.Context, id, actorID uint, organizationID uuid.UUID) (*dto.IncidentResponse, error) {
	incident, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	if incident.AcknowledgedAt != nil {
		response := mapper.MapIncident(*incident)
		return &response, nil
	}

	now := time.Now().UTC()
	incident.AcknowledgedAt = &now
	if err := s.repo.Update(incident); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		entityIDStr := strconv.FormatUint(uint64(incident.ID), 10)
		if err := s.auditRepo.LogUpdate(actorID, organizationID, "incident", entityIDStr, &incident.ProjectID, nil, "acknowledged_at", "", now.Format(time.RFC3339)); err != nil {
			logAuditFailure(ctx, "acknowledge", "incident", incident.ID, err)
		}
	}

	response := mapper.MapIncident(*incident)
	return &response, nil
}

func (s *IncidentService) DeleteIncident(ctx context.Context, id, userID uint, organizationID uuid.UUID) error {
	incident, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrIncidentNotFound
		}
		return err
	}

	// Remove dependent comments first so incident delete is not blocked by FK constraints.
	if s.commentRepo != nil {
		if err := s.commentRepo.DeleteByIncidentID(incident.ID); err != nil {
			return err
		}
	}

	// Drop audit references to the incident before deleting the row.
	if s.auditStorage != nil {
		if err := s.auditStorage.ClearIncidentReference(incident.ID, organizationID); err != nil {
			return err
		}
	}

	if err := s.repo.Delete(incident.ID, organizationID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogDelete(userID, organizationID, "incident", strconv.FormatUint(uint64(incident.ID), 10), &incident.ProjectID, nil); err != nil {
			logAuditFailure(ctx, "delete", "incident", incident.ID, err)
		}
	}

	return nil
}

func isValidSeverity(severity string) bool {
	switch severity {
	case constants.SeverityP0, constants.SeverityP1, constants.SeverityP2, constants.SeverityP3, constants.SeverityP4:
		return true
	default:
		return false
	}
}

func isValidStatus(status string) bool {
	return constants.IsValidStatus(status)
}
