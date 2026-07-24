package services

import (
	"errors"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

func (s *ProjectService) Create(name, description string, userID uint) error {

	project := &models.Project{
		Name:        name,
		Description: description,
		UserID:      userID,
	}

	return s.repo.Create(project)
}

func (s *ProjectService) GetMyProjects(userID uint) ([]models.Project, error) {
	return s.repo.GetAllByUserID(userID)
}

func (s *ProjectService) GetProjectByID(id, userID uint) (*models.Project, error) {
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

func (s *ProjectService) UpdateProject(id, userID uint, name, description string) (*models.Project, error) {
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

	project.Name = name
	project.Description = description

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) DeleteProject(id, userID uint) error {
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

	return s.repo.Delete(project.ID)
}
