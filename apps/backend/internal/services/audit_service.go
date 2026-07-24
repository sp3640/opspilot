package services

import (
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
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
