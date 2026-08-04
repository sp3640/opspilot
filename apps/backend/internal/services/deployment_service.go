package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

var namespaceRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

type DeploymentService struct {
	repo            *repository.DeploymentRepository
	applicationRepo *repository.ApplicationRepository
	projectRepo     *repository.ProjectRepository
	clusterRepo     *repository.ClusterRepository
}

func NewDeploymentService(
	repo *repository.DeploymentRepository,
	applicationRepo *repository.ApplicationRepository,
	projectRepo *repository.ProjectRepository,
	clusterRepo *repository.ClusterRepository,
) *DeploymentService {
	return &DeploymentService{
		repo:            repo,
		applicationRepo: applicationRepo,
		projectRepo:     projectRepo,
		clusterRepo:     clusterRepo,
	}
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, organizationID uuid.UUID, userID uint, req dto.CreateDeploymentRequest) (*dto.DeploymentResponse, error) {
	_ = ctx

	applicationID, err := uuid.Parse(strings.TrimSpace(req.ApplicationID))
	if err != nil {
		return nil, apperrors.ErrInvalidApplication
	}
	projectID, err := uuid.Parse(strings.TrimSpace(req.ProjectID))
	if err != nil {
		return nil, apperrors.ErrInvalidProject
	}
	clusterID, err := uuid.Parse(strings.TrimSpace(req.TargetClusterID))
	if err != nil {
		return nil, apperrors.ErrClusterNotFound
	}

	image := strings.TrimSpace(req.Image)
	if image == "" {
		return nil, apperrors.ErrInvalidDeploymentImage
	}
	if req.ReplicaCount <= 0 {
		return nil, apperrors.ErrInvalidDeploymentReplica
	}

	strategy, ok := constants.NormalizeDeploymentStrategy(req.DeploymentStrategy)
	if !ok {
		return nil, apperrors.ErrInvalidDeploymentStrategy
	}
	environment, ok := constants.NormalizeDeploymentEnvironment(req.Environment)
	if !ok {
		return nil, apperrors.ErrInvalidDeploymentEnvironment
	}
	namespace := strings.TrimSpace(req.Namespace)
	if !isValidNamespace(namespace) {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}

	if _, err := s.projectRepo.GetByIDAndOrganizationID(projectID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectForbidden
		}
		return nil, err
	}

	application, err := s.applicationRepo.GetApplication(applicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidApplication
		}
		return nil, err
	}
	if application.ProjectID != projectID {
		return nil, apperrors.ErrInvalidApplication
	}

	if _, err := s.clusterRepo.FindByID(clusterID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	deployment := &models.Deployment{
		ApplicationID:      applicationID,
		ProjectID:          projectID,
		OrganizationID:     organizationID,
		Image:              image,
		ImageTag:           strings.TrimSpace(req.ImageTag),
		Environment:        environment,
		Namespace:          namespace,
		ReplicaCount:       req.ReplicaCount,
		Status:             constants.DeploymentStatusPending,
		DeploymentStrategy: strategy,
		TargetClusterID:    clusterID,
		CreatedBy:          userID,
		UpdatedBy:          userID,
	}

	if err := s.repo.Create(deployment); err != nil {
		return nil, err
	}

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) UpdateDeployment(ctx context.Context, id, organizationID uuid.UUID, userID uint, req dto.UpdateDeploymentRequest) (*dto.DeploymentResponse, error) {
	_ = ctx
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, err
	}

	if req.Image != nil {
		image := strings.TrimSpace(*req.Image)
		if image == "" {
			return nil, apperrors.ErrInvalidDeploymentImage
		}
		deployment.Image = image
	}
	if req.ImageTag != nil {
		deployment.ImageTag = strings.TrimSpace(*req.ImageTag)
	}
	if req.Environment != nil {
		environment, ok := constants.NormalizeDeploymentEnvironment(*req.Environment)
		if !ok {
			return nil, apperrors.ErrInvalidDeploymentEnvironment
		}
		deployment.Environment = environment
	}
	if req.Namespace != nil {
		namespace := strings.TrimSpace(*req.Namespace)
		if !isValidNamespace(namespace) {
			return nil, apperrors.ErrInvalidDeploymentNamespace
		}
		deployment.Namespace = namespace
	}
	if req.ReplicaCount != nil {
		if *req.ReplicaCount <= 0 {
			return nil, apperrors.ErrInvalidDeploymentReplica
		}
		deployment.ReplicaCount = *req.ReplicaCount
	}
	if req.DeploymentStrategy != nil {
		strategy, ok := constants.NormalizeDeploymentStrategy(*req.DeploymentStrategy)
		if !ok {
			return nil, apperrors.ErrInvalidDeploymentStrategy
		}
		deployment.DeploymentStrategy = strategy
	}
	if req.TargetClusterID != nil {
		clusterID, err := uuid.Parse(strings.TrimSpace(*req.TargetClusterID))
		if err != nil {
			return nil, apperrors.ErrClusterNotFound
		}
		if _, err := s.clusterRepo.FindByID(clusterID, organizationID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.ErrClusterNotFound
			}
			return nil, err
		}
		deployment.TargetClusterID = clusterID
	}

	deployment.UpdatedBy = userID
	if err := s.repo.Update(deployment); err != nil {
		return nil, err
	}

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) CancelDeployment(ctx context.Context, id, organizationID uuid.UUID, userID uint) (*dto.DeploymentResponse, error) {
	_ = ctx
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, err
	}

	completedAt := time.Now().UTC()
	if err := s.repo.UpdateStatus(deployment.ID, organizationID, constants.DeploymentStatusCancelled, nil, &completedAt, userID); err != nil {
		return nil, err
	}

	deployment.Status = constants.DeploymentStatusCancelled
	deployment.CompletedAt = &completedAt
	deployment.UpdatedBy = userID
	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) DeleteDeployment(ctx context.Context, id, organizationID uuid.UUID) error {
	_ = ctx
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return err
	}

	return s.repo.Delete(deployment.ID, organizationID)
}

func (s *DeploymentService) GetDeployment(id, organizationID uuid.UUID) (*dto.DeploymentResponse, error) {
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, err
	}

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) ListDeployments(organizationID uuid.UUID, req *models.PaginationRequest) (*dto.DeploymentListResponse, error) {
	items, total, err := s.repo.ListByOrganization(organizationID, req)
	if err != nil {
		return nil, err
	}

	return toDeploymentListResponse(items, total, req), nil
}

func (s *DeploymentService) ListApplicationDeployments(applicationID, organizationID uuid.UUID, req *models.PaginationRequest) (*dto.DeploymentListResponse, error) {
	if _, err := s.applicationRepo.GetApplication(applicationID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrApplicationForbidden
		}
		return nil, err
	}

	items, total, err := s.repo.ListByApplication(applicationID, organizationID, req)
	if err != nil {
		return nil, err
	}

	return toDeploymentListResponse(items, total, req), nil
}

func (s *DeploymentService) ListProjectDeployments(projectID, organizationID uuid.UUID, req *models.PaginationRequest) (*dto.DeploymentListResponse, error) {
	if _, err := s.projectRepo.GetByIDAndOrganizationID(projectID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectForbidden
		}
		return nil, err
	}

	items, total, err := s.repo.ListByProject(projectID, organizationID, req)
	if err != nil {
		return nil, err
	}

	return toDeploymentListResponse(items, total, req), nil
}

func (s *DeploymentService) UpdateDeploymentStatus(ctx context.Context, id, organizationID uuid.UUID, userID uint, status string) (*dto.DeploymentResponse, error) {
	_ = ctx
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, err
	}

	normalizedStatus, ok := constants.NormalizeDeploymentStatus(status)
	if !ok {
		return nil, apperrors.ErrInvalidDeploymentStatus
	}

	var startedAt *time.Time
	var completedAt *time.Time
	now := time.Now().UTC()
	if normalizedStatus == constants.DeploymentStatusRunning {
		startedAt = &now
	}
	if normalizedStatus == constants.DeploymentStatusSucceeded ||
		normalizedStatus == constants.DeploymentStatusFailed ||
		normalizedStatus == constants.DeploymentStatusCancelled ||
		normalizedStatus == constants.DeploymentStatusRolledBack {
		completedAt = &now
	}

	if err := s.repo.UpdateStatus(deployment.ID, organizationID, normalizedStatus, startedAt, completedAt, userID); err != nil {
		return nil, err
	}

	deployment.Status = normalizedStatus
	deployment.UpdatedBy = userID
	if startedAt != nil {
		deployment.StartedAt = startedAt
	}
	if completedAt != nil {
		deployment.CompletedAt = completedAt
	}

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) GetLatestDeployment(applicationID, organizationID uuid.UUID) (*dto.DeploymentResponse, error) {
	if _, err := s.applicationRepo.GetApplication(applicationID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrApplicationNotFound
		}
		return nil, err
	}

	deployment, err := s.repo.GetLatestDeployment(applicationID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentNotFound
		}
		return nil, err
	}

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) getOwnedDeployment(id, organizationID uuid.UUID) (*models.Deployment, error) {
	deployment, err := s.repo.GetByID(id, organizationID)
	if err == nil {
		return deployment, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if deployment, anyErr := s.repo.GetByIDAnyOrganization(id); anyErr == nil {
			if deployment.OrganizationID == organizationID {
				return nil, apperrors.ErrDeploymentNotFound
			}
			return nil, apperrors.ErrDeploymentForbidden
		}
		return nil, apperrors.ErrDeploymentNotFound
	}

	return nil, err
}

func toDeploymentListResponse(items []models.Deployment, total int64, req *models.PaginationRequest) *dto.DeploymentListResponse {
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.DeploymentListResponse{
		Items:      mapper.MapDeployments(items),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func isValidNamespace(namespace string) bool {
	if len(namespace) == 0 || len(namespace) > 63 {
		return false
	}

	return namespaceRegex.MatchString(namespace)
}
