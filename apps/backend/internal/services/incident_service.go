package services

import (
	"context"
	"errors"
	"strconv"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type IncidentService struct {
	repo      *repository.IncidentRepository
	auditRepo *AuditService
}

func NewIncidentService(repo *repository.IncidentRepository, auditService *AuditService) *IncidentService {
	return &IncidentService{repo: repo, auditRepo: auditService}
}

func (s *IncidentService) CreateIncident(ctx context.Context, title, description, severity, status string, projectID, userID uint) (*models.Incident, error) {
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
		projectIDValue := projectID
		projectIDPtr := &projectIDValue
		if err := s.auditRepo.LogCreate(userID, "incident", incident.ID, projectIDPtr, nil); err != nil {
			logAuditFailure(ctx, "create", "incident", incident.ID, err)
		}
	}

	return incident, nil
}

func (s *IncidentService) GetMyIncidents(userID uint) ([]models.Incident, error) {
	return s.repo.GetAllByUserID(userID)
}

func (s *IncidentService) ListMyIncidents(userID uint, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	items, total, err := s.repo.ListByUserID(req, userID)
	if err != nil {
		return nil, err
	}

	return &models.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
		Items:      items,
	}, nil
}

func (s *IncidentService) GetIncidentByID(id, userID uint) (*models.Incident, error) {
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
	return incident, nil
}

func (s *IncidentService) UpdateIncident(ctx context.Context, id, userID uint, title, description, severity, status string, projectID uint) (*models.Incident, error) {
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
		projectIDValue := projectID
		projectIDPtr := &projectIDValue
		if previousTitle != title {
			if err := s.auditRepo.LogUpdate(userID, "incident", incident.ID, projectIDPtr, nil, "title", previousTitle, title); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousDescription != description {
			if err := s.auditRepo.LogUpdate(userID, "incident", incident.ID, projectIDPtr, nil, "description", previousDescription, description); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousSeverity != severity {
			if err := s.auditRepo.LogUpdate(userID, "incident", incident.ID, projectIDPtr, nil, "severity", previousSeverity, severity); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousStatus != status {
			if err := s.auditRepo.LogUpdate(userID, "incident", incident.ID, projectIDPtr, nil, "status", previousStatus, status); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
		if previousProjectID != projectID {
			if err := s.auditRepo.LogUpdate(userID, "incident", incident.ID, projectIDPtr, nil, "project_id", strconv.Itoa(int(previousProjectID)), strconv.Itoa(int(projectID))); err != nil {
				logAuditFailure(ctx, "update", "incident", incident.ID, err)
			}
		}
	}

	return incident, nil
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

	if err := s.repo.Delete(incident.ID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		projectIDValue := incident.ProjectID
		projectIDPtr := &projectIDValue
		if err := s.auditRepo.LogDelete(userID, "incident", incident.ID, projectIDPtr, nil); err != nil {
			logAuditFailure(ctx, "delete", "incident", incident.ID, err)
		}
	}

	return nil
}

func isValidSeverity(severity string) bool {
	switch severity {
	case "P0", "P1", "P2", "P3", "P4":
		return true
	default:
		return false
	}
}

func isValidStatus(status string) bool {
	switch status {
	case "OPEN", "INVESTIGATING", "RESOLVED":
		return true
	default:
		return false
	}
}
