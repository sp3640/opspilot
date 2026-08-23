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

// AuditEventInput is the general-purpose shape for every audit write.
// LogCreate/LogUpdate/LogDelete remain as thin, unchanged-signature
// convenience wrappers around this for the many pre-existing call sites;
// new call sites (Phase 23: login, invitations, teams, applications,
// deployments) use LogEvent directly when they need Result, IP/user-agent,
// an ApplicationID cross-reference, or a before/after snapshot.
type AuditEventInput struct {
	UserID         uint
	OrganizationID uuid.UUID
	ProjectID      *uuid.UUID
	ApplicationID  *uuid.UUID
	IncidentID     *uint
	EntityType     string
	EntityID       string
	Action         models.AuditAction
	// Result defaults to AuditResultSuccess when left blank.
	Result      models.AuditResult
	FieldName   string
	OldValue    string
	NewValue    string
	BeforeState string
	AfterState  string
	IPAddress   string
	UserAgent   string
}

// LogEvent is the single write path every audit entry ultimately goes
// through. It is best-effort by convention (callers should never fail their
// business operation because an audit write failed) - see logAuditFailure
// for the standard way callers report a write error without propagating it.
func (s *AuditService) LogEvent(input AuditEventInput) error {
	result := input.Result
	if result == "" {
		result = models.AuditResultSuccess
	}

	log := &models.AuditLog{
		UserID:         input.UserID,
		OrganizationID: input.OrganizationID,
		ProjectID:      input.ProjectID,
		ApplicationID:  input.ApplicationID,
		IncidentID:     input.IncidentID,
		EntityType:     input.EntityType,
		EntityID:       input.EntityID,
		Action:         input.Action,
		Result:         result,
		FieldName:      input.FieldName,
		OldValue:       input.OldValue,
		NewValue:       input.NewValue,
		BeforeState:    input.BeforeState,
		AfterState:     input.AfterState,
		IPAddress:      input.IPAddress,
		UserAgent:      input.UserAgent,
	}

	return s.repo.Create(log)
}

func (s *AuditService) LogCreate(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint) error {
	return s.LogEvent(AuditEventInput{
		UserID:         userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		IncidentID:     incidentID,
		EntityType:     entityType,
		EntityID:       entityID,
		Action:         models.AuditActionCreate,
	})
}

func (s *AuditService) LogUpdate(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint, fieldName string, oldValue string, newValue string) error {
	return s.LogEvent(AuditEventInput{
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
	})
}

func (s *AuditService) LogDelete(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint) error {
	return s.LogEvent(AuditEventInput{
		UserID:         userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		IncidentID:     incidentID,
		EntityType:     entityType,
		EntityID:       entityID,
		Action:         models.AuditActionDelete,
	})
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

// ListOrganizationAuditLogs returns every audit log in the organization,
// unscoped by project/incident/entity - the Organization Audit view (Phase
// 23). Visibility is organization-membership-based (mirroring every other
// org-wide list endpoint), not gated by an extra ownership check, since
// there is no narrower resource to own here.
func (s *AuditService) ListOrganizationAuditLogs(organizationID uuid.UUID, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	items, total, err := s.repo.ListByOrganization(req, organizationID)
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
