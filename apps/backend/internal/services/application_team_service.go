package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type ApplicationTeamService struct {
	applicationTeamRepo *repository.ApplicationTeamRepository
	applicationRepo     *repository.ApplicationRepository
	teamRepo            *repository.TeamRepository
	auditService        *AuditService
}

func NewApplicationTeamService(
	applicationTeamRepo *repository.ApplicationTeamRepository,
	applicationRepo *repository.ApplicationRepository,
	teamRepo *repository.TeamRepository,
) *ApplicationTeamService {
	return &ApplicationTeamService{
		applicationTeamRepo: applicationTeamRepo,
		applicationRepo:     applicationRepo,
		teamRepo:            teamRepo,
	}
}

func (s *ApplicationTeamService) WithAuditService(auditService *AuditService) *ApplicationTeamService {
	s.auditService = auditService
	return s
}

func (s *ApplicationTeamService) AssignTeam(actorRole string, actorID uint, organizationID, applicationID, teamID uuid.UUID) (*dto.ApplicationTeamResponse, error) {
	if !isApplicationTeamPlatformAdminRole(actorRole) {
		return nil, apperrors.ErrApplicationForbidden
	}

	application, err := s.applicationRepo.GetApplication(applicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrApplicationNotFound
		}
		return nil, err
	}

	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, err
	}

	if team.OrganizationID != application.OrganizationID || team.OrganizationID != organizationID {
		return nil, apperrors.ErrApplicationForbidden
	}

	isAssigned, err := s.applicationTeamRepo.IsAssigned(applicationID, teamID, organizationID)
	if err != nil {
		return nil, err
	}
	if isAssigned {
		return nil, apperrors.ErrApplicationTeamAlreadyAssigned
	}

	mapping := &models.ApplicationTeam{
		OrganizationID: organizationID,
		ApplicationID:  applicationID,
		TeamID:         teamID,
	}
	if err := s.applicationTeamRepo.AssignTeam(mapping); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         actorID,
			OrganizationID: organizationID,
			ApplicationID:  &applicationID,
			EntityType:     "application_team",
			EntityID:       mapping.ID.String(),
			Action:         models.AuditActionCreate,
			AfterState:     marshalAuditState(map[string]any{"applicationId": applicationID.String(), "teamId": teamID.String()}),
		})
	}

	response := mapApplicationTeamResponse(*mapping)
	return &response, nil
}

func (s *ApplicationTeamService) RemoveTeam(actorRole string, actorID uint, organizationID, applicationID, teamID uuid.UUID) error {
	if !isApplicationTeamPlatformAdminRole(actorRole) {
		return apperrors.ErrApplicationForbidden
	}

	if _, err := s.applicationRepo.GetApplication(applicationID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrApplicationNotFound
		}
		return err
	}

	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrTeamNotFound
		}
		return err
	}
	if team.OrganizationID != organizationID {
		return apperrors.ErrApplicationForbidden
	}

	if err := s.applicationTeamRepo.RemoveTeam(applicationID, teamID, organizationID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         actorID,
			OrganizationID: organizationID,
			ApplicationID:  &applicationID,
			EntityType:     "application_team",
			EntityID:       applicationID.String(),
			Action:         models.AuditActionDelete,
			BeforeState:    marshalAuditState(map[string]any{"applicationId": applicationID.String(), "teamId": teamID.String()}),
		})
	}

	return nil
}

func (s *ApplicationTeamService) ListApplicationTeams(actorRole string, organizationID, applicationID uuid.UUID) (*dto.ApplicationTeamListResponse, error) {
	if !isApplicationTeamOrganizationMemberRole(actorRole) {
		return nil, apperrors.ErrApplicationForbidden
	}

	if _, err := s.applicationRepo.GetApplication(applicationID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrApplicationNotFound
		}
		return nil, err
	}

	items, err := s.applicationTeamRepo.ListApplicationTeams(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ApplicationTeamResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, mapApplicationTeamResponse(item))
	}

	return &dto.ApplicationTeamListResponse{Items: responses, Total: int64(len(responses))}, nil
}

func (s *ApplicationTeamService) ListTeamApplications(actorRole string, organizationID, teamID uuid.UUID) (*dto.ApplicationTeamListResponse, error) {
	if !isApplicationTeamOrganizationMemberRole(actorRole) {
		return nil, apperrors.ErrApplicationForbidden
	}

	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, err
	}
	if team.OrganizationID != organizationID {
		return nil, apperrors.ErrApplicationForbidden
	}

	items, err := s.applicationTeamRepo.ListTeamApplications(teamID, organizationID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ApplicationTeamResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, mapApplicationTeamResponse(item))
	}

	return &dto.ApplicationTeamListResponse{Items: responses, Total: int64(len(responses))}, nil
}

func mapApplicationTeamResponse(mapping models.ApplicationTeam) dto.ApplicationTeamResponse {
	return dto.ApplicationTeamResponse{
		ID:             mapping.ID.String(),
		OrganizationID: mapping.OrganizationID.String(),
		ApplicationID:  mapping.ApplicationID.String(),
		TeamID:         mapping.TeamID.String(),
		CreatedAt:      mapping.CreatedAt,
	}
}

func isApplicationTeamPlatformAdminRole(role string) bool {
	return rbac.HasPermission(role, rbac.PermissionApplicationTeamManage)
}

func isApplicationTeamOrganizationMemberRole(role string) bool {
	return rbac.HasPermission(role, rbac.PermissionApplicationTeamRead)
}
