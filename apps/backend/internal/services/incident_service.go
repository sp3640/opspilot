package services

import (
	"context"
	"errors"
	"strconv"

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
	repo         *repository.IncidentRepository
	commentRepo  *repository.CommentRepository
	auditStorage *repository.AuditRepository
	auditRepo    *AuditService
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

// CreateIncident persists a new incident and returns its DTO representation.
func (s *IncidentService) CreateIncident(ctx context.Context, title, description, severity, status string, projectID uuid.UUID, userID uint) (*dto.IncidentResponse, error) {
	if !isValidSeverity(severity) {
		return nil, apperrors.ErrInvalidSeverity
	}
	if !isValidStatus(status) {
		return nil, apperrors.ErrInvalidStatus
	}

	belongs, err := s.repo.ProjectBelongsToUser(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	incident := &models.Incident{
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      status,
		ProjectID:   projectID,
		UserID:      userID,
	}

	if err := s.repo.Create(incident); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogCreate(userID, "incident", strconv.FormatUint(uint64(incident.ID), 10), &projectID, nil); err != nil {
			logAuditFailure(ctx, "create", "incident", incident.ID, err)
		}
	}

	response := mapper.MapIncident(*incident)
	return &response, nil
}

func (s *IncidentService) GetMyIncidents(userID uint) ([]models.Incident, error) {
	return s.repo.GetAllByUserID(userID)
}

// ListMyIncidents returns a paginated list of incidents for the user.
func (s *IncidentService) ListMyIncidents(userID uint, req *models.PaginationRequest) (*dto.IncidentListResponse, error) {
	items, total, err := s.repo.ListByUserID(req, userID)
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
func (s *IncidentService) GetIncidentByID(id, userID uint) (*dto.IncidentResponse, error) {
	incident, err := s.repo.GetByIDAndUserID(id, userID)
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
func (s *IncidentService) UpdateIncident(ctx context.Context, id, userID uint, title, description, severity, status string, projectID uuid.UUID) (*dto.IncidentResponse, error) {
	incident, err := s.repo.GetByIDAndUserID(id, userID)
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

	belongs, err := s.repo.ProjectBelongsToUser(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	previousTitle := incident.Title
	previousDescription := incident.Description
	previousSeverity := incident.Severity
	previousStatus := incident.Status
	previousProjectID := incident.ProjectID

	incident.Title = title
	incident.Description = description
	incident.Severity = severity
	incident.Status = status
	incident.ProjectID = projectID

	if err := s.repo.Update(incident); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		entityIDStr := strconv.FormatUint(uint64(incident.ID), 10)
		if previousTitle != title {
			if err := s.auditRepo.LogUpdate(userID, "incident", entityIDStr, &projectID, nil, "title", previousTitle, title); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousDescription != description {
			if err := s.auditRepo.LogUpdate(userID, "incident", entityIDStr, &projectID, nil, "description", previousDescription, description); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousSeverity != severity {
			if err := s.auditRepo.LogUpdate(userID, "incident", entityIDStr, &projectID, nil, "severity", previousSeverity, severity); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousStatus != status {
			if err := s.auditRepo.LogUpdate(userID, "incident", entityIDStr, &projectID, nil, "status", previousStatus, status); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousProjectID != projectID {
			if err := s.auditRepo.LogUpdate(userID, "incident", entityIDStr, &projectID, nil, "project_id", previousProjectID.String(), projectID.String()); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
	}

	response := mapper.MapIncident(*incident)
	return &response, nil
}

func (s *IncidentService) DeleteIncident(ctx context.Context, id, userID uint) error {
	incident, err := s.repo.GetByIDAndUserID(id, userID)
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
		if err := s.auditStorage.ClearIncidentReference(incident.ID); err != nil {
			return err
		}
	}

	if err := s.repo.Delete(incident.ID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogDelete(userID, "incident", strconv.FormatUint(uint64(incident.ID), 10), &incident.ProjectID, nil); err != nil {
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
	switch status {
	case constants.StatusOpen, constants.StatusInvestigating, constants.StatusResolved:
		return true
	default:
		return false
	}
}
