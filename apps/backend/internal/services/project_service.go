package services

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/utils"
	"gorm.io/gorm"
)

type ProjectService struct {
	repo      *repository.ProjectRepository
	userRepo  *repository.UserRepository
	auditRepo *AuditService
}

func NewProjectService(repo *repository.ProjectRepository, userRepo *repository.UserRepository, auditService *AuditService) *ProjectService {
	return &ProjectService{
		repo:      repo,
		userRepo:  userRepo,
		auditRepo: auditService,
	}
}

// Create persists a new project and returns its full DTO representation.
func (s *ProjectService) Create(ctx context.Context, name, description string, userID uint, organizationID uuid.UUID) (*dto.ProjectResponse, error) {
	name, description, err := normalizeProjectInput(name, description)
	if err != nil {
		return nil, err
	}

	project := &models.Project{
		Name:           name,
		Description:    description,
		Slug:           utils.GenerateSlug(name),
		OwnerID:        userID,
		OrganizationID: organizationID,
	}

	if err := s.repo.Create(project); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogCreate(userID, organizationID, "project", project.ID.String(), &project.ID, nil); err != nil {
			logAuditFailure(ctx, "create", "project", uint(0), err)
		}
	}

	response, err := s.mapProject(*project)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// ListMyProjects returns a paginated list of projects owned by the user.
func (s *ProjectService) ListMyProjects(userID uint, organizationID uuid.UUID, req *models.PaginationRequest) (*dto.ProjectListResponse, error) {
	items, total, err := s.repo.ListByOrganizationID(req, organizationID)
	if err != nil {
		return nil, err
	}
	owner, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &dto.ProjectListResponse{
		Items:      mapper.MapProjects(items, *owner),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetProjectByID returns a single project the user owns, as a DTO.
func (s *ProjectService) GetProjectByID(id uuid.UUID, organizationID uuid.UUID) (*dto.ProjectResponse, error) {
	project, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, err
	}

	response, err := s.mapProject(*project)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// UpdateProject applies name/description changes and returns the updated DTO.
func (s *ProjectService) UpdateProject(ctx context.Context, id uuid.UUID, userID uint, organizationID uuid.UUID, name, description string) (*dto.ProjectResponse, error) {
	name, description, err := normalizeProjectInput(name, description)
	if err != nil {
		return nil, err
	}

	project, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, err
	}

	previousName := project.Name
	previousDescription := project.Description

	project.Name = name
	project.Description = description

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		entityIDStr := project.ID.String()
		if previousName != name {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "project", entityIDStr, &project.ID, nil, "name", previousName, name); err != nil {
				logAuditFailure(ctx, "update", "project", uint(0), err)
			}
		}
		if previousDescription != description {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "project", entityIDStr, &project.ID, nil, "description", previousDescription, description); err != nil {
				logAuditFailure(ctx, "update", "project", uint(0), err)
			}
		}
	}

	response, err := s.mapProject(*project)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteProject soft-deletes a project owned by the user.
func (s *ProjectService) DeleteProject(ctx context.Context, id uuid.UUID, userID uint, organizationID uuid.UUID) error {
	project, err := s.repo.GetByIDAndOrganizationID(id, organizationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return err
	}

	if err := s.repo.Delete(project.ID, organizationID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogDelete(userID, organizationID, "project", project.ID.String(), &project.ID, nil); err != nil {
			logAuditFailure(ctx, "delete", "project", uint(0), err)
		}
	}

	return nil
}

func (s *ProjectService) mapProject(project models.Project) (dto.ProjectResponse, error) {
	owner, err := s.userRepo.GetByID(project.OwnerID)
	if err != nil {
		return dto.ProjectResponse{}, err
	}

	return mapper.MapProject(project, *owner), nil
}

func normalizeProjectInput(name, description string) (string, string, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if length := utf8.RuneCountInString(name); length < 3 || length > 100 {
		return "", "", apperrors.ErrInvalidProjectName
	}
	if utf8.RuneCountInString(description) > 300 {
		return "", "", apperrors.ErrInvalidProjectDescription
	}

	return name, description, nil
}
