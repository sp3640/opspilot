package services

import (
	"errors"

	"github.com/google/uuid"
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

func (s *AuditService) LogCreate(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint) error {
	log := &models.AuditLog{
		UserID:         userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		IncidentID:     incidentID,
		EntityType:     entityType,
		EntityID:       entityID,
		Action:         models.AuditActionCreate,
	}

	return s.repo.Create(log)
}

func (s *AuditService) LogUpdate(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint, fieldName string, oldValue string, newValue string) error {
	log := &models.AuditLog{
		UserID:         userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		IncidentID:     incidentID,
		EntityType:     entityType,
		EntityID:       entityID,
		Action:         models.AuditActionUpdate,
		FieldName:      fieldName,
		OldValue:       oldValue,
		NewValue:       newValue,
	}

	return s.repo.Create(log)
}

func (s *AuditService) LogDelete(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint) error {
	log := &models.AuditLog{
		UserID:         userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		IncidentID:     incidentID,
		EntityType:     entityType,
		EntityID:       entityID,
		Action:         models.AuditActionDelete,
	}

	return s.repo.Create(log)
}

func (s *AuditService) ListIncidentAuditLogs(userID uint, organizationID uuid.UUID, incidentID uint, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.incidentRepo.GetByIDAndOrganizationID(incidentID, organizationID); err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrIncidentNotFound
		}
		return nil, err
	}

	items, total, err := s.repo.ListByIncidentID(req, incidentID, organizationID)
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

// ListEntityAuditLogs returns the audit trail for any entity by its generic
// (entityType, entityID) identity. Unlike ListIncidentAuditLogs/
// ListProjectAuditLogs, ownership of the entity itself is the caller's
// responsibility (e.g. AlertService.ListAuditLogs verifies alert ownership
// before delegating here) since AuditService has no repo for every entity
// kind that might call this.
func (s *AuditService) ListEntityAuditLogs(organizationID uuid.UUID, entityType, entityID string, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	items, total, err := s.repo.ListByEntity(req, entityType, entityID, organizationID)
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

func (s *AuditService) ListProjectAuditLogs(userID uint, organizationID uuid.UUID, projectID uuid.UUID, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.projectRepo.GetByIDAndOrganizationID(projectID, organizationID); err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, err
	}

	items, total, err := s.repo.ListByProjectID(req, projectID, organizationID)
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
