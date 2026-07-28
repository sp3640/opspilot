package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/utils"
	"gorm.io/gorm"
)

type ProjectService struct {
	repo      *repository.ProjectRepository
	auditRepo *AuditService
}

func NewProjectService(repo *repository.ProjectRepository, auditService *AuditService) *ProjectService {
	return &ProjectService{
		repo:      repo,
		auditRepo: auditService,
	}
}

func (s *ProjectService) Create(ctx context.Context, name, description string, userID uint) error {

	project := &models.Project{
		Name:        name,
		Description: description,
		Slug:        utils.GenerateSlug(name),
		OwnerID:     userID,
	}

	if err := s.repo.Create(project); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogCreate(userID, "project", project.ID.String(), &project.ID, nil); err != nil {
			logAuditFailure(ctx, "create", "project", uint(0), err)
		}
	}

	return nil
}

func (s *ProjectService) GetMyProjects(userID uint) ([]models.Project, error) {
	return s.repo.GetAllByUserID(userID)
}

func (s *ProjectService) ListMyProjects(userID uint, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	items, total, err := s.repo.ListByUserID(req, userID)
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

func (s *ProjectService) GetProjectByID(id uuid.UUID, userID uint) (*models.Project, error) {
	project, err := s.repo.GetByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return nil, apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, id uuid.UUID, userID uint, name, description string) (*models.Project, error) {
	project, err := s.repo.GetByIDAndUserID(id, userID)
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
			if err := s.auditRepo.LogUpdate(userID, "project", entityIDStr, &project.ID, nil, "name", previousName, name); err != nil {
				logAuditFailure(ctx, "update", "project", uint(0), err)
			}
		}
		if previousDescription != description {
			if err := s.auditRepo.LogUpdate(userID, "project", entityIDStr, &project.ID, nil, "description", previousDescription, description); err != nil {
				logAuditFailure(ctx, "update", "project", uint(0), err)
			}
		}
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id uuid.UUID, userID uint) error {
	project, err := s.repo.GetByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectForbidden) {
			return apperrors.ErrProjectForbidden
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return err
	}

	if err := s.repo.Delete(project.ID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogDelete(userID, "project", project.ID.String(), &project.ID, nil); err != nil {
			logAuditFailure(ctx, "delete", "project", uint(0), err)
		}
	}

	return nil
}
