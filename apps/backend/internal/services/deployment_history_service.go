package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

type DeploymentHistoryService struct {
	repo           *repository.DeploymentHistoryRepository
	deploymentRepo *repository.DeploymentRepository
}

func NewDeploymentHistoryService(repo *repository.DeploymentHistoryRepository, deploymentRepo *repository.DeploymentRepository) *DeploymentHistoryService {
	return &DeploymentHistoryService{
		repo:           repo,
		deploymentRepo: deploymentRepo,
	}
}

func (s *DeploymentHistoryService) CreateHistoryFromDeployment(deployment *models.Deployment, changeSummary string, triggeredBy uint) error {
	if deployment == nil {
		return fmt.Errorf("deployment is required")
	}

	latestRevision, err := s.repo.GetLatestRevision(deployment.ID, deployment.OrganizationID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	nextRevision := latestRevision + 1
	history := &models.DeploymentHistory{
		DeploymentID:       deployment.ID,
		ApplicationID:      deployment.ApplicationID,
		ProjectID:          deployment.ProjectID,
		OrganizationID:     deployment.OrganizationID,
		Revision:           nextRevision,
		Image:              deployment.Image,
		ImageTag:           deployment.ImageTag,
		Environment:        deployment.Environment,
		Namespace:          deployment.Namespace,
		ReplicaCount:       deployment.ReplicaCount,
		DeploymentStrategy: deployment.DeploymentStrategy,
		Status:             deployment.Status,
		CommitSHA:          deployment.CommitSHA,
		Author:             deployment.Author,
		ChangeSummary:      changeSummary,
		TriggeredBy:        triggeredBy,
	}

	return s.repo.Create(history)
}

func (s *DeploymentHistoryService) ListDeploymentHistory(deploymentID, organizationID uuid.UUID, req *models.PaginationRequest) (*dto.DeploymentHistoryListResponse, error) {
	if _, err := s.getOwnedDeployment(deploymentID, organizationID); err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListDeploymentHistory(deploymentID, organizationID, req)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.DeploymentHistoryListResponse{
		Items:      mapper.MapDeploymentHistoryItems(items),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *DeploymentHistoryService) GetDeploymentHistoryRevision(deploymentID, organizationID uuid.UUID, revision int) (*dto.DeploymentHistoryResponse, error) {
	if revision <= 0 {
		return nil, apperrors.ErrInvalidDeploymentRevision
	}

	if _, err := s.getOwnedDeployment(deploymentID, organizationID); err != nil {
		return nil, err
	}

	history, err := s.repo.GetRevision(deploymentID, organizationID, revision)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentHistoryNotFound
		}
		return nil, err
	}

	response := mapper.MapDeploymentHistory(*history)
	return &response, nil
}

func (s *DeploymentHistoryService) GetLatestRevisionForDeployment(deploymentID, organizationID uuid.UUID) (int, error) {
	if _, err := s.getOwnedDeployment(deploymentID, organizationID); err != nil {
		return 0, err
	}

	revision, err := s.repo.GetLatestRevision(deploymentID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.ErrDeploymentHistoryNotFound
		}
		return 0, err
	}

	return revision, nil
}

func (s *DeploymentHistoryService) GetRevisionForDeployment(deploymentID, organizationID uuid.UUID, revision int) (*models.DeploymentHistory, error) {
	if revision <= 0 {
		return nil, apperrors.ErrInvalidDeploymentRevision
	}

	if _, err := s.getOwnedDeployment(deploymentID, organizationID); err != nil {
		return nil, err
	}

	history, err := s.repo.GetRevision(deploymentID, organizationID, revision)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentHistoryNotFound
		}
		return nil, err
	}

	return history, nil
}

func (s *DeploymentHistoryService) getOwnedDeployment(id, organizationID uuid.UUID) (*models.Deployment, error) {
	deployment, err := s.deploymentRepo.GetByID(id, organizationID)
	if err == nil {
		return deployment, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if deploymentAny, anyErr := s.deploymentRepo.GetByIDAnyOrganization(id); anyErr == nil {
			if deploymentAny.OrganizationID == organizationID {
				return nil, apperrors.ErrDeploymentNotFound
			}
			return nil, apperrors.ErrDeploymentForbidden
		}
		return nil, apperrors.ErrDeploymentNotFound
	}

	return nil, err
}
