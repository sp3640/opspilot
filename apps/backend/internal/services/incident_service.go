package services

import (
	"errors"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type IncidentService struct {
	repo *repository.IncidentRepository
}

func NewIncidentService(repo *repository.IncidentRepository) *IncidentService {
	return &IncidentService{repo: repo}
}

func (s *IncidentService) CreateIncident(title, description, severity, status string, projectID, userID uint) (*models.Incident, error) {
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

	return incident, nil
}

func (s *IncidentService) GetMyIncidents(userID uint) ([]models.Incident, error) {
	return s.repo.GetAllByUserID(userID)
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

func (s *IncidentService) UpdateIncident(id, userID uint, title, description, severity, status string, projectID uint) (*models.Incident, error) {
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

	incident.Title = title
	incident.Description = description
	incident.Severity = severity
	incident.Status = status
	incident.ProjectID = projectID

	if err := s.repo.Update(incident); err != nil {
		return nil, err
	}

	return incident, nil
}

func (s *IncidentService) DeleteIncident(id, userID uint) error {
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

	return s.repo.Delete(incident.ID)
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
