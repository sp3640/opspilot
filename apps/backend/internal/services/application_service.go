package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/utils"
	"gorm.io/gorm"
)

type ApplicationService struct {
	repo        *repository.ApplicationRepository
	projectRepo *repository.ProjectRepository
}

func NewApplicationService(repo *repository.ApplicationRepository, projectRepo *repository.ProjectRepository) *ApplicationService {
	return &ApplicationService{repo: repo, projectRepo: projectRepo}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, projectID uuid.UUID, organizationID uuid.UUID, req dto.CreateApplicationRequest) (*dto.ApplicationResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrInvalidApplicationName
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = utils.GenerateSlug(name)
	} else {
		slug = utils.GenerateSlug(slug)
	}
	if slug == "" {
		return nil, apperrors.ErrInvalidApplicationSlug
	}

	description := strings.TrimSpace(req.Description)
	repositoryURL := strings.TrimSpace(req.RepositoryURL)
	defaultBranch := strings.TrimSpace(req.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = constants.ApplicationDefaultBranch
	}
	runtime, ok := constants.NormalizeApplicationRuntime(req.Runtime)
	if !ok {
		return nil, apperrors.ErrInvalidApplicationRuntime
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = constants.ApplicationStatusDraft
	} else {
		var valid bool
		status, valid = constants.NormalizeApplicationStatus(status)
		if !valid {
			return nil, apperrors.ErrInvalidApplicationStatus
		}
	}
	if req.Port < 1 || req.Port > 65535 {
		return nil, apperrors.ErrInvalidApplicationPort
	}

	if _, err := s.getProjectForOrganization(projectID, organizationID); err != nil {
		return nil, err
	}

	if existing, err := s.repo.GetBySlug(projectID, organizationID, slug); err == nil && existing != nil {
		return nil, apperrors.ErrApplicationAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	application := &models.Application{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Name:           name,
		Slug:           slug,
		Description:    description,
		RepositoryURL:  repositoryURL,
		DefaultBranch:  defaultBranch,
		Runtime:        runtime,
		BuildCommand:   strings.TrimSpace(req.BuildCommand),
		StartCommand:   strings.TrimSpace(req.StartCommand),
		Port:           req.Port,
		Environment:    strings.TrimSpace(req.Environment),
		Status:         status,
	}

	if err := s.repo.CreateApplication(application); err != nil {
		return nil, err
	}

	response := mapper.MapApplication(*application)
	return &response, nil
}

func (s *ApplicationService) UpdateApplication(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req dto.UpdateApplicationRequest) (*dto.ApplicationResponse, error) {
	application, err := s.getOwnedApplication(id, organizationID)
	if err != nil {
		return nil, err
	}

	previousSlug := application.Slug

	if name := strings.TrimSpace(req.Name); name != "" {
		application.Name = name
	}
	if slug := strings.TrimSpace(req.Slug); slug != "" {
		application.Slug = utils.GenerateSlug(slug)
		if application.Slug == "" {
			return nil, apperrors.ErrInvalidApplicationSlug
		}
	}
	if description := strings.TrimSpace(req.Description); description != "" {
		application.Description = description
	}
	if repositoryURL := strings.TrimSpace(req.RepositoryURL); repositoryURL != "" {
		application.RepositoryURL = repositoryURL
	}
	if defaultBranch := strings.TrimSpace(req.DefaultBranch); defaultBranch != "" {
		application.DefaultBranch = defaultBranch
	}
	if runtime := strings.TrimSpace(req.Runtime); runtime != "" {
		resolvedRuntime, ok := constants.NormalizeApplicationRuntime(runtime)
		if !ok {
			return nil, apperrors.ErrInvalidApplicationRuntime
		}
		application.Runtime = resolvedRuntime
	}
	if buildCommand := strings.TrimSpace(req.BuildCommand); buildCommand != "" {
		application.BuildCommand = buildCommand
	}
	if startCommand := strings.TrimSpace(req.StartCommand); startCommand != "" {
		application.StartCommand = startCommand
	}
	if req.Port != 0 {
		if req.Port < 1 || req.Port > 65535 {
			return nil, apperrors.ErrInvalidApplicationPort
		}
		application.Port = req.Port
	}
	if environment := strings.TrimSpace(req.Environment); environment != "" {
		application.Environment = environment
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		resolvedStatus, ok := constants.NormalizeApplicationStatus(status)
		if !ok {
			return nil, apperrors.ErrInvalidApplicationStatus
		}
		application.Status = resolvedStatus
	}

	if application.DefaultBranch == "" {
		application.DefaultBranch = constants.ApplicationDefaultBranch
	}
	if application.Status == "" {
		application.Status = constants.ApplicationStatusDraft
	}

	if application.Slug != previousSlug {
		if existing, err := s.repo.GetBySlug(application.ProjectID, organizationID, application.Slug); err == nil && existing != nil && existing.ID != application.ID {
			return nil, apperrors.ErrApplicationAlreadyExists
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if err := s.repo.UpdateApplication(application); err != nil {
		return nil, err
	}

	response := mapper.MapApplication(*application)
	return &response, nil
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	application, err := s.getOwnedApplication(id, organizationID)
	if err != nil {
		return err
	}

	return s.repo.DeleteApplication(application.ID, organizationID)
}

func (s *ApplicationService) GetApplication(id uuid.UUID, organizationID uuid.UUID) (*dto.ApplicationResponse, error) {
	application, err := s.getOwnedApplication(id, organizationID)
	if err != nil {
		return nil, err
	}

	response := mapper.MapApplication(*application)
	return &response, nil
}

func (s *ApplicationService) ListApplicationsByProject(projectID uuid.UUID, organizationID uuid.UUID, req *models.PaginationRequest) (*dto.ApplicationListResponse, error) {
	if _, err := s.getProjectForOrganization(projectID, organizationID); err != nil {
		return nil, err
	}

	applications, total, err := s.repo.ListApplicationsByProject(projectID, organizationID, req)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.ApplicationListResponse{
		Items:      mapper.MapApplications(applications),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *ApplicationService) getOwnedApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
	application, err := s.repo.GetApplication(id, organizationID)
	if err == nil {
		return application, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if existing, anyErr := s.repo.FindByIDAnyOrganization(id); anyErr == nil && existing != nil {
			return nil, apperrors.ErrApplicationForbidden
		}
		return nil, apperrors.ErrApplicationNotFound
	}

	return nil, err
}

func (s *ApplicationService) getProjectForOrganization(projectID uuid.UUID, organizationID uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.GetByIDAndOrganizationID(projectID, organizationID)
	if err == nil {
		return project, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrProjectForbidden
	}

	return nil, err
}
