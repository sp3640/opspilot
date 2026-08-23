package services

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type TeamService struct {
	teamRepo       *repository.TeamRepository
	teamMemberRepo *repository.TeamMemberRepository
	userRepo       *repository.UserRepository
	auditService   *AuditService
}

func NewTeamService(
	teamRepo *repository.TeamRepository,
	teamMemberRepo *repository.TeamMemberRepository,
	userRepo *repository.UserRepository,
) *TeamService {
	return &TeamService{
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		userRepo:       userRepo,
	}
}

func (s *TeamService) WithAuditService(auditService *AuditService) *TeamService {
	s.auditService = auditService
	return s
}

func (s *TeamService) CreateTeam(userID uint, organizationID uuid.UUID, req dto.CreateTeamRequest) (*dto.TeamResponse, error) {
	name, description, err := normalizeTeamInput(req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	exists, err := s.teamRepo.ExistsByNameInOrganization(name, organizationID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrTeamAlreadyExists
	}

	team := &models.Team{
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
	}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogCreate(userID, organizationID, "team", team.ID.String(), nil, nil)
	}

	response := mapTeamResponse(*team)
	return &response, nil
}

func (s *TeamService) ListTeams(organizationID uuid.UUID, req *models.PaginationRequest) (*dto.TeamListResponse, error) {
	items, total, err := s.teamRepo.List(req, organizationID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TeamResponse, 0, len(items))
	for _, team := range items {
		responses = append(responses, mapTeamResponse(team))
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &dto.TeamListResponse{
		Items:      responses,
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *TeamService) GetTeamByID(id, organizationID uuid.UUID) (*dto.TeamResponse, error) {
	team, err := s.getOwnedTeam(id, organizationID)
	if err != nil {
		return nil, err
	}

	response := mapTeamResponse(*team)
	return &response, nil
}

func (s *TeamService) UpdateTeam(actorID uint, id, organizationID uuid.UUID, req dto.UpdateTeamRequest) (*dto.TeamResponse, error) {
	team, err := s.getOwnedTeam(id, organizationID)
	if err != nil {
		return nil, err
	}

	name, description, err := normalizeTeamInput(req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	if team.Name != name {
		exists, err := s.teamRepo.ExistsByNameInOrganization(name, organizationID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperrors.ErrTeamAlreadyExists
		}
	}

	previousName := team.Name
	previousDescription := team.Description
	team.Name = name
	team.Description = description

	if err := s.teamRepo.Update(team); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		if previousName != name {
			_ = s.auditService.LogUpdate(actorID, organizationID, "team", team.ID.String(), nil, nil, "name", previousName, name)
		}
		if previousDescription != description {
			_ = s.auditService.LogUpdate(actorID, organizationID, "team", team.ID.String(), nil, nil, "description", previousDescription, description)
		}
	}

	response := mapTeamResponse(*team)
	return &response, nil
}

func (s *TeamService) DeleteTeam(actorID uint, id, organizationID uuid.UUID) error {
	team, err := s.getOwnedTeam(id, organizationID)
	if err != nil {
		return err
	}

	if err := s.teamMemberRepo.RemoveAllMembers(team.ID); err != nil {
		return err
	}

	if err := s.teamRepo.Delete(team.ID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogDelete(actorID, organizationID, "team", team.ID.String(), nil, nil)
	}

	return nil
}

func (s *TeamService) AddMember(actorID uint, teamID, organizationID uuid.UUID, userID uint) (*dto.TeamMemberResponse, error) {
	team, err := s.getOwnedTeam(teamID, organizationID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}

	if user.OrganizationID == nil || *user.OrganizationID != team.OrganizationID {
		return nil, apperrors.ErrTeamForbidden
	}

	alreadyMember, err := s.teamMemberRepo.IsMember(team.ID, user.ID)
	if err != nil {
		return nil, err
	}
	if alreadyMember {
		return nil, apperrors.ErrTeamMemberAlreadyExists
	}

	member := &models.TeamMember{
		TeamID: team.ID,
		UserID: user.ID,
	}
	if err := s.teamMemberRepo.AddMember(member); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         actorID,
			OrganizationID: organizationID,
			EntityType:     "team_member",
			EntityID:       member.ID.String(),
			Action:         models.AuditActionCreate,
			AfterState:     marshalAuditState(map[string]any{"teamId": team.ID.String(), "userId": user.ID, "email": user.Email}),
		})
	}

	response := dto.TeamMemberResponse{
		ID:        member.ID.String(),
		TeamID:    member.TeamID.String(),
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: member.CreatedAt,
	}

	return &response, nil
}

func (s *TeamService) RemoveMember(actorID uint, teamID, organizationID uuid.UUID, userID uint) error {
	team, err := s.getOwnedTeam(teamID, organizationID)
	if err != nil {
		return err
	}

	isMember, err := s.teamMemberRepo.IsMember(team.ID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return apperrors.ErrTeamMemberNotFound
	}

	if err := s.teamMemberRepo.RemoveMember(team.ID, userID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogEvent(AuditEventInput{
			UserID:         actorID,
			OrganizationID: organizationID,
			EntityType:     "team_member",
			EntityID:       team.ID.String(),
			Action:         models.AuditActionDelete,
			BeforeState:    marshalAuditState(map[string]any{"teamId": team.ID.String(), "userId": userID}),
		})
	}

	return nil
}

func (s *TeamService) ListMembers(teamID, organizationID uuid.UUID) (*dto.TeamMemberListResponse, error) {
	team, err := s.getOwnedTeam(teamID, organizationID)
	if err != nil {
		return nil, err
	}

	items, err := s.teamMemberRepo.ListMembers(team.ID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TeamMemberResponse, 0, len(items))
	for _, member := range items {
		name := ""
		email := ""
		if member.User != nil {
			name = member.User.Name
			email = member.User.Email
		}
		responses = append(responses, dto.TeamMemberResponse{
			ID:        member.ID.String(),
			TeamID:    member.TeamID.String(),
			UserID:    member.UserID,
			Name:      name,
			Email:     email,
			CreatedAt: member.CreatedAt,
		})
	}

	count, err := s.teamMemberRepo.CountMembers(team.ID)
	if err != nil {
		return nil, err
	}

	return &dto.TeamMemberListResponse{Items: responses, Total: count}, nil
}

func (s *TeamService) getOwnedTeam(id, organizationID uuid.UUID) (*models.Team, error) {
	team, err := s.teamRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, err
	}

	if team.OrganizationID != organizationID {
		return nil, apperrors.ErrTeamForbidden
	}

	return team, nil
}

func normalizeTeamInput(name, description string) (string, string, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if length := utf8.RuneCountInString(name); length < 3 || length > 100 {
		return "", "", apperrors.ErrInvalidTeamName
	}
	if utf8.RuneCountInString(description) > 500 {
		return "", "", apperrors.ErrInvalidTeamDescription
	}

	return name, description, nil
}

func mapTeamResponse(team models.Team) dto.TeamResponse {
	return dto.TeamResponse{
		ID:             team.ID.String(),
		OrganizationID: team.OrganizationID.String(),
		Name:           team.Name,
		Description:    team.Description,
		CreatedAt:      team.CreatedAt,
		UpdatedAt:      team.UpdatedAt,
	}
}
