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

type ProjectTeamService struct {
	projectTeamRepo *repository.ProjectTeamRepository
	projectRepo     *repository.ProjectRepository
	teamRepo        *repository.TeamRepository
	auditService    *AuditService
}

func NewProjectTeamService(
	projectTeamRepo *repository.ProjectTeamRepository,
	projectRepo *repository.ProjectRepository,
	teamRepo *repository.TeamRepository,
) *ProjectTeamService {
	return &ProjectTeamService{
		projectTeamRepo: projectTeamRepo,
		projectRepo:     projectRepo,
		teamRepo:        teamRepo,
	}
}

func (s *ProjectTeamService) WithAuditService(auditService *AuditService) *ProjectTeamService {
	s.auditService = auditService
	return s
}

func (s *ProjectTeamService) AssignTeam(actorRole string, actorID uint, organizationID, projectID, teamID uuid.UUID) (*dto.ProjectTeamResponse, error) {
	if !isProjectTeamPlatformAdminRole(actorRole) {
		return nil, apperrors.ErrProjectForbidden
	}

	project, err := s.projectRepo.GetByIDAndOrganizationID(projectID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
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

	if team.OrganizationID != project.OrganizationID || team.OrganizationID != organizationID {
		return nil, apperrors.ErrProjectForbidden
	}

	isAssigned, err := s.projectTeamRepo.IsAssigned(projectID, teamID, organizationID)
	if err != nil {
		return nil, err
	}
	if isAssigned {
		return nil, apperrors.ErrProjectTeamAlreadyAssigned
	}

	mapping := &models.ProjectTeam{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		TeamID:         teamID,
	}
	if err := s.projectTeamRepo.AssignTeam(mapping); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogCreate(actorID, organizationID, "project_team", mapping.ID.String(), &projectID, nil)
	}

	response := mapProjectTeamResponse(*mapping)
	return &response, nil
}

func (s *ProjectTeamService) RemoveTeam(actorRole string, actorID uint, organizationID, projectID, teamID uuid.UUID) error {
	if !isProjectTeamPlatformAdminRole(actorRole) {
		return apperrors.ErrProjectForbidden
	}

	if _, err := s.projectRepo.GetByIDAndOrganizationID(projectID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
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
		return apperrors.ErrProjectForbidden
	}

	if err := s.projectTeamRepo.RemoveTeam(projectID, teamID, organizationID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         actorID,
			OrganizationID: organizationID,
			ProjectID:      &projectID,
			EntityType:     "project_team",
			EntityID:       projectID.String(),
			Action:         models.AuditActionDelete,
			BeforeState:    marshalAuditState(map[string]any{"projectId": projectID.String(), "teamId": teamID.String()}),
		})
	}

	return nil
}

func (s *ProjectTeamService) ListProjectTeams(actorRole string, organizationID, projectID uuid.UUID) (*dto.ProjectTeamListResponse, error) {
	if !isProjectTeamOrganizationMemberRole(actorRole) {
		return nil, apperrors.ErrProjectForbidden
	}

	if _, err := s.projectRepo.GetByIDAndOrganizationID(projectID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, err
	}

	items, err := s.projectTeamRepo.ListProjectTeams(projectID, organizationID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProjectTeamResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, mapProjectTeamResponse(item))
	}

	return &dto.ProjectTeamListResponse{Items: responses, Total: int64(len(responses))}, nil
}

func (s *ProjectTeamService) ListTeamProjects(actorRole string, organizationID, teamID uuid.UUID) (*dto.ProjectTeamListResponse, error) {
	if !isProjectTeamOrganizationMemberRole(actorRole) {
		return nil, apperrors.ErrProjectForbidden
	}

	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, err
	}
	if team.OrganizationID != organizationID {
		return nil, apperrors.ErrProjectForbidden
	}

	items, err := s.projectTeamRepo.ListTeamProjects(teamID, organizationID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProjectTeamResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, mapProjectTeamResponse(item))
	}

	return &dto.ProjectTeamListResponse{Items: responses, Total: int64(len(responses))}, nil
}

func mapProjectTeamResponse(mapping models.ProjectTeam) dto.ProjectTeamResponse {
	return dto.ProjectTeamResponse{
		ID:             mapping.ID.String(),
		OrganizationID: mapping.OrganizationID.String(),
		ProjectID:      mapping.ProjectID.String(),
		TeamID:         mapping.TeamID.String(),
		CreatedAt:      mapping.CreatedAt,
	}
}

func isProjectTeamPlatformAdminRole(role string) bool {
	return rbac.HasPermission(role, rbac.PermissionProjectTeamManage)
}

func isProjectTeamOrganizationMemberRole(role string) bool {
	return rbac.HasPermission(role, rbac.PermissionProjectTeamRead)
}
