package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"gorm.io/gorm"
)

var namespaceRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

type DeploymentExecutor interface {
	ExecuteDeployment(ctx context.Context, deploymentID, organizationID uuid.UUID, userID uint) (*models.Deployment, error)
}

type DeploymentService struct {
	repo            *repository.DeploymentRepository
	applicationRepo *repository.ApplicationRepository
	projectRepo     *repository.ProjectRepository
	clusterRepo     *repository.ClusterRepository
	historyService  *DeploymentHistoryService
	executor        DeploymentExecutor
	auditService    *AuditService
}

func NewDeploymentService(
	repo *repository.DeploymentRepository,
	applicationRepo *repository.ApplicationRepository,
	projectRepo *repository.ProjectRepository,
	clusterRepo *repository.ClusterRepository,
	historyService *DeploymentHistoryService,
) *DeploymentService {
	return &DeploymentService{
		repo:            repo,
		applicationRepo: applicationRepo,
		projectRepo:     projectRepo,
		clusterRepo:     clusterRepo,
		historyService:  historyService,
	}
}

func (s *DeploymentService) WithExecutor(executor DeploymentExecutor) *DeploymentService {
	s.executor = executor
	return s
}

func (s *DeploymentService) WithAuditService(auditService *AuditService) *DeploymentService {
	s.auditService = auditService
	return s
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
		CommitSHA:          normalizeOptionalString(req.CommitSHA),
		Author:             normalizeOptionalString(req.Author),
		CreatedBy:          userID,
		UpdatedBy:          userID,
	}

	if err := s.repo.Create(deployment); err != nil {
		return nil, err
	}
	if err := s.historyService.CreateHistoryFromDeployment(deployment, "Deployment created", userID); err != nil {
		return nil, err
	}
	s.logAuditEvent(ctx, deployment, userID, models.AuditActionCreate, "", deploymentAuditSnapshot(deployment))

	if s.executor != nil {
		executedDeployment, err := s.executor.ExecuteDeployment(ctx, deployment.ID, organizationID, userID)
		if err != nil {
			return nil, err
		}
		deployment = executedDeployment
	}

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) UpdateDeployment(ctx context.Context, id, organizationID uuid.UUID, userID uint, req dto.UpdateDeploymentRequest) (*dto.DeploymentResponse, error) {
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, err
	}
	beforeState := deploymentAuditSnapshot(deployment)

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
	if req.CommitSHA != nil {
		deployment.CommitSHA = normalizeOptionalString(req.CommitSHA)
	}
	if req.Author != nil {
		deployment.Author = normalizeOptionalString(req.Author)
	}

	deployment.UpdatedBy = userID
	if err := s.repo.Update(deployment); err != nil {
		return nil, err
	}
	if err := s.historyService.CreateHistoryFromDeployment(deployment, "Deployment updated", userID); err != nil {
		return nil, err
	}
	s.logAuditEvent(ctx, deployment, userID, models.AuditActionUpdate, beforeState, deploymentAuditSnapshot(deployment))

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) CancelDeployment(ctx context.Context, id, organizationID uuid.UUID, userID uint) (*dto.DeploymentResponse, error) {
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, err
	}
	previousStatus := deployment.Status

	completedAt := time.Now().UTC()
	if err := s.repo.UpdateStatus(deployment.ID, organizationID, constants.DeploymentStatusCancelled, nil, &completedAt, userID); err != nil {
		return nil, err
	}

	deployment.Status = constants.DeploymentStatusCancelled
	deployment.CompletedAt = &completedAt
	deployment.UpdatedBy = userID
	if err := s.historyService.CreateHistoryFromDeployment(deployment, "Deployment cancelled", userID); err != nil {
		return nil, err
	}
	// Cancel never touches the live cluster (there is no "stop a running
	// apply" primitive to reuse) - this audit entry records the DB-level
	// action taken, same as the executor's own status-change entries do.
	s.logAuditBestEffort(ctx, deployment, userID, "status", previousStatus, constants.DeploymentStatusCancelled)

	response := mapper.MapDeployment(*deployment)
	return &response, nil
}

func (s *DeploymentService) DeleteDeployment(ctx context.Context, id, organizationID uuid.UUID, userID uint) error {
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(deployment.ID, organizationID); err != nil {
		return err
	}

	s.logAuditEvent(ctx, deployment, userID, models.AuditActionDelete, deploymentAuditSnapshot(deployment), "")

	return nil
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
	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, err
	}
	previousStatus := deployment.Status

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

	if normalizedStatus == constants.DeploymentStatusSucceeded ||
		normalizedStatus == constants.DeploymentStatusFailed ||
		normalizedStatus == constants.DeploymentStatusRolledBack ||
		normalizedStatus == constants.DeploymentStatusCancelled {
		if err := s.historyService.CreateHistoryFromDeployment(deployment, "Deployment status changed to "+normalizedStatus, userID); err != nil {
			return nil, err
		}
	}
	if previousStatus != normalizedStatus {
		s.logAuditBestEffort(ctx, deployment, userID, "status", previousStatus, normalizedStatus)
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

func (s *DeploymentService) RollbackDeployment(ctx context.Context, id, organizationID uuid.UUID, userID uint, revision int) (*dto.RollbackDeploymentResponse, error) {
	deployment, sourceHistory, latestRevision, err := s.ValidateRollback(id, organizationID, revision)
	if err != nil {
		return nil, err
	}

	if err := s.RollbackToRevision(deployment, sourceHistory, userID); err != nil {
		return nil, err
	}
	s.logAuditBestEffort(ctx, deployment, userID, "revision", strconv.Itoa(latestRevision), fmt.Sprintf("rolled back to revision %d", sourceHistory.Revision))

	// Reuse the same executor Create already uses, so a rollback genuinely
	// re-applies the restored configuration to the live cluster rather than
	// only resetting the database record.
	if s.executor != nil {
		executedDeployment, err := s.executor.ExecuteDeployment(ctx, deployment.ID, organizationID, userID)
		if err != nil {
			return nil, err
		}
		deployment = executedDeployment
	}

	response := mapper.MapDeployment(*deployment)
	return &dto.RollbackDeploymentResponse{
		Deployment:             response,
		CurrentRevision:        latestRevision + 1,
		RollbackSourceRevision: sourceHistory.Revision,
	}, nil
}

func (s *DeploymentService) RollbackToRevision(deployment *models.Deployment, sourceHistory *models.DeploymentHistory, userID uint) error {
	deployment.Image = sourceHistory.Image
	deployment.ImageTag = sourceHistory.ImageTag
	deployment.ReplicaCount = sourceHistory.ReplicaCount
	deployment.Namespace = sourceHistory.Namespace
	deployment.Environment = sourceHistory.Environment
	deployment.DeploymentStrategy = sourceHistory.DeploymentStrategy
	deployment.CommitSHA = sourceHistory.CommitSHA
	deployment.Author = sourceHistory.Author
	deployment.Status = constants.DeploymentStatusPending
	deployment.StartedAt = nil
	deployment.CompletedAt = nil
	deployment.UpdatedBy = userID

	if err := s.repo.ApplyRollback(deployment); err != nil {
		return err
	}

	changeSummary := fmt.Sprintf("Deployment rolled back to revision %d", sourceHistory.Revision)
	if err := s.historyService.CreateHistoryFromDeployment(deployment, changeSummary, userID); err != nil {
		return err
	}

	return nil
}

func (s *DeploymentService) ValidateRollback(id, organizationID uuid.UUID, revision int) (*models.Deployment, *models.DeploymentHistory, int, error) {
	if revision <= 0 {
		return nil, nil, 0, apperrors.ErrInvalidDeploymentRevision
	}

	deployment, err := s.getOwnedDeployment(id, organizationID)
	if err != nil {
		return nil, nil, 0, err
	}

	sourceHistory, err := s.historyService.GetRevisionForDeployment(id, organizationID, revision)
	if err != nil {
		return nil, nil, 0, err
	}

	latestRevision, err := s.historyService.GetLatestRevisionForDeployment(id, organizationID)
	if err != nil {
		return nil, nil, 0, err
	}
	if revision == latestRevision {
		return nil, nil, 0, apperrors.ErrDeploymentRollbackLatest
	}

	if sourceHistory.DeploymentID != deployment.ID {
		return nil, nil, 0, apperrors.ErrDeploymentHistoryNotFound
	}
	if sourceHistory.OrganizationID != organizationID {
		return nil, nil, 0, apperrors.ErrDeploymentForbidden
	}

	return deployment, sourceHistory, latestRevision, nil
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

// ListAuditLogs returns the deployment's generic audit trail (the same
// AuditLog infrastructure Incidents/Alerts/Projects already use), covering
// every remediation action recorded against it: status changes made by the
// executor (create/rollback execution) and the explicit entries Cancel/
// Rollback write directly. Read-only, so no additional permission beyond
// organization membership + deployment ownership is required.
func (s *DeploymentService) ListAuditLogs(organizationID, id uuid.UUID, req *models.PaginationRequest) (*models.PaginationResponse, error) {
	if _, err := s.getOwnedDeployment(id, organizationID); err != nil {
		return nil, err
	}
	if s.auditService == nil {
		return &models.PaginationResponse{Page: req.Page, Limit: req.Limit, Items: []models.AuditLog{}}, nil
	}

	return s.auditService.ListEntityAuditLogs(organizationID, "deployment", id.String(), req)
}

// logAuditBestEffort records a remediation action in the generic audit log.
// Best-effort: a logging failure is recorded but never fails the
// remediation action itself, matching the executor's own
// logStatusAuditBestEffort convention.
func (s *DeploymentService) logAuditBestEffort(ctx context.Context, deployment *models.Deployment, userID uint, fieldName, oldValue, newValue string) {
	if s.auditService == nil {
		return
	}

	if err := s.auditService.LogUpdate(
		userID,
		deployment.OrganizationID,
		"deployment",
		deployment.ID.String(),
		&deployment.ProjectID,
		nil,
		fieldName,
		oldValue,
		newValue,
	); err != nil {
		logger.Error(
			ctx,
			"audit logging failed",
			slog.String("operation", "update"),
			slog.String("entity_type", "deployment"),
			slog.String("entity_id", deployment.ID.String()),
			slog.Any("error", err),
		)
	}
}

// logAuditEvent records a create/update/delete against a deployment,
// including the ApplicationID cross-reference (so the Organization/Project
// Audit views can filter "everything touching this application") and an
// optional before/after snapshot. Best-effort, matching logAuditBestEffort.
func (s *DeploymentService) logAuditEvent(ctx context.Context, deployment *models.Deployment, userID uint, action models.AuditAction, beforeState, afterState string) {
	if s.auditService == nil {
		return
	}

	if err := s.auditService.LogEvent(AuditEventInput{
		UserID:         userID,
		OrganizationID: deployment.OrganizationID,
		ProjectID:      &deployment.ProjectID,
		ApplicationID:  &deployment.ApplicationID,
		EntityType:     "deployment",
		EntityID:       deployment.ID.String(),
		Action:         action,
		BeforeState:    beforeState,
		AfterState:     afterState,
	}); err != nil {
		logger.Error(
			ctx,
			"audit logging failed",
			slog.String("operation", string(action)),
			slog.String("entity_type", "deployment"),
			slog.String("entity_id", deployment.ID.String()),
			slog.Any("error", err),
		)
	}
}

// deploymentAuditSnapshot builds the safe subset of fields worth recording
// in a deployment's audit before/after state. Excludes nothing sensitive -
// deployments carry no secrets - but stays limited to fields a reviewer
// would actually care about, not the full row.
func deploymentAuditSnapshot(deployment *models.Deployment) string {
	return marshalAuditState(map[string]any{
		"image":              deployment.Image,
		"imageTag":           deployment.ImageTag,
		"environment":        deployment.Environment,
		"namespace":          deployment.Namespace,
		"replicaCount":       deployment.ReplicaCount,
		"status":             deployment.Status,
		"deploymentStrategy": deployment.DeploymentStrategy,
		"targetClusterId":    deployment.TargetClusterID,
	})
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

// normalizeOptionalString trims value and returns nil for an absent or
// blank input, so "not available" is represented uniformly as nil rather
// than as an empty string in the database.
func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func isValidNamespace(namespace string) bool {
	if len(namespace) == 0 || len(namespace) > 63 {
		return false
	}

	return namespaceRegex.MatchString(namespace)
}
