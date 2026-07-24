package services

import (
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
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