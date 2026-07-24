package services

import (
	"errors"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type AuditService struct {
	repo         *repository.AuditRepository
	projectRepo  *repository.ProjectRepository
	incidentRepo *repository.IncidentRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) WithProjectRepo(projectRepo *repository.ProjectRepository) *AuditService {
	s.projectRepo = projectRepo
	return s
}

func (s *AuditService) WithIncidentRepo(incidentRepo *repository.IncidentRepository) *AuditService {
	s.incidentRepo = incidentRepo
	return s
}

func (s *AuditService) LogCreate(userID uint, entityType string, entityID uint, projectID *uint, incidentID *uint) error {
	log := &models.AuditLog{
		UserID:     userID,
		ProjectID:  projectID,
		IncidentID: incidentID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     models.AuditActionCreate,
	}

	return s.repo.Create(log)
}

func (s *AuditService) LogUpdate(userID uint, entityType string, entityID uint, projectID *uint, incidentID *uint, fieldName string, oldValue string, newValue string) error {
	log := &models.AuditLog{
		UserID:     userID,
		ProjectID:  projectID,
		IncidentID: incidentID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     models.AuditActionUpdate,
		FieldName:  fieldName,
		OldValue:   oldValue,
		NewValue:   newValue,
	}

	return s.repo.Create(log)
}

func (s *AuditService) LogDelete(userID uint, entityType string, entityID uint, projectID *uint, incidentID *uint) error {
	log := &models.AuditLog{
		UserID:     userID,
		ProjectID:  projectID,
		IncidentID: incidentID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     models.AuditActionDelete,
	}

	return s.repo.Create(log)
}

func (s *AuditService) GetIncidentAuditLogs(userID uint, incidentID uint) ([]models.AuditLog, error) {
	incident, err := s.incidentRepo.GetByIDAndUserID(incidentID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	_ = incident
	return s.repo.GetByIncidentID(incidentID)
}

func (s *AuditService) ListIncidentAuditLogs(userID uint, incidentID uint, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.incidentRepo.GetByIDAndUserID(incidentID, userID); err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	items, total, err := s.repo.ListByIncidentID(req, incidentID)
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

func (s *AuditService) GetProjectAuditLogs(userID uint, projectID uint) ([]models.AuditLog, error) {
	project, err := s.projectRepo.GetByIDAndUserID(projectID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, err
	}

	_ = project
	return s.repo.GetByProjectID(projectID)
}

func (s *AuditService) ListProjectAuditLogs(userID uint, projectID uint, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.projectRepo.GetByIDAndUserID(projectID, userID); err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, err
	}

	items, total, err := s.repo.ListByProjectID(req, projectID)
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
