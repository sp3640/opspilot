package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/resourcesync"
	"gorm.io/gorm"
)

type ResourceService struct {
	repo       *repository.ResourceRepository
	syncEngine *resourcesync.SyncEngine
	auditRepo  *AuditService
}

func NewResourceService(repo *repository.ResourceRepository, syncEngine *resourcesync.SyncEngine, auditService *AuditService) *ResourceService {
	if syncEngine == nil && repo != nil {
		syncEngine = resourcesync.NewSyncEngine(repo)
	}

	return &ResourceService{
		repo:       repo,
		syncEngine: syncEngine,
		auditRepo:  auditService,
	}
}

func (s *ResourceService) CreateResource(
	ctx context.Context,
	projectID uuid.UUID,
	parentResourceID *uuid.UUID,
	kind,
	name,
	displayName,
	externalID,
	provider,
	region,
	namespace,
	cluster,
	status,
	health string,
	labels,
	annotations,
	metadata json.RawMessage,
	userID uint,
) (*dto.ResourceResponse, error) {
	resource, err := s.buildResourceModel(projectID, parentResourceID, kind, name, displayName, externalID, provider, region, namespace, cluster, status, health, labels, annotations, metadata, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(resource); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogCreate(userID, "resource", resource.ID.String(), &resource.ProjectID, nil); err != nil {
			logAuditFailure(ctx, "create", "resource", 0, err)
		}
	}

	response := mapper.MapResource(*resource)
	return &response, nil
}

func (s *ResourceService) UpdateResource(
	ctx context.Context,
	id uuid.UUID,
	userID uint,
	projectID uuid.UUID,
	parentResourceID *uuid.UUID,
	kind,
	name,
	displayName,
	externalID,
	provider,
	region,
	namespace,
	cluster,
	status,
	health string,
	labels,
	annotations,
	metadata json.RawMessage,
) (*dto.ResourceResponse, error) {
	resource, err := s.getOwnedResource(id, userID)
	if err != nil {
		return nil, err
	}

	updated, err := s.buildResourceModel(projectID, parentResourceID, kind, name, displayName, externalID, provider, region, namespace, cluster, status, health, labels, annotations, metadata, userID)
	if err != nil {
		return nil, err
	}

	previousProjectID := resource.ProjectID
	previousParentResourceID := resource.ParentResourceID
	previousKind := resource.Kind
	previousName := resource.Name
	previousDisplayName := resource.DisplayName
	previousExternalID := resource.ExternalID
	previousProvider := resource.Provider
	previousRegion := resource.Region
	previousNamespace := resource.Namespace
	previousCluster := resource.Cluster
	previousStatus := resource.Status
	previousHealth := resource.Health
	previousLabels := string(resource.Labels)
	previousAnnotations := string(resource.Annotations)
	previousMetadata := string(resource.Metadata)

	resource.ProjectID = updated.ProjectID
	resource.ParentResourceID = updated.ParentResourceID
	resource.Kind = updated.Kind
	resource.Name = updated.Name
	resource.DisplayName = updated.DisplayName
	resource.ExternalID = updated.ExternalID
	resource.Provider = updated.Provider
	resource.Region = updated.Region
	resource.Namespace = updated.Namespace
	resource.Cluster = updated.Cluster
	resource.Status = updated.Status
	resource.Health = updated.Health
	resource.Labels = updated.Labels
	resource.Annotations = updated.Annotations
	resource.Metadata = updated.Metadata

	if err := s.repo.Update(resource); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		entityID := resource.ID.String()
		if previousProjectID != resource.ProjectID {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "project_id", previousProjectID.String(), resource.ProjectID.String()); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if !uuidPointerStringsEqual(previousParentResourceID, resource.ParentResourceID) {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "parent_resource_id", uuidPointerString(previousParentResourceID), uuidPointerString(resource.ParentResourceID)); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousKind != resource.Kind {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "kind", previousKind, resource.Kind); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousName != resource.Name {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "name", previousName, resource.Name); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousDisplayName != resource.DisplayName {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "display_name", previousDisplayName, resource.DisplayName); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousExternalID != resource.ExternalID {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "external_id", previousExternalID, resource.ExternalID); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousProvider != resource.Provider {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "provider", previousProvider, resource.Provider); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousRegion != resource.Region {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "region", previousRegion, resource.Region); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousNamespace != resource.Namespace {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "namespace", previousNamespace, resource.Namespace); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousCluster != resource.Cluster {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "cluster", previousCluster, resource.Cluster); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousStatus != resource.Status {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "status", previousStatus, resource.Status); err != nil {
				logAuditFailure(ctx, "status_change", "resource", 0, err)
			}
		}
		if previousHealth != resource.Health {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "health", previousHealth, resource.Health); err != nil {
				logAuditFailure(ctx, "health_change", "resource", 0, err)
			}
		}
		if previousLabels != string(resource.Labels) {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "labels", previousLabels, string(resource.Labels)); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousAnnotations != string(resource.Annotations) {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "annotations", previousAnnotations, string(resource.Annotations)); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
		if previousMetadata != string(resource.Metadata) {
			if err := s.auditRepo.LogUpdate(userID, "resource", entityID, &resource.ProjectID, nil, "metadata", previousMetadata, string(resource.Metadata)); err != nil {
				logAuditFailure(ctx, "update", "resource", 0, err)
			}
		}
	}

	response := mapper.MapResource(*resource)
	return &response, nil
}

func (s *ResourceService) DeleteResource(ctx context.Context, id uuid.UUID, userID uint) error {
	resource, err := s.getOwnedResource(id, userID)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDelete(resource.ID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogDelete(userID, "resource", resource.ID.String(), &resource.ProjectID, nil); err != nil {
			logAuditFailure(ctx, "delete", "resource", 0, err)
		}
	}

	return nil
}

func (s *ResourceService) GetResource(id uuid.UUID, userID uint) (*dto.ResourceResponse, error) {
	resource, err := s.getOwnedResource(id, userID)
	if err != nil {
		return nil, err
	}

	response := mapper.MapResource(*resource)
	return &response, nil
}

func (s *ResourceService) ListResources(userID uint, req *models.PaginationRequest) (*dto.ResourceListResponse, error) {
	items, total, err := s.repo.List(req, userID)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.ResourceListResponse{
		Items:      mapper.MapResources(items),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *ResourceService) SyncResources(ctx context.Context, projectID uuid.UUID, userID uint, discoveredResources []models.Resource) (*resourcesync.SyncResult, error) {
	belongs, err := s.repo.ProjectBelongsToUser(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	normalizedResources, err := s.normalizeDiscoveredResources(projectID, discoveredResources, userID)
	if err != nil {
		return nil, err
	}

	if s.syncEngine == nil {
		s.syncEngine = resourcesync.NewSyncEngine(s.repo)
	}

	plan, err := s.syncEngine.Plan(ctx, projectID.String(), normalizedResources)
	if err != nil {
		return nil, err
	}

	result, err := s.syncEngine.Execute(ctx, plan)
	if err != nil {
		return nil, err
	}

	s.auditSyncPlan(ctx, userID, projectID, plan, result)

	return result, nil
}

func (s *ResourceService) getOwnedResource(id uuid.UUID, userID uint) (*models.Resource, error) {
	resource, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrResourceNotFound
		}
		return nil, err
	}

	belongs, err := s.repo.ProjectBelongsToUser(resource.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrProjectForbidden
	}

	return resource, nil
}

func (s *ResourceService) buildResourceModel(
	projectID uuid.UUID,
	parentResourceID *uuid.UUID,
	kind,
	name,
	displayName,
	externalID,
	provider,
	region,
	namespace,
	cluster,
	status,
	health string,
	labels,
	annotations,
	metadata json.RawMessage,
	userID uint,
) (*models.Resource, error) {
	kind = strings.TrimSpace(kind)
	name = strings.TrimSpace(name)
	displayName = strings.TrimSpace(displayName)
	externalID = strings.TrimSpace(externalID)
	provider = strings.TrimSpace(provider)
	region = strings.TrimSpace(region)
	namespace = strings.TrimSpace(namespace)
	cluster = strings.TrimSpace(cluster)
	status = strings.TrimSpace(strings.ToUpper(status))
	health = strings.TrimSpace(strings.ToUpper(health))

	if !constants.IsValidResourceKind(kind) {
		return nil, apperrors.ErrInvalidResourceKind
	}
	if !constants.IsValidResourceStatus(status) {
		return nil, apperrors.ErrInvalidResourceStatus
	}
	if !constants.IsValidResourceHealth(health) {
		return nil, apperrors.ErrInvalidResourceHealth
	}

	belongs, err := s.repo.ProjectBelongsToUser(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	resolvedLabels := labels
	if len(resolvedLabels) == 0 {
		resolvedLabels = json.RawMessage(`{}`)
	}

	resolvedAnnotations := annotations
	if len(resolvedAnnotations) == 0 {
		resolvedAnnotations = json.RawMessage(`{}`)
	}

	resolvedMetadata := metadata
	if len(resolvedMetadata) == 0 {
		resolvedMetadata = json.RawMessage(`{}`)
	}

	resource := &models.Resource{
		ProjectID:        projectID,
		ParentResourceID: parentResourceID,
		Kind:             kind,
		Name:             name,
		DisplayName:      displayName,
		ExternalID:       externalID,
		Provider:         provider,
		Region:           region,
		Namespace:        namespace,
		Cluster:          cluster,
		Status:           status,
		Health:           health,
		Labels:           resolvedLabels,
		Annotations:      resolvedAnnotations,
		Metadata:         resolvedMetadata,
		CreatedBy:        userID,
	}

	return resource, nil
}

func (s *ResourceService) normalizeDiscoveredResources(projectID uuid.UUID, discoveredResources []models.Resource, userID uint) ([]models.Resource, error) {
	normalized := make([]models.Resource, 0, len(discoveredResources))
	for _, resource := range discoveredResources {
		resource.ProjectID = projectID
		if resource.CreatedBy == 0 {
			resource.CreatedBy = userID
		}
		resource.Kind = strings.TrimSpace(resource.Kind)
		resource.Name = strings.TrimSpace(resource.Name)
		resource.DisplayName = strings.TrimSpace(resource.DisplayName)
		resource.ExternalID = strings.TrimSpace(resource.ExternalID)
		resource.Provider = strings.TrimSpace(resource.Provider)
		resource.Region = strings.TrimSpace(resource.Region)
		resource.Namespace = strings.TrimSpace(resource.Namespace)
		resource.Cluster = strings.TrimSpace(resource.Cluster)
		resource.Status = strings.TrimSpace(strings.ToUpper(resource.Status))
		resource.Health = strings.TrimSpace(strings.ToUpper(resource.Health))

		if !constants.IsValidResourceKind(resource.Kind) {
			return nil, apperrors.ErrInvalidResourceKind
		}
		if !constants.IsValidResourceStatus(resource.Status) {
			return nil, apperrors.ErrInvalidResourceStatus
		}
		if !constants.IsValidResourceHealth(resource.Health) {
			return nil, apperrors.ErrInvalidResourceHealth
		}
		if len(resource.Labels) == 0 {
			resource.Labels = json.RawMessage(`{}`)
		}
		if len(resource.Annotations) == 0 {
			resource.Annotations = json.RawMessage(`{}`)
		}
		if len(resource.Metadata) == 0 {
			resource.Metadata = json.RawMessage(`{}`)
		}

		normalized = append(normalized, resource)
	}

	return normalized, nil
}

func (s *ResourceService) auditSyncPlan(ctx context.Context, userID uint, projectID uuid.UUID, plan *resourcesync.SyncPlan, result *resourcesync.SyncResult) {
	if s.auditRepo == nil || plan == nil || result == nil {
		return
	}

	if result.Created == len(plan.ResourcesToCreate) {
		for _, resource := range plan.ResourcesToCreate {
			if err := s.auditRepo.LogCreate(userID, "resource", resourceEntityID(resource), &projectID, nil); err != nil {
				logAuditFailure(ctx, "sync_create", "resource", 0, err)
			}
		}
	}

	if result.Updated == len(plan.ResourcesToUpdate) {
		for _, resource := range plan.ResourcesToUpdate {
			if err := s.auditRepo.LogUpdate(userID, "resource", resourceEntityID(resource), &projectID, nil, "synchronized", "stale", "current"); err != nil {
				logAuditFailure(ctx, "sync_update", "resource", 0, err)
			}
			if err := s.auditRepo.LogUpdate(userID, "resource", resourceEntityID(resource), &projectID, nil, "status", "changed", resource.Status); err != nil {
				logAuditFailure(ctx, "sync_status_change", "resource", 0, err)
			}
			if err := s.auditRepo.LogUpdate(userID, "resource", resourceEntityID(resource), &projectID, nil, "health", "changed", resource.Health); err != nil {
				logAuditFailure(ctx, "sync_health_change", "resource", 0, err)
			}
		}
	}

	if result.Deleted == len(plan.ResourcesToDelete) {
		for _, resource := range plan.ResourcesToDelete {
			if err := s.auditRepo.LogDelete(userID, "resource", resourceEntityID(resource), &projectID, nil); err != nil {
				logAuditFailure(ctx, "sync_delete", "resource", 0, err)
			}
		}
	}

	if result.Restored == len(plan.ResourcesToRestore) {
		for _, resource := range plan.ResourcesToRestore {
			if err := s.auditRepo.LogUpdate(userID, "resource", resourceEntityID(resource), &projectID, nil, "deleted_at", "set", "restored"); err != nil {
				logAuditFailure(ctx, "sync_restore", "resource", 0, err)
			}
		}
	}
}

func uuidPointerString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}

	return value.String()
}

func uuidPointerStringsEqual(left, right *uuid.UUID) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	return *left == *right
}

func resourceEntityID(resource models.Resource) string {
	if resource.ID != uuid.Nil {
		return resource.ID.String()
	}

	return resource.ProjectID.String() + ":" + resource.Kind + ":" + resource.ExternalID
}
