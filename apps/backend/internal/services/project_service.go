package services

import (
	"errors"
	"log"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
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

func (s *ProjectService) Create(name, description string, userID uint) error {

	project := &models.Project{
		Name:        name,
		Description: description,
		UserID:      userID,
	}

	if err := s.repo.Create(project); err != nil {
		return err
	}

	if s.auditRepo != nil {
		projectIDValue := project.ID
		projectIDPtr := &projectIDValue
		if err := s.auditRepo.LogCreate(userID, "project", project.ID, projectIDPtr, nil); err != nil {
			log.Printf("audit create failed: %v", err)
		}
	}

	return nil
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

	previousName := project.Name
	previousDescription := project.Description

	project.Name = name
	project.Description = description

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		projectIDValue := project.ID
		projectIDPtr := &projectIDValue
		if previousName != name {
			if err := s.auditRepo.LogUpdate(userID, "project", project.ID, projectIDPtr, nil, "name", previousName, name); err != nil {
				log.Printf("audit update failed: %v", err)
			}
		}
		if previousDescription != description {
			if err := s.auditRepo.LogUpdate(userID, "project", project.ID, projectIDPtr, nil, "description", previousDescription, description); err != nil {
				log.Printf("audit update failed: %v", err)
			}
		}
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

	if err := s.repo.Delete(project.ID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		projectIDValue := project.ID
		projectIDPtr := &projectIDValue
		if err := s.auditRepo.LogDelete(userID, "project", project.ID, projectIDPtr, nil); err != nil {
			log.Printf("audit delete failed: %v", err)
		}
	}

	return nil
}
